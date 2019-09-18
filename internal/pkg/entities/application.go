package entities

import (
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
)

func ToAppInstance(source *grpc_application_go.AppInstance) *grpc_application_manager_go.AppInstance {

	return &grpc_application_manager_go.AppInstance{
		OrganizationId:       	source.OrganizationId,
		AppDescriptorId:      	source.AppDescriptorId,
		AppInstanceId:        	source.AppInstanceId,
		Name:                 	source.Name,
		ConfigurationOptions: 	source.ConfigurationOptions,
		EnvironmentVariables: 	source.EnvironmentVariables,
		Labels:               	source.Labels,
		Rules:                	source.Rules,
		Groups:               	source.Groups,
		Status:           		source.Status,
		Metadata:			  	source.Metadata,
		Info: 				  	source.Info,
		InboundNetInterfaces: 	source.InboundNetInterfaces,
		OutboundNetInterfaces:	source.OutboundNetInterfaces,
	}
}