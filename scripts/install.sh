#!/bin/sh
# install.sh — macOS / Linux installer for mju-dataset
# Usage: curl -fsSL <install_script_url> | sh
#
# This script does NOT contain the labeling server address.
# It fetches the latest release from the distribution API and downloads
# the binary via a presigned URL.
set -e

API_BASE="https://mjudcd-grac-api.newlearn.ai.kr"
GH_REPO="mjudcd-ct-r-d-labeling/labeling_download_cli"
BINARY="mju-dataset"
INSTALL_DIR="/usr/local/bin"

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

json_get_string() {
  field="$1"

  grep -m1 -o "\"${field}\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" \
    | sed 's/.*:[[:space:]]*"//;s/"$//;s/\\u0026/\&/g;s#\\/#/#g'
}

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

# ── Resolve version ───────────────────────────────────────────────────────────
# A specific version can be passed as the first argument:
#   curl -fsSL <url>/install.sh | sh -s -- 2026.05.26.42
VERSION="${1:-}"

if [ -z "$VERSION" ]; then
  echo "Fetching latest release for ${OS_BUILD_TYPE}..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${GH_REPO}/tags?per_page=1" \
    -H "Accept: application/vnd.github+json" \
    | json_get_string name)

  if [ -z "$VERSION" ]; then
    echo "Failed to determine latest version from GitHub tags."
    exit 1
  fi
fi

# ── Fetch download URL from release server ────────────────────────────────────
DL_JSON=$(curl -s -w "\n%{http_code}" \
  "${API_BASE}/cli-releases/download/${OS_BUILD_TYPE}/${VERSION}")
HTTP_STATUS=$(printf '%s' "$DL_JSON" | tail -1)
DL_BODY=$(printf '%s' "$DL_JSON" | sed '$d')

if [ "$HTTP_STATUS" != "200" ]; then
  echo "Release ${VERSION} not found on server (HTTP ${HTTP_STATUS})."
  echo "Response: ${DL_BODY}"
  exit 1
fi

DOWNLOAD_URL=$(printf '%s' "$DL_BODY" | json_get_string download_url)
SHA256=$(printf '%s' "$DL_BODY" | json_get_string sha256)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Failed to parse download URL from server response."
  echo "Response: ${DL_BODY}"
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
  ACTUAL=$(sha256_file "${TMP}/${BINARY}")
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
