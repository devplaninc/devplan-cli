#!/bin/bash

set -e
set -o pipefail

mkdir -p build

# Common ldflags
LDFLAGS="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=${VERSION:-dev}' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=${COMMIT_HASH:-na}' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=${BUILD_DATE:-na}'"

# Define LDFLAGS for auto-update disabled builds
LDFLAGS_NOAUTOUPDATE="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=${VERSION:-dev}' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=${COMMIT_HASH:-na}' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=${BUILD_DATE:-na}' -X 'github.com/devplaninc/devplan-cli/internal/version.DisableAutoUpdate=true'"

# Build all CLIs in parallel (both auto-update enabled and disabled)
echo "Building all binaries in parallel..."

# Auto-update enabled builds
echo "Starting auto-update enabled builds..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-amd64 internal/cmd/main/main.go &
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-linux-arm64 internal/cmd/main/main.go &
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-amd64 internal/cmd/main/main.go &
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-darwin-arm64 internal/cmd/main/main.go &
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/devplan-windows-amd64.exe internal/cmd/main/main.go &
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/devplan-windows-arm64.exe internal/cmd/main/main.go &

# Auto-update disabled builds
echo "Starting auto-update disabled builds..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-linux-amd64-noautoupdate internal/cmd/main/main.go &
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-linux-arm64-noautoupdate internal/cmd/main/main.go &
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-darwin-amd64-noautoupdate internal/cmd/main/main.go &
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-darwin-arm64-noautoupdate internal/cmd/main/main.go &
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-windows-amd64-noautoupdate.exe internal/cmd/main/main.go &
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS_NOAUTOUPDATE}" -o build/devplan-windows-arm64-noautoupdate.exe internal/cmd/main/main.go &

# Wait for all background processes to complete
echo "Waiting for all builds to complete..."
wait

echo "All builds completed successfully"

# Create checksums
cd build
sha256sum * > checksums.txt

# Archive binaries
echo "Creating archives..."
# Linux archives (tar.gz) - auto-update enabled
tar -czf devplan-linux-amd64.tar.gz devplan-linux-amd64 -C ../wrappers devplan.sh
tar -czf devplan-linux-arm64.tar.gz devplan-linux-arm64 -C ../wrappers devplan.sh

# macOS archives (tar.gz) - auto-update enabled
tar -czf devplan-darwin-amd64.tar.gz devplan-darwin-amd64 -C ../wrappers devplan.sh
tar -czf devplan-darwin-arm64.tar.gz devplan-darwin-arm64 -C ../wrappers devplan.sh

# Linux archives (tar.gz) - auto-update disabled
tar -czf devplan-linux-amd64-noautoupdate.tar.gz devplan-linux-amd64-noautoupdate -C ../wrappers devplan.sh
tar -czf devplan-linux-arm64-noautoupdate.tar.gz devplan-linux-arm64-noautoupdate -C ../wrappers devplan.sh

# macOS archives (tar.gz) - auto-update disabled
tar -czf devplan-darwin-amd64-noautoupdate.tar.gz devplan-darwin-amd64-noautoupdate -C ../wrappers devplan.sh
tar -czf devplan-darwin-arm64-noautoupdate.tar.gz devplan-darwin-arm64-noautoupdate -C ../wrappers devplan.sh

# Windows archive (zip)
if command -v zip >/dev/null 2>&1; then
  # Auto-update enabled
  zip -j devplan-windows-amd64.zip devplan-windows-amd64.exe ../wrappers/devplan.ps1
  rm -f devplan-windows-amd64.exe
  zip -j devplan-windows-arm64.zip devplan-windows-arm64.exe ../wrappers/devplan.ps1
  rm -f devplan-windows-arm64.exe
  
  # Auto-update disabled
  zip -j devplan-windows-amd64-noautoupdate.zip devplan-windows-amd64-noautoupdate.exe ../wrappers/devplan.ps1
  rm -f devplan-windows-amd64-noautoupdate.exe
  zip -j devplan-windows-arm64-noautoupdate.zip devplan-windows-arm64-noautoupdate.exe ../wrappers/devplan.ps1
  rm -f devplan-windows-arm64-noautoupdate.exe
else
  echo "Warning: zip command not found, skipping Windows archive creation"
fi

# Update checksums to include archives
sha256sum * > checksums.txt

# Remove original binaries after archiving
echo "Removing original binaries..."
rm -f devplan-linux-amd64 devplan-linux-arm64 devplan-darwin-amd64 devplan-darwin-arm64
rm -f devplan-linux-amd64-noautoupdate devplan-linux-arm64-noautoupdate devplan-darwin-amd64-noautoupdate devplan-darwin-arm64-noautoupdate

echo "Archives created successfully and original binaries removed"
