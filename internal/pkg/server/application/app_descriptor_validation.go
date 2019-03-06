package application

import (
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Application Descriptor Validations", func() {

	ginkgo.Context("Valid app descriptors", func(){
		ginkgo.It("should pass the validation", func(){
			appDescriptor := utils.CreateFullAppDescriptor()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (repeated group)", func(){
			appDescriptor := utils.CreateAppDescriptorWithRepeatedGroup()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (repeated service)", func(){
			appDescriptor := utils.CreateAppDescriptorWithRepeatedService()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong group in rule)", func(){
			appDescriptor := utils.CreateAppDescriptorWrongGroupInRule()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong deploy after service)", func(){
			appDescriptor := utils.CreateAppDescriptorWrongDeployAfter()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong group deploy specs)", func(){
			appDescriptor := utils.CreateAppDescriptorWrongGroupDeploySpecs()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (service to service rule)", func(){
			appDescriptor := utils.CreateAppDescriptorServiceToService()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong environment variables)", func(){
			appDescriptor := utils.CreateAppDescriptorWrongEnvironmetVariables()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
	})
})