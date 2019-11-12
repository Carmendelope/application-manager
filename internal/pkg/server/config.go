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

package server

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Port where the gRPC API service will listen requests.
	Port int
	// ConductorAddress with the host:port to connect to Conductor.
	ConductorAddress string
	// SystemModelAddress with the host:port to connect to System Model
	SystemModelAddress string
	// Address where the queue system can be reached
	QueueAddress string
}

func (conf *Config) Validate() derrors.Error {

	if conf.ConductorAddress == "" {
		return derrors.NewInvalidArgumentError("conductorAddress must be set")
	}

	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress must be set")
	}

	if conf.QueueAddress == "" {
		return derrors.NewInvalidArgumentError("queueAddress must be set")
	}

	return nil
}

func (conf *Config) Print() {
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.ConductorAddress).Msg("Conductor")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("System Model")
	log.Info().Str("URL", conf.QueueAddress).Msg("Queue address")
}
