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

echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-windows-arm64.exe internal/cmd/main/main.go &

# Wait for all background processes to complete
wait

echo "All builds completed successfully"

# Create checksums
cd build
sha256sum * > checksums.txt

# Archive binaries
echo "Creating archives..."
# Linux archives (tar.gz)
tar -czf devplan-linux-amd64.tar.gz devplan-linux-amd64
tar -czf devplan-linux-arm64.tar.gz devplan-linux-arm64

# macOS archives (tar.gz)
tar -czf devplan-darwin-amd64.tar.gz devplan-darwin-amd64
tar -czf devplan-darwin-arm64.tar.gz devplan-darwin-arm64

# Windows archive (zip)
if command -v zip >/dev/null 2>&1; then
  zip devplan-windows-amd64.zip devplan-windows-amd64.exe
  rm -f devplan-windows-amd64.exe
  zip devplan-windows-arm64.zip devplan-windows-arm64.exe
  rm -f devplan-windows-arm64.exe
else
  echo "Warning: zip command not found, skipping Windows archive creation"
fi

# Update checksums to include archives
sha256sum * > checksums.txt

# Remove original binaries after archiving
echo "Removing original binaries..."
rm -f devplan-linux-amd64 devplan-linux-arm64 devplan-darwin-amd64 devplan-darwin-arm64

echo "Archives created successfully and original binaries removed"
