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
		cfc     *config.CfConfig
		ccg     cfCmdGenerator.CfCmdGenerator
		org     string
		space   string
		quota   string
		appName string
		appPath string

		cw CfWorkflow
	)

	BeforeEach(func() {
		cfc = &config.CfConfig{
			API:           "jigglypuff.cf-app.com",
			AppDomain:     "app.jigglypuff.cf-app.com",
			AdminUser:     "pika",
			AdminPassword: "chu",
		}
		ccg = cfCmdGenerator.New("/cfhome")
		org = "someOrg"
		space = "someSpace"
		quota = "someQuota"
		appName = "doraApp"
		appPath = "this/is/an/app/path"

		cw = New(cfc, ccg, org, space, quota, appName, appPath)
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
					ccg.CreateQuota("someQuota"),
					ccg.SetQuota("someOrg", "someQuota"),
				},
			))
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
					ccg.DeleteQuota("someQuota"),
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
