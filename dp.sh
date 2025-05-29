#!/bin/bash

set -e
# set -x
set -o pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#
#cd "$script_dir"

#go run internal/cmd/main/main.go --domain=beta "$@"
go run -C "$script_dir" internal/cmd/main/main.go "$@"
