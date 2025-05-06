#!/bin/bash

set -e
# set -x
set -o pipefail

go run internal/cmd/main/main.go --domain=local "$@"
