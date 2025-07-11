#!/bin/bash

set -e
set -o pipefail

echo "Testing build flags for auto-update functionality..."

# Clean up any existing build artifacts
rm -rf build-test
mkdir -p build-test

# Build with auto-update enabled (default)
echo "Building with auto-update enabled..."
go build -ldflags="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=test-enabled' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=test' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=test'" -o build-test/devplan-enabled internal/cmd/main/main.go

# Build with auto-update disabled
echo "Building with auto-update disabled..."
go build -ldflags="-X 'github.com/devplaninc/devplan-cli/internal/version.Version=test-disabled' -X 'github.com/devplaninc/devplan-cli/internal/version.CommitHash=test' -X 'github.com/devplaninc/devplan-cli/internal/version.BuildDate=test' -X 'github.com/devplaninc/devplan-cli/internal/version.DisableAutoUpdate=true'" -o build-test/devplan-disabled internal/cmd/main/main.go

# Test auto-update enabled binary
echo "Testing auto-update enabled binary..."
if ./build-test/devplan-enabled version | grep -q "Auto-update: enabled"; then
    echo "✓ Auto-update enabled binary works correctly"
else
    echo "✗ Auto-update enabled binary failed - expected 'Auto-update: enabled'"
    ./build-test/devplan-enabled version
    exit 1
fi

# Test auto-update disabled binary
echo "Testing auto-update disabled binary..."
if ./build-test/devplan-disabled version | grep -q "Auto-update: disabled"; then
    echo "✓ Auto-update disabled binary works correctly"
else
    echo "✗ Auto-update disabled binary failed - expected 'Auto-update: disabled'"
    ./build-test/devplan-disabled version
    exit 1
fi

# Test that auto-update disabled binary rejects update command
echo "Testing that auto-update disabled binary rejects update command..."
update_output=$(./build-test/devplan-disabled update 2>&1 || true)
if echo "$update_output" | sed 's/\x1b\[[0-9;]*m//g' | grep -q "Auto-update is disabled"; then
    echo "✓ Auto-update disabled binary correctly rejects update command"
else
    echo "✗ Auto-update disabled binary should reject update command"
    echo "Command output: $update_output"
    exit 1
fi

# Test that auto-update disabled binary rejects version --latest command
echo "Testing that auto-update disabled binary rejects version --latest command..."
latest_output=$(./build-test/devplan-disabled version --latest 2>&1 || true)
if echo "$latest_output" | sed 's/\x1b\[[0-9;]*m//g' | grep -q "Auto-update is disabled"; then
    echo "✓ Auto-update disabled binary correctly rejects version --latest command"
else
    echo "✗ Auto-update disabled binary should reject version --latest command"
    echo "Command output: $latest_output"
    exit 1
fi

# Clean up
rm -rf build-test

echo "All build flag tests passed!"