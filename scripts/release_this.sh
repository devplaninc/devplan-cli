#!/bin/bash

# Usage: ./release_this.sh <tag_name>

set -e
set -o pipefail

if [ -z "$1" ]; then
    echo "Error: Tag name is required"
    echo "Usage: $0 <tag_name>"
    exit 1
fi

TAG_NAME="$1"

# Verify that TAG_NAME starts with "v"
if [[ ! "$TAG_NAME" =~ ^v ]]; then
    echo "Error: Tag name must start with 'v'"
    echo "Usage: $0 <tag_name>"
    exit 1
fi
COMMIT="HEAD"

# Check if tag exists locally
if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    echo "Error: Tag '$TAG_NAME' already exists locally"
    exit 1
fi

# Check if tag exists remotely
if git ls-remote --tags origin "refs/tags/$TAG_NAME" | grep -q "$TAG_NAME"; then
    echo "Error: Tag '$TAG_NAME' already exists on remote"
    exit 1
fi

# Create new tag
git tag -a "$TAG_NAME" "$COMMIT" -m "Release $TAG_NAME"

# Push the new tag
git push origin "$TAG_NAME"
