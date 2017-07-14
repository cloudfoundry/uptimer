package appLogValidator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppLogValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppLogValidator Suite")
}
