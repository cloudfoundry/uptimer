package cmdRunner_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCmdRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CmdRunner Suite")
}
