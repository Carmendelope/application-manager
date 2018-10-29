/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
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
}

func (conf * Config) Validate() derrors.Error {

	if conf.ConductorAddress == "" {
		return derrors.NewInvalidArgumentError("conductorAddress must be set")
	}

	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress must be set")
	}

	return nil
}

func (conf *Config) Print() {
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.ConductorAddress).Msg("Conductor")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("System Model")
}
