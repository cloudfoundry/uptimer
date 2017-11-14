package cfWorkflow_test

import (
	"context"
	"time"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	. "github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfWorkflow", func() {
	var (
		cfc        *config.Cf
		ccg        cfCmdGenerator.CfCmdGenerator
		org        string
		space      string
		quota      string
		appName    string
		appPath    string
		appCommand string

		cw CfWorkflow
	)

	BeforeEach(func() {
		cfc = &config.Cf{
			API:           "jigglypuff.cf-app.com",
			AppDomain:     "app.jigglypuff.cf-app.com",
			AdminUser:     "pika",
			AdminPassword: "chu",

			TCPDomain:     "tcp.jigglypuff.cf-app.com",
			AvailablePort: 1025,
		}
		ccg = cfCmdGenerator.New("/cfhome")
		org = "someOrg"
		space = "someSpace"
		quota = "someQuota"
		appName = "doraApp"
		appPath = "this/is/an/app/path"
		appCommand = "./app-command"

		cw = New(cfc, org, space, quota, appName, appPath, appCommand)
	})

	It("has the correct app url", func() {
		Expect(cw.AppUrl()).To(Equal("https://doraApp.app.jigglypuff.cf-app.com"))
	})

	Describe("Push", func() {
		It("returns a series of commands to push an app", func() {
			cmds := cw.Push(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "app.jigglypuff.cf-app.com", "this/is/an/app/path", "./app-command"),
				},
			))
		})
	})

	Describe("Delete", func() {
		It("returns a series of commands to delete an app", func() {
			cmds := cw.Delete(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Delete("doraApp"),
				},
			))
		})
	})

	Describe("Setup", func() {
		It("returns a series of commands to create a new org and space", func() {
			cmds := cw.Setup(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.CreateOrg("someOrg"),
					ccg.CreateSpace("someOrg", "someSpace"),
					ccg.CreateQuota("someQuota"),
					ccg.SetQuota("someOrg", "someQuota"),
				},
			))
		})
	})

	Describe("TearDown", func() {
		It("returns a set of commands to delete an org", func() {
			cmds := cw.TearDown(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.DeleteOrg("someOrg"),
					ccg.DeleteQuota("someQuota"),
					ccg.LogOut(),
				},
			))
		})
	})

	Describe("RecentLogs", func() {
		It("returns a set of commands to get recent logs for an app", func() {
			cmds := cw.RecentLogs(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.RecentLogs("doraApp"),
				},
			))
		})
	})

	Describe("StreamLogs", func() {
		It("returns a set of commands to stream logs for an app", func() {
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			defer cancelFunc()
			cmds := cw.StreamLogs(ctx, ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.StreamLogs(ctx, "doraApp"),
				},
			))
		})
	})

	Describe("MapRoute", func() {
		It("returns a set of commands to map a route to a syslog sink app", func() {
			cmds := cw.MapRoute(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.MapRoute("doraApp", "tcp.jigglypuff.cf-app.com", 1025),
				},
			))
		})
	})

	Describe("CreateAndBindSyslogDrainService", func() {
		It("Creates and binds a user-provided syslog drain service to an app and restages the app", func() {
			cmds := cw.CreateAndBindSyslogDrainService(ccg, "syslogUPS")

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.CreateUserProvidedService("syslogUPS", "syslog://tcp.jigglypuff.cf-app.com:1025"),
					ccg.BindService("doraApp", "syslogUPS"),
					ccg.Restage("doraApp"),
				},
			))
		})
	})
})
