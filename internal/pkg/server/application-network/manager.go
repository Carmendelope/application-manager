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
 *
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

// getInboundServiceName returns the name of the service where the inbound is defined
func (m *Manager) getInboundServiceName(appInstance *grpc_application_go.AppInstance, inbound string) string {
	for _, rule := range appInstance.Rules {
		if rule.Access == grpc_application_go.PortAccess_INBOUND_APPNET && rule.InboundNetInterface == inbound {
			return rule.TargetServiceName
		}
	}
	return ""
}

// getutboundServiceName returns the name of the service where the outbound is defined
func (m *Manager) getOutboundServiceName(appInstance *grpc_application_go.AppInstance, outbound string) string {
	for _, rule := range appInstance.Rules {
		if rule.Access == grpc_application_go.PortAccess_OUTBOUND_APPNET && rule.OutboundNetInterface == outbound {
			return rule.TargetServiceName
		}
	}
	return ""
}

// AddConnection adds a new connection between one outbound and one inbound
func (m *Manager) AddConnection(addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_common_go.OpResponse, error) {

	// check it the connection already exists
	ctxGet, cancelGet := common.GetContext()
	defer cancelGet()
	exists, err := m.appNetClient.ExistsConnection(ctxGet, &grpc_application_network_go.ConnectionInstanceId{
		OrganizationId:   addRequest.OrganizationId,
		SourceInstanceId: addRequest.SourceInstanceId,
		OutboundName:     addRequest.OutboundName,
		TargetInstanceId: addRequest.TargetInstanceId,
		InboundName:      addRequest.InboundName,
	})
	if err != nil {
		return nil, err
	}
	if exists.Exists {
		return nil, conversions.ToGRPCError(derrors.NewAlreadyExistsError("connection").WithParams(
			addRequest.OrganizationId, addRequest.SourceInstanceId, addRequest.OutboundName, addRequest.TargetInstanceId, addRequest.InboundName))
	}

	ctxValidOutbounds, cancelValidOutbounds := common.GetContext()
	defer cancelValidOutbounds()
	outboundConnections, err := m.appNetClient.ListOutboundConnections(ctxValidOutbounds, &grpc_application_go.AppInstanceId{
		OrganizationId: addRequest.OrganizationId,
		AppInstanceId:  addRequest.SourceInstanceId,
	})
	if err != nil {
		return nil, err
	}
	hasOutboundUsed := false
	for _, existingConnection := range outboundConnections.Connections {
		if existingConnection.OutboundName == addRequest.OutboundName {
			hasOutboundUsed = true
			break
		}
	}
	if hasOutboundUsed {
		return nil, conversions.ToGRPCError(derrors.
			NewFailedPreconditionError("source instance already have its outbound interface connected").
			WithParams(addRequest))
	}

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

	// NP-2229. The inbound and outbound can not be in the same service
	if targetInstance.AppInstanceId == sourceInstance.AppInstanceId {
		if m.getInboundServiceName(targetInstance, addRequest.InboundName) == m.getOutboundServiceName(sourceInstance, addRequest.OutboundName) {
			return nil, conversions.ToGRPCError(derrors.NewInvalidArgumentError("Can not create a connection between an inbound and an outbound of the same service"))
		}
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

	// check if the connection exists
	ctx, cancel := common.GetContext()
	defer cancel()
	conn, vErr := m.appNetClient.GetConnection(ctx, &grpc_application_network_go.ConnectionInstanceId{
		OrganizationId:   removeRequest.OrganizationId,
		SourceInstanceId: removeRequest.SourceInstanceId,
		TargetInstanceId: removeRequest.TargetInstanceId,
		InboundName:      removeRequest.InboundName,
		OutboundName:     removeRequest.OutboundName,
	})
	if vErr != nil {
		return nil, vErr
	}

	if conn.OutboundRequired && !removeRequest.UserConfirmation {
		errorMsg := "A connection in which the outbound is required cannot be deleted; the application could have an unexpected result. If it must be deleted, user confirmation is required"
		return nil, conversions.ToGRPCError(derrors.NewFailedPreconditionError(errorMsg))
	}

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
