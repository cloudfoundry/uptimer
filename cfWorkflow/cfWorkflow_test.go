package cfWorkflow_test

import (
	"context"
	"time"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	. "github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfWorkflow", func() {
	var (
		cfc     *config.Cf
		ccg     cfCmdGenerator.CfCmdGenerator
		org     string
		space   string
		quota   string
		appName string
		appPath string

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
			TCPPort:       1026,
		}

		ccg = cfCmdGenerator.New("/cfhome", false)
		org = "someOrg"
		space = "someSpace"
		quota = "someQuota"
		appName = "doraApp"
		appPath = "this/is/an/app/path"
	})

	JustBeforeEach(func() {
		cw = New(cfc, org, space, quota, appName, appPath)
	})

	Describe("Org", func() {
		It("returns the correct org", func() {
			Expect(cw.Org()).To(Equal(org))
		})
	})

	Describe("Space", func() {
		It("returns the correct space", func() {
			Expect(cw.Space()).To(Equal(space))
		})
	})

	Describe("Quota", func() {
		It("returns the correct quota", func() {
			Expect(cw.Quota()).To(Equal(quota))
		})
	})

	Describe("AppUrl", func() {
		It("returns the correct app url", func() {
			Expect(cw.AppUrl()).To(Equal("https://doraApp.app.jigglypuff.cf-app.com"))
		})
	})

	Describe("TCPDomain", func() {
		It("returns the correct tcp domain", func() {
			Expect(cw.TCPDomain()).To(Equal("tcp.jigglypuff.cf-app.com"))
		})
	})
	Describe("TCPPort", func() {
		It("returns the correct tcp port", func() {
			Expect(cw.TCPPort()).To(Equal(1026))
		})
	})
	Describe("Push", func() {
		It("returns a series of commands to push an app with exactly 2 instances", func() {
			cmds := cw.Push(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "this/is/an/app/path", 2, false),
				},
			))
		})

		Context("when the UseSingleAppInstance flag is used", func() {
			BeforeEach(func() {
				cfc.UseSingleAppInstance = true
			})

			It("returns a series of commands to push a single instance app", func() {
				cmds := cw.Push(ccg)

				Expect(cmds).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.Target("someOrg", "someSpace"),
						ccg.Push("doraApp", "this/is/an/app/path", 1, false),
					},
				))
			})
		})
	})
	Describe("PushNoRoute", func() {
		It("calls the push command with the noStart flag as true", func() {
			cmds := cw.PushNoRoute(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "this/is/an/app/path", 2, true),
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
		When("Quotas are disabled", func() {
			BeforeEach(func() { quota = "" })

			It("returns a series of commands to create a new org and space without setting quotas", func() {
				cmds := cw.Setup(ccg)

				Expect(cmds).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.CreateOrg("someOrg"),
						ccg.CreateSpace("someOrg", "someSpace"),
					},
				))
			})
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
		When("Quotas are disabled", func() {
			BeforeEach(func() { quota = "" })
			It("returns a set of commands to delete an org without deleting the quota", func() {
				cmds := cw.TearDown(ccg)

				Expect(cmds).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.DeleteOrg("someOrg"),
						ccg.LogOut(),
					},
				))
			})
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

	Describe("MapTCPRoute", func() {
		It("returns a set of commands to map a route to a tcp app", func() {
			cmds := cw.MapTCPRoute(ccg)

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.MapRoute("doraApp", "tcp.jigglypuff.cf-app.com", 1026),
				},
			))
		})
	})
	Describe("MapSyslogRoute", func() {
		It("returns a set of commands to map a route to a syslog sink app", func() {
			cmds := cw.MapSyslogRoute(ccg)

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
