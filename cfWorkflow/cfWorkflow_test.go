package cfWorkflow_test

import (
	"os/exec"

	"github.com/cloudfoundry/uptimer/cfCmdGenerator"
	. "github.com/cloudfoundry/uptimer/cfWorkflow"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfWorkflow", func() {
	var (
		ccg cfCmdGenerator.CfCmdGenerator

		cw CfWorkflow
	)

	BeforeEach(func() {
		ccg = cfCmdGenerator.New()

		cw = New(
			"jigglypuff.cf-app.com",
			"pika",
			"chu",
			"someOrg",
			"someSpace",
			"doraApp",
			"doraPath",
			true,
			ccg,
		)
	})

	Describe("Setup", func() {
		It("returns a series of commands to push an app to a new org and space", func() {
			cmds := cw.Setup()

			Expect(cmds).To(Equal(
				[]*exec.Cmd{
					ccg.Api("jigglypuff.cf-app.com", true),
					ccg.Auth("pika", "chu"),
					ccg.CreateOrg("someOrg"),
					ccg.CreateSpace("someOrg", "someSpace"),
					ccg.Target("someOrg", "someSpace"),
					ccg.Push("doraApp", "doraPath"),
				},
			))
		})
	})

	Describe("Cleanup", func() {
		It("returns a set of commands to delete an org", func() {
			cmds := cw.Cleanup()

			Expect(cmds).To(Equal(
				[]*exec.Cmd{
					ccg.DeleteOrg("someOrg"),
				},
			))
		})
	})
})
