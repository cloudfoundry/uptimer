package cfAppGenerator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfAppGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfAppGenerator Suite")
}
