/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package unified_logging

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/rs/zerolog/log"
	"github.com/nalej/nalej-bus/pkg/queue/application/events"
	"time"
)

const (
	ApplicationManagerTimeout = time.Second * 3
	DefaultCacheEntries = 100
	unknownField = "Unknown"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	coordinatorClient         grpc_unified_logging_go.CoordinatorClient
	appsClient                grpc_application_go.ApplicationsClient
	instHelper                *utils.InstancesHelper
	unifiedLoggingClient      grpc_application_manager_go.UnifiedLoggingClient
	appHistoryLogsClient		grpc_application_history_logs_go.ApplicationHistoryLogsClient
	applicationEventsConsumer *events.ApplicationEventsConsumer
}

// NewManager creates a Manager using a set of clients.
func NewManager(coordinatorClient grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient, unifiedLoggingClient grpc_application_manager_go.UnifiedLoggingClient, appHistoryLogsClient grpc_application_history_logs_go.ApplicationHistoryLogsClient, appEventsConsumer *events.ApplicationEventsConsumer) (*Manager, derrors.Error) {
	helper, err := utils.NewInstancesHelper(appClient, DefaultCacheEntries)
	if err != nil {
		return nil, err
	}
	return &Manager{
		coordinatorClient:    coordinatorClient,
		instHelper:           helper,
		unifiedLoggingClient: unifiedLoggingClient,
		appHistoryLogsClient: appHistoryLogsClient,
		applicationEventsConsumer:appEventsConsumer,

	}, nil
}

/// TODO fill isDead field, wait until catalog is finished
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	log.Debug().Interface("request", request).Msg("search request")
	ctx, cancel := common.GetContext()
	defer cancel()

	searchResponse, err := m.coordinatorClient.Search(ctx, &grpc_unified_logging_go.SearchRequest{
		OrganizationId:         request.OrganizationId,
		AppDescriptorId:        request.AppDescriptorId,
		AppInstanceId:          request.AppInstanceId,
		ServiceGroupId:         request.ServiceGroupId,
		ServiceGroupInstanceId: request.ServiceGroupInstanceId,
		ServiceId:              request.ServiceId,
		ServiceInstanceId:      request.ServiceInstanceId,
		MsgQueryFilter:         request.MsgQueryFilter,
		From:                   request.From,
		To:                     request.To,
	})

	if err != nil {
		return nil, err
	}
	logResponse := make([]*grpc_application_manager_go.LogEntryResponse, 0)

	// convert unified_logging.LogEntryResponse to grpc_application_manager_go.LogEntryResponse
	// and expand info if proceeded
	for _, response := range searchResponse.Responses {
		for _, entry := range response.Entries {

			logResponse = append(logResponse, m.expandInformation(request.OrganizationId, &grpc_application_manager_go.LogEntryResponse{

				// IsDead: ask the catalog
				AppDescriptorId:        response.AppDescriptorId,
				AppInstanceId:          response.AppInstanceId,
				ServiceGroupId:         response.ServiceGroupId,
				ServiceGroupInstanceId: response.ServiceGroupInstanceId,
				ServiceId:              response.ServiceId,
				ServiceInstanceId:      response.ServiceInstanceId,
				Timestamp:              entry.Timestamp,
				Msg:                    entry.Msg,
			}, request.IncludeMetadata))
		}
	}

	return &grpc_application_manager_go.LogResponse{
		OrganizationId: searchResponse.OrganizationId,
		From:           searchResponse.From,
		To:             searchResponse.To,
		Entries:        logResponse,
	}, nil
}

func (m *Manager) Catalog(request *grpc_application_manager_go.AvailableLogRequest) (*grpc_application_manager_go.AvailableLogResponse, error) {
	log.Debug().Interface("request", request).Msg("available log request")
	ctx, cancel := common.GetContext()
	defer cancel()

	availableLogResponse, cErr := m.unifiedLoggingClient.Catalog(ctx, request)

	if cErr != nil {
		return nil, cErr
	}
	return availableLogResponse, nil
}

// getNamesFromSummary returns the name of the serviceGroupId and the serviceId
func (m *Manager) getNamesFromSummary(serviceGroupId string, serviceId string, inst *grpc_application_go.AppInstanceReducedSummary) (string, string) {

	groupName := unknownField
	serviceName := unknownField

	if inst == nil {
		return groupName, serviceName
	}

	for _, group := range inst.Groups {
		if group.ServiceGroupId == serviceGroupId {
			groupName = group.ServiceGroupName
			for _, service := range group.ServiceInstances {
				if service.ServiceId == serviceId {
					serviceName = service.ServiceName
					return groupName, serviceName
				}
			}
		}
	}
	return groupName, serviceName
}

// expandInformation fill the logEntry with the descriptor and names
func (m *Manager) expandInformation(organizationId string, logEntry *grpc_application_manager_go.LogEntryResponse, expand bool) *grpc_application_manager_go.LogEntryResponse {

	if !expand {
		return logEntry
	}
	if logEntry.AppInstanceId == "" {
		log.Warn().Msg("unable to expand log information, app_instance_id is empty")
		logEntry.AppDescriptorName = unknownField
		logEntry.ServiceGroupName = unknownField
		logEntry.ServiceName = unknownField
		return logEntry
	}

	summary, err := m.instHelper.RetrieveInstanceSummary(organizationId, logEntry.AppInstanceId)
	if err != nil {
		log.Warn().Interface("trace", err.StackTrace()).Str("organizationId", organizationId).Str("appInstanceId", logEntry.AppInstanceId).Msg("error getting reduced summary")
		return logEntry
	}

	logEntry.AppDescriptorName = summary.AppDescriptorName
	groupName, serviceName := m.getNamesFromSummary(logEntry.ServiceGroupId, logEntry.ServiceId, summary)
	logEntry.ServiceGroupName = groupName
	logEntry.ServiceName = serviceName

	return logEntry

}

// ManageCatalog receives DeploymentServiceUpdateRequest messages from the bus and manages the catalog entries to be sent to system-model
func (m *Manager) ManageCatalog (request *grpc_conductor_go.DeploymentServiceUpdateRequest) error {
	addCtx, addCancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer addCancel()
	for _, service := range request.List {
		appInstanceReducedSummary, sumErr := m.appsClient.GetAppInstanceReducedSummary(addCtx, &grpc_application_go.AppInstanceId{
			OrganizationId:       request.OrganizationId,
			AppInstanceId:        service.ApplicationInstanceId,
		})
		if sumErr != nil {
			return sumErr
		}

		_, addErr := m.appHistoryLogsClient.Add(addCtx, &grpc_application_history_logs_go.AddLogRequest{
			OrganizationId:         request.OrganizationId,
			AppInstanceId:          service.ApplicationInstanceId,
			AppDescriptorId:        appInstanceReducedSummary.AppDescriptorId,
			ServiceGroupId:         service.ServiceGroupId,
			ServiceGroupInstanceId:	service.ServiceGroupInstanceId,
			ServiceId:              service.ServiceId,
			ServiceInstanceId:      service.ServiceInstanceId,
			Created:                time.Now().UnixNano(),
		})
		if addErr != nil {
			return addErr
		}
	}

	return nil
}