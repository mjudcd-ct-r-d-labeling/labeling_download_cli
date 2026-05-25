#!/bin/sh
# install.sh — macOS / Linux installer for mju-dataset
# Usage: curl -fsSL <install_script_url> | sh
#
# This script does NOT contain the labeling server address.
# It only downloads the CLI binary from GitHub Releases.
set -e

REPO="mjudcd-ct-r-d-labeling/labeling_download_cli"
BINARY="mju-dataset"
INSTALL_DIR="/usr/local/bin"

# ── Detect OS ────────────────────────────────────────────────────────────────
OS=$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported operating system: $OS"
    echo "Please download the binary manually from:"
    echo "  https://github.com/${REPO}/releases/latest"
    exit 1
    ;;
esac

# ── Detect architecture ──────────────────────────────────────────────────────
ARCH=$(uname -m 2>/dev/null)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    echo "Please download the binary manually from:"
    echo "  https://github.com/${REPO}/releases/latest"
    exit 1
    ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
CHECKSUM_FILE="checksums-${OS}-${ARCH}.txt"

# ── Resolve latest release tag ───────────────────────────────────────────────
echo "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Failed to determine the latest release. Check your internet connection."
  exit 1
fi

echo "Installing ${BINARY} ${LATEST} (${OS}/${ARCH})..."

BASE_URL="https://github.com/${REPO}/releases/download/${LATEST}"

# ── Download to a temp directory ─────────────────────────────────────────────
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL --progress-bar "${BASE_URL}/${ASSET}"        -o "${TMP}/${ASSET}"
curl -fsSL               "${BASE_URL}/${CHECKSUM_FILE}" -o "${TMP}/${CHECKSUM_FILE}"

# ── Verify checksum ───────────────────────────────────────────────────────────
echo "Verifying checksum..."
( cd "$TMP" && sha256sum -c "${CHECKSUM_FILE}" --ignore-missing ) || {
  echo "Checksum verification failed. The download may be corrupt or tampered."
  exit 1
}

# ── Install ───────────────────────────────────────────────────────────────────
chmod +x "${TMP}/${ASSET}"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${ASSET}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP}/${ASSET}" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "${BINARY} installed successfully to ${INSTALL_DIR}/${BINARY}"
"${INSTALL_DIR}/${BINARY}" --version
