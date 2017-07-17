package cfCmdGenerator_test

import (
	"fmt"
	"os/exec"

	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfCmdGenerator", func() {
	var (
		cfHome string

		generator CfCmdGenerator
	)

	BeforeEach(func() {
		cfHome = "/on/the/range"

		generator = New(cfHome)
	})

	Describe("Api", func() {
		It("Generates the correct command skipping ssl validation", func() {
			rawCmd := exec.Command("cf", "api", "api.example.com", "--skip-ssl-validation")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.Api("api.example.com")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "auth", "user44", "pass55")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.Auth("user44", "pass55")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "create-org", "someOrg")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.CreateOrg("someOrg")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "create-space", "someSpace", "-o", "someOrg")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.CreateSpace("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "target", "-o", "someOrg", "-s", "someSpace")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.Target("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Push", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "push", "appName", "-p", "path/to/app", "-b", "binary_buildpack", "-c", "./app")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.Push("appName", "path/to/app")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Delete", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "delete", "appName", "-f", "-r")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.Delete("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "delete-org", "orgName", "-f")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.DeleteOrg("orgName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("LogOut", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "logout")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.LogOut()

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("RecentLogs", func() {
		It("Generates the correct command", func() {
			rawCmd := exec.Command("cf", "logs", "appName", "--recent")
			rawCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}
			expectedCmd := cmdStartWaiter.New(rawCmd)

			cmd := generator.RecentLogs("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})
})
