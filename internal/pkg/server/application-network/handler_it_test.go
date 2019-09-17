package application_network

import (
	"context"
	"fmt"
	"github.com/google/uuid"
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


func CreateAppDescriptor(organizationID string, appClient grpc_application_go.ApplicationsClient) * grpc_application_go.AppDescriptor{
	toAdd := utils.CreateAddAppDescriptorRequest(organizationID, nil, map[string]string{"lab": "value1"})
	desc, err := appClient.AddAppDescriptor(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	return desc
}

func CreateOrganization(orgClient grpc_organization_go.OrganizationsClient) * grpc_organization_go.Organization {
	toAdd := &grpc_organization_go.AddOrganizationRequest{
		Name:                 fmt.Sprintf("org-%s", uuid.New().String()),
	}
	added, err := orgClient.AddOrganization(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}

func CreateAppInstanceWithInbounds(organizationID string, appDescriptorID string, inboundName string, appClient grpc_application_go.ApplicationsClient) *grpc_application_go.AppInstance{
	appInstance := utils.CreateTestAppInstanceRequest(organizationID, appDescriptorID)
	appInstance.Name = "inbound instance"
	appInstance.InboundNetInterfaces = []*grpc_application_go.InboundNetworkInterface{{Name: inboundName}}
	inst, err := appClient.AddAppInstance(context.Background(), appInstance)
	gomega.Expect(err).To(gomega.Succeed())
	return inst
}
func CreateAppInstanceWithOutbounds(organizationID string, appDescriptorID string, outboundName string, appClient grpc_application_go.ApplicationsClient) *grpc_application_go.AppInstance{
	appInstance := utils.CreateTestAppInstanceRequest(organizationID, appDescriptorID)
	appInstance.Name = "outbound instance"
	appInstance.OutboundNetInterfaces = []*grpc_application_go.OutboundNetworkInterface{{Name: outboundName, Required:false}}
	inst, err := appClient.AddAppInstance(context.Background(), appInstance)
	gomega.Expect(err).To(gomega.Succeed())
	return inst
}

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
	var orgClient grpc_organization_go.OrganizationsClient
	var appNetClient grpc_application_network_go.ApplicationNetworkClient
	var appClient grpc_application_go.ApplicationsClient
	var smConn * grpc.ClientConn
	var client grpc_application_manager_go.ApplicationNetworkClient

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		smConn = utils.GetConnection(systemModelAddress)
		orgClient = grpc_organization_go.NewOrganizationsClient(smConn)
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
		ginkgo.It("Should be able to add a new connection", func() {
			organization := CreateOrganization( orgClient)
			// Add Descriptors
			targetAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)
			sourceAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)

			// Add Instances
			targetInstance := CreateAppInstanceWithInbounds(organization.OrganizationId, targetAppDescriptor.AppDescriptorId, "inbound", appClient)
			sourceInstance := CreateAppInstanceWithOutbounds(organization.OrganizationId, sourceAppDescriptor.AppDescriptorId, "outbound", appClient)

			// add Connection
			instance, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: organization.OrganizationId,
				TargetInstanceId: targetInstance.AppInstanceId,
				InboundName: "inbound",
				SourceInstanceId: sourceInstance.AppInstanceId,
				OutboundName: "outbound",
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instance.SourceInstanceName).Should(gomega.Equal(sourceInstance.Name))
			gomega.Expect(instance.TargetInstanceName).Should(gomega.Equal(targetInstance.Name))

		})
		ginkgo.It("Should not be able to add a new connection, target_instance_id does not exists", func() {
			organization := CreateOrganization( orgClient)

			_, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: organization.OrganizationId,
				TargetInstanceId: uuid.New().String(),
				InboundName: "inbound",
				SourceInstanceId: uuid.New().String(),
				OutboundName: "outbound",
			})
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to add a new connection, inbound does not exists", func() {
			organization := CreateOrganization( orgClient)
			// Add Descriptors
			targetAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)
			sourceAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)

			// Add Instances
			targetInstance := CreateAppInstanceWithInbounds(organization.OrganizationId, targetAppDescriptor.AppDescriptorId, "inbound", appClient)
			sourceInstance := CreateAppInstanceWithOutbounds(organization.OrganizationId, sourceAppDescriptor.AppDescriptorId, "outbound", appClient)

			// add Connection
			_, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: organization.OrganizationId,
				TargetInstanceId: targetInstance.AppInstanceId,
				InboundName: "wrong inbound",
				SourceInstanceId: sourceInstance.AppInstanceId,
				OutboundName: "outbound",
			})
			gomega.Expect(err).NotTo(gomega.Succeed())

		})
		ginkgo.It("Should not be able to add a new connection, validation error", func() {
			// add Connection
			_, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: "",
				TargetInstanceId: uuid.New().String(),
				InboundName: "",
				SourceInstanceId: uuid.New().String(),
				OutboundName: "",
			})
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
	})
	ginkgo.Context("RemoveConnection test", func() {
		ginkgo.It("Should be able to remove a connection", func() {
			organization := CreateOrganization( orgClient)
			// Add Descriptors
			targetAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)
			sourceAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)

			// Add Instances
			targetInstance := CreateAppInstanceWithInbounds(organization.OrganizationId, targetAppDescriptor.AppDescriptorId, "inbound", appClient)
			sourceInstance := CreateAppInstanceWithOutbounds(organization.OrganizationId, sourceAppDescriptor.AppDescriptorId, "outbound", appClient)

			// add Connection
			instance, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: organization.OrganizationId,
				TargetInstanceId: targetInstance.AppInstanceId,
				InboundName: "inbound",
				SourceInstanceId: sourceInstance.AppInstanceId,
				OutboundName: "outbound",
			})
			gomega.Expect(err).To(gomega.Succeed())

			// Remove Connection
			success, err := client.RemoveConnection(context.Background(), &grpc_application_network_go.RemoveConnectionRequest{
				OrganizationId: instance.OrganizationId,
				SourceInstanceId: instance.SourceInstanceId,
				TargetInstanceId: instance.TargetInstanceId,
				InboundName: "inbound",
				OutboundName: "outbound",
				UserConfirmation: true,
			})
			gomega.Expect(success).ShouldNot(gomega.BeNil())
			gomega.Expect(err).To(gomega.Succeed())

		})
		ginkgo.It("Should not be able to remove connection, validation error", func() {
			// Remove Connection
			_, err := client.RemoveConnection(context.Background(), &grpc_application_network_go.RemoveConnectionRequest{
				OrganizationId: "",
				SourceInstanceId: "",
				TargetInstanceId: "",
				InboundName: "inbound",
				OutboundName: "outbound",
				UserConfirmation: true,
			})
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("Should not be able to remove connection if it does not exists", func() {
			// Remove Connection
			_, err := client.RemoveConnection(context.Background(), &grpc_application_network_go.RemoveConnectionRequest{
				OrganizationId: uuid.New().String(),
				SourceInstanceId: uuid.New().String(),
				TargetInstanceId: uuid.New().String(),
				InboundName: "inbound",
				OutboundName: "outbound",
				UserConfirmation: true,
			})
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
	})
	ginkgo.Context("ListConnection test", func() {
		ginkgo.It("Should be able to list connections of an organization", func() {
			organization := CreateOrganization( orgClient)
			// Add Descriptors
			targetAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)
			sourceAppDescriptor := CreateAppDescriptor(organization.OrganizationId, appClient)

			// Add Instances
			targetInstance := CreateAppInstanceWithInbounds(organization.OrganizationId, targetAppDescriptor.AppDescriptorId, "inbound", appClient)
			sourceInstance := CreateAppInstanceWithOutbounds(organization.OrganizationId, sourceAppDescriptor.AppDescriptorId, "outbound", appClient)

			// add Connection
			instance, err := client.AddConnection(context.Background(), &grpc_application_network_go.AddConnectionRequest{
				OrganizationId: organization.OrganizationId,
				TargetInstanceId: targetInstance.AppInstanceId,
				InboundName: "inbound",
				SourceInstanceId: sourceInstance.AppInstanceId,
				OutboundName: "outbound",
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instance.SourceInstanceName).Should(gomega.Equal(sourceInstance.Name))
			gomega.Expect(instance.TargetInstanceName).Should(gomega.Equal(targetInstance.Name))

			list, err := client.ListConnections(context.Background(), &grpc_organization_go.OrganizationId{
				OrganizationId: organization.OrganizationId,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(list).NotTo(gomega.BeNil())
			gomega.Expect(len(list.Connections)).Should(gomega.Equal(1))
		})
		ginkgo.It("Should be able to list an empty list of connections of an organization", func() {
			organization := CreateOrganization( orgClient)
			list, err := client.ListConnections(context.Background(), &grpc_organization_go.OrganizationId{
				OrganizationId: organization.OrganizationId,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(list).NotTo(gomega.BeNil())
			gomega.Expect(len(list.Connections)).Should(gomega.Equal(0))
		})

	})
})
