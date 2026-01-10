#!/bin/bash
# Install repoman:
# curl -sSL https://raw.githubusercontent.com/liffiton/repoman/main/install.sh | bash

set -e

# Configuration
OWNER="liffiton"
REPO="repoman"
BINARY_NAME="repoman"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "${OS}" in
  linux*)  OS='linux' ;;
  darwin*) OS='darwin' ;;
  *) echo "Error: Unsupported OS: ${OS}"; exit 1 ;;
esac

# Detect Architecture
ARCH=$(uname -m)
case "${ARCH}" in
  x86_64) ARCH='amd64' ;;
  aarch64|arm64) ARCH='arm64' ;;
  *) echo "Error: Unsupported architecture: ${ARCH}"; exit 1 ;;
esac

# Construct Asset Name (matches internal/update/update.go)
# Pattern: repoman-{os}-{arch}
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${OWNER}/${REPO}/releases/latest/download/${ASSET_NAME}"

echo "Downloading ${BINARY_NAME} for ${OS}/${ARCH}..."

# Create install directory if it doesn't exist
mkdir -p "${INSTALL_DIR}"

# Download the binary
if ! curl -L -f -o "${INSTALL_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"; then
  echo "Error: Failed to download binary. The release may not exist yet or the architecture is unsupported."
  exit 1
fi

# Make it executable
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}"

# Check if the install directory is in the PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  echo ""
  echo "Warning: ${INSTALL_DIR} is not in your PATH."
  echo "To use ${BINARY_NAME} from anywhere, add it to your shell profile:"
  
  # Identify shell profile
  SHELL_TYPE=$(basename "$SHELL")
  PROFILE_FILE=""
  if [ "$SHELL_TYPE" = "zsh" ]; then
    PROFILE_FILE="${HOME}/.zshrc"
    PRINT_PROFILE="~/.zshrc"
  elif [ "$SHELL_TYPE" = "bash" ]; then
    PROFILE_FILE="${HOME}/.bashrc"
    PRINT_PROFILE="~/.bashrc"
    if [ -f "$HOME/.bash_profile" ]; then
      PROFILE_FILE="${HOME}/.bash_profile"
      PRINT_PROFILE="~/.bash_profile"
    fi
  fi

  if [ -n "$PROFILE_FILE" ]; then
    echo "  echo 'export PATH=\"\$PATH:${INSTALL_DIR}\"' >> ${PRINT_PROFILE}"
    echo "  source ${PRINT_PROFILE}"
  else
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
  fi
fi
