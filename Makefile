.PHONY: all build test clean proto generate

ROOT := $(shell git rev-parse --show-toplevel)
export GOBIN := $(ROOT)/bin
export GOPRIVATE=github.com/devplaninc/webapp

bin/protoc-gen-go:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go

# Default target
all: generate

build:
	@echo "Building Devplan CLI..."
	@mkdir -p out
	@go build -ldflags="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=dev' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=$(shell git rev-parse HEAD)' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")'" -o out/devplan-cli internal/cmd/main/main.go

test:
	@echo "Running tests..."
	@go test ./...

proto: bin/protoc-gen-go
	@echo "Generating protobuf files..."
	@chmod +x scripts/generate_proto.sh
	@./scripts/generate_proto.sh

deps:
	@echo "Installing dependencies..."
	@go mod tidy

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf out

generate: proto
