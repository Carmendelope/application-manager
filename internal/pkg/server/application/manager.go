package application

import (
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-conductor-go"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	appClient       grpc_application_go.ApplicationsClient
	conductorClient grpc_conductor_go.ConductorClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	appClient grpc_application_go.ApplicationsClient,
	conductorClient grpc_conductor_go.ConductorClient,
) Manager {
	return Manager{appClient, conductorClient}
}
