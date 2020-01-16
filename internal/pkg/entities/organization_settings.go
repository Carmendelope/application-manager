/*
 * Copyright 2020 Nalej
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

package entities

import (
	"github.com/nalej/application-manager/internal/pkg/server/common"
	"github.com/nalej/grpc-organization-go"
	orgMng "github.com/nalej/grpc-organization-manager-go"
	"github.com/rs/zerolog/log"
	"strconv"
)

// OrganizationSettings contains all settings that can be applied in descriptors and instances
// For now -> DEFAULT_STORAGE_SIZE
type OrganizationSettings struct {
	StoreSize int64
}

// NewOrganizationSettings generates a OrganizationSetting for a received organization
func NewOrganizationSettings(organizationID string, client orgMng.OrganizationsClient) *OrganizationSettings {
	log.Debug().Str("organizationId", organizationID).Msg("creating a new OrganizationSettings")

	if client == nil {
		return nil
	}

	defaultStore := int64(0)

	ctx, cancel := common.GetContext()
	defer cancel()

	setting, err := client.GetSetting(ctx, &grpc_organization_go.SettingKey{
		OrganizationId: organizationID,
		Key:            grpc_organization_go.AllowedSettingKey_DEFAULT_STORAGE_SIZE.String(),
	})
	if err != nil {
		log.Warn().Str("error", err.Error()).Str("setting", grpc_organization_go.AllowedSettingKey_DEFAULT_STORAGE_SIZE.String()).
			Msg("error getting setting")
	}else{
		defaultStore, err = strconv.ParseInt(setting.Value, 10, 64)
		if err != nil {
			log.Warn().Str("error", err.Error()).Str("value", setting.Value).Msg("error converting the value to int")
		}
	}

	return &OrganizationSettings{defaultStore}
}
