package cfCmdGenerator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfCmdGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfCmdGenerator Suite")
}
