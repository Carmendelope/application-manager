/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
		"org1", "desc1", "inst1", map[string]string{"l1":"v1"}, []string{"g1"})
	var app2 = utils.CreateTestAppInstance(
		"org1", "desc1", "inst1", map[string]string{"l1":"v1", "l2":"v2"}, []string{"g1"})
	var app3 = utils.CreateTestAppInstance(
		"org1", "desc1", "inst1", map[string]string{"l3":"v3"}, []string{"g3"})

	var allApps = &grpc_application_go.AppInstanceList{
		Instances:            []*grpc_application_go.AppInstance{app1, app2, app3},
	}

	ginkgo.Context("ApplyFilters", func(){

		var emptyFilter = &grpc_application_manager_go.ApplicationFilter{
			OrganizationId:       "org1",
			DeviceGroupId:        "g1",
			MatchLabels:          nil,
		}

		ginkgo.It("should return empty on empty list", func(){
			result, err := ApplyFilter(allApps, emptyFilter)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(result).ShouldNot(gomega.BeNil())
			gomega.Expect(len(result.Instances)).Should(gomega.Equal(0))
		})

		ginkgo.It("should return empty if group does not match", func(){

		})

		ginkgo.It("should return empty if the labels do not match", func(){

		})

		ginkgo.It("should return all apps on empty labels filter with proper group", func(){

		})

		ginkgo.It("should filter applications based on labels", func(){

		})

	})

	ginkgo.Context("ToApplicationLabelsList", func(){
		ginkgo.It("should transform an empty list", func(){

		})
		ginkgo.It("should transform a list with apps", func(){

		})
	})


})