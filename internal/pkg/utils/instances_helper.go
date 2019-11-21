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
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

type InstancesHelper struct {
	appClient grpc_application_go.ApplicationsClient
	cache     lru.Cache
}

func NewInstancesHelper(appClient grpc_application_go.ApplicationsClient, numCachedEntries int) (*InstancesHelper, derrors.Error) {
	lruCache, err := lru.New(numCachedEntries)
	if err != nil {
		return nil, derrors.AsError(err, "cannot create cache")
	}
	return &InstancesHelper{
		appClient: appClient,
		cache:     *lruCache,
	}, nil
}

func (i *InstancesHelper) composePK(organizationId string, instanceId string) string {
	return fmt.Sprintf("%s#%s", organizationId, instanceId)
}

// RetrieveInstanceSummary looks for a Instance Summary in the cache. If it does not exists, retrieves it from the database
func (i *InstancesHelper) RetrieveInstanceSummary(organizationId string, appInstanceId string) (*grpc_application_go.AppInstanceReducedSummary, derrors.Error) {

	pk := i.composePK(organizationId, appInstanceId)

	summary, found := i.cache.Get(pk)
	if found {
		return summary.(*grpc_application_go.AppInstanceReducedSummary), nil
	}

	// else -> ask to system-model
	ctx, cancel := common.GetContext()
	defer cancel()
	retrievedSummary, err := i.appClient.GetAppInstanceReducedSummary(ctx, &grpc_application_go.AppInstanceId{
		OrganizationId: organizationId,
		AppInstanceId:  appInstanceId,
	})
	if err != nil {
		return nil, conversions.ToDerror(err)
	}

	_ = i.cache.Add(pk, retrievedSummary)
	return retrievedSummary, nil

}
