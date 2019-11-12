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
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func createAllowedPathParameters() []*grpc_application_go.AppParameter {

	return []*grpc_application_go.AppParameter{
		{Path: "groups.0.services.0.environment_variables.ORGANIZATION_ID"},
		{Path: "groups.10.services.0.specs.cpu"},
		{Path: "groups.10.services.0.storage.1.size"},
		{Path: "rules.11.device_group_names.0"},
		{Path: "configuration_options.CONF_ID"},
		{Path: "environment_variables.HOST_ID"},
		{Path: "labels.LABEL_NAME"},
		{Path: "rules.11.device_group_names.0"},
		{Path: "groups.0.policy.0.environment_variables.ORGANIZATION_ID"},
		{Path: "groups.0.specs.replicas"},
	}
}

func createNotAllowedPathParameters() []*grpc_application_go.AppParameter {

	return []*grpc_application_go.AppParameter{
		{Path: "groups.10.services.0.storage.1.mount_path"},
		{Path: "rules.11.auth_service_group_name"},
		{Path: "groups.0.services.0.credentials.username"},
		{Path: "groups.10.services.0.exposed_ports.2.internal_port"},
		{Path: "groups.10.services.0.deploy_after.2"},
	}
}

var _ = ginkgo.Describe("Parameter tests", func() {

	ginkgo.Context("Parametrized descriptor", func() {
		ginkgo.It("Should be able to create a Parametrized descriptor from GRPC ", func() {

			descriptor := utils.CreateTestDescriptor()

			parametrized := newParametrizedDescriptorFromDescriptor(descriptor)
			gomega.Expect(parametrized).NotTo(gomega.BeNil())

			// update descriptor and parametrized to check they are not the same
			descriptor.Rules[0].Name = "name modified"
			parametrized.Groups[0].Services[0].Name = "service name modified"

			gomega.Expect(descriptor.Rules[0].Name).ShouldNot(gomega.Equal(parametrized.Rules[0].Name))
			gomega.Expect(parametrized.Groups[0].Services[0].Name).ShouldNot(gomega.Equal(descriptor.Groups[0].Services[0].Name))

		})
		ginkgo.It("should be able to create parametrized Descriptor with parameters", func() {
			descriptor := utils.CreateTestDescriptorWithParameters()
			parameters := grpc_application_go.InstanceParameterList{
				Parameters: []*grpc_application_go.InstanceParameter{{ParameterName: "replicas", Value: "10"}, {ParameterName: "env1", Value: "modified"}},
			}
			parametrized, err := CreateParametrizedDescriptor(descriptor, &parameters)
			gomega.Expect(parametrized).NotTo(gomega.BeNil())
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(parametrized.Groups[0].Specs.Replicas).Should(gomega.Equal(int32(10)))
			gomega.Expect(parametrized.Groups[0].Services[0].EnvironmentVariables["env1"]).Should(gomega.Equal("modified"))
		})
		ginkgo.It("should not be able to create parametrized Descriptor with parameters (invalid type)", func() {
			descriptor := utils.CreateTestDescriptorWithParameters()
			parameters := grpc_application_go.InstanceParameterList{
				Parameters: []*grpc_application_go.InstanceParameter{{ParameterName: "replicas", Value: "replicas test"}},
			}
			_, err := CreateParametrizedDescriptor(descriptor, &parameters)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not be able to create parametrized Descriptor with invalid parameter", func() {
			descriptor := utils.CreateTestDescriptorWithParameters()
			parameters := grpc_application_go.InstanceParameterList{
				Parameters: []*grpc_application_go.InstanceParameter{{ParameterName: "invalid_param", Value: "10"}},
			}
			_, err := CreateParametrizedDescriptor(descriptor, &parameters)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("must recognize the parameters as valid", func() {

			descriptor := utils.CreateTestAddDescriptorWithParameters()
			err := ValidateDescriptorParameters(descriptor)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("must recognize the parameters as valid (name defined twice)", func() {

			descriptor := utils.CreateTestAddDescriptorWithParameters()
			descriptor.Parameters = append(descriptor.Parameters, &grpc_application_go.AppParameter{
				Name: descriptor.Parameters[0].Name,
				Type: descriptor.Parameters[0].Type,
				Path: descriptor.Parameters[0].Path})
			err := ValidateDescriptorParameters(descriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to validate allowed parameters", func() {
			params := createAllowedPathParameters()
			for _, p := range params {
				err := validateAllowedParameter(p)
				gomega.Expect(err).To(gomega.BeNil())
			}

		})
		ginkgo.It("should not be able to validate allowed parameters", func() {
			params := createNotAllowedPathParameters()
			for _, p := range params {
				err := validateAllowedParameter(p)
				gomega.Expect(err).NotTo(gomega.BeNil())
			}

		})

	})

	ginkgo.Context("Inbound and Outbound validation", func() {

		ginkgo.It("Should be able to Valid the descriptor ", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).To(gomega.Succeed())
		})

		ginkgo.It("Should not be able to Valid the descriptor, inbound defined twice ", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.InboundNetInterfaces = append(appDesc.InboundNetInterfaces, &grpc_application_go.InboundNetworkInterface{Name: "inbound1"})
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to Valid the descriptor, inbound and outbound with the same name", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.InboundNetInterfaces = append(appDesc.InboundNetInterfaces, &grpc_application_go.InboundNetworkInterface{Name: "outbound1"})
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

		ginkgo.It("Should not be able to Valid the descriptor, outbound defined twice ", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.OutboundNetInterfaces = append(appDesc.OutboundNetInterfaces, &grpc_application_go.OutboundNetworkInterface{Name: "outbound1", Required: false})
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to Valid the descriptor, inbound and outbound with the same name", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.OutboundNetInterfaces = append(appDesc.OutboundNetInterfaces, &grpc_application_go.OutboundNetworkInterface{Name: "inbound1", Required: false})
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

		// The rules must be informed with existing interfaces
		ginkgo.It("Should not be able to Valid the descriptor, invalid interface in the inbound rule", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Rules[0].InboundNetInterface = "wrong inbound"
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to Valid the descriptor, invalid interface in the outbound rule", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Rules[1].InboundNetInterface = "wrong outbound"
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

		ginkgo.It("Should not be able to Valid the descriptor, invalid ACCESS. Outbound rule with inbound interface", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Rules[0].Access = grpc_application_go.PortAccess_OUTBOUND_APPNET
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to Valid the descriptor, invalid ACCESS. Inbound rule with outbound interface", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Rules[1].Access = grpc_application_go.PortAccess_INBOUND_APPNET
			err := ValidDescriptorLogic(appDesc)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

		ginkgo.It("Is not valid to have inbound interfaces not linked to a rule", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.InboundNetInterfaces = append(appDesc.InboundNetInterfaces, &grpc_application_go.InboundNetworkInterface{Name: "inbound2"})
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})
		ginkgo.It("Is not valid to have outbound interfaces not linked to a rule", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.OutboundNetInterfaces = append(appDesc.OutboundNetInterfaces, &grpc_application_go.OutboundNetworkInterface{Name: "inbound2", Required: false})
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})

		ginkgo.It("Is not valid to have a rule with both inboundNetInterface and outboundNetInterface fields informed", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Rules[0].OutboundNetInterface = "invalidOutbound"
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})

		ginkgo.It("Is not valid to have an inbound rule without inbound interfaces", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.InboundNetInterfaces = nil
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})
		ginkgo.It("Is not valid to have an outbound rule without outbound interfaces", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.OutboundNetInterfaces = nil
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})

		ginkgo.It("Is not valid to link an inbound interface to a multi replica group service", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Groups[0].Specs = &grpc_application_go.ServiceGroupDeploymentSpecs{Replicas: 2}
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})
		ginkgo.It("Is not valid to link an inbound interface to a multiClusterReplica enabled group service", func() {
			appDesc := utils.CreateAppDescriptorWithInboundAndOutbounds()
			appDesc.Groups[0].Specs = &grpc_application_go.ServiceGroupDeploymentSpecs{MultiClusterReplica: true}
			gomega.Expect(ValidDescriptorLogic(appDesc)).ToNot(gomega.Succeed())
		})
	})

})
