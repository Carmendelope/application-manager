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
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"sync"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	unifiedLogging grpc_unified_logging_go.CoordinatorClient
	appClient      grpc_application_go.ApplicationsClient
	cache          utils.InstanceCache
}

// NewManager creates a Manager using a set of clients.
func NewManager(unifiedLogging grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient) Manager {
	return Manager{
		unifiedLogging: unifiedLogging,
		cache:          utils.NewInstCache(appClient)}
}

// callToSearch call to unified-logging search and store the result in the channel
func (m *Manager) callToSearch(respond chan<- grpc_unified_logging_go.LogResponse, wg *sync.WaitGroup,
	request *grpc_unified_logging_go.SearchRequest,	instance string) {

	defer wg.Done()

	ctx, cancel := common.GetContext()
	defer cancel()

	list, err := m.unifiedLogging.Search(ctx, &grpc_unified_logging_go.SearchRequest{
		OrganizationId:         request.OrganizationId,
		AppInstanceId:          instance,
		ServiceGroupId:         request.ServiceGroupId,
		ServiceGroupInstanceId: request.ServiceGroupInstanceId,
		ServiceId:              request.ServiceId,
		ServiceInstanceId:      request.ServiceInstanceId,
		MsgQueryFilter:         request.MsgQueryFilter,
		From:                   request.From,
		To:                     request.To,
	})

	if err != nil {
		log.Error().Err(err).Msg("error sending search to unified-logging")
	}

	respond <- *list
}


// TODO fill isDead field, wait until catalog is finished
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	// array to store all appInstanceIds for which we have to ask
	var instances []string
	if request.AppInstanceId != "" {
		instances = []string{request.AppInstanceId}
	} else {
		// update memory structure
		err := m.cache.UpdateInstanceList(request.OrganizationId)
		if err != nil {
			return nil, conversions.ToGRPCError(err)
		}

		if request.AppDescriptorId != "" {
			instances = m.cache.RetrieveInstancesOfADescriptor(request.OrganizationId, request.AppDescriptorId)
		} else {
			instances = m.cache.RetrieveInstancesOfAnOrganization(request.OrganizationId)
		}
	}

	log.Debug().Interface("instances", instances).Msg("instances to ask for")

	logResponse := make([]*grpc_application_manager_go.LogEntryResponse, 0)

	if len(instances) == 0 {
		// TODO: reguntar a Dani si devuelvo error o vacio
		return &grpc_application_manager_go.LogResponse{
			OrganizationId: request.OrganizationId,
			From:           request.From,
			To:             request.To,
			Entries:        logResponse,
		}, nil
	}

	respond := make(chan grpc_unified_logging_go.LogResponse, len(instances))
	var wg sync.WaitGroup
	wg.Add(len(instances))

	for _, instance := range instances {

		// call to unified-logging search
		go m.callToSearch(respond, &wg, &grpc_unified_logging_go.SearchRequest{
			OrganizationId:         request.OrganizationId,
			AppInstanceId:          instance,
			ServiceGroupId:         request.ServiceGroupId,
			ServiceGroupInstanceId: request.ServiceGroupInstanceId,
			ServiceId:              request.ServiceId,
			ServiceInstanceId:      request.ServiceInstanceId,
			MsgQueryFilter:         request.MsgQueryFilter,
			From:                   request.From,
			To:                     request.To,
		}, instance)


	}

	wg.Wait()
	close(respond)

	for response := range respond {
		for _, entry := range response.Entries {
			logResponse = append(logResponse, m.expandInformation(request.OrganizationId, &grpc_application_manager_go.LogEntryResponse{
				//AppDescriptorId
				//AppDescriptorName
				//AppInstanceName:
				//ServiceGroupName:
				//ServiceName:
				// IsDead: ask the catalog
				AppInstanceId:          response.AppInstanceId,
				ServiceGroupId:         response.ServiceGroupId,
				ServiceGroupInstanceId: response.ServiceGroupInstanceId,
				ServiceId:              response.ServiceId,
				ServiceInstanceId:      response.ServiceInstanceId,
				Timestamp:              entry.Timestamp,
				Msg:                    entry.Msg,
			}, true))//request.IncludeMetadata))
		}
	}

	log.Debug().Int("len", len(logResponse)).Msg(" TOTAL: Search result")

	return &grpc_application_manager_go.LogResponse{
		OrganizationId: request.OrganizationId,
		From:           request.From,
		To:             request.To,
		Entries:        logResponse,
	}, nil
}

// getServiceGroupName returns the name of the serviceGroupId and the serviceId
func (m *Manager) getNames (serviceGroupId string, serviceId string, inst *grpc_application_go.AppInstance) (string, string) {

	groupName := ""
	serviceName := ""

	if inst == nil {
		return groupName, serviceName
	}

	for _, group := range inst.Groups {
		if group.ServiceGroupId == serviceGroupId {
			groupName = group.Name
			for _, service := range group.ServiceInstances {
				if service.ServiceId == serviceId {
					serviceName = service.Name
					return groupName, serviceName
				}
			}
		}
	}
	return groupName, serviceName
}

// expandInformation fill the logEntry with the descriptor and names
func (m *Manager) expandInformation(organizationId string, logEntry *grpc_application_manager_go.LogEntryResponse, expand bool)*grpc_application_manager_go.LogEntryResponse{

	inst, err := m.cache.RetrieveInstanceInformation(organizationId, logEntry.AppInstanceId)
	if err != nil {
		log.Error().Err(err).Str("organizationId", organizationId).Str("appInstanceId", logEntry.AppInstanceId).
			Msg("error getting instance information")
	}

	logEntry.AppDescriptorId = inst.AppDescriptorId

	if expand {
		name, err := m.cache.RetrieveDescriptorName(organizationId, logEntry.AppDescriptorId)
		if err != nil {
			logEntry.AppDescriptorName = name
		}
		logEntry.AppInstanceName = inst.Name
		if logEntry.ServiceGroupId != "" {
			groupName, serviceName := m.getNames(logEntry.ServiceGroupId, logEntry.ServiceId, inst)
			logEntry.ServiceGroupName = groupName
			logEntry.ServiceName = serviceName
		}

	}
	return logEntry

}