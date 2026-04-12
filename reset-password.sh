#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 重置管理员密码
#  用法: ./reset-password.sh [新密码]
#        不传参数则进入交互式输入
# =============================================================
set -euo pipefail

# ── 颜色 ─────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
C='\033[0;36m'; W='\033[1;37m'; DIM='\033[2m'; N='\033[0m'

info()    { echo -e "  ${C}▸${N} $*"; }
ok()      { echo -e "  ${G}✓${N} $*"; }
warn()    { echo -e "  ${Y}⚠${N}  $*"; }
err()     { echo -e "  ${R}✗${N} $*"; exit 1; }

# ── 定位安装目录 ──────────────────────────────────────────────
_SCRIPT="${BASH_SOURCE[0]}"
while [ -L "$_SCRIPT" ]; do
  _DIR="$(cd "$(dirname "$_SCRIPT")" && pwd)"
  _SCRIPT="$(readlink "$_SCRIPT")"
  [[ "$_SCRIPT" != /* ]] && _SCRIPT="$_DIR/$_SCRIPT"
done
INSTALL_DIR="$(cd "$(dirname "$_SCRIPT")" && pwd)"

if [ -f "$INSTALL_DIR/config.yaml" ]; then
  CONFIG="$INSTALL_DIR/config.yaml"
elif [ -f "$HOME/smtp-lite/config.yaml" ]; then
  INSTALL_DIR="$HOME/smtp-lite"
  CONFIG="$INSTALL_DIR/config.yaml"
else
  err "未找到 config.yaml，请在 smtp-lite 安装目录下运行"
fi

SERVICE_NAME="smtp-lite"

# ── 辅助函数 ──────────────────────────────────────────────────

yaml_set() {
  local key="$1" value="$2" file="${3:-$CONFIG}"
  if grep -qE "^\s*${key}:" "$file" 2>/dev/null; then
    sed -i.bak "s|^\(\s*${key}:\).*|\1 ${value}|" "$file"
    rm -f "${file}.bak"
  else
    echo "  ${key}: ${value}" >> "$file"
  fi
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux"  ;;
    *)      echo "unknown" ;;
  esac
}

restart_service() {
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    sudo systemctl restart "$SERVICE_NAME"
    ok "Systemd 服务已重启"
  elif [ "$os_type" = "darwin" ] && launchctl list 2>/dev/null | grep -q "com.smtp-lite"; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    launchctl unload "$plist" 2>/dev/null || true
    launchctl load "$plist"
    ok "LaunchAgent 已重启"
  else
    local pid
    pid=$(pgrep -f "${INSTALL_DIR}/smtp-lite" 2>/dev/null | head -1 || true)
    if [ -n "$pid" ]; then
      kill "$pid" 2>/dev/null || true
      sleep 1
      nohup "$INSTALL_DIR/smtp-lite" >> "$INSTALL_DIR/smtp-lite.log" 2>&1 &
      ok "服务已重启 (PID: $!)"
    else
      warn "服务未运行，跳过重启"
    fi
  fi
}

# ── 主流程 ────────────────────────────────────────────────────

echo ""
echo -e "  ${W}SMTP Lite - 重置管理员密码${N}"
echo -e "  ${DIM}───────────────────────────────────────────${N}"

# 获取新密码：命令行参数或交互输入
if [ $# -ge 1 ]; then
  NEW_PASS="$1"
else
  while true; do
    echo -ne "  ${C}→${N} 新密码（至少 8 位）: "
    read -rs NEW_PASS; echo
    [ ${#NEW_PASS} -ge 8 ] || { warn "密码至少 8 位，请重新输入"; continue; }

    echo -ne "  ${C}→${N} 确认新密码: "
    read -rs CONFIRM_PASS; echo
    [ "$NEW_PASS" = "$CONFIRM_PASS" ] && break
    warn "两次密码不一致，请重新输入"
  done
fi

# 校验长度
[ ${#NEW_PASS} -ge 8 ] || err "密码至少 8 位"

# 写入明文密码（服务启动时自动迁移为 bcrypt hash）
yaml_set "password" "$NEW_PASS"
ok "密码已重置"

# 重启服务
echo ""
echo -ne "  ${C}→${N} 立即重启服务使密码生效？ ${DIM}[Y/n]${N}: "
read -r ans
ans="${ans:-y}"
if [[ "$ans" =~ ^[Yy] ]]; then
  restart_service
else
  warn "请手动重启服务使新密码生效"
fi

echo ""
