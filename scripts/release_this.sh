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

# Delete the tag if it exists (both local and remote)
git tag -d "$TAG_NAME" 2>/dev/null || true
git push origin ":refs/tags/$TAG_NAME" 2>/dev/null || true

# Create new tag
git tag -a "$TAG_NAME" "$COMMIT" -m "Release $TAG_NAME"

# Push the new tag
git push origin "$TAG_NAME"
