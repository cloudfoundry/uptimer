//go:build tools
// +build tools

// This package imports things required by build, and test, scripts, to force
// `go mod` to see them as dependencies.
package tools

import (
	_ "github.com/onsi/ginkgo/v2/ginkgo"
)
