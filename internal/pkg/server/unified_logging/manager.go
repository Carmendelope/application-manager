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
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/rs/zerolog/log"
)

const DefaultCacheEntries = 100

// Manager structure with the required clients for roles operations.
type Manager struct {
	unifiedLogging grpc_unified_logging_go.CoordinatorClient
	appClient      grpc_application_go.ApplicationsClient
	instHelper     *utils.InstancesHelper
}

// NewManager creates a Manager using a set of clients.
func NewManager(unifiedLogging grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient) Manager {
	helper, _ := utils.NewInstancesHelper(appClient, DefaultCacheEntries)
	return Manager{
		unifiedLogging: unifiedLogging,
		instHelper:     helper,
	}
}

/// TODO fill isDead field, wait until catalog is finished
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	log.Debug().Interface("request", request).Msg("search request")
	ctx, cancel := common.GetContext()
	defer cancel()

	response, err := m.unifiedLogging.Search(ctx, &grpc_unified_logging_go.SearchRequest{
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
	logResponse := make([]*grpc_application_manager_go.LogEntryResponse, len(response.Entries))

	// convert unified_logging.LogEntryResponse to grpc_application_manager_go.LogEntryResponse
	// and expand info if proceeded
	for i, entry := range response.Entries {
		logResponse[i] = m.expandInformation(request.OrganizationId, &grpc_application_manager_go.LogEntryResponse{

			// IsDead: ask the catalog
			AppDescriptorId:        response.AppDescriptorId,
			AppInstanceId:          response.AppInstanceId,
			ServiceGroupId:         response.ServiceGroupId,
			ServiceGroupInstanceId: response.ServiceGroupInstanceId,
			ServiceId:              response.ServiceId,
			ServiceInstanceId:      response.ServiceInstanceId,
			Timestamp:              entry.Timestamp,
			Msg:                    entry.Msg,
		}, request.IncludeMetadata)
	}

	return &grpc_application_manager_go.LogResponse{
		OrganizationId: request.OrganizationId,
		From:           request.From,
		To:             request.To,
		Entries:        logResponse,
	}, nil
}

// getNamesFromSummary returns the name of the serviceGroupId and the serviceId
func (m *Manager) getNamesFromSummary(serviceGroupId string, serviceId string, inst *grpc_application_go.AppInstanceReducedSummary) (string, string) {

	groupName := ""
	serviceName := ""

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
