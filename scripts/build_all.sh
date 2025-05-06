#!/bin/bash

set -ex
set -o pipefail

mkdir -p build

# Common ldflags
LDFLAGS="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=${VERSION:-dev}' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=${COMMIT_HASH:-na}' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=${BUILD_DATE:-na}'"

# Build for Linux
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-amd64 internal/cmd/main/main.go
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-arm64 internal/cmd/main/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-amd64 internal/cmd/main/main.go
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-arm64 internal/cmd/main/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-windows-amd64.exe internal/cmd/main/main.go

# Create checksums
cd build
sha256sum * > checksums.txt
