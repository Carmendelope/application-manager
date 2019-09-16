/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application_network

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	appNetClient grpc_application_network_go.ApplicationNetworkClient
	appClient grpc_application_go.ApplicationsClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(appNet grpc_application_network_go.ApplicationNetworkClient, appClient grpc_application_go.ApplicationsClient) Manager{
	return Manager{
		appNetClient: appNet,
		appClient: appClient,
	}
}

// AddConnection adds a new connection between one outbound and one inbound
func (m *Manager) AddConnection(addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_application_network_go.ConnectionInstance, error){

	ctx, cancel := common.GetContext()
	defer cancel()


	instanceList, err := m.appClient.ListAppInstances(ctx, &grpc_organization_go.OrganizationId{
		OrganizationId: addRequest.OrganizationId,
	})

	if err != nil {
		return nil, err
	}

	// check if source_instance_id exists
	if ! common.InstanceExists(instanceList, addRequest.SourceInstanceId) {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("source_instance_id does not exist").WithParams(addRequest.SourceInstanceId))
	}
	// check if target_instance_id exists
	if ! common.InstanceExists(instanceList, addRequest.TargetInstanceId) {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("target_instance_id does not exist").WithParams(addRequest.TargetInstanceId))
	}

	// check if the inboundName exists (the inboundName should be defined in the targetInstance)
	if ! common.InboundExists(instanceList, addRequest.TargetInstanceId, addRequest.InboundName) {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("inbound_name does not exist").WithParams(addRequest.InboundName, addRequest.TargetInstanceId))
	}
	// check if the outboundName exists (the outboundName should be defined in the sourceInstance)
	if ! common.OutboundExists(instanceList, addRequest.SourceInstanceId, addRequest.OutboundName) {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("outbound_name does not exist").WithParams(addRequest.OutboundName, addRequest.SourceInstanceId))
	}
	return m.appNetClient.AddConnection(context.Background(), addRequest)
}
// RemoveConnection removes a connection
func (m *Manager) RemoveConnection(removeRequest *grpc_application_network_go.RemoveConnectionRequest) (*grpc_common_go.Success, error) {
	ctx, cancel := common.GetContext()
	defer cancel()

	return m.appNetClient.RemoveConnection(ctx, removeRequest)
}
// ListConnections retrieves a list all the established connections of an organization
func (m *Manager) ListConnections(orgID *grpc_organization_go.OrganizationId) (*grpc_application_network_go.ConnectionInstanceList, error){
	ctx, cancel := common.GetContext()
	defer cancel()

	return m.appNetClient.ListConnections(ctx, orgID)
}