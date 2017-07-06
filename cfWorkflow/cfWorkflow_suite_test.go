package cfWorkflow_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfWorkflow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfWorkflow Suite")
}
