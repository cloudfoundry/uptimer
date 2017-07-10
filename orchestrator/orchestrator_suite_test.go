package orchestrator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUptimer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orchestrator Suite")
}
