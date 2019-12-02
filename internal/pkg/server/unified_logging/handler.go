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
	"context"
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

// Handler structure for the user requests.
type Handler struct {
	Manager Manager
}

// NewHandler creates a new Handler with a linked manager.
func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

func (h *Handler) Search(_ context.Context, in *grpc_application_manager_go.SearchRequest) (*grpc_application_manager_go.LogResponse, error) {
	vErr := entities.ValidSearchRequest(in)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	return h.Manager.Search(in)
}


func (h *Handler) Catalog(_ context.Context, in *grpc_application_manager_go.AvailableLogRequest) (*grpc_application_manager_go.AvailableLogResponse, error) {
	return nil, conversions.ToGRPCError(derrors.NewUnimplementedError("not implemented yet"))
}
