#!/bin/sh
set -e

# aix installation script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/thoreinstein/aix/main/install.sh | sh

GITHUB_REPO="thoreinstein/aix"
BINARY_NAME="aix"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)   OS='linux';;
  darwin*)  OS='darwin';;
  msys*|cygwin*|mingw*) OS='windows';;
  *)        echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64) ARCH='amd64';;
  arm64|aarch64) ARCH='arm64';;
  *)      echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

# Get latest version
echo "Checking for the latest version of aix..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_VERSION}" ]; then
  echo "Error: Could not find latest version for ${GITHUB_REPO}"
  exit 1
fi

# Clean version (remove 'v' prefix if present)
VERSION_NUMBER=$(echo "${LATEST_VERSION}" | sed 's/^v//')

echo "Latest version is ${LATEST_VERSION}"

# Determine extension
EXT="tar.gz"
if [ "${OS}" = "windows" ]; then
  EXT="zip"
fi

# Template: aix_{{.Version}}_{{.Os}}_{{.Arch}}
FILENAME="aix_${VERSION_NUMBER}_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${FILENAME}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading ${DOWNLOAD_URL}..."
curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${FILENAME}"

echo "Extracting..."
if [ "${EXT}" = "zip" ]; then
  unzip -q "${TMP_DIR}/${FILENAME}" -d "${TMP_DIR}"
else
  tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"
fi

# Install binary
INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
else
  echo "Installing to ${INSTALL_DIR}..."
  mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
fi

echo "aix ${LATEST_VERSION} installed successfully to ${INSTALL_DIR}/${BINARY_NAME}"
