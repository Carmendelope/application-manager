/*
 * Copyright 2020 Nalej
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

package entities

import (
	"github.com/nalej/grpc-application-manager-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"strings"
)

func CreateMinimalDeploymentRequest(organizationID string, appDescriptorID string, name string) *grpc_application_manager_go.DeployRequest{
	return &grpc_application_manager_go.DeployRequest{
		OrganizationId:       organizationID,
		AppDescriptorId:      appDescriptorID,
		Name:                 name,
	}
}

var _ = ginkgo.Describe("Validation tests", func() {
	// NP-2237 Use the same validation as in the web component.
	ginkgo.Context("Deployment request names", func(){
		ginkgo.It("should fail on empty names", func(){
		    request := CreateMinimalDeploymentRequest("org", "desc", "")
		    err := ValidDeployRequestName(request)
		    gomega.Expect(err).Should(gomega.HaveOccurred())
		})
		ginkgo.It("should require a minimal length for the name", func(){
			name := strings.Repeat("a", MinDeployRequestNameLength -1)
			request := CreateMinimalDeploymentRequest("org", "desc", name)
			err := ValidDeployRequestName(request)
			gomega.Expect(err).Should(gomega.HaveOccurred())
		})
		ginkgo.It("Should accept valid names", func(){
			names := []string{"this", "isThe", "validName001"}
			for _, n := range names{
				request := CreateMinimalDeploymentRequest("org", "desc", n)
				err := ValidDeployRequestName(request)
				gomega.Expect(err).Should(gomega.Succeed())
			}
		})
	})
})
