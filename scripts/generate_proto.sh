#!/bin/bash

set -ex
set -o pipefail

export ROOT=$(git rev-parse --show-toplevel)
GOBIN="$ROOT"/bin

PROTO_FILES=$(find proto -name "*.proto")

protoc \
  --proto_path=proto \
  --plugin="$GOBIN"/protoc-gen-go \
  --go_out="$ROOT" \
  --go_opt=default_api_level=API_OPAQUE,paths=import \
  --go_opt=module=github.com/devplaninc/devplan-cli \
  ${PROTO_FILES}
