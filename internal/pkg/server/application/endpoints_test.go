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

package application

import (
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Endpoints utils", func() {

	var app1 = utils.CreateTestAppInstance(
		"org1", "desc1", "inst1", map[string]string{"l1": "v1"}, []string{"g1"})
	var app2 = utils.CreateTestAppInstance(
		"org1", "desc1", "inst1", map[string]string{"l1": "v1", "l2": "v2"}, []string{"g1"})
	var app3 = utils.CreateTestAppInstance(
		"org1", "desc1", "inst1", map[string]string{"l3": "v3"}, []string{"g3"})

	var allApps = &grpc_application_go.AppInstanceList{
		Instances: []*grpc_application_go.AppInstance{app1, app2, app3},
	}

	ginkgo.Context("ApplyFilters", func() {

		var emptyFilter = &grpc_application_manager_go.ApplicationFilter{
			OrganizationId:  "org1",
			DeviceGroupName: "desc1",
			MatchLabels:     nil,
		}

		var noMatchGroupFilter = &grpc_application_manager_go.ApplicationFilter{
			OrganizationId:  "org1",
			DeviceGroupName: "g2",
			MatchLabels:     map[string]string{"l1": "v1"},
		}

		var noMatchLabelFilter = &grpc_application_manager_go.ApplicationFilter{
			OrganizationId:  "org1",
			DeviceGroupName: "g1",
			MatchLabels:     map[string]string{"l4": "v4"},
		}

		ginkgo.It("should return empty on empty list", func() {
			emptyApps := &grpc_application_go.AppInstanceList{
				Instances: []*grpc_application_go.AppInstance{},
			}

			result := ApplyFilter(emptyApps, emptyFilter)
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(0))
		})

		ginkgo.It("should return empty if group does not match", func() {
			result := ApplyFilter(allApps, noMatchGroupFilter)
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(0))
		})

		ginkgo.It("should return empty if the labels do not match", func() {
			result := ApplyFilter(allApps, noMatchLabelFilter)
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(0))
		})

		ginkgo.It("should return all apps on empty labels filter with proper group", func() {
			var filter = &grpc_application_manager_go.ApplicationFilter{
				OrganizationId:  "org1",
				DeviceGroupName: "g1",
				MatchLabels:     nil,
			}

			result := ApplyFilter(allApps, filter)
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(2))
		})

		ginkgo.It("should filter applications based on labels", func() {
			var filter = &grpc_application_manager_go.ApplicationFilter{
				OrganizationId:  "org1",
				DeviceGroupName: "g1",
				MatchLabels:     map[string]string{"l1": "v1"},
			}
			result := ApplyFilter(allApps, filter)
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(2))
		})

	})

	ginkgo.Context("ToApplicationLabelsList", func() {
		ginkgo.It("should transform an empty list", func() {
			emptyList := &grpc_application_go.AppInstanceList{
				Instances: []*grpc_application_go.AppInstance{},
			}

			result, err := ToApplicationLabelsList(emptyList)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Applications)).Should(gomega.Equal(0))

		})
		ginkgo.It("should transform a list with apps", func() {
			result, err := ToApplicationLabelsList(allApps)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Applications)).Should(gomega.Equal(3))
		})
	})

})
