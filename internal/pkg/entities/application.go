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

package entities

import (
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
)

func ToAppInstance(source *grpc_application_go.AppInstance) *grpc_application_manager_go.AppInstance {

	return &grpc_application_manager_go.AppInstance{
		OrganizationId:        source.OrganizationId,
		AppDescriptorId:       source.AppDescriptorId,
		AppInstanceId:         source.AppInstanceId,
		Name:                  source.Name,
		ConfigurationOptions:  source.ConfigurationOptions,
		EnvironmentVariables:  source.EnvironmentVariables,
		Labels:                source.Labels,
		Rules:                 source.Rules,
		Groups:                source.Groups,
		Status:                source.Status,
		Metadata:              source.Metadata,
		Info:                  source.Info,
		InboundNetInterfaces:  source.InboundNetInterfaces,
		OutboundNetInterfaces: source.OutboundNetInterfaces,
	}
}
type AppDescriptorLogSummary struct {
	OrganizationId string
	AppDescriptorId string
	AppDescriptorName string
	CurrentLabels map[string]string
	Instances []AppInstanceLogSummary
}

type AppInstanceLogSummary struct {
	OrganizationId string
	AppInstanceId string
	AppInstanceName string
	AppDescriptorId string
	AppDescriptorName string
	CurrentLabels map[string]string
	Groups []ServiceGroupInstanceLogSummary
}

type ServiceGroupInstanceLogSummary struct {
	ServiceGroupId string
	ServiceGroupInstanceId string
	Name string
	ServiceInstances []ServiceInstanceLogSummary
}

type ServiceInstanceLogSummary struct {
	ServiceId string
	ServiceInstanceId string
	Name string
}