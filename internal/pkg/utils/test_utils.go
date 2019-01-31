/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nalej/grpc-application-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"os"
)

// RunIntegrationTests checks whether integration tests should be executed.
func RunIntegrationTests() bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	return runIntegration == "true"
}

func GetConnection(address string) *grpc.ClientConn {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	gomega.Expect(err).To(gomega.Succeed())
	return conn
}

func CreateTestAppInstance(organizationID string, appDescriptorID string, appInstanceID string, labels map[string]string, groups []string) *grpc_application_go.AppInstance {
	service := &grpc_application_go.ServiceInstance{
		OrganizationId:      "",
		AppDescriptorId:     "",
		AppInstanceId:       "",
		ServiceId:           "service1",
		Endpoints:           nil,
		DeployedOnClusterId: "",
	}
	sr := &grpc_application_go.SecurityRule{
		OrganizationId:       organizationID,
		AppDescriptorId:      appDescriptorID,
		RuleId:               "rule1",
		Name:                 "",
		SourceServiceId:      "service1",
		SourcePort:           80,
		Access:               grpc_application_go.PortAccess_DEVICE_GROUP,
		DeviceGroups:         groups,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	return &grpc_application_go.AppInstance{
		OrganizationId:  organizationID,
		AppDescriptorId: appDescriptorID,
		AppInstanceId:   appInstanceID,
		Labels:          labels,
		Rules:           []*grpc_application_go.SecurityRule{sr},
		Services:        []*grpc_application_go.ServiceInstance{service},
	}
}

func CreateTestAppInstanceRequest (organizationID string, appDescriptorID string) *grpc_application_go.AddAppInstanceRequest {

	return &grpc_application_go.AddAppInstanceRequest{
		OrganizationId:  organizationID,
		AppDescriptorId: appDescriptorID,
		Name:   "test",
	}
}

func CreateAddAppDescriptorRequest(organizationID string, groups []string, labels map[string]string) * grpc_application_go.AddAppDescriptorRequest{
	service := &grpc_application_go.Service{
		OrganizationId:       organizationID,
		ServiceId:            uuid.New().String(),
		Name:                 "Service-test",
		Description:          "minimal IT nginx",
		Type:                 grpc_application_go.ServiceType_DOCKER,
		Image:                "nginx:1.12",
		Specs:                &grpc_application_go.DeploySpecs{
			Replicas:             1,
		},
	}
	rules := make([]*grpc_application_go.SecurityRule, 0)
	rule := &grpc_application_go.SecurityRule {
		OrganizationId: organizationID,
		Name: "SecurityRule (it)",
		Access: grpc_application_go.PortAccess_DEVICE_GROUP,
		DeviceGroups: groups,
	}
	rules = append(rules, rule)


	toAdd := &grpc_application_go.AddAppDescriptorRequest{
		RequestId:            fmt.Sprintf("application-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		OrganizationId:       organizationID,
		Name:                 fmt.Sprintf("app-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		Description:          "Device app descriptor descriptor",
		ConfigurationOptions: nil,
		EnvironmentVariables: nil,
		Labels:               labels,
		Rules:                rules,
		Groups:               nil,
		Services:             []*grpc_application_go.Service{service},
	}
	return toAdd
}