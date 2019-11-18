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

package unified_logging

import (
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-unified-logging-go"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	unifiedLogging grpc_unified_logging_go.CoordinatorClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(unifiedLogging grpc_unified_logging_go.CoordinatorClient) Manager {
	return Manager{unifiedLogging }
}

func (m *Manager) Search(request *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error){

	// TODO: por aquí
	return nil, nil
}
