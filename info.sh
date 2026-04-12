#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 系统信息面板
#  用法: bash info.sh
# =============================================================
set -euo pipefail

# ── 颜色 ─────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; DIM='\033[2m'; N='\033[0m'

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
  echo -e "  ${R}✗${N} 未找到 config.yaml"; exit 1
fi

SERVICE_NAME="smtp-lite"

# ── 辅助函数 ──────────────────────────────────────────────────

yaml_get() {
  local key="$1" file="${2:-$CONFIG}"
  grep -E "^\s*${key}:" "$file" 2>/dev/null | head -1 | sed "s/.*${key}:\s*//" | sed 's/#.*//' | xargs
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux"  ;;
    *)      echo "unknown" ;;
  esac
}

get_pid() {
  pgrep -f "${INSTALL_DIR}/smtp-lite" 2>/dev/null | head -1 || true
}

get_version() {
  if [ -x "$INSTALL_DIR/smtp-lite" ]; then
    "$INSTALL_DIR/smtp-lite" --version 2>/dev/null || echo "unknown"
  elif [ -f "$INSTALL_DIR/internal/version/version.go" ]; then
    grep 'Version' "$INSTALL_DIR/internal/version/version.go" 2>/dev/null \
      | grep -o '"v[^"]*"' | tr -d '"' || echo "unknown"
  else
    echo "unknown"
  fi
}

get_nginx_conf() {
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "darwin" ]; then
    echo "/usr/local/etc/nginx/servers/smtp-lite.conf"
  else
    echo "/etc/nginx/sites-available/smtp-lite"
  fi
}

get_current_domain() {
  local conf
  conf=$(get_nginx_conf)
  if [ -f "$conf" ]; then
    grep "server_name" "$conf" 2>/dev/null | head -1 | awk '{print $2}' | tr -d ';'
  else
    echo "(未配置)"
  fi
}

row() {
  printf "  ${DIM}%-14s${N}→ %b\n" "$1" "$2"
}

divider() { echo -e "  ${DIM}───────────────────────────────────────────${N}"; }

human_size() {
  du -h "$1" 2>/dev/null | awk '{print $1}'
}

# ── 进程运行时长 ──────────────────────────────────────────────
get_uptime() {
  local pid="$1" os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ]; then
    local start_sec now_sec diff
    start_sec=$(stat -c %Y "/proc/$pid" 2>/dev/null || echo "")
    [ -z "$start_sec" ] && return
    now_sec=$(date +%s)
    diff=$((now_sec - start_sec))
  elif [ "$os_type" = "darwin" ]; then
    local start_ts diff
    start_ts=$(ps -p "$pid" -o lstart= 2>/dev/null || echo "")
    [ -z "$start_ts" ] && return
    local start_sec now_sec
    start_sec=$(date -j -f "%c" "$start_ts" +%s 2>/dev/null || echo "")
    [ -z "$start_sec" ] && return
    now_sec=$(date +%s)
    diff=$((now_sec - start_sec))
  else
    return
  fi
  local d=$((diff/86400)) h=$(( (diff%86400)/3600 )) m=$(( (diff%3600)/60 ))
  if [ $d -gt 0 ]; then
    echo "${d}天${h}时${m}分"
  elif [ $h -gt 0 ]; then
    echo "${h}时${m}分"
  else
    echo "${m}分钟"
  fi
}

# ── 数据库统计 ────────────────────────────────────────────────
db_count() {
  local table="$1" db_file="$INSTALL_DIR/smtp-lite.db"
  local driver
  driver=$(yaml_get "driver")
  [ -z "$driver" ] && driver="sqlite"

  if [ "$driver" = "sqlite" ] && [ -f "$db_file" ] && command -v sqlite3 &>/dev/null; then
    sqlite3 "$db_file" "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "?"
  elif [ "$driver" = "mysql" ] && command -v mysql &>/dev/null; then
    local db_host db_port db_user db_pass db_name
    db_host=$(yaml_get "host"); db_port=$(yaml_get "port")
    db_user=$(yaml_get "user"); db_pass=$(yaml_get "password")
    db_name=$(yaml_get "name")
    mysql -h"$db_host" -P"${db_port:-3306}" -u"$db_user" -p"$db_pass" "$db_name" \
      -sNe "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "?"
  else
    echo "?"
  fi
}

# ══════════════════════════════════════════════════════════════
#  输出面板
# ══════════════════════════════════════════════════════════════

echo ""
echo -e "  ${W}SMTP Lite 系统信息${N}"
divider

# ── 基本信息 ──────────────────────────────────────────────────
version=$(get_version)
port=$(yaml_get "port")
username=$(yaml_get "username")
domain=$(get_current_domain)

row "版本"      "${C}${version}${N}"
row "安装目录"  "${C}${INSTALL_DIR}${N}"
row "监听端口"  "${C}${port}${N}"
row "管理账号"  "${C}${username}${N}"
row "绑定域名"  "${C}${domain}${N}"

divider

# ── 服务状态 ──────────────────────────────────────────────────
pid=$(get_pid)
if [ -n "$pid" ]; then
  uptime_str=$(get_uptime "$pid")
  if [ -n "$uptime_str" ]; then
    row "服务状态"  "${G}运行中${N}  PID: ${pid}  已运行: ${uptime_str}"
  else
    row "服务状态"  "${G}运行中${N}  PID: ${pid}"
  fi

  # 内存占用
  mem=$(ps -p "$pid" -o rss= 2>/dev/null | xargs || true)
  if [ -n "$mem" ]; then
    mem_mb=$(awk "BEGIN{printf \"%.1f\", $mem/1024}")
    row "内存占用"  "${C}${mem_mb} MB${N}"
  fi
else
  row "服务状态"  "${R}已停止${N}"
fi

# 开机自启
os_type=$(detect_os)
if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
  enabled=$(systemctl is-enabled "$SERVICE_NAME" 2>/dev/null || echo "未配置")
  if [ "$enabled" = "enabled" ]; then
    row "开机自启"  "${G}已启用${N} (systemd)"
  else
    row "开机自启"  "${DIM}${enabled}${N}"
  fi
elif [ "$os_type" = "darwin" ] && [ -f "$HOME/Library/LaunchAgents/com.smtp-lite.plist" ]; then
  row "开机自启"  "${G}已配置${N} (launchd)"
fi

divider

# ── Nginx & SSL ──────────────────────────────────────────────
nginx_conf=$(get_nginx_conf)
if command -v nginx &>/dev/null; then
  if [ -f "$nginx_conf" ]; then
    row "Nginx"     "${G}已配置${N}"
  else
    row "Nginx"     "${Y}已安装但未配置${N}"
  fi
else
  row "Nginx"       "${DIM}未安装${N}"
fi

if [ -f "$nginx_conf" ] && grep -q "ssl_certificate" "$nginx_conf" 2>/dev/null; then
  cert_domain=$(grep "server_name" "$nginx_conf" | head -1 | awk '{print $2}' | tr -d ';')
  cert_path="/etc/letsencrypt/live/${cert_domain}/fullchain.pem"
  if [ -f "$cert_path" ]; then
    expiry=$(openssl x509 -in "$cert_path" -noout -enddate 2>/dev/null | cut -d= -f2 || echo "未知")
    # 判断是否即将过期（30 天内）
    expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$expiry" +%s 2>/dev/null || echo "0")
    now_epoch=$(date +%s)
    days_left=$(( (expiry_epoch - now_epoch) / 86400 ))
    if [ "$days_left" -le 0 ] 2>/dev/null; then
      row "SSL 证书"  "${R}已过期${N}  ${DIM}${expiry}${N}"
    elif [ "$days_left" -le 30 ] 2>/dev/null; then
      row "SSL 证书"  "${Y}${days_left} 天后过期${N}  ${DIM}${expiry}${N}"
    else
      row "SSL 证书"  "${G}有效${N}  剩余 ${days_left} 天  ${DIM}${expiry}${N}"
    fi
  else
    row "SSL 证书"  "${Y}证书文件不存在${N}"
  fi
else
  row "SSL 证书"    "${DIM}未配置${N}"
fi

divider

# ── 数据库 ────────────────────────────────────────────────────
db_driver=$(yaml_get "driver")
[ -z "$db_driver" ] && db_driver="sqlite"

if [ "$db_driver" = "sqlite" ]; then
  db_file="$INSTALL_DIR/smtp-lite.db"
  if [ -f "$db_file" ]; then
    row "数据库"    "${C}SQLite${N}  ${DIM}$(human_size "$db_file")${N}"
  else
    row "数据库"    "${C}SQLite${N}  ${Y}文件不存在${N}"
  fi
else
  db_host=$(yaml_get "host")
  db_name=$(yaml_get "name")
  row "数据库"      "${C}MySQL${N}  ${DIM}${db_host} / ${db_name}${N}"
fi

# 业务数据统计
smtp_count=$(db_count "smtp_accounts")
apikey_count=$(db_count "api_keys")
log_count=$(db_count "send_logs")
template_count=$(db_count "email_templates")

row "SMTP 账号"   "${C}${smtp_count}${N} 个"
row "API 密钥"    "${C}${apikey_count}${N} 个"
row "发信记录"    "${C}${log_count}${N} 条"
row "邮件模板"    "${C}${template_count}${N} 个"

divider

# ── 日志 & 磁盘 ──────────────────────────────────────────────
log_file="$INSTALL_DIR/smtp-lite.log"
if [ -f "$log_file" ]; then
  row "日志文件"  "${C}$(human_size "$log_file")${N}  ${DIM}${log_file}${N}"
fi

# 磁盘总占用
if [ -d "$INSTALL_DIR" ]; then
  dir_size=$(du -sh "$INSTALL_DIR" 2>/dev/null | awk '{print $1}')
  row "磁盘占用"  "${C}${dir_size}${N}  ${DIM}${INSTALL_DIR}${N}"
fi

# ── 系统资源 ──────────────────────────────────────────────────
divider
if [ "$os_type" = "linux" ]; then
  # 磁盘使用率
  disk_usage=$(df -h / 2>/dev/null | awk 'NR==2{printf "%s / %s (%s)", $3, $2, $5}')
  row "系统磁盘"  "${C}${disk_usage}${N}"
  # 内存
  mem_info=$(free -h 2>/dev/null | awk 'NR==2{printf "%s / %s", $3, $2}')
  row "系统内存"  "${C}${mem_info}${N}"
  # 负载
  load=$(uptime 2>/dev/null | awk -F'load average:' '{print $2}' | xargs)
  row "系统负载"  "${C}${load}${N}"
elif [ "$os_type" = "darwin" ]; then
  disk_usage=$(df -h / 2>/dev/null | awk 'NR==2{printf "%s / %s (%s)", $3, $2, $5}')
  row "系统磁盘"  "${C}${disk_usage}${N}"
  load=$(sysctl -n vm.loadavg 2>/dev/null | tr -d '{}' | xargs)
  row "系统负载"  "${C}${load}${N}"
fi

echo ""

# ── 最近日志（最后 5 行）────────────────────────────────────
if [ -f "$log_file" ]; then
  echo -e "  ${W}最近日志${N}"
  divider
  tail -5 "$log_file" 2>/dev/null | while IFS= read -r line; do
    echo -e "  ${DIM}${line}${N}"
  done
  echo ""
fi
