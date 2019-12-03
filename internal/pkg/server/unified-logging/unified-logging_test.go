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
	grpc_application_go "github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-utils/pkg/test"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"time"
)

func createLogResponse() *grpc_application_history_logs_go.LogResponse {
	organizationId := "org"
	s11 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:   "1",
		ServiceGroupId:  "1",
	}
	s12 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:   "1",
		ServiceGroupId:  "2",
	}
	s13 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:   "2",
		ServiceGroupId:  "1",
	}
	s21 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:   "1",
		ServiceGroupId:  "1",
	}
	s22 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:   "2",
		ServiceGroupId:  "1",
	}
	s23 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:   "2",
		ServiceGroupId:  "2",
	}
	s31 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "3",
		AppInstanceId:   "1",
		ServiceGroupId:  "1",
	}
	events := []*grpc_application_history_logs_go.ServiceInstanceLog{s11, s12, s13, s21, s22, s23, s31}
	return &grpc_application_history_logs_go.LogResponse{
		OrganizationId: organizationId,
		From:           time.Now().UnixNano() - 50*time.Hour.Nanoseconds(),
		To:             time.Now().UnixNano() + 50*time.Hour.Nanoseconds(),
		Events:         events,
	}
}

var _ = ginkgo.Describe("Test", func() {
	// gRPC server
	var server *grpc.Server
	// grpc test listener
	var listener *bufconn.Listener
	// manager
	var manager *Manager

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		appcConn, err := test.GetConn(*listener)
		gomega.Expect(err).Should(gomega.Succeed())
		appsClient := grpc_application_go.NewApplicationsClient(appcConn)

		// Register the service
		manager, _ = NewManager(nil, appsClient, nil, nil)

		test.LaunchServer(server, listener)
	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		_ = listener.Close()
	})

	ginkgo.Context("U-L", func() {
		ginkgo.It("-----", func() {
			availableLogResponse := manager.Organize(createLogResponse())
			gomega.Expect(availableLogResponse).NotTo(gomega.BeNil())
		})
	})
})
