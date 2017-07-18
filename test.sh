#!/bin/bash

go generate $(go list ./... | grep -v /vendor/)
ginkgo -r