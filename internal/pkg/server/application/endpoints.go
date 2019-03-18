/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
)



func compareDeviceGroupId (instance *grpc_application_go.AppInstance, filter *grpc_application_manager_go.ApplicationFilter) bool {

	if instance.Rules == nil {
		return false
	}

	if filter == nil || filter.DeviceGroupId == "" {
		return true
	}

	for _, rule := range instance.Rules {
		if rule.Access == grpc_application_go.PortAccess_DEVICE_GROUP {
			for _, groupId := range rule.DeviceGroupIds {
				if groupId == filter.DeviceGroupId {
					return true
				}
			}
		}
	}
	return false
}

func compareLabels (instance *grpc_application_go.AppInstance, filter *grpc_application_manager_go.ApplicationFilter) bool {
	// if filter has no labels to match -> true
	if filter.MatchLabels == nil {
		return true
	}
	// if instance has no labels (and filter does) -> false
	if instance.Labels == nil {
		return false
	}

	// match labels...
	for key, value := range filter.MatchLabels {
		label, ok := instance.Labels[key]
		if !ok || label != value {
			return false
		}
	}
	return true

}

// ApplyFilter filters out applications that do not allow the device group and whose labels do not match the filter.
func ApplyFilter(appList *grpc_application_go.AppInstanceList, filter *grpc_application_manager_go.ApplicationFilter) (*grpc_application_go.AppInstanceList){

	appInstances := make ([]*grpc_application_go.AppInstance, 0)

	for _, instance := range appList.Instances {

		if filter.OrganizationId == instance.OrganizationId && compareDeviceGroupId(instance, filter) && compareLabels(instance, filter) {
			appInstances = append(appInstances, instance)
		}

	}
	return &grpc_application_go.AppInstanceList{
		Instances: appInstances,
	}

}

// ToApplicationLabelsList transforms the result from the filter into the TargetApplications object.
func ToApplicationLabelsList(appList *grpc_application_go.AppInstanceList) (*grpc_application_manager_go.TargetApplicationList, derrors.Error){

	applications := make([]*grpc_application_manager_go.TargetApplication, 0)

	for _, app := range appList.Instances {
		target := &grpc_application_manager_go.TargetApplication{
			AppInstanceId: app.AppInstanceId,
			Labels: app.Labels,
		}
		applications = append(applications, target)
	}

	return &grpc_application_manager_go.TargetApplicationList	{
		Applications:applications,
	}, nil

}
