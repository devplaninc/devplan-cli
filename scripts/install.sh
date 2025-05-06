#!/bin/bash

# Devplan CLI Installer
# This script downloads and installs the Devplan CLI

set -e

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture to Go architecture
case "${ARCH}" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: ${ARCH}"; exit 1 ;;
esac

# Set installation directory
if [ "${OS}" = "darwin" ] || [ "${OS}" = "linux" ]; then
    # For Linux and Mac, we'll use ~/.devplan/cli/versions/<version> for the actual binary
    # and create a symlink in ${HOME}/bin, /usr/local/bin, or ${HOME}/.local/bin
    # Check for ${HOME}/bin first
    if [ -d "${HOME}/bin" ] && [ -w "${HOME}/bin" ]; then
        SYMLINK_DIR="${HOME}/bin"
    # If ${HOME}/bin doesn't exist but parent directory is writable, create it
    elif [ ! -d "${HOME}/bin" ] && [ -w "${HOME}" ]; then
        SYMLINK_DIR="${HOME}/bin"
        mkdir -p "${SYMLINK_DIR}"
    # Then check for /usr/local/bin
    elif [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        SYMLINK_DIR="/usr/local/bin"
    # Finally, use ${HOME}/.local/bin
    else
        SYMLINK_DIR="${HOME}/.local/bin"
        mkdir -p "${SYMLINK_DIR}"
    fi
    # We'll set INSTALL_DIR after we know the version
elif [ "${OS}" = "windows" ]; then
    INSTALL_DIR="${HOME}/bin"
    mkdir -p "${INSTALL_DIR}"
else
    echo "Unsupported operating system: ${OS}"
    exit 1
fi

# Add installation directory to PATH if needed
if [ "${OS}" = "darwin" ] || [ "${OS}" = "linux" ]; then
    # For Linux and Mac, we need to add SYMLINK_DIR to PATH
    if [[ ":$PATH:" != *":${SYMLINK_DIR}:"* ]]; then
        echo "Adding ${SYMLINK_DIR} to PATH"
        if [ -f "${HOME}/.bashrc" ]; then
            echo "export PATH=\"\${PATH}:${SYMLINK_DIR}\"" >> "${HOME}/.bashrc"
            echo "Please run 'source ~/.bashrc' after installation to update your PATH"
        elif [ -f "${HOME}/.zshrc" ]; then
            echo "export PATH=\"\${PATH}:${SYMLINK_DIR}\"" >> "${HOME}/.zshrc"
            echo "Please run 'source ~/.zshrc' after installation to update your PATH"
        else
            echo "Please add ${SYMLINK_DIR} to your PATH manually"
        fi
    fi
elif [ "${OS}" = "windows" ]; then
    # For Windows, we need to add INSTALL_DIR to PATH
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        echo "Adding ${INSTALL_DIR} to PATH"
        echo "Please add ${INSTALL_DIR} to your PATH manually"
    fi
fi

# Set binary name based on OS
if [ "${OS}" = "windows" ]; then
    BINARY_NAME="devplan.exe"
else
    BINARY_NAME="devplan"
fi

# Set download URL
VERSION="latest"
if [ -n "$1" ]; then
    VERSION="$1"
fi

if [ "${VERSION}" = "latest" ]; then
    # Fetch the latest version from the config
    CONFIG_URL="https://devplan-cli.sfo3.digitaloceanspaces.com/releases/version.json"
    echo "Fetching latest version from ${CONFIG_URL}..."

    if command -v curl > /dev/null 2>&1; then
        VERSION_JSON=$(curl -s "${CONFIG_URL}")
    elif command -v wget > /dev/null 2>&1; then
        VERSION_JSON=$(wget -q -O - "${CONFIG_URL}")
    else
        echo "Neither curl nor wget found. Please install one of them and try again."
        exit 1
    fi

    # Extract the production_version field from the JSON
    # This is a simple extraction that assumes the JSON is well-formed
    # and that the production_version field is the only field with that name
    PRODUCTION_VERSION=$(echo "${VERSION_JSON}" | grep -o '"production_version"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"[^"]*"$' | tr -d '"')

    if [ -z "${PRODUCTION_VERSION}" ]; then
        echo "Failed to extract production version from config. Exiting."
        exit 1
    else
        echo "Latest version is ${PRODUCTION_VERSION}"
        VERSION="${PRODUCTION_VERSION}"
    fi
fi

# Set up the installation directory based on OS and version
if [ "${OS}" = "darwin" ] || [ "${OS}" = "linux" ]; then
    INSTALL_DIR="${HOME}/.devplan/cli/versions/${VERSION}"
    mkdir -p "${INSTALL_DIR}"
fi

DOWNLOAD_URL="https://devplan-cli.sfo3.digitaloceanspaces.com/releases/versions/${VERSION}/devplan-${OS}-${ARCH}"
if [ "${OS}" = "windows" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
fi

echo "Downloading Devplan CLI ${VERSION} for ${OS}/${ARCH}..."
echo "From: ${DOWNLOAD_URL}"

# Download the binary
if command -v curl > /dev/null 2>&1; then
    curl -L -o "${INSTALL_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"
elif command -v wget > /dev/null 2>&1; then
    wget -O "${INSTALL_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"
else
    echo "Neither curl nor wget found. Please install one of them and try again."
    exit 1
fi

# Make the binary executable
if [ "${OS}" != "windows" ]; then
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Create symlink for Linux and Mac
if [ "${OS}" = "darwin" ] || [ "${OS}" = "linux" ]; then
    # Remove existing symlink or file if it exists
    if [ -L "${SYMLINK_DIR}/${BINARY_NAME}" ] || [ -f "${SYMLINK_DIR}/${BINARY_NAME}" ]; then
        rm -f "${SYMLINK_DIR}/${BINARY_NAME}"
    fi

    # Create symlink
    ln -s "${INSTALL_DIR}/${BINARY_NAME}" "${SYMLINK_DIR}/${BINARY_NAME}"
    echo "Devplan CLI has been installed to ${INSTALL_DIR}/${BINARY_NAME}"
    echo "Symlink created at ${SYMLINK_DIR}/${BINARY_NAME}"
else
    echo "Devplan CLI has been installed to ${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "Run 'devplan --help' to get started"
