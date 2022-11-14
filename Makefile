# Define the target to run if make is called with no arguments.
.PHONY: default
default: build test

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=
export GO111MODULE=on
export GODEBUG=x509ignoreCN=0

.PHONY: build
build:
	go build $(BUILD_OPTS )-o gen-crd-api-reference-docs

test:
	go test `go list ./...`

clean:
	go clean