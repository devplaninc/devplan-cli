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
    INSTALL_DIR="/usr/local/bin"
    if [ ! -d "${INSTALL_DIR}" ] || [ ! -w "${INSTALL_DIR}" ]; then
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "${INSTALL_DIR}"
    fi
elif [ "${OS}" = "windows" ]; then
    INSTALL_DIR="${HOME}/bin"
    mkdir -p "${INSTALL_DIR}"
else
    echo "Unsupported operating system: ${OS}"
    exit 1
fi

# Add installation directory to PATH if needed
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    echo "Adding ${INSTALL_DIR} to PATH"
    if [ -f "${HOME}/.bashrc" ]; then
        echo "export PATH=\"\${PATH}:${INSTALL_DIR}\"" >> "${HOME}/.bashrc"
        echo "Please run 'source ~/.bashrc' after installation to update your PATH"
    elif [ -f "${HOME}/.zshrc" ]; then
        echo "export PATH=\"\${PATH}:${INSTALL_DIR}\"" >> "${HOME}/.zshrc"
        echo "Please run 'source ~/.zshrc' after installation to update your PATH"
    else
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
    VERSION="v0.1.0" # Default to v0.1.0 if no version is specified
fi

DOWNLOAD_URL="https://devplan-cli.sfo3.digitaloceanspaces.com/releases/${VERSION}/devplan-${OS}-${ARCH}"
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

echo "Devplan CLI has been installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo "Run 'devplan --help' to get started"