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
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-application-manager-go"
)

func (m *Manager) createDescriptorLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.AppDescriptorLogSummary {
	instanceNames := m.instHelper.GetNames(event.OrganizationId, event.AppInstanceId, event.ServiceGroupId, event.ServiceId)
	labels := m.instHelper.GetLabels(event.OrganizationId, event.AppDescriptorId)
	return &grpc_application_manager_go.AppDescriptorLogSummary{
		OrganizationId:    event.OrganizationId,
		AppDescriptorId:   event.AppDescriptorId,
		AppDescriptorName: instanceNames.AppDescriptorName,
		CurrentLabels:     labels,
		Instances:         []*grpc_application_manager_go.AppInstanceLogSummary{m.createInstanceLogSummary(event)},
	}
}

func (m *Manager) createInstanceLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.AppInstanceLogSummary {
	instanceNames := m.instHelper.GetNames(event.OrganizationId, event.AppInstanceId, event.ServiceGroupId, event.ServiceId)
	labels := m.instHelper.GetLabels(event.OrganizationId, event.AppDescriptorId)
	return &grpc_application_manager_go.AppInstanceLogSummary{
		OrganizationId:    event.OrganizationId,
		AppInstanceId:     event.AppInstanceId,
		AppInstanceName:   instanceNames.AppInstanceName,
		AppDescriptorId:   event.AppDescriptorId,
		AppDescriptorName: instanceNames.AppDescriptorName,
		CurrentLabels:     labels,
		Groups:            []*grpc_application_manager_go.ServiceGroupInstanceLogSummary{m.createServiceGroupLogSummary(event)},
	}
}

func (m *Manager) createServiceGroupLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.ServiceGroupInstanceLogSummary {
	instanceNames := m.instHelper.GetNames(event.OrganizationId, event.AppInstanceId, event.ServiceGroupId, event.ServiceId)
	return &grpc_application_manager_go.ServiceGroupInstanceLogSummary{
		ServiceGroupId:         event.ServiceGroupId,
		ServiceGroupInstanceId: event.ServiceGroupInstanceId,
		Name:                   instanceNames.ServiceGroupName,
		ServiceInstances:       []*grpc_application_manager_go.ServiceInstanceLogSummary{m.createServiceInstanceLogSummary(event)},
	}
}

func (m *Manager) createServiceInstanceLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.ServiceInstanceLogSummary {
	instanceNames := m.instHelper.GetNames(event.OrganizationId, event.AppInstanceId, event.ServiceGroupId, event.ServiceId)
	return &grpc_application_manager_go.ServiceInstanceLogSummary{
		ServiceId:         event.ServiceId,
		ServiceInstanceId: event.ServiceInstanceId,
		Name:              instanceNames.ServiceName,
	}
}

func (m *Manager) Organize(logResponse *grpc_application_history_logs_go.LogResponse) *grpc_application_manager_go.AvailableLogResponse {
	// LogResponse entries organized according to the structure needed
	//var appDescriptorLogSummary *grpc_application_manager_go.AppDescriptorLogSummary
	//var appInstanceLogSummary *grpc_application_manager_go.AppInstanceLogSummary
	//var serviceGroupInstanceLogSummary *grpc_application_manager_go.ServiceGroupInstanceLogSummary

	appDescriptorLogSummaries := make([]*grpc_application_manager_go.AppDescriptorLogSummary, 0)
	appInstanceLogSummaries := make([]*grpc_application_manager_go.AppInstanceLogSummary, 0)

	for _, event := range logResponse.Events {

		var appDescriptorLogSummary *grpc_application_manager_go.AppDescriptorLogSummary
		var appInstanceLogSummary *grpc_application_manager_go.AppInstanceLogSummary
		var serviceGroupInstanceLogSummary *grpc_application_manager_go.ServiceGroupInstanceLogSummary

		// Descriptor
		found := false
		for i := 0; i < len(appDescriptorLogSummaries) && !found && appDescriptorLogSummaries != nil; i++ {
			if appDescriptorLogSummaries[i].AppDescriptorId == event.AppDescriptorId {
				found = true
				appDescriptorLogSummary = appDescriptorLogSummaries[i]
			}
		}
		if !found {
			appDescriptorLogSummaries = append(appDescriptorLogSummaries, m.createDescriptorLogSummary(event))
			continue
		}

		// Instance
		found = false
		for i := 0; i < len(appDescriptorLogSummary.Instances) && !found && appDescriptorLogSummary.Instances != nil; i++ {
			if appDescriptorLogSummary.Instances[i].AppInstanceId == event.AppInstanceId {
				found = true
				appInstanceLogSummary = appDescriptorLogSummary.Instances[i]
			}
		}
		if !found {
			appDescriptorLogSummary.Instances = append(appDescriptorLogSummary.Instances, m.createInstanceLogSummary(event))
			continue
		}

		// Service Group
		found = false
		for i := 0; appInstanceLogSummary != nil && i < len(appInstanceLogSummary.Groups) && !found; i++ {
			if appInstanceLogSummary.Groups[i].ServiceGroupId == event.ServiceGroupId {
				found = true
				serviceGroupInstanceLogSummary = appInstanceLogSummary.Groups[i]
			}
		}
		if !found {
			appInstanceLogSummary.Groups = append(appInstanceLogSummary.Groups, m.createServiceGroupLogSummary(event))
			continue
		}

		// Service Instance
		found = false
		for i := 0; serviceGroupInstanceLogSummary != nil && i < len(serviceGroupInstanceLogSummary.ServiceInstances) && !found; i++ {
			if serviceGroupInstanceLogSummary.ServiceInstances[i].ServiceInstanceId == event.ServiceGroupId {
				found = true
			}
		}
		if !found {
			serviceGroupInstanceLogSummary.ServiceInstances = append(serviceGroupInstanceLogSummary.ServiceInstances, m.createServiceInstanceLogSummary(event))
			continue
		}
	}

	for _, appDescriptorLogSummary := range appDescriptorLogSummaries {
		for _, appInstanceLogSummary := range appDescriptorLogSummary.Instances {
			appInstanceLogSummaries = append(appInstanceLogSummaries, appInstanceLogSummary)
		}
	}

	availableLogResponse := grpc_application_manager_go.AvailableLogResponse{
		OrganizationId:          logResponse.OrganizationId,
		AppDescriptorLogSummary: appDescriptorLogSummaries,
		AppInstanceLogSummary:   appInstanceLogSummaries,
		From:                    logResponse.From,
		To:                      logResponse.To,
	}

	return &availableLogResponse
}
