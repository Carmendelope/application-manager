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
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/rs/zerolog/log"
)

const DefaultCacheEntries = 100
const unknownField = "Unknown"

// Manager structure with the required clients for roles operations.
type Manager struct {
	unifiedLogging grpc_unified_logging_go.CoordinatorClient
	appClient      grpc_application_go.ApplicationsClient
	instHelper     *utils.InstancesHelper
}

// NewManager creates a Manager using a set of clients.
func NewManager(unifiedLogging grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient) (*Manager, derrors.Error) {
	helper, err := utils.NewInstancesHelper(appClient, DefaultCacheEntries)
	if err != nil {
		return nil, err
	}
	return &Manager{
		unifiedLogging: unifiedLogging,
		instHelper:     helper,
	}, nil
}

/// TODO fill isDead field, wait until catalog is finished
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	log.Debug().Interface("request", request).Msg("search request")
	ctx, cancel := common.GetContext()
	defer cancel()

	searchResponse, err := m.unifiedLogging.Search(ctx, &grpc_unified_logging_go.SearchRequest{
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
			})
		}
	}

	return &grpc_application_manager_go.LogResponse{
		OrganizationId: searchResponse.OrganizationId,
		From:           searchResponse.From,
		To:             searchResponse.To,
		Entries:        logResponse,
		FailedClusterIds: searchResponse.FailedClusterIds,
	}, nil
}

