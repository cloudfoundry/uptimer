package cfCmdGenerator_test

import (
	"context"
	"os/exec"
	"time"

	"github.com/cloudfoundry/uptimer/cmdStartWaiter"

	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfCmdGenerator", func() {
	const (
		cfHome       = "/on/the/range"
		cfHomeEnvVar = "CF_HOME=" + cfHome
	)

	var (
		useBuildpackDetection bool
		generator             CfCmdGenerator
	)

	BeforeEach(func() {
		useBuildpackDetection = false

	})

	JustBeforeEach(func() {
		generator = New(cfHome, useBuildpackDetection)
	})

	Describe("Api", func() {
		It("Generates the correct command skipping ssl validation", func() {
			expectedCmd := exec.Command("cf", "api", "api.example.com", "--skip-ssl-validation")
			cmd := generator.Api("api.example.com")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "auth", "user44", "pass55")
			cmd := generator.Auth("user44", "pass55")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-org", "someOrg")
			cmd := generator.CreateOrg("someOrg")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
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
			cmd := generator.CreateQuota("someQuota")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("SetQuota", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "set-quota", "someQuota", "someOrg")
			cmd := generator.SetQuota("someQuota", "someOrg")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-space", "someSpace", "-o", "someOrg")
			cmd := generator.CreateSpace("someOrg", "someSpace")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "target", "-o", "someOrg", "-s", "someSpace")

			cmd := generator.Target("someOrg", "someSpace")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("Push", func() {

		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "push", "appName", "-f", "manifest.yml", "-i", "3", "-b", "go_buildpack")
			expectedCmd.Dir = "path/to/app"

			cmd := generator.Push("appName", "path/to/app", 3, false)
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar, "CF_STAGING_TIMEOUT=5")
		})

		Context("given buildpack detection is turned on", func() {

			BeforeEach(func() {
				useBuildpackDetection = true
			})

			It("should not specify the go_buildpack", func() {
				expectedCmd := exec.Command("cf", "push", "appName", "-f", "manifest.yml", "-i", "3")
				expectedCmd.Dir = "path/to/app"
				cmd := generator.Push("appName", "path/to/app", 3, false)
				expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar, "CF_STAGING_TIMEOUT=5")
			})
		})
		Context("when the noStart flag is true", func() {

			It("append the right flag to the cf cli", func() {
				expectedCmd := exec.Command("cf", "push", "appName", "-f", "manifest.yml", "-i", "3", "-b", "go_buildpack", "--no-route")
				expectedCmd.Dir = "path/to/app"
				cmd := generator.Push("appName", "path/to/app", 3, true)
				expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar, "CF_STAGING_TIMEOUT=5")
			})
		})
	})

	Describe("Delete", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete", "appName", "-f", "-r")
			cmd := generator.Delete("appName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-org", "orgName", "-f")
			cmd := generator.DeleteOrg("orgName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("DeleteQuota", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "delete-quota", "quotaName", "-f")
			cmd := generator.DeleteQuota("quotaName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("LogOut", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logout")
			cmd := generator.LogOut()
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("RecentLogs", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "logs", "appName", "--recent")
			cmd := generator.RecentLogs("appName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("StreamLogs", func() {
		It("Generates the correct command", func() {
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			defer cancelFunc()
			expectedCmd := exec.CommandContext(ctx, "cf", "logs", "appName")
			cmd := generator.StreamLogs(ctx, "appName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("MapRoute", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "map-route", "appName", "tcp.example.com", "--port", "1025")
			cmd := generator.MapRoute("appName", "tcp.example.com", 1025)
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("CreateUserProvidedService", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "create-user-provided-service", "serviceName", "-l", "syslog://tcp.example.com:54321")
			cmd := generator.CreateUserProvidedService("serviceName", "syslog://tcp.example.com:54321")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("BindService", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "bind-service", "appName", "serviceName")
			cmd := generator.BindService("appName", "serviceName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})

	Describe("Restage", func() {
		It("Generates the correct command", func() {
			expectedCmd := exec.Command("cf", "restage", "appName")
			cmd := generator.Restage("appName")
			expectCommandToBeEquivalent(cmd, expectedCmd, cfHomeEnvVar)
		})
	})
})

func expectCommandToBeEquivalent(cmd cmdStartWaiter.CmdStartWaiter, expectedCmd *exec.Cmd, envIncludes ...string) {
	ExpectWithOffset(1, cmd).To(BeAssignableToTypeOf(expectedCmd))
	rawCmd := cmd.(*exec.Cmd)

	ExpectWithOffset(1, rawCmd.Path).To(Equal(expectedCmd.Path))
	ExpectWithOffset(1, rawCmd.Args).To(Equal(expectedCmd.Args))
	ExpectWithOffset(1, rawCmd.Dir).To(Equal(expectedCmd.Dir))
	ExpectWithOffset(1, rawCmd.Env).To(ContainElements(envIncludes))
}
