package appLogValidator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppLogValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppLogValidator Suite")
}
