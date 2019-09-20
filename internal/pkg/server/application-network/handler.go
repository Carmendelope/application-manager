/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application_network

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

// Handler structure for the user requests.
type Handler struct {
	Manager Manager
}

// NewHandler creates a new Handler with a linked manager.
func NewHandler(manager Manager) *Handler{
	return &Handler{manager}
}

// AddConnection adds a new connection between one outbound and one inbound
func (h *Handler) AddConnection(_ context.Context, addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_common_go.OpResponse, error){

	vErr := entities.ValidAddConnectionRequest(addRequest)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.AddConnection(addRequest)
}
// RemoveConnection removes a connection
func (h *Handler) RemoveConnection(_ context.Context, removeRequest *grpc_application_network_go.RemoveConnectionRequest) (*grpc_common_go.OpResponse, error) {
	vErr := entities.ValidRemoveConnectionRequest(removeRequest)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.RemoveConnection(removeRequest)
}

// ListConnections retrieves a list all the established connections of an organization
func (h *Handler) ListConnections(_ context.Context, orgID *grpc_organization_go.OrganizationId) (*grpc_application_network_go.ConnectionInstanceList, error){
	vErr := entities.ValidOrganizationId(orgID)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.ListConnections(orgID)
}

