#!/bin/sh
# Devplan CLI Wrapper Script
# This script wraps the devplan binary to support follow-up instructions execution.
# It forwards all arguments to the CLI with an additional --instructions-file parameter,
# then executes any command specified in the instructions file.

set -e

# Resolve the directory where this script is located
# Try readlink -f first for canonical path, fall back to basic dirname
if command -v readlink >/dev/null 2>&1 && readlink -f "$0" >/dev/null 2>&1; then
    SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
else
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
fi

# Determine binary name based on OS
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) BINARY_NAME="devplan-cli.exe" ;;
    *) BINARY_NAME="devplan-cli" ;;
esac

BINARY_RELATIVE_DIR=""

DEVPLAN_BIN="${SCRIPT_DIR}/${BINARY_RELATIVE_DIR}${BINARY_NAME}"

# Check if binary exists
if [ ! -f "${DEVPLAN_BIN}" ]; then
    echo "Error: devplan binary not found at ${DEVPLAN_BIN}" >&2
    exit 1
fi

# Create temporary instructions file
INSTRUCTIONS_FILE=$(mktemp)
if [ -z "${INSTRUCTIONS_FILE}" ] || [ ! -f "${INSTRUCTIONS_FILE}" ]; then
    echo "Error: Failed to create temporary instructions file" >&2
    exit 1
fi

# Cleanup function to remove temp file
cleanup() {
    if [ -f "${INSTRUCTIONS_FILE}" ]; then
        rm -f "${INSTRUCTIONS_FILE}" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Run the devplan CLI with all arguments plus the instructions file
set +e
"${DEVPLAN_BIN}" "$@" --instructions-file="${INSTRUCTIONS_FILE}"
CLI_EXIT_CODE=$?
set -e

# If instructions file doesn't exist or is empty, exit with CLI exit code
if [ ! -f "${INSTRUCTIONS_FILE}" ] || [ ! -s "${INSTRUCTIONS_FILE}" ]; then
    exit "${CLI_EXIT_CODE}"
fi

# Parse the first non-empty line for exec: "command" pattern
# Read each line and check for exec pattern
EXEC_CMD=""
while IFS= read -r line || [ -n "$line" ]; do
    # Skip empty lines
    case "$line" in
        "") continue ;;
    esac

    # Check if line matches exec: "..." pattern
    # Extract content between quotes after "exec:"
    if echo "$line" | grep -q '^[[:space:]]*exec=[[:space:]]*'; then
        # Extract the command from between the quotes
        EXEC_CMD=$(echo "$line" | sed 's/^[[:space:]]*exec=[[:space:]]*\(.*\)/\1/')
        break
    fi

    # Only check first non-empty line
    break
done < "${INSTRUCTIONS_FILE}"

# If no valid exec command found, exit with CLI exit code
if [ -z "${EXEC_CMD}" ]; then
    exit "${CLI_EXIT_CODE}"
fi

# Execute the extracted command
set +e

echo "Starting: ${EXEC_CMD}"

sh -c "${EXEC_CMD}"

EXEC_EXIT_CODE=$?
set -e

exit "${EXEC_EXIT_CODE}"
