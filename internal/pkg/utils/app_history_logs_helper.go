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

package utils

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-history-logs-go"
)

type AppHistoryLogsHelper struct {
	appHistoryLogsClient grpc_application_history_logs_go.ApplicationHistoryLogsClient
	cache     lru.Cache
}

func NewAppHistoryLogsHelper (appHistoryLogsClient grpc_application_history_logs_go.ApplicationHistoryLogsClient, numCachedEntries int) (*AppHistoryLogsHelper, derrors.Error) {
	lruCache, err := lru.New(numCachedEntries)
	if err != nil {
		return nil, derrors.AsError(err, "cannot create cache")
	}
	return &AppHistoryLogsHelper{
		appHistoryLogsClient: appHistoryLogsClient,
		cache:     *lruCache,
	}, nil
}

func (a * AppHistoryLogsHelper) RetrieveAppHistoryLogs (searchLogRequest *grpc_application_history_logs_go.SearchLogRequest) (*grpc_application_history_logs_go.LogResponse, error) {
	organization, found := a.cache.Get(searchLogRequest.OrganizationId)
	if found {
		return organization.(*grpc_application_history_logs_go.LogResponse), nil
	}

	// else: ask system-model
	ctx, cancel := common.GetContext()
	defer cancel()
	retrievedLogResponse, err := a.appHistoryLogsClient.Search(ctx, searchLogRequest)
	if err != nil {
		return nil, err
	}

	_ = a.cache.Add(searchLogRequest.OrganizationId, retrievedLogResponse)

	return retrievedLogResponse, nil
}