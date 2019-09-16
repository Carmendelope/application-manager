/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */
package common

import "github.com/nalej/grpc-application-go"

// this file contains some util functions to check inbounds and outbounds of an organization

// getInstance returns the instance whose name matches the instanceId if it exists in the list of instances
func getInstance (list *grpc_application_go.AppInstanceList, appInstanceID string) *grpc_application_go.AppInstance{
	for _, instance := range list.Instances {
		if instance.AppInstanceId == appInstanceID {
			return instance
		}
	}
	return nil
}

// InstanceExists check if an instance exist in the list
func InstanceExists(list *grpc_application_go.AppInstanceList, appInstanceID string) bool {

	instance := getInstance (list, appInstanceID)

	if instance != nil {
		return true
	}

	return false
}

// InboundExists check if the exists an instance with appInstanceID identifier and it has an inbound with the inboundName defined
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

// OutboundExists check if the exists an instance with appInstanceID identifier and it has an outbound with the inboundName defined
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