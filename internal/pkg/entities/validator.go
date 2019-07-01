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
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

const emptyRequestId = "request_id cannot be empty"
const emptyOrganizationId = "organization_id cannot be empty"
const emptyDescriptorId = "app_descriptor_id cannot be empty"
const emptyInstanceId = "app_instance_id cannot be empty"
const emptyName = "name cannot be empty"
const emptyDeviceGroupId = "device_group_id cannot be empty"
const emptyDeviceGroupName = "device_group_name cannot be empty"
const emptyAppDescriptorId = "app_descriptor_id cannot be empty"

const NalejEnvironmentVariablePrefix = "NALEJ_SERV_"


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

	err = ValidateStoragePathAppRequest(toAdd)
	if err != nil {
		return err
	}

	// logic validation
	err = ValidDescriptorLogic(toAdd)
	return err
}

// getPath returns a path equivalent to 'path' but without indirections
func GetPath(path string) string {
	res := ""

	directories := strings.Split(path, "/")

	if directories[0] == ".." {
		// unable to get path, we dont know what is the current path
		return path
	}

	dirRes := make([]string, len(directories))

	index := 1
	dirRes[0] = directories[0]

	if len(directories) > 1 {
		for i := 1; i < len(directories); i++ {
			if directories[i] == "."{
				// nothing to do
			}else if directories[i] == ".."{
				index --
				if index < 0 {
					log.Warn().Str("path", path).Msg("unable to validate path")
					return path
				}
			} else {
				dirRes[index] = directories[i]
				index ++
			}
		}
	}

	res = dirRes[0]
	for i:=1; i< index; i++ {
		res = fmt.Sprintf("%s/%s", res, dirRes[i])
	}

	return res
}

// ValidateStoragePathAppRequest validate if the same storage path is added more than once
func ValidateStoragePathAppRequest(toAdd * grpc_application_go.AddAppDescriptorRequest) derrors.Error {

	for _, group := range toAdd.Groups {
		for _, service := range group.Services {
			// map to store the storage paths
			pathMap := make (map[string] bool, 0)
			for _, sto := range service.Storage {

				path := GetPath(sto.MountPath)
				// check if the mountPath is used before
				_, exists := pathMap[path]
				if exists{
					return derrors.NewInvalidArgumentError("mounthPath defined twice").WithParams(sto.MountPath)
				}
				pathMap[path] = true

			}
		}
	}


	return nil
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
	if filter.DeviceGroupId == "" {
		return derrors.NewInvalidArgumentError(emptyDeviceGroupId)
	}

	if filter.DeviceGroupName == ""{
		return derrors.NewInvalidArgumentError(emptyDeviceGroupName)
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

func ValidAppDescriptorEnvironmentVariables (appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]bool) derrors.Error {

	// - Environment variables must be checked with existing service names
	// TODO: the enviroment variables only looks the service name (servicegroup should be included)
	for _, value := range appDescriptor.EnvironmentVariables {
		serviceValue := strings.Trim(value, " ")
		if strings.Index(serviceValue, NalejEnvironmentVariablePrefix) == 0 {
			// check the service exists.
			pos := strings.Index(serviceValue, ":")
			if pos == -1 {
				pos = len(serviceValue)
			}
			nalejService :=value[len(NalejEnvironmentVariablePrefix):pos]

			// find the service
			_, exists := appServices[strings.ToLower(nalejService)]
			if !exists {
				return derrors.NewFailedPreconditionError("Environment variable error, service does not exist").WithParams(value)
			}
		}
	}
	return nil
}

func ValidAppDescriptorGroupSpecs (appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]bool) derrors.Error{

	// TODO: the DeployAfter only looks the service name (servicegroup should be included)
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
			if group.Specs.MultiClusterReplica && group.Specs.Replicas > 0 {
				return derrors.NewFailedPreconditionError("Multireplicate set cannot be set with number of replicas").WithParams(group.Name)
			}
		}
	}
	return nil
}

func ValidAppDescriptorRules(appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]bool, appGroups map[string]bool) derrors.Error{

	// - Rules refer to existing services
	for _, rule := range appDescriptor.Rules{

		// RuleId is filed by system model
		if rule.RuleId != "" {
			return derrors.NewFailedPreconditionError("Rule Id cannot be filled").WithParams(rule.Name)
		}

		_, exists := appGroups[rule.TargetServiceGroupName]
		if ! exists {
			return derrors.NewFailedPreconditionError("Target Service Group Name in rule not found in groups definition").WithParams(rule.Name, rule.TargetServiceGroupName)
		}
		_, exists = appServices[rule.TargetServiceName]
		if ! exists {
			return derrors.NewFailedPreconditionError("Target Service Name in rule not found in services definition").WithParams(rule.Name, rule.TargetServiceGroupName, rule.TargetServiceName)
		}

		// only rules referring to PortAccess_APP_SERVICES AuthServiceGroupName and AuthServiceName should be specified or expected.
		if rule.Access == grpc_application_go.PortAccess_APP_SERVICES {

			_, exists = appGroups[rule.AuthServiceGroupName]
			if ! exists {
				return derrors.NewFailedPreconditionError("Auth Service Group Name in rule not found in groups definition").WithParams(rule.Name, rule.AuthServiceGroupName)
			}
			for _, serviceName := range rule.AuthServices {
				_, exists = appServices[serviceName]
				if ! exists {
					return derrors.NewFailedPreconditionError("Auth Service Name in rule not found in services definition").WithParams(rule.Name, rule.AuthServiceGroupName, serviceName)
				}
			}
		}else{
			if rule.AuthServiceGroupName != ""{
				return derrors.NewFailedPreconditionError("Auth Service Group Name should no be specified for selected access rule").WithParams(rule.Name)
			}
			if len(rule.AuthServices) > 0  {
				return derrors.NewFailedPreconditionError("Auth Services should no be specified for selected access rule").WithParams(rule.Name)
			}
			if rule.Access == grpc_application_go.PortAccess_DEVICE_GROUP{
				// at least should be defined one device
				if rule.DeviceGroupNames == nil || len(rule.DeviceGroupNames) <= 0 {
					return derrors.NewFailedPreconditionError("Device group names in rule not defined").WithParams(rule.Name)
				}
			}
		}

		// RuleIds is filled by system-model
		if rule.DeviceGroupIds != nil && len(rule.DeviceGroupIds) > 0 {
			return derrors.NewFailedPreconditionError("Device Group Ids cannot be filled").WithParams(rule.Name)
		}
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

	// at least one group should be defined
	if appDescriptor.Groups == nil || len (appDescriptor.Groups) <=0 {
		return derrors.NewFailedPreconditionError("at least one group should be defined")
	}

	// - Service and service group name are uniq (and they cannot be empty)
	for _, appGroup := range appDescriptor.Groups {

		if appGroup.Name == "" {
			return derrors.NewFailedPreconditionError("Service Group Name cannot be empty")
		}

		// ServiceGroupId is filled by system-model
		if appGroup.ServiceGroupId != ""{
			return derrors.NewFailedPreconditionError("Service Group Id cannot be filled").WithParams(appGroup.Name)
		}

		_, exists := appGroups[appGroup.Name]
		if exists {
			return derrors.NewFailedPreconditionError("Service Group Name defined twice").WithParams(appGroup.Name)
		}

		for _, service := range appGroup.Services {
			if service.Name == "" {
				return derrors.NewFailedPreconditionError("Service Name cannot be empty")
			}
			// ServiceId is filled by system-model
			if service.ServiceId != "" {
				return derrors.NewFailedPreconditionError("Service Id cannot be filled").WithParams(service.Name)
			}
			if service.ServiceGroupId != "" {
				return derrors.NewFailedPreconditionError("Service Group Id cannot be filled").WithParams(service.Name)
			}

			_ , exists := appServices[service.Name]
			if exists {
				return derrors.NewFailedPreconditionError("Service Name defined twice").WithParams(appGroup.Name, service.Name)
			}
			appServices[service.Name] = true

			// ConfigFiledId is filled by system-model
			if service.Configs != nil && len(service.Configs) > 0 {
				for _, config := range service.Configs {
					if config.ConfigFileId != "" {
						return derrors.NewFailedPreconditionError("Config File Id cannot be filled").WithParams(config.Name)
					}
				}
			}

		}
		appGroups[appGroup.Name] = true
	}

	// - Rules refer to existing services
	err := ValidAppDescriptorRules(appDescriptor, appServices, appGroups)
	if err != nil {
		return err
	}

	// ValidGroupSpecs:
	// - Deploy after should point to existing services
	// - Multireplicate set cannot be set with number of replicas
	err = ValidAppDescriptorGroupSpecs(appDescriptor, appServices)
	if err != nil {
		return err
	}

	// - Environment variables must be checked with existing service names
	err = ValidAppDescriptorEnvironmentVariables(appDescriptor, appServices)
	if err != nil {
		return err
	}

	return nil
}