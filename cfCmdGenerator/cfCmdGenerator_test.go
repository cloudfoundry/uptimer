package cfCmdGenerator_test

import (
	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfCmdGenerator", func() {
	var (
		generator CfCmdGenerator
	)

	BeforeEach(func() {
		generator = New()
	})

	Describe("Api", func() {
		It("Generates the correct command skipping ssl validation", func() {
			expectedCmd := exec.Command("cf", "api", "api.example.com", "--skip-ssl-validation")

			cmd := generator.Api("api.example.com")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "auth", "user44", "pass55")

			cmd := generator.Auth("user44", "pass55")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-org", "someOrg")

			cmd := generator.CreateOrg("someOrg")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-space", "someSpace", "-o", "someOrg")

			cmd := generator.CreateSpace("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "target", "-o", "someOrg", "-s", "someSpace")

			cmd := generator.Target("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Push", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "push", "appName", "-p", "path/to/app", "-b", "binary_buildpack", "-c", "./app")

			cmd := generator.Push("appName", "path/to/app")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-org", "orgName", "-f")

			cmd := generator.DeleteOrg("orgName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("LogOut", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logout")

			cmd := generator.LogOut()

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("RecentLogs", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logs", "appName", "--recent")

			cmd := generator.RecentLogs("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})
})
