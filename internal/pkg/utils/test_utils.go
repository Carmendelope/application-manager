/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package utils

import (
	"github.com/nalej/grpc-application-go"
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
