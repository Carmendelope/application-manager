/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */
package common

import "github.com/nalej/grpc-application-go"

// this file contains some util functions to check inbounds and outbounds of an organization


func getInstance (list *grpc_application_go.AppInstanceList, appInstanceID string) *grpc_application_go.AppInstance{
	for _, instance := range list.Instances {
		if instance.AppInstanceId == appInstanceID {
			return instance
		}
	}
	return nil
}

func InstanceExists(list *grpc_application_go.AppInstanceList, appInstanceID string) bool {

	instance := getInstance (list, appInstanceID)

	if instance != nil {
		return true
	}

	return false
}

func InboundExists (list *grpc_application_go.AppInstanceList, appInstanceID string, inboundName string) bool {

	instance := getInstance (list, appInstanceID)

	if instance != nil {
		return true
	}
	if instance.InboundNetInterfaces != nil {
		for _, inbound := range instance.InboundNetInterfaces {
			if inbound.Name == inboundName {
				return true
			}
		}
	}

	return false
}

func OutboundExists (list *grpc_application_go.AppInstanceList, appInstanceID string, outboundName string) bool {

	instance := getInstance (list, appInstanceID)

	if instance != nil {
		return true
	}
	if instance.OutboundNetInterfaces != nil {
		for _, outbound := range instance.OutboundNetInterfaces {
			if outbound.Name == outboundName {
				return true
			}
		}
	}

	return false
}