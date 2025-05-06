#!/bin/bash

set -e
set -o pipefail

mkdir -p build

# Common ldflags
LDFLAGS="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=${VERSION:-dev}' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=${COMMIT_HASH:-na}' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=${BUILD_DATE:-na}'"

# Build all CLIs in parallel
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-amd64 internal/cmd/main/main.go &

echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-arm64 internal/cmd/main/main.go &

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-amd64 internal/cmd/main/main.go &

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-arm64 internal/cmd/main/main.go &

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-windows-amd64.exe internal/cmd/main/main.go &

# Wait for all background processes to complete
wait

echo "All builds completed successfully"

# Create checksums
cd build
sha256sum * > checksums.txt
