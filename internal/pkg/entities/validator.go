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
 */

package entities

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/validation"
	"regexp"
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
const emptySourceInstanceId = "source_instance_id cannot be empty"
const emptyTargetInstanceId = "target_instance_id cannot be empty"
const emptyInboundName = "inbound_name cannot be empty"
const emptyOutboundName = "outbound_name cannot be empty"
const emptyServiceGroupId = "service_group_id cannot be empty"
const emptyServiceGroupInstanceId = "service_group_instance_id cannot be empty"
const emptyServiceId = "service_id cannot be empty"
const impossibleDuration = "to cannot be greater than from"

const NalejEnvironmentVariablePrefix = "NALEJ_SERV_"
const EnvironmentVariableRegex = "[._a-zA-Z][._a-zA-Z0-9]*"
// DeployNameRegex with the regular expresion for application names.
const DeployNameRegex = "^[a-zA-Z0-9]+$"
// MinDeployRequestNameLength with the minimal required length for an application name.
const MinDeployRequestNameLength = 3

// Map containing port numbers used by Nalej that cannot be used by any application.
var NalejUsedPorts = map[int32]bool{
	// Port used by Zt-sidecars redirection
	1576: true,
	// Port used by Zt-sidecars zt daemon
	9993: true,
}

func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

// TODO: include the endpoint options validation
func ValidAddAppDescriptorRequest(toAdd *grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	if toAdd.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if toAdd.RequestId == "" {
		return derrors.NewInvalidArgumentError(emptyRequestId)
	}

	if toAdd.Name == "" {
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
	err := ValidateAppRequestForNames(toAdd)

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
			if directories[i] == "." {
				// nothing to do
			} else if directories[i] == ".." {
				index--
				if index < 0 {
					log.Warn().Str("path", path).Msg("unable to validate path")
					return path
				}
			} else {
				dirRes[index] = directories[i]
				index++
			}
		}
	}

	res = dirRes[0]
	for i := 1; i < index; i++ {
		res = fmt.Sprintf("%s/%s", res, dirRes[i])
	}

	return res
}

// ValidateStoragePathAppRequest validate if the same storage path is added more than once
func ValidateStoragePathAppRequest(toAdd *grpc_application_go.AddAppDescriptorRequest) derrors.Error {

	for _, group := range toAdd.Groups {
		for _, service := range group.Services {
			// map to store the storage paths
			pathMap := make(map[string]bool, 0)
			for _, sto := range service.Storage {

				path := GetPath(sto.MountPath)
				// check if the mountPath is used before
				_, exists := pathMap[path]
				if exists {
					return derrors.NewInvalidArgumentError("mounthPath defined twice").WithParams(sto.MountPath)
				}
				pathMap[path] = true

			}
		}
	}

	return nil
}

// ValidateDescriptor checks validity of object names, ports name, port numbers  meeting Kubernetes specs.
func ValidateAppRequestForNames(toAdd *grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	var errs []string
	// for each group
	for _, group := range toAdd.Groups {
		for _, service := range group.Services {

			// Validate service name
			kerr := validation.IsDNS1123Label(service.Name)
			if len(kerr) > 0 {
				errs = append(errs, "serviceName", service.Name)
				errs = append(errs, kerr...)
			}
			// validate Exposed Port Name and Number
			for _, port := range service.ExposedPorts {
				kerr = validation.IsValidPortName(port.Name)
				if len(kerr) > 0 {
					errs = append(errs, "PortName", port.Name)
					errs = append(errs, kerr...)
				}
				kerr = validation.IsValidPortNum(int(port.ExposedPort))
				if len(kerr) > 0 {
					errs = append(errs, "ExposedPort")
					errs = append(errs, kerr...)
				}

				found, _ := NalejUsedPorts[port.ExposedPort]
				if found {
					errs = append(errs, "ExposedPort")
					errs = append(errs, "this is a reserved port")
				}

				kerr = validation.IsValidPortNum(int(port.InternalPort))
				if len(kerr) > 0 {
					errs = append(errs, "InternalPort")
					errs = append(errs, kerr...)
				}

				found, _ = NalejUsedPorts[port.InternalPort]
				if found {
					log.Error().Msg("internal port error")
					errs = append(errs, "InternalPort")
					errs = append(errs, "this is a reserved port")
				}

			}
		}
	}
	if len(errs) > 0 {
		err := derrors.NewFailedPreconditionError(fmt.Sprintf("%s: %v", "App descriptor validation failed", errs))
		return err
	}
	return nil
}

func ValidAppDescriptorID(descriptorID *grpc_application_go.AppDescriptorId) derrors.Error {
	if descriptorID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if descriptorID.AppDescriptorId == "" {
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}
	return nil
}

func ValidUpdateAppDescriptorRequest(request *grpc_application_go.UpdateAppDescriptorRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.AppDescriptorId == "" {
		return derrors.NewInvalidArgumentError(emptyAppDescriptorId)
	}
	return nil
}

func ValidAppInstanceID(instanceID *grpc_application_go.AppInstanceId) derrors.Error {
	if instanceID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if instanceID.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	return nil
}

func ValidDeployRequest(deployRequest *grpc_application_manager_go.DeployRequest) derrors.Error {
	if deployRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if deployRequest.AppDescriptorId == "" {
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}

	nameErr := ValidDeployRequestName(deployRequest)
	if nameErr != nil{
		return nameErr
	}

	return nil
}

// ValidDeployRequestName checks the conditions for the name of the deployed application.
func ValidDeployRequestName(deployRequest * grpc_application_manager_go.DeployRequest) derrors.Error{
	if len(deployRequest.Name) < MinDeployRequestNameLength {
		return derrors.NewInvalidArgumentError(fmt.Sprintf("name cannot be empty and has a minimum length of %d", MinDeployRequestNameLength))
	}

	re := regexp.MustCompile(DeployNameRegex)
	if re.FindString(deployRequest.Name) == "" {
		return derrors.NewFailedPreconditionError(fmt.Sprintf("Invalid application name; it must adhere to %s", DeployNameRegex))
	}

	return nil
}

func ValidUndeployRequest(request *grpc_application_manager_go.UndeployRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if request.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}

	return nil
}

func ValidAppFilter(filter *grpc_application_manager_go.ApplicationFilter) derrors.Error {
	if filter.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if filter.DeviceGroupId == "" {
		return derrors.NewInvalidArgumentError(emptyDeviceGroupId)
	}

	if filter.DeviceGroupName == "" {
		return derrors.NewInvalidArgumentError(emptyDeviceGroupName)
	}
	return nil
}

func ValidRetrieveEndpointsRequest(request *grpc_application_manager_go.RetrieveEndpointsRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if request.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	return nil
}

func ValidAppDescriptorEnvironmentVariables(appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]*grpc_application_go.Service) derrors.Error {

	// - Environment variables must be checked with existing service names
	// TODO: the enviroment variables only looks the service name (servicegroup should be included)
	re := regexp.MustCompile(EnvironmentVariableRegex)
	for key, value := range appDescriptor.EnvironmentVariables {

		// a valid environment variable name must consist of alphabetic characters, digits, '_', '', or '.', and must not start with a digit
		// regex used for validation is '[._a-zA-Z][._a-zA-Z0-9]*'

		match := re.FindString(key)
		if match != key {
			return derrors.NewFailedPreconditionError("A valid environment variable name must consist of alphabetic characters, digits, '_', '', or '.', and must not start with a digit").WithParams(key)
		}

		serviceValue := strings.Trim(value, " ")
		if strings.Index(serviceValue, NalejEnvironmentVariablePrefix) == 0 {
			// check the service exists.
			pos := strings.Index(serviceValue, ":")
			if pos == -1 {
				pos = len(serviceValue)
			}
			nalejService := value[len(NalejEnvironmentVariablePrefix):pos]

			// find the service
			_, exists := appServices[strings.ToLower(nalejService)]
			if !exists {
				return derrors.NewFailedPreconditionError("Environment variable error, service does not exist").WithParams(value)
			}
		}
	}
	return nil
}

func ValidAppDescriptorGroupSpecs(appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]*grpc_application_go.Service) derrors.Error {

	// TODO: the DeployAfter only looks the service name (servicegroup should be included)
	// - Deploy after should point to existing services
	for _, group := range appDescriptor.Groups {
		for _, service := range group.Services {
			for _, after := range service.DeployAfter {
				_, exists := appServices[after]
				if !exists {
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

func ValidAppDescriptorRules(appDescriptor *grpc_application_go.AddAppDescriptorRequest, appServices map[string]*grpc_application_go.Service, appGroups map[string]*grpc_application_go.ServiceGroup) derrors.Error {

	// NP-1962 Descriptor Validation for inbound and outbound connections
	// check the interface names are unique for each descriptor (in both inbound and outbound)
	// interfaceNames is a map with all the names of the inbound and outbound names
	// the value will be true if it is an inbound and false in other case
	interfaceNames := make(map[string]bool, 0)
	for _, inbound := range appDescriptor.InboundNetInterfaces {
		_, exists := interfaceNames[inbound.Name]
		if exists {
			return derrors.NewFailedPreconditionError("Inbound/outbound name defined twice").WithParams(inbound.Name)
		}
		interfaceNames[inbound.Name] = true
	}
	for _, outbound := range appDescriptor.OutboundNetInterfaces {
		_, exists := interfaceNames[outbound.Name]
		if exists {
			return derrors.NewFailedPreconditionError("Inbound/outbound name defined twice").WithParams(outbound.Name)
		}
		interfaceNames[outbound.Name] = false
	}

	// - Rules refer to existing services
	for _, rule := range appDescriptor.Rules {

		// RuleId is filed by system model
		if rule.RuleId != "" {
			return derrors.NewFailedPreconditionError("Rule Id cannot be filled").WithParams(rule.Name)
		}

		ruleGroup, exists := appGroups[rule.TargetServiceGroupName]
		if !exists {
			return derrors.NewFailedPreconditionError("Target Service Group Name in rule not found in groups definition").WithParams(rule.Name, rule.TargetServiceGroupName)
		}
		_, exists = appServices[rule.TargetServiceName]
		if !exists {
			return derrors.NewFailedPreconditionError("Target Service Name in rule not found in services definition").WithParams(rule.Name, rule.TargetServiceGroupName, rule.TargetServiceName)
		}

		// only rules referring to PortAccess_APP_SERVICES AuthServiceGroupName and AuthServiceName should be specified or expected.
		if rule.Access == grpc_application_go.PortAccess_APP_SERVICES {

			_, exists = appGroups[rule.AuthServiceGroupName]
			if !exists {
				return derrors.NewFailedPreconditionError("Auth Service Group Name in rule not found in groups definition").WithParams(rule.Name, rule.AuthServiceGroupName)
			}
			for _, serviceName := range rule.AuthServices {
				_, exists = appServices[serviceName]
				if !exists {
					return derrors.NewFailedPreconditionError("Auth Service Name in rule not found in services definition").WithParams(rule.Name, rule.AuthServiceGroupName, serviceName)
				}
			}
		} else {
			if rule.AuthServiceGroupName != "" {
				return derrors.NewFailedPreconditionError("Auth Service Group Name should no be specified for selected access rule").WithParams(rule.Name)
			}
			if len(rule.AuthServices) > 0 {
				return derrors.NewFailedPreconditionError("Auth Services should no be specified for selected access rule").WithParams(rule.Name)
			}
			if rule.Access == grpc_application_go.PortAccess_DEVICE_GROUP {
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

		// NP-1962 Descriptor Validation for inbound and outbound connections
		// if the rule is related to an inbound interface -> it should be defined in inboundInterfaces
		if rule.InboundNetInterface != "" {
			if rule.Access != grpc_application_go.PortAccess_INBOUND_APPNET {
				return derrors.NewFailedPreconditionError("inbound_net_interface defined but the access is not INBOUND").WithParams(rule.InboundNetInterface, rule.Access)
			}
			inbound, exists := interfaceNames[rule.InboundNetInterface]
			if !exists {
				return derrors.NewFailedPreconditionError("inbound_net_interface found in security rule is not defined").WithParams(rule.InboundNetInterface)
			}
			if inbound == false {
				// the interface named rule.InboundInterface is an outbound
				return derrors.NewFailedPreconditionError("inbound_net_interface found in security rule is defined as Outbound").WithParams(rule.InboundNetInterface)
			}
			if ruleGroup.Specs != nil && (ruleGroup.Specs.Replicas > 1 || ruleGroup.Specs.MultiClusterReplica) {
				return derrors.NewFailedPreconditionError("Inbound rule linked to a multireplica service group").WithParams(rule.InboundNetInterface)
			}
		}
		// if the rule is related to an outbound interface -> it should be defined in outboundInterfaces
		if rule.OutboundNetInterface != "" {
			if rule.Access != grpc_application_go.PortAccess_OUTBOUND_APPNET {
				return derrors.NewFailedPreconditionError("outbound_net_interface defined but the access is not OUTBOUND").WithParams(rule.OutboundNetInterface, rule.Access)
			}
			outbound, exists := interfaceNames[rule.OutboundNetInterface]
			if !exists {
				return derrors.NewFailedPreconditionError("outbound_net_interface found in security rule is not defined").WithParams(rule.OutboundNetInterface)
			}
			if outbound == true {
				// the interface named rule.outboundInterface is an inbound
				return derrors.NewFailedPreconditionError("outbound_net_interface found in security rule is defined as inbound").WithParams(rule.OutboundNetInterface)
			}
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
		    - Validate that the inbound and outbound net interfaces are always linked to specific rules
	*/

	// TODO: the service should be unique per group, not unique in the descriptor
	appServices := make(map[string]*grpc_application_go.Service)
	appGroups := make(map[string]*grpc_application_go.ServiceGroup)

	// at least one group should be defined
	if appDescriptor.Groups == nil || len(appDescriptor.Groups) <= 0 {
		return derrors.NewFailedPreconditionError("at least one group should be defined")
	}

	// - Service and service group name are uniq (and they cannot be empty)
	for _, appGroup := range appDescriptor.Groups {

		if appGroup.Name == "" {
			return derrors.NewFailedPreconditionError("Service Group Name cannot be empty")
		}

		// ServiceGroupId is filled by system-model
		if appGroup.ServiceGroupId != "" {
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

			_, exists := appServices[service.Name]
			if exists {
				return derrors.NewFailedPreconditionError("Service Name defined twice").WithParams(appGroup.Name, service.Name)
			}
			appServices[service.Name] = service

			// ConfigFiledId is filled by system-model
			if service.Configs != nil && len(service.Configs) > 0 {
				for _, config := range service.Configs {
					if config.ConfigFileId != "" {
						return derrors.NewFailedPreconditionError("Config File Id cannot be filled").WithParams(config.Name)
					}
				}
			}
			// validate the options in the endpoints
			// the allowed values are: CLIENT_MAX_BODY_SIZE or HOST_HEADER_CONFIGURATION
			// if and only if Type = WEB
			for _, port := range service.ExposedPorts {
				for _, endpoint := range port.Endpoints {
					if endpoint.Options != nil && len(endpoint.Options) > 0 {
						if endpoint.Type != grpc_application_go.EndpointType_WEB {
							return derrors.NewFailedPreconditionError("Endpoint options not allowed for this endpoint type").WithParams(endpoint.Type.String())
						}
						for key := range endpoint.Options {
							if key != grpc_application_go.EndpointOptions_CLIENT_MAX_BODY_SIZE.String() && key != grpc_application_go.EndpointOptions_HOST_HEADER_CONFIGURATION.String() {
								return derrors.NewFailedPreconditionError("Endpoint option not allowed").WithParams(key)
							}
						}
					}
				}
			}

		}
		appGroups[appGroup.Name] = appGroup
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

	// - Validate the inbound and outbound net interfaces
	err = ValidAppDescriptorNetInterfaces(appDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// ValidAppDescriptorNetInterfaces This function validates that:
// - All inbound and outbound net interfaces are linked to a security rule. This is because rules can not be generated
// after deployment and a net interface without rule does not make sense.
func ValidAppDescriptorNetInterfaces(addAppDescriptorRequest *grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	var missedInterface string
	found := true
	for _, inboundNetInterface := range addAppDescriptorRequest.InboundNetInterfaces {
		found = false
		for _, rule := range addAppDescriptorRequest.Rules {
			if rule.InboundNetInterface == inboundNetInterface.Name {
				found = true
			}
		}
		if !found {
			missedInterface = inboundNetInterface.Name
			break
		}
	}
	if !found {
		return derrors.NewFailedPreconditionError("The inbound net interface is not linked to a rule").WithParams(missedInterface)
	}
	for _, outboundNetInterface := range addAppDescriptorRequest.OutboundNetInterfaces {
		found = false
		for _, rule := range addAppDescriptorRequest.Rules {
			if rule.OutboundNetInterface == outboundNetInterface.Name {
				found = true
			}
		}
		if !found {
			missedInterface = outboundNetInterface.Name
			break
		}
	}
	if !found {
		return derrors.NewFailedPreconditionError("The outbound net interface is not linked to a rule").WithParams(missedInterface)
	}
	return nil
}

func ValidAddConnectionRequest(addRequest *grpc_application_network_go.AddConnectionRequest) derrors.Error {
	if addRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if addRequest.TargetInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyTargetInstanceId)
	}
	if addRequest.SourceInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptySourceInstanceId)
	}
	if addRequest.InboundName == "" {
		return derrors.NewInvalidArgumentError(emptyInboundName)
	}
	if addRequest.OutboundName == "" {
		return derrors.NewInvalidArgumentError(emptyOutboundName)
	}

	return nil
}

func ValidRemoveConnectionRequest(removeRequest *grpc_application_network_go.RemoveConnectionRequest) derrors.Error {
	if removeRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if removeRequest.TargetInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyTargetInstanceId)
	}
	if removeRequest.SourceInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptySourceInstanceId)
	}
	if removeRequest.InboundName == "" {
		return derrors.NewInvalidArgumentError(emptyInboundName)
	}
	if removeRequest.OutboundName == "" {
		return derrors.NewInvalidArgumentError(emptyOutboundName)
	}
	return nil
}

func ValidSearchRequest(request *grpc_application_manager_go.SearchRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	// validate the field dependencies,
	if request.ServiceGroupId != "" && request.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	if request.ServiceGroupInstanceId != "" {
		if request.AppInstanceId == "" {
			return derrors.NewInvalidArgumentError(emptyInstanceId)
		} else if request.ServiceGroupId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceGroupId)
		}
	}
	if request.ServiceId != "" {
		if request.AppInstanceId == "" {
			return derrors.NewInvalidArgumentError(emptyInstanceId)
		} else if request.ServiceGroupId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceGroupId)
		} else if request.ServiceGroupInstanceId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceGroupInstanceId)
		}
	}
	if request.ServiceInstanceId != "" {
		if request.AppInstanceId == "" {
			return derrors.NewInvalidArgumentError(emptyInstanceId)
		} else if request.ServiceGroupId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceGroupId)
		} else if request.ServiceGroupInstanceId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceGroupInstanceId)
		} else if request.ServiceId == "" {
			return derrors.NewInvalidArgumentError(emptyServiceId)
		}
	}

	return nil
}

func ValidAvailableLogRequest(request *grpc_application_manager_go.AvailableLogRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if request.To < request.From {
		return derrors.NewInvalidArgumentError(impossibleDuration)
	}

	return nil
}
