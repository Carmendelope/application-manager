/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package application

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
)

// Handler structure for the user requests.
type Handler struct {
	Manager Manager
}

// NewHandler creates a new Handler with a linked manager.
func NewHandler(manager Manager) *Handler{
	return &Handler{manager}
}

// AddAppDescriptor adds a new application descriptor to a given organization.
func (h * Handler) AddAppDescriptor(ctx context.Context, addDescriptorRequest *grpc_application_go.AddAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {
	log.Debug().Str("organizationID", addDescriptorRequest.OrganizationId).
		Str("name", addDescriptorRequest.Name).Msg("add application descriptor")
	vErr := entities.ValidAddAppDescriptorRequest(addDescriptorRequest)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.AddAppDescriptor(addDescriptorRequest)
}

// ListAppDescriptors retrieves a list of application descriptors.
func (h * Handler) ListAppDescriptors(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppDescriptorList, error) {
	vErr := entities.ValidOrganizationId(organizationID)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.ListAppDescriptors(organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (h * Handler) GetAppDescriptor(ctx context.Context, appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppDescriptor, error) {
	vErr := entities.ValidAppDescriptorID(appDescriptorID)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.GetAppDescriptor(appDescriptorID)
}

// UpdateAppDescriptor allows the user to update the information of a registered descriptor.
func (h * Handler) UpdateAppDescriptor(ctx context.Context, request *grpc_application_go.UpdateAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {
	vErr := entities.ValidUpdateAppDescriptorRequest(request)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.UpdateAppDescriptor(request)
}


// RemoveAppDescriptor removes an application descriptor from the system.
func (h*Handler) RemoveAppDescriptor(ctx context.Context, appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_common_go.Success, error){
	vErr := entities.ValidAppDescriptorID(appDescriptorID)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.RemoveAppDescriptor(appDescriptorID)
}

// Deploy an application descriptor.
func (h * Handler) Deploy(ctx context.Context, deployRequest *grpc_application_manager_go.DeployRequest) (*grpc_application_manager_go.DeploymentResponse, error) {
	log.Debug().Str("organizationID", deployRequest.OrganizationId).
		Str("appDescriptorId", deployRequest.AppDescriptorId).Msg("deploy application")
	vErr := entities.ValidDeployRequest(deployRequest)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.Deploy(deployRequest)
}

// Undeploy a running application instance.
func (h * Handler) Undeploy(ctx context.Context, appInstanceID *grpc_application_go.AppInstanceId) (*grpc_common_go.Success, error) {
	log.Debug().Str("organizationID", appInstanceID.OrganizationId).
		Str("appInstanceId", appInstanceID.AppInstanceId).Msg("undeploy application")
	vErr := entities.ValidAppInstanceID(appInstanceID)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.Undeploy(appInstanceID)
}

// ListAppInstances retrieves a list of application descriptors.
func (h * Handler) ListAppInstances(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AppInstanceList, error) {
	vErr := entities.ValidOrganizationId(organizationID)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.ListAppInstances(organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (h * Handler) GetAppInstance(ctx context.Context, appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_manager_go.AppInstance, error) {
	vErr := entities.ValidAppInstanceID(appInstanceID)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.GetAppInstance(appInstanceID)
}

// ListInstanceParameters retrieves a list of instance parameters
func (h * Handler)  ListInstanceParameters (ctx context.Context, appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.InstanceParameterList, error) {
	vErr := entities.ValidAppInstanceID(appInstanceID)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.ListInstanceParameters(appInstanceID)
}

func (h * Handler)  ListDescriptorAppParameters (ctx context.Context, descriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppParameterList, error) {
	vErr := entities.ValidAppDescriptorID(descriptorID)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.ListDescriptorAppParameters(descriptorID)
}

func (h * Handler) RetrieveTargetApplications(ctx context.Context, filter *grpc_application_manager_go.ApplicationFilter) (*grpc_application_manager_go.TargetApplicationList, error){
	vErr := entities.ValidAppFilter(filter)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.RetrieveTargetApplications(filter)
}

func (h * Handler) RetrieveEndpoints(ctx context.Context, filter *grpc_application_manager_go.RetrieveEndpointsRequest) (*grpc_application_manager_go.ApplicationEndpoints, error){
	vErr := entities.ValidRetrieveEndpointsRequest(filter)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.RetrieveEndpoints(filter)
}

// ListAvailableInstanceInbounds retrieves a list of available inbounds of an organization
func (h * Handler) ListAvailableInstanceInbounds(ctx context.Context, in *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AvailableInstanceInboundList, error){

	return nil, conversions.ToGRPCError(derrors.NewUnimplementedError("not implemented yet"))
}
// ListAvailableInstanceOutbounds retrieves a list of available outbounds of an organization
func (h * Handler) ListAvailableInstanceOutbounds(ctx context.Context, in *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AvailableInstanceOutboundList, error){
	return nil, conversions.ToGRPCError(derrors.NewUnimplementedError("not implemented yet"))

}