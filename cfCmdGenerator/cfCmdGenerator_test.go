package cfCmdGenerator_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfCmdGenerator", func() {
	var (
		cfHome                string
		useBuildpackDetection bool

		generator CfCmdGenerator
	)

	BeforeEach(func() {
		cfHome = "/on/the/range"
		useBuildpackDetection = false

	})

	JustBeforeEach(func() {
		generator = New(cfHome, useBuildpackDetection)
	})

	Describe("Api", func() {
		It("Generates the correct command skipping ssl validation", func() {
			expectedCmd := exec.Command("cf", "api", "api.example.com", "--skip-ssl-validation")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.Api("api.example.com")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "auth", "user44", "pass55")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.Auth("user44", "pass55")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-org", "someOrg")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.CreateOrg("someOrg")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateQuota", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command(
				"cf", "create-quota", "someQuota",
				"-m", "10G",
				"-i", "1G",
				"-r", "1000",
				"-s", "100",
				"--reserved-route-ports", "1",
			)
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.CreateQuota("someQuota")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("SetQuota", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "set-quota", "someQuota", "someOrg")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.SetQuota("someQuota", "someOrg")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-space", "someSpace", "-o", "someOrg")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.CreateSpace("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "target", "-o", "someOrg", "-s", "someSpace")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.Target("someOrg", "someSpace")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Push", func() {

		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "push", "appName", "-f", "manifest.yml", "-i", "3", "-b", "go_buildpack")
			expectedCmd.Env = append(expectedCmd.Env, fmt.Sprintf("CF_HOME=%s", cfHome))
			expectedCmd.Env = append(expectedCmd.Env, "CF_STAGING_TIMEOUT=5")
			expectedCmd.Dir = "path/to/app"

			cmd := generator.Push("appName", "path/to/app", 3)

			Expect(cmd).To(Equal(expectedCmd))
		})

		Context("given buildpack detection is turned on", func() {

			BeforeEach(func() {
				useBuildpackDetection = true
			})

			It("should not specify the go_buildpack", func() {
				expectedCmd := exec.Command("cf", "push", "appName", "-f", "manifest.yml", "-i", "3")
				expectedCmd.Env = append(expectedCmd.Env, fmt.Sprintf("CF_HOME=%s", cfHome))
				expectedCmd.Env = append(expectedCmd.Env, "CF_STAGING_TIMEOUT=5")
				expectedCmd.Dir = "path/to/app"

				cmd := generator.Push("appName", "path/to/app", 3)

				Expect(cmd).To(Equal(expectedCmd))
			})
		})
	})

	Describe("Delete", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete", "appName", "-f", "-r")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.Delete("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-org", "orgName", "-f")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.DeleteOrg("orgName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("DeleteQuota", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-quota", "quotaName", "-f")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.DeleteQuota("quotaName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("LogOut", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logout")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.LogOut()

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("RecentLogs", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logs", "appName", "--recent")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.RecentLogs("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("StreamLogs", func() {
		It("Generates the correct command", func() {
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			defer cancelFunc()
			expectedCmd := exec.CommandContext(ctx, "cf", "logs", "appName")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.StreamLogs(ctx, "appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("MapRoute", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "map-route", "appName", "tcp.example.com", "--port", "1025")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.MapRoute("appName", "tcp.example.com", 1025)

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("CreateUserProvidedService", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-user-provided-service", "serviceName", "-l", "syslog://tcp.example.com:54321")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.CreateUserProvidedService("serviceName", "syslog://tcp.example.com:54321")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("BindService", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "bind-service", "appName", "serviceName")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.BindService("appName", "serviceName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})

	Describe("Restage", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "restage", "appName")
			expectedCmd.Env = []string{fmt.Sprintf("CF_HOME=%s", cfHome)}

			cmd := generator.Restage("appName")

			Expect(cmd).To(Equal(expectedCmd))
		})
	})
})
