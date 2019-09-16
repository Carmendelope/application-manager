package application_network

import (
	"context"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/test"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"os"
)

var _ = ginkgo.Describe("Application Manager service", func() {

	if ! utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var (
		systemModelAddress= os.Getenv("IT_SM_ADDRESS")
	)

	if systemModelAddress == "" {
		ginkgo.Fail("missing environment variables")
	}

	// gRPC server
	var server *grpc.Server
	// grpc test listener
	var listener *bufconn.Listener
	// client
	var appNetClient grpc_application_network_go.ApplicationNetworkClient
	var appClient grpc_application_go.ApplicationsClient
	var smConn * grpc.ClientConn
	var client grpc_application_manager_go.ApplicationNetworkClient

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		smConn = utils.GetConnection(systemModelAddress)
		appNetClient = grpc_application_network_go.NewApplicationNetworkClient(smConn)
		appClient = grpc_application_go.NewApplicationsClient(smConn)

		test.LaunchServer(server, listener)

		// Register the service
		manager := NewManager(appNetClient, appClient)
		handler := NewHandler(manager)
		grpc_application_network_go.RegisterApplicationNetworkServer(server, handler)

		conn, err := test.GetConn(*listener)
		gomega.Expect(err).Should(gomega.Succeed())
		client = grpc_application_manager_go.NewApplicationNetworkClient(conn)
	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		listener.Close()
	})

	ginkgo.Context("AddConnection test", func() {
		ginkgo.PIt("Should be able to add a new connection", func() {
		})
		ginkgo.PIt("Should not be able to add a new connection, target_instance_id does not exists", func() {

		})
		ginkgo.PIt("Should not be able to add a new connection, inbound does not exists", func() {

		})
		ginkgo.PIt("Should not be able to add a new connection, validation error", func() {

		})
	})
	ginkgo.Context("RemoveConnection test", func() {
		ginkgo.PIt("Should be able to remove a connection", func() {

		})
		ginkgo.PIt("Should not be able to remove connection, validation error", func() {

		})
		ginkgo.PIt("Should not be able to remove connection if it does not exists", func() {

		})
	})
	ginkgo.Context("ListConnection test", func() {
		ginkgo.PIt("Should be able to list connections of an organization", func() {
			client.ListConnections(context.Background(), &grpc_organization_go.OrganizationId{
				OrganizationId: "org1",
			})
		})
		ginkgo.PIt("Should be able to list an empty list of connections of an organization", func() {

		})

	})
})
