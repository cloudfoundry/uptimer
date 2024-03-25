package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var uptimerPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := Build("github.com/cloudfoundry/uptimer")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(CleanupBuildArtifacts)
	return []byte(path)
}, func(data []byte) {
	uptimerPath = string(data)
})
