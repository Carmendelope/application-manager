/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package application

import (
	"context"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-organization-go"
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
	panic("implement me")
}

// ListAppDescriptors retrieves a list of application descriptors.
func (h * Handler) ListAppDescriptors(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppDescriptorList, error) {
	panic("implement me")
}

// GetAppDescriptor retrieves a given application descriptor.
func (h * Handler) GetAppDescriptor(ctx context.Context, appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppDescriptor, error) {
	panic("implement me")
}

// Deploy an application descriptor.
func (h * Handler) Deploy(ctx context.Context, deployRequest *grpc_application_manager_go.DeployRequest) (*grpc_conductor_go.DeploymentResponse, error) {
	panic("implement me")
}

// Undeploy a running application instance.
func (h * Handler) Undeploy(ctx context.Context, appInstanceID *grpc_application_go.AppInstanceId) (*grpc_common_go.Success, error) {
	panic("implement me")
}

// ListAppInstances retrieves a list of application descriptors.
func (h * Handler) ListAppInstances(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppInstanceList, error) {
	panic("implement me")
}

// GetAppDescriptor retrieves a given application descriptor.
func (h * Handler) GetAppInstance(ctx context.Context, appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.AppInstance, error) {
	panic("implement me")
}



