/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

/*
RUN_INTEGRATION_TEST=true
IT_SM_ADDRESS=localhost:8800
IT_CONDUCTOR_ADDRESS=localhost:5000
 */

package application

import (
	"context"
	"fmt"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/grpc-utils/pkg/test"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"math/rand"
	"os"
)

func GetAddAppDescriptorRequest(name string, organizationID string) * grpc_application_go.AddAppDescriptorRequest{
	service := &grpc_application_go.Service{
		OrganizationId:       organizationID,
		ServiceId:            "1",
		Name:                 "minimal-nginx",
		Description:          "minimal IT nginx",
		Type:                 grpc_application_go.ServiceType_DOCKER,
		Image:                "nginx:1.12",
		Specs:                &grpc_application_go.DeploySpecs{
			Replicas:             1,
		},
	}

	toAdd := &grpc_application_go.AddAppDescriptorRequest{
		RequestId:            fmt.Sprintf("application-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		OrganizationId:       organizationID,
		Name:                 fmt.Sprintf("%s-app-manager-it-%d", name, ginkgo.GinkgoRandomSeed()),
		Description:          "IT minimal descriptor",
		ConfigurationOptions: nil,
		EnvironmentVariables: nil,
		Labels:               nil,
		Rules:                nil,
		Groups:               nil,
		Services:             []*grpc_application_go.Service{service},
	}
	return toAdd
}

func CreateAppDescriptor(name string, organizationID string, appClient grpc_application_go.ApplicationsClient) * grpc_application_go.AppDescriptor{
	toAdd := GetAddAppDescriptorRequest(name, organizationID)
	desc, err := appClient.AddAppDescriptor(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	return desc
}

func CreateOrganization(name string, orgClient grpc_organization_go.OrganizationsClient) * grpc_organization_go.Organization {
	toAdd := &grpc_organization_go.AddOrganizationRequest{
		Name:                 fmt.Sprintf("%s-%d-%d", name, ginkgo.GinkgoRandomSeed(), rand.Int()),
	}
	added, err := orgClient.AddOrganization(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}

var _ = ginkgo.Describe("Application Manager service", func() {

	if ! utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var (
		systemModelAddress = os.Getenv("IT_SM_ADDRESS")
		conductorAddress = os.Getenv("IT_CONDUCTOR_ADDRESS")
	)

	if systemModelAddress == "" || conductorAddress == "" {
		ginkgo.Fail("missing environment variables")
	}

	// gRPC server
	var server *grpc.Server
	// grpc test listener
	var listener *bufconn.Listener
	// client
	var orgClient grpc_organization_go.OrganizationsClient
	var appClient grpc_application_go.ApplicationsClient
	var conductorClient grpc_conductor_go.ConductorClient
	var smConn * grpc.ClientConn
	var conductorConn * grpc.ClientConn
	var client grpc_application_manager_go.ApplicationManagerClient

	// Target organization.
	var targetOrganization *grpc_organization_go.Organization
	var targetAppDescriptor *grpc_application_go.AppDescriptor

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		smConn = utils.GetConnection(systemModelAddress)
		orgClient = grpc_organization_go.NewOrganizationsClient(smConn)
		appClient = grpc_application_go.NewApplicationsClient(smConn)
		conductorConn = utils.GetConnection(conductorAddress)
		conductorClient = grpc_conductor_go.NewConductorClient(conductorConn)

		test.LaunchServer(server, listener)

		// Register the service
		manager := NewManager(appClient, conductorClient)
		handler := NewHandler(manager)
		grpc_application_manager_go.RegisterApplicationManagerServer(server, handler)

		conn, err := test.GetConn(*listener)
		gomega.Expect(err).Should(gomega.Succeed())
		client = grpc_application_manager_go.NewApplicationManagerClient(conn)
	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		listener.Close()
	})

	ginkgo.BeforeEach(func(){
		ginkgo.By("creating target entities", func(){
			// Initial data
			targetOrganization = CreateOrganization("app-manager-it", orgClient)
			targetAppDescriptor = CreateAppDescriptor("app-manager-it", targetOrganization.OrganizationId, appClient)
		})
	})

	ginkgo.It("Should be able to add a new application descriptor", func(){
	    added, err := client.AddAppDescriptor(context.Background(),
	    	GetAddAppDescriptorRequest("add-test", targetOrganization.OrganizationId))
	    gomega.Expect(err).To(gomega.Succeed())
	    gomega.Expect(added.AppDescriptorId).ShouldNot(gomega.BeEmpty())
	})

	ginkgo.It("should be able to list existing app descriptors", func(){
		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId:       targetOrganization.OrganizationId,
		}
	    descriptors, err := client.ListAppDescriptors(context.Background(), organizationID)
	    gomega.Expect(err).To(gomega.Succeed())
	    gomega.Expect(len(descriptors.Descriptors)).Should(gomega.Equal(1))
	})

	ginkgo.It("should be able to get an existing app descriptor", func(){
		appDescriptorID := &grpc_application_go.AppDescriptorId{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
		}
	    retrieved, err := client.GetAppDescriptor(context.Background(), appDescriptorID)
	    gomega.Expect(err).To(gomega.Succeed())
	    gomega.Expect(retrieved.AppDescriptorId).Should(gomega.Equal(targetAppDescriptor.AppDescriptorId))
	})

	ginkgo.It("should be able to delete a descriptor without instances", func(){
		appDescriptorID := &grpc_application_go.AppDescriptorId{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
		}
		success, err := client.RemoveAppDescriptor(context.Background(), appDescriptorID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(success).ShouldNot(gomega.BeNil())
	})

	ginkgo.It("should be able to deploy an instance", func(){
		deployRequest := &grpc_application_manager_go.DeployRequest{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
			Name:                 "test-deploy-app-manager",
			Description:          "Test deploy from app mananager IT",
		}
	    response, err := client.Deploy(context.Background(), deployRequest)
	    if err != nil {
	    	fmt.Println(conversions.ToDerror(err).DebugReport())
		}
	    gomega.Expect(err).To(gomega.Succeed())
	    gomega.Expect(response.AppInstanceId).ShouldNot(gomega.BeEmpty())
	})

	ginkgo.It("should not be able to delete a descriptor with instances", func(){
		deployRequest := &grpc_application_manager_go.DeployRequest{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
			Name:                 "test-deploy-app-manager",
			Description:          "Test deploy from app mananager IT",
		}
		response, err := client.Deploy(context.Background(), deployRequest)
		if err != nil {
			fmt.Println(conversions.ToDerror(err).DebugReport())
		}
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(response.AppInstanceId).ShouldNot(gomega.BeEmpty())

		appDescriptorID := &grpc_application_go.AppDescriptorId{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
		}
		success, err := client.RemoveAppDescriptor(context.Background(), appDescriptorID)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(success).Should(gomega.BeNil())
	})

	ginkgo.PIt("should be able to undeploy a running instance", func(){
		deployRequest := &grpc_application_manager_go.DeployRequest{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
			Name:                 "test-deploy-app-manager",
			Description:          "Test deploy from app mananager IT",
		}
		response, err := client.Deploy(context.Background(), deployRequest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(response.AppInstanceId).ShouldNot(gomega.BeEmpty())
		instanceID := &grpc_application_go.AppInstanceId{
			OrganizationId:       targetOrganization.OrganizationId,
			AppInstanceId:        response.AppInstanceId,
		}
		client.Undeploy(context.Background(), instanceID)
	})

	ginkgo.It("should be able to get a running application instance", func(){
		deployRequest := &grpc_application_manager_go.DeployRequest{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
			Name:                 "test-deploy-app-manager",
			Description:          "Test deploy from app mananager IT",
		}
		response, err := client.Deploy(context.Background(), deployRequest)
		gomega.Expect(err).To(gomega.Succeed())

		appInstanceID := &grpc_application_go.AppInstanceId{
			OrganizationId:       targetOrganization.OrganizationId,
			AppInstanceId:        response.AppInstanceId,
		}

		retrieved, err := client.GetAppInstance(context.Background(), appInstanceID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(retrieved.AppInstanceId).Should(gomega.Equal(response.AppInstanceId))
	})

	ginkgo.It("should be able to list running applications", func(){
		deployRequest := &grpc_application_manager_go.DeployRequest{
			OrganizationId:       targetAppDescriptor.OrganizationId,
			AppDescriptorId:      targetAppDescriptor.AppDescriptorId,
			Name:                 "test-deploy-app-manager",
			Description:          "Test deploy from app mananager IT",
		}
		_, err := client.Deploy(context.Background(), deployRequest)
		gomega.Expect(err).To(gomega.Succeed())
		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId:       targetOrganization.OrganizationId,
		}
		instances, err := client.ListAppInstances(context.Background(), organizationID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(len(instances.Instances)).Should(gomega.Equal(1))
	})

})