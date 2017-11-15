#!/bin/bash

echo "Generating fakes..."
go generate $(go list ./... | grep -v /vendor/)
echo
echo "Running tests..."
echo ginkgo -r -randomizeSuites -randomizeAllSpecs
ginkgo -r -randomizeSuites -randomizeAllSpecs
echo
echo "Running go vet..."
go vet $(go list ./... | grep -v /vendor/)