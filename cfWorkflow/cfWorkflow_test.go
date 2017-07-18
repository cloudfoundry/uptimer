package cfWorkflow_test

import (
	"context"
	"time"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	"github.com/cloudfoundry/uptimer/cfCmdGenerator/cfCmdGeneratorfakes"
	. "github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
	"github.com/cloudfoundry/uptimer/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfWorkflow", func() {
	var (
		cfc               *config.CfConfig
		ccg               cfCmdGenerator.CfCmdGenerator
		appPath           string
		guidMatchingRegex string

		cw CfWorkflow
	)

	BeforeEach(func() {
		cfc = &config.CfConfig{
			API:           "jigglypuff.cf-app.com",
			AppDomain:     "app.jigglypuff.cf-app.com",
			AdminUser:     "pika",
			AdminPassword: "chu",
			Org:           "someOrg",
			Space:         "someSpace",
			AppName:       "doraApp",
		}
		ccg = cfCmdGenerator.New("/cfhome")
		appPath = "this/is/an/app/path"
		guidMatchingRegex = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}"

		cw = New(cfc, ccg, appPath)
	})

	It("has the correct app url", func() {
		Expect(cw.AppUrl()).To(Equal("https://doraApp.app.jigglypuff.cf-app.com"))
	})

	Describe("Push", func() {
		It("returns a series of commands to push an app", func() {
			cmds := cw.Push()

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "this/is/an/app/path"),
				},
			))
		})

		Context("when the cf name data isn't provided in the config", func() {
			var (
				fakeCfCmdGenerator *cfCmdGeneratorfakes.FakeCfCmdGenerator
			)

			BeforeEach(func() {
				cfc = &config.CfConfig{
					API:           "jigglypuff.cf-app.com",
					AdminUser:     "pika",
					AdminPassword: "chu",
					Org:           "someOrg",
					Space:         "someSpace",
				}
				fakeCfCmdGenerator = &cfCmdGeneratorfakes.FakeCfCmdGenerator{
					ApiStub:         ccg.Api,
					AuthStub:        ccg.Auth,
					CreateOrgStub:   ccg.CreateOrg,
					CreateSpaceStub: ccg.CreateSpace,
					TargetStub:      ccg.Target,
					PushStub:        ccg.Push,
				}

				cw = New(cfc, fakeCfCmdGenerator, appPath)
			})

			It("generates a unique appName", func() {
				cmds := cw.Push()

				generatedAppName, _ := fakeCfCmdGenerator.PushArgsForCall(0)
				Expect(generatedAppName).To(MatchRegexp("^uptimer-app-%s$", guidMatchingRegex))
				Expect(cmds).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.Target("someOrg", "someSpace"),
						ccg.Push(generatedAppName, "this/is/an/app/path"),
					},
				))
			})
		})
	})

	Describe("Delete", func() {
		It("returns a series of commands to delete an app", func() {
			cmds := cw.Delete()

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
			cmds := cw.Setup()

			Expect(cmds).To(Equal(
				[]cmdStartWaiter.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.CreateOrg("someOrg"),
					ccg.CreateSpace("someOrg", "someSpace"),
				},
			))
		})

		Context("when the cf name data isn't provided in the config", func() {
			var (
				fakeCfCmdGenerator *cfCmdGeneratorfakes.FakeCfCmdGenerator
			)

			BeforeEach(func() {
				cfc = &config.CfConfig{
					API:           "jigglypuff.cf-app.com",
					AdminUser:     "pika",
					AdminPassword: "chu",
				}
				fakeCfCmdGenerator = &cfCmdGeneratorfakes.FakeCfCmdGenerator{
					ApiStub:         ccg.Api,
					AuthStub:        ccg.Auth,
					CreateOrgStub:   ccg.CreateOrg,
					CreateSpaceStub: ccg.CreateSpace,
					TargetStub:      ccg.Target,
					PushStub:        ccg.Push,
				}

				cw = New(cfc, fakeCfCmdGenerator, appPath)
			})

			It("generates identifiable but non-colliding names", func() {
				cmds := cw.Setup()

				generatedOrg := fakeCfCmdGenerator.CreateOrgArgsForCall(0)
				_, generatedSpace := fakeCfCmdGenerator.CreateSpaceArgsForCall(0)

				Expect(generatedOrg).To(MatchRegexp("^uptimer-org-%s$", guidMatchingRegex))
				Expect(generatedSpace).To(MatchRegexp("^uptimer-space-%s$", guidMatchingRegex))

				Expect(cmds).To(Equal(
					[]cmdStartWaiter.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.CreateOrg(generatedOrg),
						ccg.CreateSpace(generatedOrg, generatedSpace),
					},
				))
			})
		})
	})

	Describe("TearDown", func() {
		It("returns a set of commands to delete an org", func() {
			cmds := cw.TearDown()

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

	Describe("RecentLogs", func() {
		It("returns a set of commands to get recent logs for an app", func() {
			cmds := cw.RecentLogs()

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
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			cmds := cw.StreamLogs(ctx)

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
})
