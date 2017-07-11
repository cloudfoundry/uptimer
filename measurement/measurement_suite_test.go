package measurement_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMeasurement(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Measurement Suite")
}
