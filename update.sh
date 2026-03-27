#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 一键更新脚本
#  用法: bash update.sh          # 检测到新版本后更新
#        bash update.sh --force  # 强制重新编译并重启
# =============================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="smtp-lite"
GITHUB_REPO="DoBestone/smtp-lite"

G='\033[0;32m'; Y='\033[1;33m'; B='\033[0;34m'
R='\033[0;31m'; W='\033[1;37m'; N='\033[0m'

info()    { echo -e "  ${B}▸${N} $1"; }
ok()      { echo -e "  ${G}✓${N} $1"; }
warn()    { echo -e "  ${Y}⚠${N}  $1"; }
err()     { echo -e "  ${R}✗${N} $1"; exit 1; }
step()    { echo -e "\n  ${W}── $1 ──${N}"; }

FORCE=false
[[ "${1:-}" == "--force" ]] && FORCE=true

cd "$SCRIPT_DIR"

# ---- 检查运行环境 ----
step "环境检查"
command -v git &>/dev/null || err "未找到 git 命令"
command -v go  &>/dev/null || err "未找到 go 命令（请先运行 install.sh）"
ok "git $(git --version | awk '{print $3}')"
ok "go  $(go version | awk '{print $3}')"

# ---- 获取版本信息 ----
step "版本检测"

CURRENT=$(grep 'Version' internal/version/version.go 2>/dev/null \
  | grep -o '"v[^"]*"' | tr -d '"' || echo "unknown")
info "当前版本: ${W}${CURRENT}${N}"

# 查询 GitHub 最新 release
LATEST=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null \
  | grep '"tag_name"' | grep -o '"v[^"]*"' | tr -d '"' || echo "")

if [ -z "$LATEST" ]; then
  warn "无法获取 GitHub 最新版本（网络问题），继续执行本地更新"
else
  info "最新版本: ${W}${LATEST}${N}"
  if [ "$LATEST" = "$CURRENT" ] && [ "$FORCE" = false ]; then
    ok "已是最新版本，无需更新"
    echo ""
    echo -e "  提示：使用 ${W}bash update.sh --force${N} 可强制重新编译"
    exit 0
  fi
  [ "$LATEST" != "$CURRENT" ] && info "发现新版本 ${LATEST}，开始更新..."
fi

# ---- 拉取代码 ----
step "拉取代码 (git pull)"
git pull || err "git pull 失败，请检查网络连接或手动解决冲突"
ok "代码拉取完成"

# ---- 编译 ----
step "重新编译"
info "go build ./cmd/server/ ..."
go build -o smtp-lite ./cmd/server/ || err "编译失败，请检查错误日志"
NEW_VER=$(grep 'Version' internal/version/version.go 2>/dev/null \
  | grep -o '"v[^"]*"' | tr -d '"' || echo "unknown")
ok "编译完成 → ${NEW_VER}"

# ---- 重启服务 ----
step "重启服务"

_restart_systemd() {
  sudo systemctl restart "$SERVICE_NAME"
  ok "Systemd 服务已重启 (systemctl status ${SERVICE_NAME})"
}

_restart_launchd() {
  PLIST="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
  if [ -f "$PLIST" ]; then
    launchctl unload "$PLIST" 2>/dev/null || true
    launchctl load   "$PLIST"
    ok "LaunchAgent 已重启"
  else
    warn "未找到 LaunchAgent 配置，尝试直接重启进程"
    _restart_process
  fi
}

_restart_process() {
  # 找到运行中的进程并发送 SIGTERM，然后重新启动
  PID=$(pgrep -f "$SCRIPT_DIR/smtp-lite" 2>/dev/null || true)
  if [ -n "$PID" ]; then
    info "停止旧进程 (PID: ${PID})..."
    kill "$PID" 2>/dev/null || true
    sleep 2
  fi
  info "启动新进程..."
  nohup "$SCRIPT_DIR/smtp-lite" >> "$SCRIPT_DIR/smtp-lite.log" 2>&1 &
  NEW_PID=$!
  sleep 1
  kill -0 "$NEW_PID" 2>/dev/null && ok "服务已启动 (PID: ${NEW_PID})" \
    || err "服务启动失败，请检查 smtp-lite.log"
}

# 按优先级检测服务管理器
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
  _restart_systemd
elif launchctl list 2>/dev/null | grep -q "com.smtp-lite"; then
  _restart_launchd
else
  _restart_process
fi

# ---- 完成 ----
echo ""
echo -e "  ${G}┌──────────────────────────────────┐${N}"
echo -e "  ${G}│${N}   ${W}✓ 更新完成！${N}                  ${G}│${N}"
echo -e "  ${G}│${N}   版本: ${CURRENT} → ${W}${NEW_VER}${N}       ${G}│${N}"
echo -e "  ${G}└──────────────────────────────────┘${N}"
echo ""
