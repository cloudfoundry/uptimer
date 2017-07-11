package cfWorkflow_test

import (
	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	. "github.com/cloudfoundry/uptimer/cfWorkflow"
	"github.com/cloudfoundry/uptimer/config"
	"github.com/cloudfoundry/uptimer/fakes"

	"github.com/cloudfoundry/uptimer/cmdRunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfWorkflow", func() {
	var (
		cfc *config.CfConfig
		ccg cfCmdGenerator.CfCmdGenerator

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
			AppPath:       "doraPath",
		}
		ccg = cfCmdGenerator.New()

		cw = New(cfc, ccg)
	})

	It("has the correct app url", func() {
		Expect(cw.AppUrl()).To(Equal("doraApp.app.jigglypuff.cf-app.com"))
	})

	Describe("Setup", func() {
		It("returns a series of commands to push an app to a new org and space", func() {
			cmds := cw.Setup()

			Expect(cmds).To(Equal(
				[]cmdRunner.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.CreateOrg("someOrg"),
					ccg.CreateSpace("someOrg", "someSpace"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "doraPath"),
				},
			))
		})

		Context("when the cf name data isn't provided in the config", func() {
			var (
				fakeCfCmdGenerator *fakes.FakeCfCmdGenerator
			)

			BeforeEach(func() {
				cfc = &config.CfConfig{
					API:           "jigglypuff.cf-app.com",
					AdminUser:     "pika",
					AdminPassword: "chu",
					AppPath:       "doraPath",
				}
				fakeCfCmdGenerator = &fakes.FakeCfCmdGenerator{
					ApiStub:         ccg.Api,
					AuthStub:        ccg.Auth,
					CreateOrgStub:   ccg.CreateOrg,
					CreateSpaceStub: ccg.CreateSpace,
					TargetStub:      ccg.Target,
					PushStub:        ccg.Push,
				}

				cw = New(cfc, fakeCfCmdGenerator)
			})

			It("generates identifiable but non-colliding names", func() {
				cmds := cw.Setup()

				generatedOrg := fakeCfCmdGenerator.CreateOrgArgsForCall(0)
				_, generatedSpace := fakeCfCmdGenerator.CreateSpaceArgsForCall(0)
				generatedAppName, _ := fakeCfCmdGenerator.PushArgsForCall(0)

				guidMatchingRegex := "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}"
				Expect(generatedOrg).To(MatchRegexp("^uptimer-org-%s$", guidMatchingRegex))
				Expect(generatedSpace).To(MatchRegexp("^uptimer-space-%s$", guidMatchingRegex))
				Expect(generatedAppName).To(MatchRegexp("^uptimer-app-%s$", guidMatchingRegex))

				Expect(cmds).To(Equal(
					[]cmdRunner.CmdStartWaiter{
						ccg.Api("jigglypuff.cf-app.com"),
						ccg.Auth("pika", "chu"),
						ccg.CreateOrg(generatedOrg),
						ccg.CreateSpace(generatedOrg, generatedSpace),
						ccg.Target(generatedOrg, generatedSpace),
						ccg.Push(generatedAppName, "doraPath"),
					},
				))
			})
		})
	})

	Describe("TearDown", func() {
		It("returns a set of commands to delete an org", func() {
			cmds := cw.TearDown()

			Expect(cmds).To(Equal(
				[]cmdRunner.CmdStartWaiter{
					ccg.Api("jigglypuff.cf-app.com"),
					ccg.Auth("pika", "chu"),
					ccg.DeleteOrg("someOrg"),
					ccg.LogOut(),
				},
			))
		})
	})
})
