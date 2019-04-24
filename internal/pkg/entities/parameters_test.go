/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)


var _ = ginkgo.Describe("Parameter tests", func() {

	ginkgo.Context("Parametrized descriptor", func(){
		ginkgo.It("Should be able to create a Parametrized descriptor from GRPC ", func(){

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
				Parameters:[]*grpc_application_go.InstanceParameter{{ParameterName:"replicas", Value:"10"}, {ParameterName:"env1", Value:"modified"}},
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
				Parameters:[]*grpc_application_go.InstanceParameter{{ParameterName:"replicas", Value:"replicas test"}},
			}
			_, err := CreateParametrizedDescriptor(descriptor, &parameters)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not be able to create parametrized Descriptor with invalid parameter", func() {
			descriptor := utils.CreateTestDescriptorWithParameters()
			parameters := grpc_application_go.InstanceParameterList{
				Parameters:[]*grpc_application_go.InstanceParameter{{ParameterName:"invalid_param", Value:"10"}},
			}
			_, err := CreateParametrizedDescriptor(descriptor, &parameters)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

	})


})
