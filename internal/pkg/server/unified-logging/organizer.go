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
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-application-manager-go"
)

func createDescriptorLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.AppDescriptorLogSummary{
	return &grpc_application_manager_go.AppDescriptorLogSummary{
		OrganizationId:       event.OrganizationId,
		AppDescriptorId:      event.AppDescriptorId,
		AppDescriptorName:    "missing",
		CurrentLabels:        nil,
		Instances:            []*grpc_application_manager_go.AppInstanceLogSummary{createInstanceLogSummary(event)},
	}
}

func createInstanceLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.AppInstanceLogSummary{
	return &grpc_application_manager_go.AppInstanceLogSummary{
		OrganizationId:       event.OrganizationId,
		AppInstanceId:        event.AppInstanceId,
		AppInstanceName:      "missing",
		AppDescriptorId:      event.AppDescriptorId,
		AppDescriptorName:    "missing",
		CurrentLabels:        nil,
		Groups:               []*grpc_application_manager_go.ServiceGroupInstanceLogSummary{createServiceGroupLogSummary(event)},
	}
}

func createServiceGroupLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.ServiceGroupInstanceLogSummary{
	return &grpc_application_manager_go.ServiceGroupInstanceLogSummary{
		ServiceGroupId:         event.ServiceGroupId,
		ServiceGroupInstanceId: event.ServiceGroupInstanceId,
		Name:                   "missing",
		ServiceInstances:       []*grpc_application_manager_go.ServiceInstanceLogSummary{createServiceInstanceLogSummary(event)},
	}
}

func createServiceInstanceLogSummary(event *grpc_application_history_logs_go.ServiceInstanceLog) *grpc_application_manager_go.ServiceInstanceLogSummary {
	return &grpc_application_manager_go.ServiceInstanceLogSummary{
		ServiceId:            event.ServiceId,
		ServiceInstanceId:    event.ServiceInstanceId,
		Name:                 "missing",
	}
}

func Organize(logResponse *grpc_application_history_logs_go.LogResponse) *grpc_application_manager_go.AvailableLogResponse {
	// LogResponse entries organized according to the structure needed
	var appDescriptorLogSummary *grpc_application_manager_go.AppDescriptorLogSummary
	var appInstanceLogSummary *grpc_application_manager_go.AppInstanceLogSummary
	var serviceGroupInstanceLogSummary *grpc_application_manager_go.ServiceGroupInstanceLogSummary

	appDescriptorLogSummaries := make([]*grpc_application_manager_go.AppDescriptorLogSummary, 0)
	appInstanceLogSummaries := make([]*grpc_application_manager_go.AppInstanceLogSummary, 0)

	for _, event := range logResponse.Events {

		// Descriptor
		found := false
		for i := 0; i < len(appDescriptorLogSummaries) && !found; i++ {
			if appDescriptorLogSummaries[i].AppDescriptorId == event.AppDescriptorId {
				found = true
				appDescriptorLogSummary = appDescriptorLogSummaries[i]
			}
		}
		if !found {
			appDescriptorLogSummaries = append(appDescriptorLogSummaries, createDescriptorLogSummary(event))
			continue
		}

		// Instance
		for i:=0; i<len(appDescriptorLogSummary.Instances) && !found; i++ {
			if appDescriptorLogSummary.Instances[i].AppInstanceId == event.AppInstanceId {
				found = true
				appInstanceLogSummary = appDescriptorLogSummary.Instances[i]
			}
		}
		if !found {
			appDescriptorLogSummary.Instances = append(appDescriptorLogSummary.Instances, createInstanceLogSummary(event))
			continue
		}

		// Service Group
		for i:=0; i<len(appInstanceLogSummary.Groups) && !found; i++ {
			if appInstanceLogSummary.Groups[i].ServiceGroupId == event.ServiceGroupId {
				found = true
				serviceGroupInstanceLogSummary = appInstanceLogSummary.Groups[i]
			}
		}
		if !found {
			appInstanceLogSummary.Groups = append(appInstanceLogSummary.Groups, createServiceGroupLogSummary(event))
			continue
		}

		// Service Instance
		for i:=0; i<len(serviceGroupInstanceLogSummary.ServiceInstances) && !found; i++ {
			if serviceGroupInstanceLogSummary.ServiceInstances[i].ServiceInstanceId == event.ServiceGroupId {
				found = true
			}
		}
		if !found {
			serviceGroupInstanceLogSummary.ServiceInstances = append(serviceGroupInstanceLogSummary.ServiceInstances, createServiceInstanceLogSummary(event))
			continue
		}
	}

	availableLogResponse := grpc_application_manager_go.AvailableLogResponse{
		OrganizationId:          logResponse.OrganizationId,
		AppDescriptorLogSummary: appDescriptorLogSummaries,
		AppInstanceLogSummary:   appInstanceLogSummaries,
		From:                    0,
		To:                      0,
	}

	return &availableLogResponse
}
