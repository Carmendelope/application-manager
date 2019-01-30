/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
)

// ApplyFilter filters out applications that do not allow the device group and whose labels do not match the filter.
func ApplyFilter(appList *grpc_application_go.AppInstanceList, filter *grpc_application_manager_go.ApplicationFilter) (*grpc_application_go.AppInstanceList, derrors.Error){
	return nil, nil
}

// ToApplicationLabelsList transforms the result from the filter into the TargetApplications object.
func ToApplicationLabelsList(appList *grpc_application_go.AppInstanceList) (*grpc_application_manager_go.TargetApplications, derrors.Error){
	return nil, nil
}