/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application

import (
	"context"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"math/rand"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	appClient       grpc_application_go.ApplicationsClient
	conductorClient grpc_conductor_go.ConductorClient
	clusterClient   grpc_infrastructure_go.ClustersClient
	deviceClient    grpc_device_go.DevicesClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	appClient grpc_application_go.ApplicationsClient,
	conductorClient grpc_conductor_go.ConductorClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	deviceClient grpc_device_go.DevicesClient) Manager {
	return Manager{appClient, conductorClient, clusterClient, deviceClient}
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

// UpdateAppDescriptor allows the user to update the information of a registered descriptor.
func (m * Manager) UpdateAppDescriptor(request *grpc_application_go.UpdateAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {
	return m.appClient.UpdateAppDescriptor(context.Background(), request)
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
	undeployRequest := &grpc_conductor_go.UndeployRequest{
		OrganizationId:       appInstanceID.OrganizationId,
		AppInstanceId:            appInstanceID.AppInstanceId,
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

func (m*Manager) RetrieveTargetApplications(filter *grpc_application_manager_go.ApplicationFilter) (*grpc_application_manager_go.TargetApplicationList, error){

	// check if the device_group_id and device_group_name are correct
	group, err := m.deviceClient.GetDeviceGroup(context.Background(), &grpc_device_go.DeviceGroupId{
		OrganizationId: filter.OrganizationId,
		DeviceGroupId: filter.DeviceGroupId,
	})
	if err != nil {
		return nil, err
	}
	if group.Name != filter.DeviceGroupName {
		return nil, conversions.ToGRPCError(derrors.NewPermissionDeniedError("cannot access device_group_name"))
	}

	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId:       filter.OrganizationId,
	}
	// TODO allow filtering on the list request
	allApps, err := m.appClient.ListAppInstances(context.Background(), orgID)
	if err != nil{
		return nil, err
	}
	filtered := ApplyFilter(allApps, filter)

	result, fErr := ToApplicationLabelsList(filtered)
	if fErr != nil{
		return nil, conversions.ToGRPCError(fErr)
	}
	return result, nil
}

func (m*Manager) fillEndpoints(endpoints []*grpc_application_go.EndpointInstance) {
	for i:=0; i<len(endpoints); i++{
		endpoints[i].Fqdn = fmt.Sprintf("%s:%d", endpoints[i].Fqdn, endpoints[i].Port)
	}
}

func (m*Manager) RetrieveEndpoints(request *grpc_application_manager_go.RetrieveEndpointsRequest) (*grpc_application_manager_go.ApplicationEndpoints, error){

	instanceID := &grpc_application_go.AppInstanceId{
		OrganizationId: request.OrganizationId,
		AppInstanceId:  request.AppInstanceId,
	}
	// get the instance requested
	instance, err := m.appClient.GetAppInstance(context.Background(), instanceID)
	if err != nil{
		return nil, err
	}

	appClusterEndPoints := make ([]*grpc_application_manager_go.ApplicationClusterEndpoints, 0)

	//foreach serviceInstance in appInstance -> get endPoints and DeployedClusterId
	for _, group := range instance.Groups {
		for _, service := range group.ServiceInstances {

			// get the clusterHost (if the service is RUNNING)
			if service.Status == grpc_application_go.ServiceStatus_SERVICE_RUNNING  &&
				len(service.Endpoints) > 0  { // the service has endpoints

				clusterId := &grpc_infrastructure_go.ClusterId{
					OrganizationId: request.OrganizationId,
					ClusterId:      service.DeployedOnClusterId,
				}
				cluster, err := m.clusterClient.GetCluster(context.Background(), clusterId)
				if err != nil {
					return nil, err
				}

				m.fillEndpoints(service.Endpoints)

				clusterEndPoint := &grpc_application_manager_go.ApplicationClusterEndpoints{
					DeviceControllerUrl: fmt.Sprintf("device-controller.%s", cluster.Hostname),
					Endpoints:           service.Endpoints,
				}
				appClusterEndPoints = append(appClusterEndPoints, clusterEndPoint)
			}
		}
	}

	return  &grpc_application_manager_go.ApplicationEndpoints{
		ClusterEndpoints: appClusterEndPoints,
	} , nil

}