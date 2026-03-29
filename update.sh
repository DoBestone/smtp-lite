#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - Auto-update script (downloads GitHub Release binary)
#  Requires: curl only - NO git or go needed on the server
#  Usage:
#    bash update.sh           check for new version, update if found
#    bash update.sh --force   force download & restart even if same version
# =============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="smtp-lite"
GITHUB_REPO="DoBestone/smtp-lite"
BINARY="$SCRIPT_DIR/smtp-lite"

FORCE=false
[[ "${1:-}" == "--force" ]] && FORCE=true

# ---- Detect platform ----
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
ASSET_NAME="smtp-lite-${OS}-${ARCH}"
echo "Platform: ${OS}/${ARCH}"

# ---- Read port from config.yaml ----
PORT=8090
if [ -f "$SCRIPT_DIR/config.yaml" ]; then
  _port=$(grep -E '^[[:space:]]+port:' "$SCRIPT_DIR/config.yaml" 2>/dev/null | head -1 | awk '{print $2}' || true)
  [ -n "$_port" ] && PORT=$_port
fi

# ---- Current version: query local API ----
CURRENT=""
CURRENT=$(curl -fsSL "http://localhost:${PORT}/api/v1/version" 2>/dev/null   | grep -o '"v[0-9][^"]*"' | tr -d '"' || true)
if [ -z "$CURRENT" ] && [ -x "$BINARY" ]; then
  CURRENT=$("$BINARY" --version 2>/dev/null || true)
fi
[ -z "$CURRENT" ] && CURRENT="unknown"
echo "Current version: ${CURRENT}"

# ---- Latest release from GitHub ----
RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null || true)
LATEST=$(echo "$RELEASE_JSON" | grep '"tag_name"' | grep -o '"v[^"]*"' | tr -d '"' || true)
DOWNLOAD_URL=$(echo "$RELEASE_JSON"   | grep '"browser_download_url"'   | grep "${ASSET_NAME}"   | grep -o '"https://[^"]*"' | tr -d '"' | head -1 || true)

[ -z "$LATEST" ] && { echo "ERROR: cannot get latest release from GitHub - check network"; exit 1; }
echo "Latest version:  ${LATEST}"

if [ "$LATEST" = "$CURRENT" ] && [ "$FORCE" = false ]; then
  echo "Already up to date."
  echo "Tip: use 'bash update.sh --force' to re-download"
  exit 0
fi

[ -z "$DOWNLOAD_URL" ] && {
  echo "ERROR: Release ${LATEST} has no asset ${ASSET_NAME}"
  echo "The CI build may still be running. Check: https://github.com/${GITHUB_REPO}/releases"
  exit 1
}

# ---- Download and atomic replace ----
echo "Downloading ${LATEST} (${ASSET_NAME})..."
echo "URL: $DOWNLOAD_URL"

TMP_BINARY=$(mktemp "${SCRIPT_DIR}/.smtp-lite.tmp.XXXXXX")
trap 'rm -f "$TMP_BINARY"' EXIT

curl -fL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BINARY" || { echo "Download failed"; exit 1; }
chmod +x "$TMP_BINARY"

# Backup old binary
if [ -f "$BINARY" ]; then
  cp "$BINARY" "${BINARY}.bak"
  echo "Old binary backed up to smtp-lite.bak"
fi

# Atomic replace (mv within same filesystem is atomic)
mv "$TMP_BINARY" "$BINARY"
echo "Binary replaced -> ${LATEST}"

# ---- Restart service ----
echo "Restarting service..."

_restart_process() {
  PID=$(pgrep -f "${BINARY}" 2>/dev/null | head -1 || true)
  if [ -n "$PID" ]; then
    echo "Stopping process (PID: ${PID})..."
    kill "$PID" 2>/dev/null || true
    for _i in 1 2 3 4 5; do
      kill -0 "$PID" 2>/dev/null || break
      sleep 1
    done
  fi
  echo "Starting new process..."
  nohup "$BINARY" >> "$SCRIPT_DIR/smtp-lite.log" 2>&1 &
  NEW_PID=$!
  sleep 1
  kill -0 "$NEW_PID" 2>/dev/null     && echo "Service started (PID: ${NEW_PID})"     || { echo "Service failed to start - check smtp-lite.log"; exit 1; }
}

if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
  sudo systemctl restart "$SERVICE_NAME"
  echo "systemd service restarted"
elif launchctl list 2>/dev/null | grep -q "com.smtp-lite"; then
  PLIST="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
  if [ -f "$PLIST" ]; then
    launchctl unload "$PLIST" 2>/dev/null || true
    launchctl load   "$PLIST"
    echo "LaunchAgent restarted"
  else
    _restart_process
  fi
else
  _restart_process
fi

echo ""
echo "Update complete: ${CURRENT} -> ${LATEST}"
echo ""
