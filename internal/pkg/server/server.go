/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
	"fmt"
	"github.com/nalej/application-manager/internal/pkg/bus"
	"github.com/nalej/application-manager/internal/pkg/server/application"
	"github.com/nalej/application-manager/internal/pkg/server/application-network"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

// Service structure with the configuration and the gRPC server.
type Service struct {
	Configuration Config
}

// NewService creates a new system model service.
func NewService(conf Config) *Service {
	return &Service{
		conf,
	}
}

// Clients structure with the gRPC clients for remote services.
type Clients struct {
	AppClient grpc_application_go.ApplicationsClient
	ConductorClient grpc_conductor_go.ConductorClient
	ClusterClient grpc_infrastructure_go.ClustersClient
	DeviceClient  grpc_device_go.DevicesClient
	AppNetClient grpc_application_network_go.ApplicationNetworkClient
}

// GetClients creates the required connections with the remote clients.
func (s * Service) GetClients() (* Clients, derrors.Error) {
	conductorConn, err := grpc.Dial(s.Configuration.ConductorAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the conductor component")
	}

	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	aClient := grpc_application_go.NewApplicationsClient(smConn)
	cClient := grpc_conductor_go.NewConductorClient(conductorConn)
	clClient := grpc_infrastructure_go.NewClustersClient(smConn)
	dvClient := grpc_device_go.NewDevicesClient(smConn)
	appNetClient := grpc_application_network_go.NewApplicationNetworkClient(smConn)

	return &Clients{aClient, cClient, clClient, dvClient, appNetClient}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() error {
	cErr := s.Configuration.Validate()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("invalid configuration")
	}
	s.Configuration.Print()
	clients, cErr := s.GetClients()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	log.Info().Msg("instatiating bus client...")
	// Create bus client
	queueClient := pulsar_comcast.NewClient(s.Configuration.QueueAddress)
	if err != nil {
		log.Panic().Err(err).Msg("impossible to create bus client instance")
		return err
	}
	log.Info().Msg("done")
	// Instantiate the bus manager
	log.Info().Msg("instantiating bus manager...")
	busManager, err := bus.NewBusManager(queueClient, "ApplicationManager")
	if err != nil {
		log.Panic().Err(err).Msg("impossible to create bus manager instance")
		return err
	}
	log.Info().Msg("done")

	// Create handlers
	manager := application.NewManager(clients.AppClient, clients.ConductorClient, clients.ClusterClient, clients.DeviceClient, busManager)
	handler := application.NewHandler(manager)

	appNetManager := application_network.NewManager(clients.AppNetClient, clients.AppClient)
	appNetHandler := application_network.NewHandler(appNetManager)

	grpcServer := grpc.NewServer()
	grpc_application_manager_go.RegisterApplicationManagerServer(grpcServer, handler)
	grpc_application_network_go.RegisterApplicationNetworkServer(grpcServer, appNetHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
	return nil
}