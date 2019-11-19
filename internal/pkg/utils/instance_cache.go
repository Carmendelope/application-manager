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

package utils

import (
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"sync"
)

// InstCache manage the instances cache.
type InstanceCache struct {
	appClient grpc_application_go.ApplicationsClient
	// map of instance identifiers indexed by organizationId, descriptorID
	instanceIds map[string]map[string][]string
	// map of instances indexed by organizationId, instanceId
	// TODO: preguntar a Dani si nuestro compromiso es instanceId único por Organization o por appDescriptor
	instances map[string]map[string]*grpc_application_go.AppInstance
	// map of descriptor names indexed by organizationId, descriptorId
	// TODO: ask where remove this structure
	descriptors map [string]map[string]string
	sync.Mutex
}

// NewManager creates a Manager using a set of clients.
func NewInstCache(appClient grpc_application_go.ApplicationsClient) InstanceCache {
	ids := make(map[string]map[string][]string, 0)
	instances := make(map[string]map[string]*grpc_application_go.AppInstance, 0)
	descriptors := make(map [string]map[string]string, 0)
	return InstanceCache{
		appClient:   appClient,
		instanceIds: ids,
		instances:   instances,
		descriptors: descriptors,
	}
}

func (ic *InstanceCache) UpdateInstanceList(organizationId string) derrors.Error {
	// get all the instances
	ctx, cancel := common.GetContext()
	defer cancel()
	list, err := ic.appClient.ListAppInstances(ctx, &grpc_organization_go.OrganizationId{OrganizationId: organizationId})
	if err != nil {
		return conversions.ToDerror(err)
	}
	ic.Lock()
	defer ic.Unlock()

	// delete the organization entry
	delete(ic.instanceIds, organizationId)
	delete(ic.instances, organizationId)

	if len(list.Instances) > 0 {
		ids := make(map[string][]string, 0)
		instances := make(map[string]*grpc_application_go.AppInstance, 0)
		for _, inst := range list.Instances {
			// Ids
			_, exists := ids[inst.AppDescriptorId]
			if exists {
				ids[inst.AppDescriptorId] = append(ids[inst.AppDescriptorId], inst.AppInstanceId)
			} else {
				ids[inst.AppDescriptorId] = []string{inst.AppInstanceId}
			}
			// Instances
			instances[inst.AppInstanceId] = inst
		}
		ic.instanceIds[organizationId] = ids
		ic.instances[organizationId] = instances
	}

	return nil
}

// retrieveInstancesOfADescriptor returns a list with the appInstance identifiers of a descriptor
func (ic *InstanceCache) RetrieveInstancesOfADescriptor(organizationId string, descriptorId string) []string {
	ic.Lock()
	defer ic.Unlock()

	list, exists := ic.instanceIds[organizationId]
	if !exists {
		return make([]string, 0)
	}
	instances, exists := list[descriptorId]
	if !exists {
		return make([]string, 0)
	}
	return instances
}

// retrieveInstancesOfAnOrganization returns a list with the appInstance identifiers of an organization
func (ic *InstanceCache) RetrieveInstancesOfAnOrganization(organizationId string) []string {
	ic.Lock()
	defer ic.Unlock()

	res := make([]string, 0)
	list, exists := ic.instanceIds[organizationId]
	if !exists {
		return res
	}
	for _, instances := range list {
		for _, id := range instances {
			res = append(res, id)
		}
	}

	return res
}

func (ic *InstanceCache) RetrieveInstanceInformation(organizationId string, instanceId string) (*grpc_application_go.AppInstance, derrors.Error){

	// find the instance
	ic.Lock()
	defer ic.Unlock()

	list, exists := ic.instances[organizationId]
	if exists {
		instance, exists := list[instanceId]
		if exists {
			return instance, nil // TODO: ask, whats happend if after return the instance, it is removed from memory (delete)
		}
	}

	// not exists -> get it
	ctx, cancel := common.GetContext()
	defer cancel()
	instance, err := ic.appClient.GetAppInstance(ctx, &grpc_application_go.AppInstanceId{OrganizationId: organizationId, AppInstanceId: instanceId})
	if err != nil {
		return nil, conversions.ToDerror(err)
	}
	// TODO: add the instance in memory structures
	return instance, nil
}

func (ic *InstanceCache) RetrieveDescriptorName (organizationId string, descriptorId string) (string, derrors.Error) {

	log.Debug().Str("organizationId", organizationId).Str("descriptorId", descriptorId).Msg("RetrieveDescriptorName")

	ic.Lock()
	defer ic.Unlock()

	list, exists := ic.descriptors[organizationId]
	if exists {
		name, exists := list[descriptorId]
		if exists {
			log.Debug().Str("name", name).Msg("exists")
			return name, nil
		}
	}

	// not exists
	ctx, cancel := common.GetContext()
	defer cancel()
	descriptor, err := ic.appClient.GetAppDescriptor(ctx, &grpc_application_go.AppDescriptorId{OrganizationId: organizationId, AppDescriptorId: descriptorId})
	if err != nil {
		return "", conversions.ToDerror(err)
	}

	// add in the cache
	list, exists = ic.descriptors[organizationId]
	if ! exists{
		ic.descriptors[organizationId] = map[string]string {descriptorId:descriptor.Name}
	}else{
		ic.descriptors[organizationId][descriptorId] = descriptor.Name
	}
	log.Debug().Str("name", descriptor.Name).Msg("NON exists")
	return descriptor.Name, nil
}