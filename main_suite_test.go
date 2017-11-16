package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
	"os/exec"
)

func TestUptimer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var commandPath string

var _ = BeforeSuite(func() {
	executablePath, err := Build("github.com/cloudfoundry/uptimer")
	Expect(err).NotTo(HaveOccurred())
	commandPath = string(executablePath)
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})

func runCommand(args ...string) *Session {
	cmd := exec.Command(commandPath, args...)
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	<-session.Exited

	return session
}