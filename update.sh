#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 自动更新脚本 v2.1
#  优先下载 GitHub Release 预编译二进制，失败则源码编译
#  用法:
#    bash update.sh           检查并更新到最新版本
#    bash update.sh --force   强制更新（即使版本相同）
# =============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="smtp-lite"
GITHUB_REPO="DoBestone/smtp-lite"
REPO_URL="https://github.com/${GITHUB_REPO}.git"
BINARY="$SCRIPT_DIR/smtp-lite"

# ── 颜色 ─────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; DIM='\033[2m'; N='\033[0m'

info()    { echo -e "  ${B}▸${N} $*"; }
ok()      { echo -e "  ${G}✓${N} $*"; }
warn()    { echo -e "  ${Y}⚠${N}  $*"; }
err()     { echo -e "  ${R}✗${N} $*"; exit 1; }

FORCE=false
[[ "${1:-}" == "--force" ]] && FORCE=true

# ── 检测平台 ─────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)    ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  armv7*|armhf)    ARCH="armv7" ;;
  i686|i386)       ARCH="386" ;;
  mips64el|mips64) ARCH="mips64le" ;;
  riscv64)         ARCH="riscv64" ;;
  *)               err "不支持的架构: $ARCH" ;;
esac
ASSET_NAME="smtp-lite-${OS}-${ARCH}"
info "平台: ${OS}/${ARCH}"

# ── 读取端口 ─────────────────────────────────────────────────
PORT=8090
if [ -f "$SCRIPT_DIR/config.yaml" ]; then
  _port=$(grep -E '^[[:space:]]+port:' "$SCRIPT_DIR/config.yaml" 2>/dev/null | head -1 | awk '{print $2}' || true)
  [ -n "$_port" ] && PORT=$_port
fi

# ── 当前版本 ─────────────────────────────────────────────────
CURRENT=""
CURRENT=$(curl -fsSL "http://localhost:${PORT}/api/v1/version" 2>/dev/null \
  | grep -o '"v[0-9][^"]*"' | tr -d '"' || true)
if [ -z "$CURRENT" ] && [ -x "$BINARY" ]; then
  CURRENT=$("$BINARY" --version 2>/dev/null || true)
fi
[ -z "$CURRENT" ] && CURRENT="unknown"
info "当前版本: ${W}${CURRENT}${N}"

# ── 最新版本 ─────────────────────────────────────────────────
RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null || true)
LATEST=$(echo "$RELEASE_JSON" | grep '"tag_name"' | grep -o '"v[^"]*"' | tr -d '"' || true)
DOWNLOAD_URL=$(echo "$RELEASE_JSON" \
  | grep '"browser_download_url"' \
  | grep "${ASSET_NAME}" \
  | grep -o '"https://[^"]*"' | tr -d '"' | head -1 || true)

[ -z "$LATEST" ] && { err "无法获取最新版本，请检查网络"; }
info "最新版本: ${W}${LATEST}${N}"

if [ "$LATEST" = "$CURRENT" ] && [ "$FORCE" = false ]; then
  ok "已是最新版本"
  echo -e "  提示：使用 ${W}bash update.sh --force${N} 强制更新"
  exit 0
fi

# ── 更新方式1：下载预编译二进制 ──────────────────────────────
update_binary() {
  if [ -z "$DOWNLOAD_URL" ]; then
    warn "Release ${LATEST} 没有 ${ASSET_NAME} 资源"
    return 1
  fi

  info "下载 ${LATEST} (${ASSET_NAME})..."
  TMP_BINARY=$(mktemp "${SCRIPT_DIR}/.smtp-lite.tmp.XXXXXX")
  trap 'rm -f "$TMP_BINARY"' EXIT

  if ! curl -fL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BINARY"; then
    warn "下载失败"
    rm -f "$TMP_BINARY"
    return 1
  fi

  chmod +x "$TMP_BINARY"

  # 验证可执行
  if ! "$TMP_BINARY" --version &>/dev/null; then
    warn "下载的二进制文件无法执行"
    rm -f "$TMP_BINARY"
    return 1
  fi

  # 备份旧文件
  if [ -f "$BINARY" ]; then
    cp "$BINARY" "${BINARY}.bak"
    info "旧版本已备份为 smtp-lite.bak"
  fi

  mv "$TMP_BINARY" "$BINARY"
  ok "二进制文件已更新 → ${LATEST}"
  return 0
}

# ── 更新方式2：源码编译 ──────────────────────────────────────
update_source() {
  info "使用源码编译方式更新..."

  # 检查依赖
  if ! command -v git &>/dev/null; then
    err "源码编译需要 Git，请先安装"
  fi
  if ! command -v go &>/dev/null; then
    if [ -x /usr/local/go/bin/go ]; then
      export PATH=$PATH:/usr/local/go/bin
    else
      err "源码编译需要 Go，请先安装"
    fi
  fi

  cd "$SCRIPT_DIR"

  if [ -d ".git" ]; then
    info "拉取最新代码..."
    git checkout -- . 2>/dev/null || true
    git pull || err "git pull 失败"
  else
    warn "非 Git 仓库，克隆源码到临时目录..."
    local tmp_src
    tmp_src=$(mktemp -d)
    git clone --depth 1 "$REPO_URL" "$tmp_src"
    # 复制源码（保留 config.yaml 和数据文件）
    cp -r "$tmp_src/cmd" "$SCRIPT_DIR/"
    cp -r "$tmp_src/internal" "$SCRIPT_DIR/"
    cp -r "$tmp_src/web" "$SCRIPT_DIR/"
    cp -r "$tmp_src/frontend" "$SCRIPT_DIR/"
    cp "$tmp_src/go.mod" "$tmp_src/go.sum" "$SCRIPT_DIR/"
    rm -rf "$tmp_src"
  fi

  # 编译前端
  if [ -d "frontend" ] && command -v npm &>/dev/null; then
    info "构建前端..."
    (cd frontend && npm install --silent && npm run build) || warn "前端构建失败，跳过"
  fi

  # 备份旧文件
  if [ -f "$BINARY" ]; then
    cp "$BINARY" "${BINARY}.bak"
    info "旧版本已备份为 smtp-lite.bak"
  fi

  info "编译..."
  go build -ldflags="-s -w" -o smtp-lite ./cmd/server/ || err "编译失败"
  ok "源码编译完成 → ${LATEST}"
}

# ── 重启服务 ─────────────────────────────────────────────────
restart_service() {
  info "重启服务..."

  if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    sudo systemctl restart "$SERVICE_NAME"
    ok "Systemd 服务已重启"
  elif launchctl list 2>/dev/null | grep -q "com.smtp-lite"; then
    PLIST="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    if [ -f "$PLIST" ]; then
      launchctl unload "$PLIST" 2>/dev/null || true
      launchctl load   "$PLIST"
      ok "LaunchAgent 已重启"
    else
      _restart_process
    fi
  else
    _restart_process
  fi
}

_restart_process() {
  PID=$(pgrep -f "${BINARY}" 2>/dev/null | head -1 || true)
  if [ -n "$PID" ]; then
    info "停止旧进程 (PID: ${PID})..."
    kill "$PID" 2>/dev/null || true
    for _i in 1 2 3 4 5; do
      kill -0 "$PID" 2>/dev/null || break
      sleep 1
    done
  fi
  info "启动新进程..."
  cd "$SCRIPT_DIR"
  nohup "$BINARY" >> "$SCRIPT_DIR/smtp-lite.log" 2>&1 &
  NEW_PID=$!
  sleep 1
  kill -0 "$NEW_PID" 2>/dev/null \
    && ok "服务已启动 (PID: ${NEW_PID})" \
    || err "服务启动失败，请检查日志: $SCRIPT_DIR/smtp-lite.log"
}

# ── 主流程 ────────────────────────────────────────────────────
echo ""
echo -e "  ${W}SMTP Lite 更新${N}"
echo ""

if update_binary; then
  info "使用: 预编译二进制"
else
  warn "预编译二进制不可用，尝试源码编译..."
  update_source
  info "使用: 源码编译"
fi

restart_service

echo ""
echo -e "  ${G}更新完成: ${CURRENT} → ${W}${LATEST}${N}"
echo ""
