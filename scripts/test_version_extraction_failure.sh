#!/bin/bash

# Test script to verify the failure behavior when version extraction fails

# Simulate an empty or malformed version.json
echo "Simulating malformed version.json..."
VERSION_JSON='{}'
echo "JSON: ${VERSION_JSON}"

# Extract the production_version field from the JSON
PRODUCTION_VERSION=$(echo "${VERSION_JSON}" | grep -o '"production_version"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"[^"]*"$' | tr -d '"')

if [ -z "${PRODUCTION_VERSION}" ]; then
    echo "Failed to extract production version from config. Exiting."
    exit 1
else
    echo "Extracted version: ${PRODUCTION_VERSION}"
fi

echo "This line should not be reached if extraction fails."