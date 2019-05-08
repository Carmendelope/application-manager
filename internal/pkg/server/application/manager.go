/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package application

import (
	"context"
	"fmt"
	"github.com/nalej/application-manager/internal/pkg/bus"
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"math/rand"
	"time"
)


const DefaultTimeout =  time.Minute
const RequiredParamNotFilled = "Required parameter not filled"

// Manager structure with the required clients for roles operations.
type Manager struct {
	appClient       grpc_application_go.ApplicationsClient
	conductorClient grpc_conductor_go.ConductorClient
	clusterClient   grpc_infrastructure_go.ClustersClient
	deviceClient    grpc_device_go.DevicesClient
	busManager		*bus.BusManager
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	appClient grpc_application_go.ApplicationsClient,
	conductorClient grpc_conductor_go.ConductorClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	deviceClient grpc_device_go.DevicesClient,
	busManager *bus.BusManager) Manager {
	return Manager{appClient, conductorClient, clusterClient, deviceClient, busManager}
}

// AddAppDescriptor adds a new application descriptor to a given organization.
func (m * Manager) AddAppDescriptor(addDescriptorRequest *grpc_application_go.AddAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {


	// before add appDescriptor, validate parameters
	err := entities.ValidateDescriptorParameters(addDescriptorRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}

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

// checkAllRequiredParametersAreFilled checks all the params defined as required are filled in deploy request
func (m * Manager) checkAllRequiredParametersAreFilled(desc *grpc_application_go.AppDescriptor, params  *grpc_application_go.InstanceParameterList) error {

	// get all the required parameters
	for _, p := range desc.Parameters {
		if p.Required == true {
			find := false
			// look for it in deploy params
			for _, deployParam := range params.Parameters {

				if deployParam.ParameterName == p.Name{
					find = true
					break
				}

			}
			if !find {
				return derrors.NewFailedPreconditionError(RequiredParamNotFilled)
			}
		}
	}

	return nil
}
// Deploy an application descriptor.
func (m * Manager) Deploy(deployRequest *grpc_application_manager_go.DeployRequest) (*grpc_application_manager_go.DeploymentResponse, error) {

	log.Debug().Interface("request", deployRequest).Msg("received deployment request")

	// Retrieve descriptor by descriptorID
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	desc, err := m.appClient.GetAppDescriptor(ctx, &grpc_application_go.AppDescriptorId{
		OrganizationId: deployRequest.OrganizationId,
		AppDescriptorId: deployRequest.AppDescriptorId,
	})
	if err!= nil {
		log.Error().Err(err).Msgf("error getting application descriptor %s", deployRequest.AppDescriptorId)
		return nil,err
	}

	// check if all required params are filled
	err = m.checkAllRequiredParametersAreFilled(desc, deployRequest.Parameters)
	if err != nil {
		return nil, err
	}

	// Create it parametrized descriptor
	parametrizedDesc, err := entities.CreateParametrizedDescriptor(desc, deployRequest.Parameters)
	if err != nil {
		log.Error().Err(err).Msgf("error creating  parametrized descriptor %s.", deployRequest.AppDescriptorId)
		return nil, err
	}

	// Create new application instance
	addReq := &grpc_application_go.AddAppInstanceRequest{
		OrganizationId: deployRequest.OrganizationId,
		AppDescriptorId: deployRequest.AppDescriptorId,
		Name: deployRequest.Name,
		Parameters: deployRequest.Parameters,
	}

	// Add instance, by default this is created with bus status
	ctxInstance, cancelInstance := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelInstance()
	instance, err := m.appClient.AddAppInstance(ctxInstance, addReq)
	if err != nil {
		log.Error().Err(err).Msg("error adding application instance")
		return nil, err
	}

	// fill the instance_id in the parametrized descriptor
	parametrizedDesc.AppInstanceId = instance.AppInstanceId

	appInstanceID := &grpc_application_go.AppInstanceId{
		OrganizationId:     deployRequest.OrganizationId,
		AppInstanceId:      instance.AppInstanceId,
	}

	// Add parametrizedDescriptor in the system
	ctxParametrized, cancelParametrized := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelParametrized()
	newDesc , err := m.appClient.AddParametrizedDescriptor(ctxParametrized, parametrizedDesc)
	if err != nil {
		log.Error().Err(err).Msgf("error adding  parametrized descriptor %s. Delete instance", instance.AppInstanceId)
		_, rollbackErr := m.appClient.RemoveAppInstance(context.Background(), appInstanceID )
		if rollbackErr != nil {
			log.Error().Err(err).Msgf("error in rollback deleting the instance %s", instance.AppInstanceId)
		}
		return nil, err
	}

	// update the instance with the rules parametrized
	if len(parametrizedDesc.Rules) > 0 {
		ctxUpdateInstance, cancelUpdate := context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancelUpdate()
		// update the instance
		instance.Rules = newDesc.Rules
		instance.ConfigurationOptions = newDesc.ConfigurationOptions
		instance.EnvironmentVariables = newDesc.EnvironmentVariables
		instance.Labels = newDesc.Labels
		_, err := m.appClient.UpdateAppInstance(ctxUpdateInstance, instance)

		if err != nil {
			log.Error().Err(err).Msgf("error updating instance %s. Delete instance", instance.AppInstanceId)
			_, rollbackErr := m.appClient.RemoveAppInstance(context.Background(), appInstanceID )
			if rollbackErr != nil {
				log.Error().Err(err).Msgf("error in rollback deleting the instance %s", instance.AppInstanceId)
			}
			return nil, err
		}

	}

	// send deploy command to conductor
	request := &grpc_conductor_go.DeploymentRequest{
		RequestId:            fmt.Sprintf("app-mngr-%d", rand.Int()),
		AppInstanceId:        appInstanceID,
		Name:                 deployRequest.Name,
	}

	/*
	// TODO remove legacy interaction with the conductor API
	ctxConductor, cancelConductor := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelConductor()
	_, err = m.conductorClient.Deploy(ctxConductor, request)
	if err != nil {
		log.Error().Err(err).Msgf("problems deploying application %s", instance.AppInstanceId)
		return nil, err
	}
	*/

	ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	err = m.busManager.Send(ctx, request)
	if err != nil {
		log.Error().Err(err).Str("appInstanceId", instance.AppInstanceId).
			Msg("error when sending deployment request to the queue")
		return nil, err
	}

	toReturn := grpc_application_manager_go.DeploymentResponse{
		RequestId:     fmt.Sprintf("app-mngr-%d", rand.Int()),
		AppInstanceId: instance.AppInstanceId,
		Status:        grpc_application_go.ApplicationStatus_QUEUED}

	log.Debug().Interface("deploymentResponse", toReturn).Msg("Response")

	return &toReturn, nil

}

// Undeploy a running application instance.
func (m * Manager) Undeploy(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_common_go.Success, error) {
	undeployRequest := &grpc_conductor_go.UndeployRequest{
		OrganizationId:       appInstanceID.OrganizationId,
		AppInstanceId:            appInstanceID.AppInstanceId,
	}

	// TODO: remove legacy interaction with the conductor API
	//return  m.conductorClient.Undeploy(context.Background(), undeployRequest)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	err := m.busManager.Send(ctx, undeployRequest)
	if err != nil {
		log.Error().Err(err).Str("appInstanceId", undeployRequest.AppInstanceId).
			Msg("error when sending the undeploy request to the queue")
		return nil, err
	}

	return &grpc_common_go.Success{}, nil

}

// ListAppInstances retrieves a list of application descriptors.
func (m * Manager) ListAppInstances(organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppInstanceList, error) {
	return m.appClient.ListAppInstances(context.Background(), organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (m * Manager) GetAppInstance(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.AppInstance, error) {
	return m.appClient.GetAppInstance(context.Background(), appInstanceID)
}

func (m * Manager)  ListInstanceParameters (appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.InstanceParameterList, error) {
	return m.appClient.GetInstanceParameters(context.Background(), appInstanceID)
}

func (m * Manager)  ListDescriptorAppParameters (descriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppParameterList, error) {
	return m.appClient.GetDescriptorAppParameters(context.Background(), descriptorID)
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