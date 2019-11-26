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
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

const DefaultCacheEntries = 100
const unknownField = "Unknown"

// Manager structure with the required clients for roles operations.
type Manager struct {
	unifiedLogging grpc_unified_logging_go.CoordinatorClient
	appClient      grpc_application_go.ApplicationsClient
	catalogClient  grpc_application_history_logs_go.ApplicationHistoryLogsClient
	instHelper     *utils.InstancesHelper
}

// NewManager creates a Manager using a set of clients.
func NewManager(unifiedLogging grpc_unified_logging_go.CoordinatorClient, appClient grpc_application_go.ApplicationsClient,
	catClient grpc_application_history_logs_go.ApplicationHistoryLogsClient) (*Manager, derrors.Error) {
	helper, err := utils.NewInstancesHelper(appClient, DefaultCacheEntries)
	if err != nil {
		return nil, err
	}
	return &Manager{
		unifiedLogging: unifiedLogging,
		instHelper:     helper,
		catalogClient:  catClient,
	}, nil
}

func (m *Manager) checkConditions(request grpc_application_manager_go.SearchRequest, event grpc_application_history_logs_go.ServiceInstanceLog) bool {
	review := true

	if request.ServiceGroupInstanceId != "" && request.ServiceGroupInstanceId != event.ServiceGroupInstanceId {
		review = false
	}
	if request.ServiceInstanceId != "" && request.ServiceInstanceId != event.ServiceInstanceId {
		review = false
	}
	if request.ServiceGroupId != "" && request.ServiceGroupId != event.ServiceGroupId {
		review = false
	}
	if request.ServiceId != "" && request.ServiceId != event.ServiceId {
		review = false
	}
	if request.AppInstanceId != "" && request.AppInstanceId != event.AppInstanceId {
		review = false
	}
	if request.AppDescriptorId != "" && request.AppDescriptorId != event.AppDescriptorId {
		review = false
	}

	return review
}

// grepMsgQueryFilter returns a map with the identifiers of those fields whose name contains the string msgQueryFilter
func (m *Manager) grepMsgQueryFilter(request *grpc_application_manager_go.SearchRequest) (map[string]*grpc_unified_logging_go.IdList /*map[string][]string*/, derrors.Error) {

	descriptorList := make([]string, 0)
	instanceList := make([]string, 0)
	serviceGroupList := make([]string, 0)
	serviceList := make([]string, 0)

	// initialize the identifiers map
	res := map[string]*grpc_unified_logging_go.IdList{}

	ctxCat, cancelCat := common.GetContext()
	defer cancelCat()

	// TODO: store the catalog response in a cache
	response, err := m.catalogClient.Search(ctxCat, &grpc_application_history_logs_go.SearchLogRequest{
		OrganizationId: request.OrganizationId,
		From:           request.From,
		To:             request.To,
	})
	if err != nil {
		return res, conversions.ToDerror(err)
	}

	for _, event := range response.Events {
		if m.checkConditions(*request, *event) {
			summary, err := m.instHelper.RetrieveInstanceSummary(event.OrganizationId, event.AppInstanceId)
			if err != nil {
				log.Warn().Str("trace", err.DebugReport()).Str("organizationId", event.OrganizationId).
					Str("app_instance_id", event.AppInstanceId).Msg("error getting summary")
			} else {
				// TODO: summary and request.MsgQueryFilter to lowerCase
				// APP_DESCRIPTOR_NAME
				if strings.Contains(summary.AppDescriptorName, request.MsgQueryFilter) {
					descriptorList = append(descriptorList, event.AppDescriptorId)
				}
				// APP_INSTANCE_NAME
				if strings.Contains(summary.AppInstanceName, request.MsgQueryFilter) {
					instanceList = append(instanceList, event.AppInstanceId)
				}
				// SERVICE_GROUP_NAME
				for _, group := range summary.Groups {
					if strings.Contains(group.ServiceGroupName, request.MsgQueryFilter) {
						serviceGroupList = append(serviceGroupList, group.ServiceGroupId)
					}
					// SERVICE_NAME
					for _, service := range group.ServiceInstances {
						if strings.Contains(service.ServiceName, request.MsgQueryFilter) {
							serviceList = append(serviceList, service.ServiceId)
						}
					}
				}
			}
		}
	}
	if len(descriptorList) > 0 {
		res[common.NALEJ_ANNOTATION_APP_DESCRIPTOR] = &grpc_unified_logging_go.IdList{
			Ids: descriptorList,
		}
	}
	if len(instanceList) > 0 {
		res[common.NALEJ_ANNOTATION_APP_INSTANCE_ID] = &grpc_unified_logging_go.IdList{
			Ids: instanceList,
		}
	}
	if len(serviceGroupList) > 0 {
		res[common.NALEJ_ANNOTATION_SERVICE_GROUP_ID] = &grpc_unified_logging_go.IdList{
			Ids: serviceGroupList,
		}
	}
	if len(serviceList) > 0 {
		res[common.NALEJ_ANNOTATION_SERVICE_ID] = &grpc_unified_logging_go.IdList{
			Ids: serviceList,
		}
	}

	return res, nil
}

/// TODO fill isDead field, wait until catalog is finished
func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {

	log.Debug().Interface("request", request).Msg("search request")
	// if we don't have date range, we ask for the last hour
	if request.From == 0 && request.To == 0 {
		oneHourAgo := time.Now().Add(time.Hour * (-1))
		request.From = oneHourAgo.Unix()
	}

	k8sQuery := make(map[string]*grpc_unified_logging_go.IdList, 0)

	// if request.MsgQueryFilter is filled -> ask to the catalog
	if request.MsgQueryFilter != "" {
		queryIds, err := m.grepMsgQueryFilter(request)
		if err != nil {
			k8sQuery = queryIds
		}
	}

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
		K8SIdQueryFilter:       k8sQuery,
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
