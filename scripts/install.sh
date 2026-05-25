#!/bin/sh
# install.sh — macOS / Linux installer for mju-dataset
# Usage: curl -fsSL <install_script_url> | sh
#
# This script does NOT contain the labeling server address.
# It fetches the latest release from the distribution API and downloads
# the binary via a presigned URL.
set -e

API_BASE="https://mjudcd-grac-api.newlearn.ai.kr"
BINARY="mju-dataset"
INSTALL_DIR="/usr/local/bin"

# ── Detect OS ────────────────────────────────────────────────────────────────
OS=$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported operating system: $OS"
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
    exit 1
    ;;
esac

OS_BUILD_TYPE="${OS}-${ARCH}"

# ── Fetch latest version info ─────────────────────────────────────────────────
echo "Fetching latest release for ${OS_BUILD_TYPE}..."
LATEST_JSON=$(curl -fsSL \
  "${API_BASE}/cli-releases/latest?os_build_type=${OS_BUILD_TYPE}&current_version=0.0.0")

VERSION=$(printf '%s' "$LATEST_JSON" | grep -o '"version":"[^"]*"' \
  | sed 's/"version":"//;s/"$//')
DOWNLOAD_URL=$(printf '%s' "$LATEST_JSON" | grep -o '"download_url":"[^"]*"' \
  | sed 's/"download_url":"//;s/"$//')
SHA256=$(printf '%s' "$LATEST_JSON" | grep -o '"sha256":"[^"]*"' \
  | sed 's/"sha256":"//;s/"$//')

if [ -z "$VERSION" ] || [ -z "$DOWNLOAD_URL" ]; then
  echo "Failed to fetch latest release information."
  echo "Response: $LATEST_JSON"
  exit 1
fi

echo "Installing ${BINARY} v${VERSION} (${OS_BUILD_TYPE})..."

# ── Download to a temp directory ─────────────────────────────────────────────
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL --progress-bar "${DOWNLOAD_URL}" -o "${TMP}/${BINARY}"

# ── Verify checksum ───────────────────────────────────────────────────────────
if [ -n "$SHA256" ]; then
  echo "Verifying checksum..."
  ACTUAL=$(sha256sum "${TMP}/${BINARY}" | awk '{print $1}')
  if [ "$ACTUAL" != "$SHA256" ]; then
    echo "Checksum verification failed. The download may be corrupt or tampered."
    echo "Expected: $SHA256"
    echo "Actual:   $ACTUAL"
    exit 1
  fi
  echo "Checksum OK"
fi

# ── Install ───────────────────────────────────────────────────────────────────
chmod +x "${TMP}/${BINARY}"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "${BINARY} v${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
"${INSTALL_DIR}/${BINARY}" --version
