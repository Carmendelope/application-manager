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

package server

import (
	"fmt"
	"github.com/nalej/application-manager/internal/pkg/queue"
	"github.com/nalej/application-manager/internal/pkg/server/application"
	"github.com/nalej/application-manager/internal/pkg/server/application-network"
	"github.com/nalej/application-manager/internal/pkg/server/unified-logging"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	grpc_application_history_logs_go "github.com/nalej/grpc-application-history-logs-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-unified-logging-go"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/application/events"
	"github.com/nalej/nalej-bus/pkg/queue/application/ops"
	networkOps "github.com/nalej/nalej-bus/pkg/queue/network/ops"
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
	AppClient            grpc_application_go.ApplicationsClient
	ConductorClient      grpc_conductor_go.ConductorClient
	ClusterClient        grpc_infrastructure_go.ClustersClient
	DeviceClient         grpc_device_go.DevicesClient
	AppNetClient         grpc_application_network_go.ApplicationNetworkClient
	CoordinatorClient    grpc_unified_logging_go.CoordinatorClient
	UnifiedLoggingClient grpc_application_manager_go.UnifiedLoggingClient
	AppHistoryLogsClient grpc_application_history_logs_go.ApplicationHistoryLogsClient
}

type BusClients struct {
	AppOpsProducer    *ops.ApplicationOpsProducer
	NetOpsProducer    *networkOps.NetworkOpsProducer
	AppEventsConsumer *events.ApplicationEventsConsumer
}

// GetBusClients creates the required connections with the bus
func (s *Service) GetBusClients() (*BusClients, derrors.Error) {
	queueClient := pulsar_comcast.NewClient(s.Configuration.QueueAddress, nil)

	appOpsProducer, err := ops.NewApplicationOpsProducer(queueClient, "ApplicationManager-app_ops")
	if err != nil {
		return nil, err
	}

	netOpsProducer, err := networkOps.NewNetworkOpsProducer(queueClient, "ApplicationManager-network_ops")
	if err != nil {
		return nil, err
	}

	appEventsConfig := events.NewConfigApplicationEventsConsumer(5, events.ConsumableStructsApplicationEventsConsumer{
		DeploymentServiceUpdateRequest: true,
	})
	appEventsConsumer, err := events.NewApplicationEventsConsumer(queueClient, "network-manager-application-events", true, appEventsConfig)

	return &BusClients{
		AppOpsProducer:    appOpsProducer,
		NetOpsProducer:    netOpsProducer,
		AppEventsConsumer: appEventsConsumer,
	}, nil
}

// GetClients creates the required connections with the remote clients.
func (s *Service) GetClients() (*Clients, derrors.Error) {
	conductorConn, err := grpc.Dial(s.Configuration.ConductorAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the conductor component")
	}

	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	coordConn, err := grpc.Dial(s.Configuration.UnifiedLoggingAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with unified logging coordinator")
	}

	ulConn, err := grpc.Dial(s.Configuration.UnifiedLoggingAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with unified logging")
	}

	ahlConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	aClient := grpc_application_go.NewApplicationsClient(smConn)
	cClient := grpc_conductor_go.NewConductorClient(conductorConn)
	clClient := grpc_infrastructure_go.NewClustersClient(smConn)
	dvClient := grpc_device_go.NewDevicesClient(smConn)
	appNetClient := grpc_application_network_go.NewApplicationNetworkClient(smConn)
	coordClient := grpc_unified_logging_go.NewCoordinatorClient(coordConn)
	ulClient := grpc_application_manager_go.NewUnifiedLoggingClient(ulConn)
	ahlClient := grpc_application_history_logs_go.NewApplicationHistoryLogsClient(ahlConn)

	return &Clients{aClient, cClient, clClient,
		dvClient, appNetClient, coordClient, ulClient, ahlClient}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() error {
	// Configuration
	cErr := s.Configuration.Validate()
	if cErr != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("invalid configuration")
	}
	s.Configuration.Print()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	// Clients
	clients, cErr := s.GetClients()
	if cErr != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	// BusClients
	busClients, bErr := s.GetBusClients()
	if err != nil {
		log.Fatal().Str("err", bErr.DebugReport()).Msg("Cannot create bus clients")
	}

	// Create handlers
	appNetManager := application_network.NewManager(clients.AppNetClient, clients.AppClient, busClients.NetOpsProducer)
	appNetHandler := application_network.NewHandler(appNetManager)

	unifiedLoggingManager, err := unified_logging.NewManager(clients.CoordinatorClient, clients.AppClient, clients.AppHistoryLogsClient, busClients.AppEventsConsumer)
	if err != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create unified-logging manager")
	}
	unifiedLogHandler := unified_logging.NewHandler(*unifiedLoggingManager)

	manager := application.NewManager(clients.AppClient, clients.ConductorClient, clients.ClusterClient, clients.DeviceClient, clients.AppNetClient, busClients.AppOpsProducer, appNetManager)
	handler := application.NewHandler(manager)

	appEventsHandler := queue.NewAppEventsHandler(unifiedLoggingManager, busClients.AppEventsConsumer)
	appEventsHandler.Run()

	grpcServer := grpc.NewServer()
	grpc_application_manager_go.RegisterApplicationManagerServer(grpcServer, handler)
	grpc_application_manager_go.RegisterApplicationNetworkServer(grpcServer, appNetHandler)
	grpc_application_manager_go.RegisterUnifiedLoggingServer(grpcServer, unifiedLogHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
	return nil
}
