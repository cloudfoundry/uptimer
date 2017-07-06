package cfCmdGenerator_test

import (
	. "github.com/cloudfoundry/uptimer/cfCmdGenerator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfHelper", func() {
	var (
		helper CfCmdGenerator
	)

	BeforeEach(func() {
		helper = New()
	})

	Describe("Api", func() {
		It("Generates the correct command not skipping ssl validation", func() {
			cmd := helper.Api("api.example.com", false)

			Expect(cmd.Args).To(Equal([]string{"cf", "api", "api.example.com"}))
		})

		It("Generates the correct command skipping ssl validation", func() {
			cmd := helper.Api("api.example.com", true)

			Expect(cmd.Args).To(Equal([]string{"cf", "api", "api.example.com", "--skip-ssl-validation"}))
		})
	})

	Describe("Auth", func() {
		It("Generates the correct command", func() {
			cmd := helper.Auth("user44", "pass55")

			Expect(cmd.Args).To(Equal([]string{"cf", "auth", "user44", "pass55"}))
		})
	})

	Describe("CreateOrg", func() {
		It("Generates the correct command", func() {
			cmd := helper.CreateOrg("someOrg")

			Expect(cmd.Args).To(Equal([]string{"cf", "create-org", "someOrg"}))
		})
	})

	Describe("CreateSpace", func() {
		It("Generates the correct command", func() {
			cmd := helper.CreateSpace("someOrg", "someSpace")

			Expect(cmd.Args).To(Equal([]string{"cf", "create-space", "someOrg", "someSpace"}))
		})
	})

	Describe("Target", func() {
		It("Generates the correct command", func() {
			cmd := helper.Target("someOrg", "someSpace")

			Expect(cmd.Args).To(Equal([]string{"cf", "target", "-o", "someOrg", "-s", "someSpace"}))
		})
	})

	Describe("Push", func() {
		It("Generates the correct command", func() {
			cmd := helper.Push("appName", "path/to/app")

			Expect(cmd.Args).To(Equal([]string{"cf", "push", "appName", "-p", "path/to/app"}))
		})
	})

	Describe("DeleteOrg", func() {
		It("Generates the correct command", func() {
			cmd := helper.DeleteOrg("orgName")

			Expect(cmd.Args).To(Equal([]string{"cf", "delete-org", "orgName", "-f"}))
		})
	})
})
