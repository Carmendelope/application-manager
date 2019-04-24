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
	"github.com/google/uuid"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
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
		Name:                 "minimal-nginx",
		Type:                 grpc_application_go.ServiceType_DOCKER,
		Image:                "nginx:1.12",
		Specs:                &grpc_application_go.DeploySpecs{
			Replicas:             1,
		},
	}

	group := &grpc_application_go.ServiceGroup{
		OrganizationId:       organizationID,
		Name:                 "g1",
		Services:             []*grpc_application_go.Service{service},
		Policy:               0,
		Specs:                nil,
		Labels:               nil,
	}

	toAdd := &grpc_application_go.AddAppDescriptorRequest{
		RequestId:            fmt.Sprintf("application-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		OrganizationId:       organizationID,
		Name:                 fmt.Sprintf("%s-app-manager-it-%d", name, ginkgo.GinkgoRandomSeed()),
		ConfigurationOptions: nil,
		EnvironmentVariables: nil,
		Labels:               nil,
		Rules:                nil,
		Groups:               []*grpc_application_go.ServiceGroup{group},
	}
	return toAdd
}

func GetAddAppDescriptorWithParametersRequest(name string, organizationID string) * grpc_application_go.AddAppDescriptorRequest{
	service := &grpc_application_go.Service{
		OrganizationId:       organizationID,
		Name:                 "minimal-nginx",
		Type:                 grpc_application_go.ServiceType_DOCKER,
		Image:                "nginx:1.12",
		Specs:                &grpc_application_go.DeploySpecs{
			Replicas:             1,
		},
	}

	group := &grpc_application_go.ServiceGroup{
		OrganizationId:       organizationID,
		Name:                 "g1",
		Services:             []*grpc_application_go.Service{service},
		Policy:               0,
		Specs:                nil,
		Labels:               nil,
	}
	parameter := &grpc_application_go.AppParameter{
		Name: "replicas",
		Description: "replicas",
		Path: "groups.0.services.0.specs.replicas",
		Type: grpc_application_go.ParamDataType_INTEGER,
		Category:grpc_application_go.ParamCategory_BASIC,

	}

	toAdd := &grpc_application_go.AddAppDescriptorRequest{
		RequestId:            fmt.Sprintf("application-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		OrganizationId:       organizationID,
		Name:                 fmt.Sprintf("%s-app-manager-it-%d", name, ginkgo.GinkgoRandomSeed()),
		ConfigurationOptions: nil,
		EnvironmentVariables: nil,
		Labels:               nil,
		Rules:                nil,
		Groups:               []*grpc_application_go.ServiceGroup{group},
		Parameters:           []*grpc_application_go.AppParameter{parameter},
	}
	return toAdd
}

func CreateAppDescriptor(name string, organizationID string, appClient grpc_application_go.ApplicationsClient) * grpc_application_go.AppDescriptor{
	toAdd := GetAddAppDescriptorRequest(name, organizationID)
	desc, err := appClient.AddAppDescriptor(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	return desc
}

func CreateAppDescriptorWithParameters (name string, organizationID string, appClient grpc_application_go.ApplicationsClient) * grpc_application_go.AppDescriptor{
	toAdd := GetAddAppDescriptorWithParametersRequest(name, organizationID)
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
	var clusterClient grpc_infrastructure_go.ClustersClient
	var deviceClient grpc_device_go.DevicesClient
	var smConn * grpc.ClientConn
	var conductorConn * grpc.ClientConn
	var client grpc_application_manager_go.ApplicationManagerClient

	// Target organization.
	var targetOrganization *grpc_organization_go.Organization
	var targetAppDescriptor *grpc_application_go.AppDescriptor

	var deviceGroupNames  []string
	var deviceGroupIds []string

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		smConn = utils.GetConnection(systemModelAddress)
		orgClient = grpc_organization_go.NewOrganizationsClient(smConn)
		appClient = grpc_application_go.NewApplicationsClient(smConn)
		conductorConn = utils.GetConnection(conductorAddress)
		conductorClient = grpc_conductor_go.NewConductorClient(conductorConn)
		clusterClient = grpc_infrastructure_go.NewClustersClient(smConn)
		deviceClient = grpc_device_go.NewDevicesClient(smConn)

		test.LaunchServer(server, listener)

		// Register the service
		manager := NewManager(appClient, conductorClient, clusterClient, deviceClient)
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

	ginkgo.Context("App decriptors and instances", func(){
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

		ginkgo.FIt("Should be able to add a descriptor with parameters", func(){
			// add the descriptor with params
			// TODO: Fill the default value
			added, err := client.AddAppDescriptor(context.Background(),
				GetAddAppDescriptorWithParametersRequest("Descriptor with parameter", targetOrganization.OrganizationId))
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added.AppDescriptorId).ShouldNot(gomega.BeEmpty())

		})
		ginkgo.It("Should be able to deploy a instance of a descriptor withs params", func() {
			// add the descriptor with params
			added, err := client.AddAppDescriptor(context.Background(),
				GetAddAppDescriptorWithParametersRequest("Descriptor with parameter", targetOrganization.OrganizationId))
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added.AppDescriptorId).ShouldNot(gomega.BeEmpty())

			// deploy it
			deployRequest := &grpc_application_manager_go.DeployRequest{
				OrganizationId:       targetAppDescriptor.OrganizationId,
				AppDescriptorId:      added.AppDescriptorId,
				Name:                 "test-deploy-app-manager-with-params",
				Parameters:           &grpc_application_go.InstanceParameterList {
					Parameters: []*grpc_application_go.InstanceParameter{{ParameterName:"replicas", Value:"2"}},
				},
			}
			// TODO: waiting to conductor to be updated to check the deploy
			response, err := client.Deploy(context.Background(), deployRequest)
			if err != nil {
				fmt.Println(conversions.ToDerror(err).DebugReport())
			}
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(response.AppInstanceId).ShouldNot(gomega.BeEmpty())

		})

	})

	ginkgo.Context("Devices", func(){
		ginkgo.BeforeEach(func() {
			targetOrganization = CreateOrganization("app-manager-it(device)", orgClient)
			deviceGroupNames = []string{"dg1", "dg2", "dg3"}
			deviceGroupIds = make([]string, 0)

			for _, dg := range deviceGroupNames {
				// create deviceGroup
				added, err := deviceClient.AddDeviceGroup(context.Background(), &grpc_device_go.AddDeviceGroupRequest{
					RequestId: uuid.New().String(),
					OrganizationId: targetOrganization.OrganizationId,
					Name: dg,
					Labels: map[string]string{"l1": "v1"},
				})
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(added).NotTo(gomega.BeNil())

				deviceGroupIds = append(deviceGroupIds, added.DeviceGroupId)

			}
		})

		// -- RetrieveTargetApplications
		ginkgo.It("Should be able to retrieve target applications", func(){


			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())


			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: targetOrganization.OrganizationId,
				DeviceGroupName: deviceGroupNames[0],
				DeviceGroupId: deviceGroupIds[0],
				MatchLabels: map[string]string{"l1":"v1"},
			}
			// RetrieveTargetApplications
			request, err := client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(request).NotTo(gomega.BeNil())
			gomega.Expect(len(request.Applications)).Should(gomega.Equal(1))

		})
		ginkgo.It("Should be able to retrieve target applications without labels filering", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: targetOrganization.OrganizationId,
				DeviceGroupName: deviceGroupNames[0],
				DeviceGroupId: deviceGroupIds[0],
				MatchLabels: map[string]string{},
			}
			// RetrieveTargetApplications
			request, err := client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(request).NotTo(gomega.BeNil())
			gomega.Expect(len(request.Applications)).Should(gomega.Equal(1))

		})
		ginkgo.It("Should not be able to retrieve target applications of a non existing organization", func(){

			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: uuid.New().String(),
				DeviceGroupName: uuid.New().String(),
				DeviceGroupId:uuid.New().String(),
				MatchLabels: map[string]string{"l1":"v1"},
			}
			// RetrieveTargetApplications
			_, err := client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).NotTo(gomega.Succeed())

		})
		ginkgo.It("Should be able to retrieve an empty list (no match deviceGroupId)", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, []string{deviceGroupNames[2]},
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: targetOrganization.OrganizationId,
				DeviceGroupName: deviceGroupNames[1],
				DeviceGroupId: deviceGroupIds[1],
				MatchLabels: map[string]string{"l1":"v1"},
			}
			// RetrieveTargetApplications
			request, err := client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(request).NotTo(gomega.BeNil())
			gomega.Expect(len(request.Applications)).Should(gomega.Equal(0))

		})
		ginkgo.It("Should be able to retrieve an empty list (no match labels)", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: targetOrganization.OrganizationId,
				DeviceGroupName: deviceGroupNames[0],
				DeviceGroupId: deviceGroupIds[0],
				MatchLabels: map[string]string{"l1":"v2"},
			}
			// RetrieveTargetApplications
			request, err := client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(request).NotTo(gomega.BeNil())
			gomega.Expect(len(request.Applications)).Should(gomega.Equal(0))

		})
		ginkgo.It("Should be able return an error when the deviceGroupID does not match the deviceGroupName", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			filters := &grpc_application_manager_go.ApplicationFilter{
				OrganizationId: targetOrganization.OrganizationId,
				DeviceGroupName: deviceGroupNames[0],
				DeviceGroupId: deviceGroupIds[1],
				MatchLabels: map[string]string{"l1":"v2"},
			}
			// RetrieveTargetApplications
			_, err = client.RetrieveTargetApplications(context.Background(), filters)
			gomega.Expect(err).NotTo(gomega.Succeed())

		})

		// -- RetrieveEndpoints(ctx context.Context, filter *grpc_application_manager_go.RetrieveEndpointsRequest)
		ginkgo.It("Should be able to retrieve endpoints", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			list, err := appClient.AddServiceGroupInstances(context.Background(), &grpc_application_go.AddServiceGroupInstancesRequest{
				OrganizationId: targetOrganization.OrganizationId,
				AppDescriptorId: added.AppDescriptorId,
				AppInstanceId: instAdded.AppInstanceId,
				ServiceGroupId: added.Groups[0].ServiceGroupId,
				NumInstances: 1,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(list).NotTo(gomega.BeNil())

			// add cluster
			cluster, err := clusterClient.AddCluster(context.Background(), &grpc_infrastructure_go.AddClusterRequest{
				RequestId: uuid.New().String(),
				OrganizationId: targetOrganization.OrganizationId,
				Name:"test cluster",
				Hostname: "URL_daisho",
				ControlPlaneHostname: "ControlPlaneHostname",
				Labels: map[string]string {"label1": "eti1"},
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(cluster).NotTo(gomega.BeNil())


			ei := &grpc_application_go.EndpointInstance{
				Type:                 0,
				Fqdn:                 "target.instance",
			}


			// update service status
			success, err := appClient.UpdateServiceStatus(context.Background(), &grpc_application_go.UpdateServiceStatusRequest{
				OrganizationId: targetOrganization.OrganizationId,
				AppInstanceId: instAdded.AppInstanceId,
				ServiceGroupInstanceId: list.ServiceGroupInstances[0].ServiceGroupInstanceId,
				ServiceInstanceId: list.ServiceGroupInstances[0].ServiceInstances[0].ServiceInstanceId,
				Status: grpc_application_go.ServiceStatus_SERVICE_RUNNING,
				Endpoints: []*grpc_application_go.EndpointInstance{ei},
				DeployedOnClusterId: cluster.ClusterId,
			})
			gomega.Expect(success).NotTo(gomega.BeNil())
			gomega.Expect(err).To(gomega.Succeed())

			filter := &grpc_application_manager_go.RetrieveEndpointsRequest{
				OrganizationId: targetOrganization.OrganizationId,
				AppInstanceId: instAdded.AppInstanceId,
			}
			endPoints, err := client.RetrieveEndpoints(context.Background(), filter)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(endPoints).NotTo(gomega.BeNil())

		})
		ginkgo.It("Should be able to retrieve an empty endpoints (service is waiting)", func(){

			// create descriptor
			descriptor := utils.CreateAddAppDescriptorRequest(targetOrganization.OrganizationId, deviceGroupNames,
				map[string]string{"l1":"v1", "l2":"v2"})
			added, err := appClient.AddAppDescriptor(context.Background(), descriptor)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added).NotTo(gomega.BeNil())

			// add instance
			instance := utils.CreateTestAppInstanceRequest(targetOrganization.OrganizationId, added.AppDescriptorId)
			instAdded, err := appClient.AddAppInstance(context.Background(), instance)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(instAdded).NotTo(gomega.BeNil())

			filter := &grpc_application_manager_go.RetrieveEndpointsRequest{
				OrganizationId: targetOrganization.OrganizationId,
				AppInstanceId: instAdded.AppInstanceId,
			}
			endPoints, err := client.RetrieveEndpoints(context.Background(), filter)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(endPoints).NotTo(gomega.BeNil())
			gomega.Expect(endPoints.ClusterEndpoints).Should(gomega.BeEmpty())

		})

	})
})