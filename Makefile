# Define the target to run if make is called with no arguments.
.PHONY: default
default: fmt build test

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=
export GO111MODULE=on
export GODEBUG=x509ignoreCN=0

.PHONY: build
build:
	go build $(BUILD_OPTS )-o gen-crd-api-reference-docs

.PHONY: test
test:
	go test `go list ./...`

.PHONY: clean
clean:
	go clean

.PHONY: fmt
fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	find . -name '*.go' | xargs gofmt -s -l -w