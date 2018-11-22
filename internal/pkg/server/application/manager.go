package application

import (
	"context"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-organization-go"
	"math/rand"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	appClient       grpc_application_go.ApplicationsClient
	conductorClient grpc_conductor_go.ConductorClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	appClient grpc_application_go.ApplicationsClient,
	conductorClient grpc_conductor_go.ConductorClient,
) Manager {
	return Manager{appClient, conductorClient}
}

// AddAppDescriptor adds a new application descriptor to a given organization.
func (m * Manager) AddAppDescriptor(addDescriptorRequest *grpc_application_go.AddAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {
	return m.appClient.AddAppDescriptor(context.Background(), addDescriptorRequest)
}

// ListAppDescriptors retrieves a list of application descriptors.
func (m * Manager) ListAppDescriptors(organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppDescriptorList, error) {
	return m.appClient.ListAppDescriptors(context.Background(), organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (m * Manager) GetAppDescriptor(appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppDescriptor, error) {
	return m.appClient.GetAppDescriptor(context.Background(), appDescriptorID)
}

// RemoveAppDescriptor removes an application descriptor from the system.
func (m * Manager)  RemoveAppDescriptor(appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_common_go.Success, error) {
	// Check if there are instances running with that descriptor
	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId: appDescriptorID.OrganizationId,
	}
	instances, err := m.appClient.ListAppInstances(context.Background(), orgID)
	if err != nil{
		return nil, err
	}
	for _, inst := range instances.Instances {
		if inst.AppDescriptorId == appDescriptorID.AppDescriptorId {
			return nil, derrors.NewFailedPreconditionError("application instances must be removed before deleting the descriptor")
		}
	}
	return m.appClient.RemoveAppDescriptor(context.Background(), appDescriptorID)
}

// Deploy an application descriptor.
func (m * Manager) Deploy(deployRequest *grpc_application_manager_go.DeployRequest) (*grpc_conductor_go.DeploymentResponse, error) {
	//TODO NP-249 Create the instance and pass it to conductor

	appDescriptorID := &grpc_application_go.AppDescriptorId{
		OrganizationId:       deployRequest.OrganizationId,
		AppDescriptorId:      deployRequest.AppDescriptorId,
	}

	request := &grpc_conductor_go.DeploymentRequest{
		RequestId:            fmt.Sprintf("app-mngr-%d", rand.Int()),
		AppId:                appDescriptorID,
		Name:                 deployRequest.Name,
		Description:          deployRequest.Description,
	}

	return m.conductorClient.Deploy(context.Background(), request)
}

// Undeploy a running application instance.
func (m * Manager) Undeploy(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_common_go.Success, error) {
	// TODO NP-223 Add organization Id to the message.
	undeployRequest := &grpc_conductor_go.UndeployRequest{
		InstaceId:            appInstanceID.AppInstanceId,
	}
	return  m.conductorClient.Undeploy(context.Background(), undeployRequest)
}

// ListAppInstances retrieves a list of application descriptors.
func (m * Manager) ListAppInstances(organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppInstanceList, error) {
	return m.appClient.ListAppInstances(context.Background(), organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (m * Manager) GetAppInstance(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.AppInstance, error) {
	return m.appClient.GetAppInstance(context.Background(), appInstanceID)
}
