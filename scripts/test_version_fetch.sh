#!/bin/bash

# Test script to verify the JSON parsing in install.sh

# Simulate fetching the version.json file
echo "Simulating fetch of version.json..."
VERSION_JSON='{"production_version":"v1.2.3"}'
echo "JSON: ${VERSION_JSON}"

# Extract the production_version field from the JSON
PRODUCTION_VERSION=$(echo "${VERSION_JSON}" | grep -o '"production_version"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"[^"]*"$' | tr -d '"')

if [ -z "${PRODUCTION_VERSION}" ]; then
    echo "Failed to extract production version from config."
else
    echo "Extracted version: ${PRODUCTION_VERSION}"
fi

# Test with a more complex JSON
echo -e "\nTesting with more complex JSON..."
VERSION_JSON='{"production_version":"v2.0.0","other_field":"value"}'
echo "JSON: ${VERSION_JSON}"

# Extract the production_version field from the JSON
PRODUCTION_VERSION=$(echo "${VERSION_JSON}" | grep -o '"production_version"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"[^"]*"$' | tr -d '"')

if [ -z "${PRODUCTION_VERSION}" ]; then
    echo "Failed to extract production version from config."
else
    echo "Extracted version: ${PRODUCTION_VERSION}"
fi