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
 */

package unified_logging

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/nalej-bus/pkg/queue/application/events"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	ApplicationManagerTimeout = time.Second * 3
	DefaultCacheEntries       = 100
	unknownField              = "Unknown"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	coordinatorClient         grpc_unified_logging_go.CoordinatorClient
	appsClient                grpc_application_go.ApplicationsClient
	instHelper                *utils.InstancesHelper
	appHistoryLogsClient      grpc_application_history_logs_go.ApplicationHistoryLogsClient
	applicationEventsConsumer *events.ApplicationEventsConsumer
}

// NewManager creates a Manager using a set of clients.
func NewManager(coordinatorClient grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient, appHistoryLogsClient grpc_application_history_logs_go.ApplicationHistoryLogsClient, appEventsConsumer *events.ApplicationEventsConsumer) (*Manager, derrors.Error) {
	instHelper, err := utils.NewInstancesHelper(appClient, DefaultCacheEntries)
	if err != nil {
		return nil, err
	}
	return &Manager{
		coordinatorClient:         coordinatorClient,
		instHelper:                instHelper,
		appHistoryLogsClient:      appHistoryLogsClient,
		applicationEventsConsumer: appEventsConsumer,
	}, nil
}

func (m *Manager) isDead(response *grpc_unified_logging_go.LogResponse, catalog *grpc_application_history_logs_go.LogResponse) bool {
	isDead := false

	if catalog != nil {
		for _, desc := range catalog.Events {
			if desc.AppDescriptorId == response.AppDescriptorId && desc.AppInstanceId == response.AppInstanceId &&
				desc.ServiceGroupId == response.ServiceGroupId && desc.ServiceGroupInstanceId == response.ServiceGroupInstanceId &&
				desc.ServiceId == response.ServiceId && desc.ServiceInstanceId == response.ServiceInstanceId {
				if desc.Terminated != 0 {
					return true
				} else {
					return false
				}
			}
		}
	}
	return isDead
}

// Search returns a list of log entries that follow the conditions of the request and
// returns two lists of services that have been active in the time period of the logs returned
// one group by application descriptors and another one group application instances
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	log.Debug().Interface("request", request).Msg("search request")

	ctx, cancel := common.GetContext()
	defer cancel()
	// 1.- call to unified logging to retrieve log entries
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
		NFirst: request.NFirst,
	})

	if err != nil {
		return nil, err
	}


	// 2.- go to system model to retrieve a list of service_history_logs
	searchRequest := &grpc_application_history_logs_go.SearchLogRequest{
		OrganizationId: request.OrganizationId,
		From: request.From ,
		To:   request.To,
	}

	descriptors := make([]*grpc_application_manager_go.AppDescriptorLogSummary, 0)
	instances := make([]*grpc_application_manager_go.AppInstanceLogSummary, 0)

	logHistoryResponse, cErr := m.appHistoryLogsClient.Search(ctx, searchRequest)
	if cErr != nil {
		log.Error().Err(err).Msg("unable to return the catalog")
		// TODO: ask what to do here
	}else{
		// 3.- Fill the list returned by system-model with the names and labels
		availableList := m.Organize(logHistoryResponse)

		descriptors = availableList.AppDescriptorLogSummary
		instances = availableList.AppInstanceLogSummary
	}

	logResponse := make([]*grpc_application_manager_go.LogEntryResponse, 0)

	// convert unified_logging.LogEntryResponse to grpc_application_manager_go.LogEntryResponse
	// and expand info if proceeded
	for _, response := range searchResponse.Responses {
		// check if the service_instance_id is dead or not
		isDead := m.isDead(response, logHistoryResponse)
		for _, entry := range response.Entries {
			logResponse = append(logResponse, &grpc_application_manager_go.LogEntryResponse{

				// IsDead: ask the catalog
				AppDescriptorId:        response.AppDescriptorId,
				AppDescriptorName:      response.AppDescriptorName,
				AppInstanceId:          response.AppInstanceId,
				AppInstanceName:        response.AppInstanceName,
				ServiceGroupId:         response.ServiceGroupId,
				ServiceGroupName:       response.ServiceGroupName,
				ServiceGroupInstanceId: response.ServiceGroupInstanceId,
				ServiceId:              response.ServiceId,
				ServiceName:            response.ServiceName,
				ServiceInstanceId:      response.ServiceInstanceId,
				Timestamp:              entry.Timestamp,
				Msg:                    entry.Msg,
				IsDead:                 isDead,
			})
		}
	}

	// Return
	return &grpc_application_manager_go.LogResponse{
		OrganizationId:          searchResponse.OrganizationId,
		From:                    searchResponse.From,
		To:                      searchResponse.To,
		Entries:                 logResponse,
		AppDescriptorLogSummary: descriptors,
		AppInstanceLogSummary:   instances,
		FailedClusterIds:        searchResponse.FailedClusterIds,
	}, nil
}

func (m *Manager) Catalog(request *grpc_application_manager_go.AvailableLogRequest) (*grpc_application_manager_go.AvailableLogResponse, error) {
	log.Debug().Interface("request", request).Msg("available log request")
	ctx, cancel := common.GetContext()
	defer cancel()

	searchRequest := &grpc_application_history_logs_go.SearchLogRequest{
		OrganizationId: request.OrganizationId,
		From:           request.From,
		To:             request.To,
	}

	logResponse, cErr := m.appHistoryLogsClient.Search(ctx, searchRequest)
	if cErr != nil {
		return nil, cErr
	}

	availableLogResponse := m.Organize(logResponse)

	return availableLogResponse, nil
}

// ManageCatalog receives DeploymentServiceUpdateRequest messages from the bus and manages the catalog entries to be sent to system-model
func (m *Manager) ManageCatalog(request *grpc_conductor_go.DeploymentServiceUpdateRequest) error {
	addCtx, addCancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer addCancel()
	for _, service := range request.List {
		log.Debug().Str("app instance id", service.ApplicationInstanceId).Msg("incoming service update request")
		if service.Status == grpc_application_go.ServiceStatus_SERVICE_DEPLOYING {
			log.Debug().Str("service instance id", service.ServiceInstanceId).Msg("adding service to service history logs")
			appInstanceReducedSummary, sumErr := m.instHelper.RetrieveInstanceSummary(request.OrganizationId, service.ApplicationInstanceId)
			if sumErr != nil {
				log.Debug().Msg("error retrieving service instance id")
				return conversions.ToGRPCError(sumErr)
			}

			_, addErr := m.appHistoryLogsClient.Add(addCtx, &grpc_application_history_logs_go.AddLogRequest{
				OrganizationId:         request.OrganizationId,
				AppInstanceId:          service.ApplicationInstanceId,
				AppDescriptorId:        appInstanceReducedSummary.AppDescriptorId,
				ServiceGroupId:         service.ServiceGroupId,
				ServiceGroupInstanceId: service.ServiceGroupInstanceId,
				ServiceId:              service.ServiceId,
				ServiceInstanceId:      service.ServiceInstanceId,
				Created:                time.Now().UnixNano(),
			})
			if addErr != nil {
				log.Debug().Msg("error adding service instance log")
				return addErr
			}
		}

		if service.Status == grpc_application_go.ServiceStatus_SERVICE_ERROR || service.Status == grpc_application_go.ServiceStatus_SERVICE_TERMINATING {
			log.Debug().Str("service instance id", service.ServiceInstanceId).Msg("updating service from service history logs")
			_, updateErr := m.appHistoryLogsClient.Update(addCtx, &grpc_application_history_logs_go.UpdateLogRequest{
				OrganizationId:    request.OrganizationId,
				AppInstanceId:     service.ApplicationInstanceId,
				ServiceInstanceId: service.ServiceInstanceId,
				Terminated:        time.Now().UnixNano(),
			})
			if updateErr != nil {
				log.Debug().Msg("error updating service instance log")
				return updateErr
			}
		}
	}

	return nil
}
