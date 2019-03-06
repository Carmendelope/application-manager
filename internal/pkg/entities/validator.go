/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package entities

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-organization-go"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

const emptyRequestId = "request_id cannot be empty"
const emptyOrganizationId = "organization_id cannot be empty"
const emptyDescriptorId = "app_descriptor_id cannot be empty"
const emptyInstanceId = "app_instance_id cannot be empty"
const emptyName = "name cannot be empty"
const emptyDeviceGroupId = "device_group_id cannot be empty"
const emptyAppDescriptorId = "app_descriptor_id cannot be empty"


func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidAddAppDescriptorRequest(toAdd * grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	if toAdd.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if toAdd.RequestId == "" {
		return derrors.NewInvalidArgumentError(emptyRequestId)
	}

	if toAdd.Name == ""{
		return derrors.NewInvalidArgumentError(emptyName)
	}

	// At least one service group
	if len(toAdd.Groups) == 0 {
		return derrors.NewInvalidArgumentError("expecting at least one service group")
	}

	// Every service group must have at least one service
	for _, g := range toAdd.Groups {
		if len(g.Services) == 0 {
			return derrors.NewInvalidArgumentError("service group must have at least one service").WithParams(g.Name)
		}
	}
	err :=  ValidateAppRequestForNames(toAdd)

	if err != nil {
		return err
	}

	// logic validation
	err = ValidDescriptorLogic(toAdd)
	return err
}

// ValidateDescriptor checks validity of object names, ports name, port numbers  meeting Kubernetes specs.
func  ValidateAppRequestForNames(toAdd * grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	var errs []string
	// for each group
	for _, group := range toAdd.Groups {
		for _,service := range group.Services {
			// Validate service name
			kerr := validation.IsDNS1123Label(service.Name)
			if len(kerr) > 0 {
				errs = append(errs, "serviceName", service.Name)
				errs = append(errs, kerr...)
			}
			// validate Exposed Port Name and Number
			for _,port := range service.ExposedPorts {
				kerr = validation.IsValidPortName(port.Name)
				if len(kerr) > 0 {
					errs = append(errs,"PortName", port.Name)
					errs = append(errs, kerr...)
				}
				kerr = validation.IsValidPortNum(int(port.ExposedPort))
				if len(kerr) > 0 {
					errs = append(errs, "ExposedPort")
					errs = append(errs, kerr...)
				}
				kerr = validation.IsValidPortNum(int(port.InternalPort))
				if len(kerr) > 0 {
					errs = append(errs, "InternalPort")
					errs = append(errs, kerr...)
				}
			}
		}
	}
	if len(errs) > 0 {
		err := derrors.NewFailedPreconditionError(fmt.Sprintf("%s: %v","App descriptor validation failed",errs))
		return err
	}
	return nil
}

func ValidAppDescriptorID(descriptorID * grpc_application_go.AppDescriptorId) derrors.Error {
	if descriptorID.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if descriptorID.AppDescriptorId == "" {
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}
	return nil
}

func ValidUpdateAppDescriptorRequest(request * grpc_application_go.UpdateAppDescriptorRequest) derrors.Error{
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.AppDescriptorId == ""{
		return derrors.NewInvalidArgumentError(emptyAppDescriptorId)
	}
	return nil
}

func ValidAppInstanceID(instanceID * grpc_application_go.AppInstanceId) derrors.Error {
	if instanceID.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if instanceID.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	return nil
}

func ValidDeployRequest(deployRequest *grpc_application_manager_go.DeployRequest) derrors.Error {
	if deployRequest.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if deployRequest.AppDescriptorId == ""{
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}

	if deployRequest.Name == ""{
		return derrors.NewInvalidArgumentError(emptyName)
	}

	return nil
}

func ValidAppFilter (filter *grpc_application_manager_go.ApplicationFilter) derrors.Error{
	if filter.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if filter.DeviceGroupId == ""{
		return derrors.NewInvalidArgumentError(emptyDeviceGroupId)
	}
	return nil
}

func ValidRetrieveEndpointsRequest (request *grpc_application_manager_go.RetrieveEndpointsRequest) derrors.Error{
	if request.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if request.AppInstanceId == ""{
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	return nil
}

// ValidAppDescriptor checks validity related to the descriptor logic
func ValidDescriptorLogic(appDescriptor *grpc_application_go.AddAppDescriptorRequest) derrors.Error {

	/*
	   	- Rules refer to existing services
		- Rules specifying service to service restrictions are not supported yet ???
		- Service and service group name are uniq
		- Environment variables must be checked with existing service names
		- Deploy after should point to existing services
		- Multireplicate set cannot be set with number of replicas
	 */

	 // TODO: the service should be unique per group, not unique in the descriptor
	appServices := make(map[string]bool)
	appGroups := make(map[string]bool)

	// - Service and service group name are uniq (and they cannot be empty)
	for _, appGroup := range appDescriptor.Groups {
		if appGroup.Name == "" {
			return derrors.NewFailedPreconditionError("Service Group Name cannot be empty")
		}
		_, exists := appGroups[appGroup.Name]
		if exists {
			return derrors.NewFailedPreconditionError("Service Group Name defined twice").WithParams(appGroup.Name)
		}

		for _, service := range appGroup.Services {
			if service.Name == "" {
				return derrors.NewFailedPreconditionError("Service Name cannot be empty")
			}
			_ , exists := appServices[service.Name]
			if exists {
				return derrors.NewFailedPreconditionError("Service Name defined twice").WithParams(appGroup.Name, service.Name)
			}
			appServices[service.Name] = true

		}
		appGroups[appGroup.Name] = true
	}

	// - Rules refer to existing services
	for _, rule := range appDescriptor.Rules{

		if rule.Access == grpc_application_go.PortAccess_APP_SERVICES{
			return derrors.NewFailedPreconditionError("Service to service restrictions are not supported yet")
		}

		_, exists := appGroups[rule.TargetServiceGroupName]
		if ! exists {
			return derrors.NewFailedPreconditionError("Service Group Name in rule not defined").WithParams(rule.Name, rule.TargetServiceGroupName)
		}
		_, exists = appServices[rule.TargetServiceName]
		if ! exists {
			return derrors.NewFailedPreconditionError("Service Name in rule not defined").WithParams(rule.Name, rule.TargetServiceGroupName, rule.TargetServiceName)
		}

		_, exists = appGroups[rule.AuthServiceGroupName]
		if ! exists {
			return derrors.NewFailedPreconditionError("Service Group Name in rule not defined").WithParams(rule.Name, rule.TargetServiceGroupName)
		}
		for _, serviceName := range rule.AuthServices {
			_, exists = appServices[serviceName]
			if ! exists {
				return derrors.NewFailedPreconditionError("Service Name in rule not defined").WithParams(rule.Name, rule.AuthServiceGroupName, serviceName)
			}
		}

	}

	// - Deploy after should point to existing services
	for _, group := range appDescriptor.Groups{
		for _, service := range group.Services {
			for _, after := range service.DeployAfter {
				_, exists := appServices[after]
				if ! exists {
					return derrors.NewFailedPreconditionError("Service indicated in deploy after field does not exist").WithParams(after)
				}
			}
		}
		// - Multireplicate set cannot be set with number of replicas
		if group.Specs != nil {
			if group.Specs.MultiClusterReplica && group.Specs.NumReplicas > 0 {
				return derrors.NewFailedPreconditionError("Multireplicate set cannot be set with number of replicas").WithParams(group.Name)
			}
		}
	}

	// - Environment variables must be checked with existing service names
	// TODO: the enviroment variables only looks the service name (servicegroup should be included)
	for _, value := range appDescriptor.EnvironmentVariables {
		serviceValue := strings.Trim(value, " ")
		if strings.Index(serviceValue, "NALEJ_SERV_") == 0 {
			// check the service exists.
			pos := strings.Index(serviceValue, ":")
			if pos == -1 {
				pos = len(serviceValue)
			}
			nalejService :=value[11:pos]

			// find the service
			_, exists := appServices[nalejService]
			if !exists {
				return derrors.NewFailedPreconditionError("Environment variable error, service does not exist").WithParams(value)
			}
		}
	}

	return nil
}