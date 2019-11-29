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
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func create() *grpc_application_history_logs_go.LogResponse {
	s11 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:"1",
		ServiceGroupId: "1",
	}
	s12 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:"1",
		ServiceGroupId: "2",
	}
	s13 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "1",
		AppInstanceId:"2",
		ServiceGroupId: "1",
	}
	s21 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:"1",
		ServiceGroupId: "1",
	}
	s22 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:"2",
		ServiceGroupId: "1",
	}
	s23 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "2",
		AppInstanceId:"2",
		ServiceGroupId: "2",
	}
	s31 := &grpc_application_history_logs_go.ServiceInstanceLog{
		AppDescriptorId: "3",
		AppInstanceId:"1",
		ServiceGroupId: "1",
	}
	events := []*grpc_application_history_logs_go.ServiceInstanceLog{s11,s12,s13, s21, s22, s23, s31}
	return &grpc_application_history_logs_go.LogResponse{
		Events:events,
	}
}

var _ = ginkgo.Describe("Test", func() {
	ginkgo.Context("U-L", func() {
		ginkgo.It("-----", func() {
		 response := Organize(create())
		 gomega.Expect(response).NotTo(gomega.BeNil())

		})
	})
})