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
	@go build -o out/devplan-cli

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
