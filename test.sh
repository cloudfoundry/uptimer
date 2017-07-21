#!/bin/bash

echo "Generating fakes..."
go generate $(go list ./... | grep -v /vendor/)
echo
echo "Running tests..."
ginkgo -r