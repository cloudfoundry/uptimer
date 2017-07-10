package cfCmdGenerator_test

import (
	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfHelper", func() {
	var (
		helper CfCmdGenerator
	)

	BeforeEach(func() {
		helper = New()
	})

	Describe("Api", func() {
		It("Generates the correct command skipping ssl validation", func() {
			expectedCmd := exec.Command("cf", "api", "api.example.com", "--skip-ssl-validation")

			cmd := helper.Api("api.example.com")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "auth", "user44", "pass55")

			cmd := helper.Auth("user44", "pass55")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-org", "someOrg")

			cmd := helper.CreateOrg("someOrg")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-space", "someSpace", "-o", "someOrg")

			cmd := helper.CreateSpace("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "target", "-o", "someOrg", "-s", "someSpace")

			cmd := helper.Target("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Push", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "push", "appName", "-p", "path/to/app")

			cmd := helper.Push("appName", "path/to/app")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-org", "orgName", "-f")

			cmd := helper.DeleteOrg("orgName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})
})
