/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application_network

import (
	"github.com/google/uuid"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/nalej-bus/pkg/queue/network/ops"
	"github.com/rs/zerolog/log"
	"time"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	appNetClient   grpc_application_network_go.ApplicationNetworkClient
	appClient      grpc_application_go.ApplicationsClient
	netOpsProducer *ops.NetworkOpsProducer
}

// NewManager creates a Manager using a set of clients.
func NewManager(appNet grpc_application_network_go.ApplicationNetworkClient,
	appClient grpc_application_go.ApplicationsClient,
	netOpsProducer *ops.NetworkOpsProducer) Manager {
	return Manager{
		appNetClient:   appNet,
		appClient:      appClient,
		netOpsProducer: netOpsProducer,
	}
}

// AddConnection adds a new connection between one outbound and one inbound
func (m *Manager) AddConnection(addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_common_go.OpResponse, error) {

	ctxSource, cancelSource := common.GetContext()
	defer cancelSource()

	// Source & Outbound
	sourceInstance, err := m.appClient.GetAppInstance(ctxSource, &grpc_application_go.AppInstanceId{
		OrganizationId: addRequest.OrganizationId,
		AppInstanceId:  addRequest.SourceInstanceId,
	})
	if err != nil {
		return nil, err
	}

	if sourceInstance.OutboundNetInterfaces == nil {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("outbound_name does not exist").WithParams(addRequest.SourceInstanceId, addRequest.OutboundName))
	}

	outBoundFound := false
	for _, outbound := range sourceInstance.OutboundNetInterfaces {
		if outbound.Name == addRequest.OutboundName {
			outBoundFound = true
		}
	}
	if !outBoundFound {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("outbound_name does not exist").WithParams(addRequest.SourceInstanceId, addRequest.OutboundName))
	}

	// Target & Inbound
	ctxTarget, cancelTarget := common.GetContext()
	defer cancelTarget()
	targetInstance, err := m.appClient.GetAppInstance(ctxTarget, &grpc_application_go.AppInstanceId{
		OrganizationId: addRequest.OrganizationId,
		AppInstanceId:  addRequest.TargetInstanceId,
	})
	if err != nil {
		return nil, err
	}
	inBoundFound := false
	for _, inbound := range targetInstance.InboundNetInterfaces {
		if inbound.Name == addRequest.InboundName {
			inBoundFound = true
		}
	}
	if !inBoundFound {
		return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("inbound_name does not exist").WithParams(addRequest.TargetInstanceId, addRequest.InboundName))
	}

	// send the message to the queue
	ctxSend, cancelSend := common.GetContext()
	defer cancelSend()
	err = m.netOpsProducer.Send(ctxSend, addRequest)
	if err != nil {
		log.Error().Interface("connection", addRequest).Msg("error sending addConnection to the queue")
		return nil, err
	}

	return &grpc_common_go.OpResponse{
		OrganizationId: addRequest.OrganizationId,
		RequestId:      uuid.New().String(),
		Timestamp:      time.Now().Unix(),
		Status:         grpc_common_go.OpStatus_SCHEDULED,
		Info:           "Add Connection queued",
	}, nil
}

// RemoveConnection removes a connection
func (m *Manager) RemoveConnection(removeRequest *grpc_application_network_go.RemoveConnectionRequest) (*grpc_common_go.OpResponse, error) {

	// send the message to the queue
	ctxSend, cancelSend := common.GetContext()
	defer cancelSend()
	err := m.netOpsProducer.Send(ctxSend, removeRequest)
	if err != nil {
		log.Error().Interface("connection", removeRequest).Msg("error sending removeConnection to the queue")
		return nil, err
	}
	return &grpc_common_go.OpResponse{
		OrganizationId: removeRequest.OrganizationId,
		RequestId:      uuid.New().String(),
		Timestamp:      time.Now().Unix(),
		Status:         grpc_common_go.OpStatus_SCHEDULED,
		Info:           "Remove Connection queued",
	}, nil
}

// ListConnections retrieves a list all the established connections of an organization
func (m *Manager) ListConnections(orgID *grpc_organization_go.OrganizationId) (*grpc_application_network_go.ConnectionInstanceList, error) {
	ctx, cancel := common.GetContext()
	defer cancel()

	return m.appNetClient.ListConnections(ctx, orgID)
}