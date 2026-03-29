#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 交互式管理工具 v2.1
#  用法: smtp-lite              进入交互式菜单
#        smtp-lite <命令>       直接执行命令（兼容旧版）
# =============================================================
set -euo pipefail

# ── 自动检测安装目录 ──────────────────────────────────────────
_SCRIPT="${BASH_SOURCE[0]}"
while [ -L "$_SCRIPT" ]; do
  _DIR="$(cd "$(dirname "$_SCRIPT")" && pwd)"
  _SCRIPT="$(readlink "$_SCRIPT")"
  [[ "$_SCRIPT" != /* ]] && _SCRIPT="$_DIR/$_SCRIPT"
done
_REAL_DIR="$(cd "$(dirname "$_SCRIPT")" && pwd)"

if [ -f "$_REAL_DIR/config.yaml" ]; then
  INSTALL_DIR="$_REAL_DIR"
elif [ -f "$HOME/smtp-lite/config.yaml" ]; then
  INSTALL_DIR="$HOME/smtp-lite"
else
  echo "错误: 未找到 smtp-lite 安装目录，请在安装目录下运行或确认已安装"
  exit 1
fi

CONFIG="$INSTALL_DIR/config.yaml"
SERVICE_NAME="smtp-lite"
GITHUB_REPO="DoBestone/smtp-lite"

# ── 颜色 ─────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; DIM='\033[2m'; N='\033[0m'

# ── 日志函数 ──────────────────────────────────────────────────
info()    { echo -e "  ${B}▸${N} $*"; }
ok()      { echo -e "  ${G}✓${N} $*"; }
warn()    { echo -e "  ${Y}⚠${N}  $*"; }
err()     { echo -e "  ${R}✗${N} $*"; exit 1; }
divider() { echo -e "  ${DIM}───────────────────────────────────────────${N}"; }

# ── 辅助函数 ──────────────────────────────────────────────────

yaml_get() {
  local key="$1" file="${2:-$CONFIG}"
  grep -E "^\s*${key}:" "$file" 2>/dev/null | head -1 | sed "s/.*${key}:\s*//" | sed 's/#.*//' | xargs
}

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
    echo "(未配置域名)"
  fi
}

prompt_input() {
  local __var="$1" __prompt="$2" __default="${3:-}"
  local __display_default=""
  [ -n "$__default" ] && __display_default=" ${DIM}[${__default}]${N}"
  echo -ne "  ${C}→${N} ${__prompt}${__display_default}: "
  read -r __val
  eval "$__var='${__val:-$__default}'"
}

prompt_secret() {
  local __var="$1" __prompt="$2"
  echo -ne "  ${C}→${N} ${__prompt}: "
  read -rs __val; echo
  eval "$__var='$__val'"
}

prompt_yn() {
  local __prompt="$1" __default="${2:-n}"
  local __hint; [ "$__default" = "y" ] && __hint="[Y/n]" || __hint="[y/N]"
  echo -ne "  ${C}→${N} ${__prompt} ${DIM}${__hint}${N}: "
  read -r __ans
  __ans="${__ans:-$__default}"
  [[ "$__ans" =~ ^[Yy] ]]
}

# 等待按键返回菜单
press_enter() {
  echo ""
  echo -ne "  ${DIM}按 Enter 返回菜单...${N}"
  read -r
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
    pid=$(get_pid)
    if [ -n "$pid" ]; then
      kill "$pid" 2>/dev/null || true
      sleep 2
    fi
    cd "$INSTALL_DIR"
    nohup "$INSTALL_DIR/smtp-lite" >> "$INSTALL_DIR/smtp-lite.log" 2>&1 &
    sleep 1
    local new_pid
    new_pid=$(get_pid)
    if [ -n "$new_pid" ]; then
      ok "服务已重启 (PID: ${new_pid})"
    else
      err "服务启动失败，请检查日志: ${INSTALL_DIR}/smtp-lite.log"
    fi
  fi
}

reload_nginx() {
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ]; then
    sudo nginx -t && sudo systemctl reload nginx
  else
    nginx -t && (brew services restart nginx 2>/dev/null || nginx -s reload)
  fi
}

# ══════════════════════════════════════════════════════════════
#  命令实现
# ══════════════════════════════════════════════════════════════

cmd_start() {
  echo -e "\n  ${W}── 启动服务 ──${N}"
  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    warn "服务已在运行 (PID: ${pid})"
    return
  fi
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
    sudo systemctl start "$SERVICE_NAME" && ok "Systemd 服务已启动"
  elif [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    if [ -f "$plist" ]; then
      launchctl load "$plist" 2>/dev/null && ok "LaunchAgent 已启动"
    else
      cd "$INSTALL_DIR"
      nohup "$INSTALL_DIR/smtp-lite" >> "$INSTALL_DIR/smtp-lite.log" 2>&1 &
      sleep 1
      pid=$(get_pid)
      [ -n "$pid" ] && ok "服务已启动 (PID: ${pid})" || err "启动失败"
    fi
  else
    cd "$INSTALL_DIR"
    nohup "$INSTALL_DIR/smtp-lite" >> "$INSTALL_DIR/smtp-lite.log" 2>&1 &
    sleep 1
    pid=$(get_pid)
    [ -n "$pid" ] && ok "服务已启动 (PID: ${pid})" || err "启动失败"
  fi
}

cmd_stop() {
  echo -e "\n  ${W}── 停止服务 ──${N}"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    sudo systemctl stop "$SERVICE_NAME" && ok "Systemd 服务已停止"
  elif [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    if launchctl list 2>/dev/null | grep -q "com.smtp-lite"; then
      launchctl unload "$plist" 2>/dev/null && ok "LaunchAgent 已停止"
    fi
  fi
  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    kill "$pid" 2>/dev/null || true
    sleep 1
    ok "进程已停止 (PID: ${pid})"
  else
    warn "服务未在运行"
  fi
}

cmd_restart() {
  echo -e "\n  ${W}── 重启服务 ──${N}"
  restart_service
}

cmd_status() {
  echo ""
  echo -e "  ${W}SMTP Lite 状态${N}"
  divider

  local version
  version=$(get_version)
  echo -e "  ${DIM}版本        ${N}→ ${C}${version}${N}"
  echo -e "  ${DIM}安装目录    ${N}→ ${C}${INSTALL_DIR}${N}"

  local port
  port=$(yaml_get "port")
  echo -e "  ${DIM}监听端口    ${N}→ ${C}${port}${N}"

  local username
  username=$(yaml_get "username")
  echo -e "  ${DIM}管理账号    ${N}→ ${C}${username}${N}"

  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    echo -e "  ${DIM}服务状态    ${N}→ ${G}运行中${N} (PID: ${pid})"
  else
    echo -e "  ${DIM}服务状态    ${N}→ ${R}已停止${N}"
  fi

  local domain
  domain=$(get_current_domain)
  echo -e "  ${DIM}绑定域名    ${N}→ ${C}${domain}${N}"

  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ -f "$nginx_conf" ] && grep -q "ssl_certificate" "$nginx_conf" 2>/dev/null; then
    local cert_domain
    cert_domain=$(grep "server_name" "$nginx_conf" | head -1 | awk '{print $2}' | tr -d ';')
    local cert_path="/etc/letsencrypt/live/${cert_domain}/fullchain.pem"
    if [ -f "$cert_path" ]; then
      local expiry
      expiry=$(sudo openssl x509 -in "$cert_path" -noout -enddate 2>/dev/null | cut -d= -f2 || echo "未知")
      echo -e "  ${DIM}SSL 证书    ${N}→ ${G}已配置${N} (到期: ${expiry})"
    else
      echo -e "  ${DIM}SSL 证书    ${N}→ ${Y}证书文件不存在${N}"
    fi
  else
    echo -e "  ${DIM}SSL 证书    ${N}→ ${DIM}未配置${N}"
  fi

  if command -v nginx &>/dev/null; then
    if [ -f "$nginx_conf" ]; then
      echo -e "  ${DIM}Nginx       ${N}→ ${G}已配置${N}"
    else
      echo -e "  ${DIM}Nginx       ${N}→ ${Y}已安装但未配置${N}"
    fi
  else
    echo -e "  ${DIM}Nginx       ${N}→ ${DIM}未安装${N}"
  fi

  # 数据库信息
  local db_driver
  db_driver=$(yaml_get "driver")
  [ -z "$db_driver" ] && db_driver="sqlite"
  echo -e "  ${DIM}数据库      ${N}→ ${C}${db_driver}${N}"

  local db_file="$INSTALL_DIR/smtp-lite.db"
  if [ "$db_driver" = "sqlite" ] && [ -f "$db_file" ]; then
    local db_size
    db_size=$(du -h "$db_file" | awk '{print $1}')
    echo -e "  ${DIM}数据库大小  ${N}→ ${C}${db_size}${N}"
  fi

  local log_file="$INSTALL_DIR/smtp-lite.log"
  if [ -f "$log_file" ]; then
    local log_size
    log_size=$(du -h "$log_file" | awk '{print $1}')
    echo -e "  ${DIM}日志大小    ${N}→ ${C}${log_size}${N}"
  fi

  divider

  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && systemctl is-enabled "$SERVICE_NAME" &>/dev/null; then
    local enabled
    enabled=$(systemctl is-enabled "$SERVICE_NAME" 2>/dev/null || echo "unknown")
    echo -e "  ${DIM}开机自启    ${N}→ ${C}${enabled}${N} (systemd)"
  elif [ "$os_type" = "darwin" ] && [ -f "$HOME/Library/LaunchAgents/com.smtp-lite.plist" ]; then
    echo -e "  ${DIM}开机自启    ${N}→ ${G}已配置${N} (launchd)"
  fi
  echo ""
}

cmd_port() {
  echo -e "\n  ${W}── 修改端口 ──${N}"
  local current_port
  current_port=$(yaml_get "port")
  info "当前端口: ${W}${current_port}${N}"

  local new_port
  while true; do
    prompt_input new_port "新端口" "$current_port"
    [[ "$new_port" =~ ^[0-9]+$ ]] && [ "$new_port" -ge 1 ] && [ "$new_port" -le 65535 ] && break
    warn "请输入有效端口号 (1-65535)"
  done

  if [ "$new_port" = "$current_port" ]; then
    warn "端口未变更"
    return
  fi

  yaml_set "port" "$new_port"
  ok "端口已修改: ${current_port} → ${new_port}"

  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ -f "$nginx_conf" ]; then
    info "同步更新 Nginx 配置..."
    sudo sed -i.bak "s|proxy_pass.*http://127.0.0.1:[0-9]*|proxy_pass         http://127.0.0.1:${new_port}|g" "$nginx_conf"
    sudo rm -f "${nginx_conf}.bak"
    reload_nginx && ok "Nginx 已重载"
  fi

  if prompt_yn "立即重启服务使端口生效？" "y"; then
    restart_service
  else
    warn "请手动重启服务"
  fi
}

cmd_domain() {
  echo -e "\n  ${W}── 换绑域名 ──${N}"
  local current_domain
  current_domain=$(get_current_domain)
  info "当前域名: ${W}${current_domain}${N}"

  local new_domain
  while true; do
    prompt_input new_domain "新域名（如 smtp.example.com）" ""
    [[ "$new_domain" =~ ^[a-zA-Z0-9]([a-zA-Z0-9\-]*\.)+[a-zA-Z]{2,}$ ]] && break
    warn "域名格式不正确，请重新输入"
  done

  if [ "$new_domain" = "$current_domain" ]; then
    warn "域名未变更"
    return
  fi

  local nginx_conf
  nginx_conf=$(get_nginx_conf)

  if ! command -v nginx &>/dev/null; then
    err "未安装 Nginx，无法配置域名"
  fi

  local port
  port=$(yaml_get "port")

  local use_ssl=false
  if prompt_yn "是否为新域名申请 SSL 证书？" "y"; then
    use_ssl=true
    if ! command -v certbot &>/dev/null; then
      err "未安装 certbot，请先安装: sudo apt install certbot"
    fi
  fi

  if $use_ssl; then
    local webroot="/var/www/certbot"
    sudo mkdir -p "${webroot}/.well-known/acme-challenge"
    sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
      || sudo chown -R "$(whoami)" "$webroot"

    info "配置临时 HTTP 用于证书验证..."
    local tee_cmd="sudo tee"
    [ "$(detect_os)" = "darwin" ] && tee_cmd="tee"
    $tee_cmd "$nginx_conf" > /dev/null <<EOF
server {
    listen 80;
    server_name ${new_domain};

    location /.well-known/acme-challenge/ {
        root ${webroot};
        allow all;
    }

    location / {
        proxy_pass         http://127.0.0.1:${port};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }
}
EOF
    if [ "$(detect_os)" = "linux" ]; then
      sudo ln -sf "$nginx_conf" "/etc/nginx/sites-enabled/smtp-lite"
    fi
    reload_nginx

    info "申请 SSL 证书..."
    local certbot_email=""
    prompt_input certbot_email "证书通知邮箱（可留空）" ""
    local certbot_args=("certonly" "--webroot" "-w" "$webroot" "-d" "$new_domain" "--non-interactive" "--agree-tos")
    [ -n "$certbot_email" ] && certbot_args+=("--email" "$certbot_email") \
                            || certbot_args+=("--register-unsafely-without-email")

    if sudo certbot "${certbot_args[@]}"; then
      ok "SSL 证书申请成功"
      local cert_dir="/etc/letsencrypt/live/${new_domain}"
      $tee_cmd "$nginx_conf" > /dev/null <<EOF
server {
    listen 80;
    server_name ${new_domain};

    location /.well-known/acme-challenge/ {
        root ${webroot};
        allow all;
    }

    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name ${new_domain};

    ssl_certificate     ${cert_dir}/fullchain.pem;
    ssl_certificate_key ${cert_dir}/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache   shared:SSL:10m;
    ssl_session_timeout 1d;

    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;

    location / {
        proxy_pass         http://127.0.0.1:${port};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF
      reload_nginx
      ok "域名已切换: ${current_domain} → ${new_domain} (HTTPS)"
    else
      warn "SSL 申请失败，保留 HTTP 配置"
      ok "域名已切换: ${current_domain} → ${new_domain} (HTTP)"
    fi
  else
    local tee_cmd="sudo tee"
    [ "$(detect_os)" = "darwin" ] && tee_cmd="tee"
    $tee_cmd "$nginx_conf" > /dev/null <<EOF
server {
    listen 80;
    server_name ${new_domain};

    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;

    location / {
        proxy_pass         http://127.0.0.1:${port};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF
    if [ "$(detect_os)" = "linux" ]; then
      sudo ln -sf "$nginx_conf" "/etc/nginx/sites-enabled/smtp-lite"
    fi
    reload_nginx
    ok "域名已切换: ${current_domain} → ${new_domain} (HTTP)"
  fi
}

cmd_password() {
  echo -e "\n  ${W}── 修改管理员密码 ──${N}"
  local current_user
  current_user=$(yaml_get "username")
  info "当前账号: ${W}${current_user}${N}"

  local new_pass confirm_pass
  while true; do
    prompt_secret new_pass "新密码（至少 8 位）"
    [ ${#new_pass} -ge 8 ] || { warn "密码至少 8 位"; continue; }
    prompt_secret confirm_pass "确认新密码"
    [ "$new_pass" = "$confirm_pass" ] && break
    warn "两次密码不一致，请重新输入"
  done

  yaml_set "password" "$new_pass"
  ok "密码已修改"

  if prompt_yn "立即重启服务使密码生效？" "y"; then
    restart_service
  fi
}

cmd_username() {
  echo -e "\n  ${W}── 修改管理员用户名 ──${N}"
  local current_user
  current_user=$(yaml_get "username")
  info "当前用户名: ${W}${current_user}${N}"

  local new_user
  prompt_input new_user "新用户名" "$current_user"

  if [ "$new_user" = "$current_user" ]; then
    warn "用户名未变更"
    return
  fi

  yaml_set "username" "$new_user"
  ok "用户名已修改: ${current_user} → ${new_user}"

  if prompt_yn "立即重启服务使更改生效？" "y"; then
    restart_service
  fi
}

# ── SSL 子命令 ────────────────────────────────────────────────

cmd_ssl_status() {
  echo -e "\n  ${W}── SSL 证书状态 ──${N}"
  local domain
  domain=$(get_current_domain)
  if [ "$domain" = "(未配置域名)" ]; then
    warn "未配置域名，无 SSL 信息"
    return
  fi
  local cert_path="/etc/letsencrypt/live/${domain}/fullchain.pem"
  if [ -f "$cert_path" ]; then
    ok "域名: ${domain}"
    local info_output
    info_output=$(sudo openssl x509 -in "$cert_path" -noout -subject -issuer -dates 2>/dev/null || echo "无法读取证书信息")
    echo "$info_output" | while IFS= read -r line; do
      echo -e "  ${DIM}${line}${N}"
    done
  else
    warn "未找到证书文件: ${cert_path}"
  fi
}

cmd_ssl_apply() {
  echo -e "\n  ${W}── 申请 SSL 证书 ──${N}"
  local domain
  domain=$(get_current_domain)
  if [ "$domain" = "(未配置域名)" ]; then
    warn "请先配置域名"
    return
  fi

  if ! command -v certbot &>/dev/null; then
    err "未安装 certbot，请先安装"
  fi

  local webroot="/var/www/certbot"
  sudo mkdir -p "${webroot}/.well-known/acme-challenge"
  sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
    || sudo chown -R "$(whoami)" "$webroot"

  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if ! grep -q "acme-challenge" "$nginx_conf" 2>/dev/null; then
    info "在 Nginx 配置中添加验证路径..."
    sudo sed -i.bak "/location \//i\\
    location /.well-known/acme-challenge/ {\\
        root ${webroot};\\
        allow all;\\
    }\\
" "$nginx_conf"
    sudo rm -f "${nginx_conf}.bak"
    reload_nginx
  fi

  local email=""
  prompt_input email "证书通知邮箱（可留空）" ""
  local certbot_args=("certonly" "--webroot" "-w" "$webroot" "-d" "$domain" "--non-interactive" "--agree-tos")
  [ -n "$email" ] && certbot_args+=("--email" "$email") \
                   || certbot_args+=("--register-unsafely-without-email")

  if sudo certbot "${certbot_args[@]}"; then
    ok "SSL 证书申请成功"
    cmd_domain
  else
    err "SSL 证书申请失败"
  fi
}

cmd_ssl_renew() {
  echo -e "\n  ${W}── 续期 SSL 证书 ──${N}"
  local webroot="/var/www/certbot"
  if sudo certbot renew --webroot -w "$webroot" --deploy-hook "nginx -s reload" 2>&1; then
    ok "证书续期检查完成"
  else
    err "证书续期失败"
  fi
}

cmd_ssl_remove() {
  echo -e "\n  ${W}── 移除 SSL 配置 ──${N}"
  local domain
  domain=$(get_current_domain)
  if [ "$domain" = "(未配置域名)" ]; then
    warn "未配置域名"
    return
  fi

  prompt_yn "确认移除 SSL 配置，降级为 HTTP？" "n" || return

  local port
  port=$(yaml_get "port")
  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  local tee_cmd="sudo tee"
  [ "$(detect_os)" = "darwin" ] && tee_cmd="tee"
  $tee_cmd "$nginx_conf" > /dev/null <<EOF
server {
    listen 80;
    server_name ${domain};

    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
        allow all;
    }

    location / {
        proxy_pass         http://127.0.0.1:${port};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF
  reload_nginx
  ok "已移除 SSL，降级为 HTTP"
}

cmd_ssl() {
  local subcmd="${1:-}"
  case "$subcmd" in
    renew)  cmd_ssl_renew ;;
    status) cmd_ssl_status ;;
    apply)  cmd_ssl_apply ;;
    remove) cmd_ssl_remove ;;
    *)      menu_ssl ;;
  esac
}

# ── 日志子命令 ────────────────────────────────────────────────

cmd_log_tail() {
  echo -e "\n  ${W}── 实时日志 ──${N} ${DIM}(Ctrl+C 退出)${N}"
  local log_file="$INSTALL_DIR/smtp-lite.log"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    sudo journalctl -u "$SERVICE_NAME" -f --no-pager
  elif [ -f "$log_file" ]; then
    tail -f "$log_file"
  else
    warn "未找到日志文件"
  fi
}

cmd_log_show() {
  local lines="${1:-50}"
  echo -e "\n  ${W}── 最近 ${lines} 行日志 ──${N}"
  local log_file="$INSTALL_DIR/smtp-lite.log"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    sudo journalctl -u "$SERVICE_NAME" -n "$lines" --no-pager
  elif [ -f "$log_file" ]; then
    tail -n "$lines" "$log_file"
  else
    warn "未找到日志文件"
  fi
}

cmd_log_clear() {
  echo -e "\n  ${W}── 清空日志 ──${N}"
  prompt_yn "确认清空日志文件？" "n" || return
  local log_file="$INSTALL_DIR/smtp-lite.log"
  if [ -f "$log_file" ]; then
    > "$log_file"
    ok "日志已清空"
  fi
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && command -v journalctl &>/dev/null; then
    sudo journalctl --vacuum-time=1s -u "$SERVICE_NAME" &>/dev/null || true
    ok "Journald 日志已清理"
  fi
}

cmd_log() {
  local subcmd="${1:-tail}"
  case "$subcmd" in
    tail|follow|-f) cmd_log_tail ;;
    show|view)      cmd_log_show "${2:-50}" ;;
    clear)          cmd_log_clear ;;
    *)              menu_log ;;
  esac
}

# ── update ────────────────────────────────────────────────────
cmd_update() {
  echo -e "\n  ${W}── 检查更新 ──${N}"

  # 优先使用 update.sh 脚本
  if [ -f "$INSTALL_DIR/update.sh" ]; then
    local force_flag=""
    [ "${1:-}" = "--force" ] && force_flag="--force"
    bash "$INSTALL_DIR/update.sh" $force_flag
    return
  fi

  # 回退：内联更新逻辑
  local force=false
  [ "${1:-}" = "--force" ] && force=true

  cd "$INSTALL_DIR"
  local current
  current=$(get_version)
  info "当前版本: ${W}${current}${N}"

  local latest
  latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | grep -o '"v[^"]*"' | tr -d '"' || echo "")

  if [ -n "$latest" ]; then
    info "最新版本: ${W}${latest}${N}"
    if [ "$latest" = "$current" ] && [ "$force" = false ]; then
      ok "已是最新版本"
      return
    fi
  else
    warn "无法获取最新版本号"
  fi

  info "拉取代码..."
  git pull || err "git pull 失败"
  ok "代码更新完成"

  info "重新编译..."
  if [ -d "$INSTALL_DIR/frontend" ] && command -v npm &>/dev/null; then
    (cd "$INSTALL_DIR/frontend" && npm install --silent && npm run build) || warn "前端构建失败，跳过"
  fi
  go build -o smtp-lite ./cmd/server/ || err "编译失败"
  local new_ver
  new_ver=$(get_version)
  ok "编译完成 → ${new_ver}"
  restart_service
}

# ── config ────────────────────────────────────────────────────

cmd_config_show() {
  echo -e "\n  ${W}── 当前配置 ──${N}"
  echo ""
  while IFS= read -r line; do
    if echo "$line" | grep -qiE "(password|secret|key):" ; then
      local key val
      key=$(echo "$line" | cut -d: -f1)
      val=$(echo "$line" | cut -d: -f2-)
      val=$(echo "$val" | xargs)
      if [ ${#val} -gt 4 ]; then
        echo -e "  ${DIM}${key}:${N} ${val:0:4}$( printf '*%.0s' $(seq 1 $((${#val}-4))) )"
      else
        echo -e "  ${DIM}${key}:${N} ****"
      fi
    else
      echo -e "  ${DIM}${line}${N}"
    fi
  done < "$CONFIG"
  echo ""
}

cmd_config_edit() {
  local editor="${EDITOR:-vi}"
  echo -e "\n  ${W}── 编辑配置 ──${N}"
  info "使用编辑器: ${editor}"
  warn "修改后需重启服务"
  $editor "$CONFIG"
}

cmd_config_reset_jwt() {
  echo -e "\n  ${W}── 重置 JWT Secret ──${N}"
  prompt_yn "重置 JWT Secret 将使所有现有登录失效，确认？" "n" || return
  local new_secret
  new_secret=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
  yaml_set "secret" "$new_secret"
  ok "JWT Secret 已重置"
  restart_service
}

cmd_config_reset_enc() {
  echo -e "\n  ${W}── 重置加密密钥 ──${N}"
  warn "重置加密密钥后，已加密的 SMTP 密码将无法解密！"
  warn "建议先导出 SMTP 账号信息，重置后重新添加"
  prompt_yn "确认重置加密密钥？此操作不可逆！" "n" || return
  local new_key
  new_key=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
  yaml_set "key" "$new_key"
  ok "加密密钥已重置"
  restart_service
}

cmd_config() {
  local subcmd="${1:-show}"
  case "$subcmd" in
    show|view)           cmd_config_show ;;
    edit)                cmd_config_edit ;;
    reset-jwt)           cmd_config_reset_jwt ;;
    reset-encryption-key) cmd_config_reset_enc ;;
    *)                   menu_config ;;
  esac
}

# ── backup ────────────────────────────────────────────────────
cmd_backup() {
  echo -e "\n  ${W}── 备份数据 ──${N}"
  local backup_dir="${1:-$INSTALL_DIR/backups}"
  local timestamp
  timestamp=$(date +%Y%m%d_%H%M%S)
  local backup_path="${backup_dir}/smtp-lite_${timestamp}"

  mkdir -p "$backup_path"

  cp "$CONFIG" "$backup_path/config.yaml"
  ok "配置文件已备份"

  local db_file="$INSTALL_DIR/smtp-lite.db"
  if [ -f "$db_file" ]; then
    cp "$db_file" "$backup_path/smtp-lite.db"
    ok "数据库已备份"
  fi

  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ -f "$nginx_conf" ]; then
    cp "$nginx_conf" "$backup_path/nginx-smtp-lite.conf"
    ok "Nginx 配置已备份"
  fi

  local archive="${backup_dir}/smtp-lite_${timestamp}.tar.gz"
  tar -czf "$archive" -C "$backup_dir" "smtp-lite_${timestamp}"
  rm -rf "$backup_path"

  ok "备份完成: ${archive}"
  local size
  size=$(du -h "$archive" | awk '{print $1}')
  info "大小: ${size}"
}

# ── restore ───────────────────────────────────────────────────
cmd_restore() {
  echo -e "\n  ${W}── 恢复数据 ──${N}"
  local backup_dir="$INSTALL_DIR/backups"

  if [ -n "${1:-}" ] && [ -f "$1" ]; then
    local archive="$1"
  else
    if [ ! -d "$backup_dir" ] || [ -z "$(ls -A "$backup_dir"/*.tar.gz 2>/dev/null)" ]; then
      warn "未找到备份文件"
      return
    fi

    echo ""
    info "可用备份:"
    local i=0
    local -a backups=()
    for f in "$backup_dir"/smtp-lite_*.tar.gz; do
      i=$((i+1))
      backups+=("$f")
      local fname size
      fname=$(basename "$f")
      size=$(du -h "$f" | awk '{print $1}')
      echo -e "    ${C}${i}${N}) ${fname} (${size})"
    done
    echo ""

    local choice
    prompt_input choice "选择备份编号" ""
    if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#backups[@]} ]; then
      err "无效选择"
    fi
    local archive="${backups[$((choice-1))]}"
  fi

  prompt_yn "确认恢复？当前配置和数据库将被覆盖" "n" || return

  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    info "停止服务..."
    cmd_stop
  fi

  local tmp_dir
  tmp_dir=$(mktemp -d)
  tar -xzf "$archive" -C "$tmp_dir"
  local extracted
  extracted=$(ls -d "$tmp_dir"/smtp-lite_* | head -1)

  if [ -f "$extracted/config.yaml" ]; then
    cp "$extracted/config.yaml" "$CONFIG"
    ok "配置文件已恢复"
  fi

  if [ -f "$extracted/smtp-lite.db" ]; then
    cp "$extracted/smtp-lite.db" "$INSTALL_DIR/smtp-lite.db"
    ok "数据库已恢复"
  fi

  if [ -f "$extracted/nginx-smtp-lite.conf" ]; then
    local nginx_conf
    nginx_conf=$(get_nginx_conf)
    sudo cp "$extracted/nginx-smtp-lite.conf" "$nginx_conf"
    if [ "$(detect_os)" = "linux" ]; then
      sudo ln -sf "$nginx_conf" "/etc/nginx/sites-enabled/smtp-lite"
    fi
    reload_nginx && ok "Nginx 配置已恢复并重载"
  fi

  rm -rf "$tmp_dir"
  restart_service
  ok "数据恢复完成"
}

# ── nginx ─────────────────────────────────────────────────────

cmd_nginx_status() {
  echo -e "\n  ${W}── Nginx 状态 ──${N}"
  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if ! command -v nginx &>/dev/null; then
    warn "Nginx 未安装"
    return
  fi
  ok "Nginx 已安装: $(nginx -v 2>&1 | grep -o '[0-9].*' || echo 'unknown')"
  if [ -f "$nginx_conf" ]; then
    ok "配置文件: ${nginx_conf}"
    echo ""
    cat "$nginx_conf"
  else
    warn "未找到 smtp-lite Nginx 配置"
  fi
}

cmd_nginx_reload() {
  echo -e "\n  ${W}── 重载 Nginx ──${N}"
  reload_nginx && ok "Nginx 已重载"
}

cmd_nginx_edit() {
  echo -e "\n  ${W}── 编辑 Nginx 配置 ──${N}"
  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ ! -f "$nginx_conf" ]; then
    err "未找到 Nginx 配置文件"
  fi
  local editor="${EDITOR:-vi}"
  sudo $editor "$nginx_conf"
  if prompt_yn "是否重载 Nginx？" "y"; then
    reload_nginx && ok "Nginx 已重载"
  fi
}

cmd_nginx() {
  local subcmd="${1:-status}"
  case "$subcmd" in
    status) cmd_nginx_status ;;
    reload) cmd_nginx_reload ;;
    edit)   cmd_nginx_edit ;;
    *)      menu_nginx ;;
  esac
}

# ── build ─────────────────────────────────────────────────────
cmd_build() {
  echo -e "\n  ${W}── 重新编译 ──${N}"
  cd "$INSTALL_DIR"
  if [ -d "frontend" ] && command -v npm &>/dev/null; then
    info "npm run build (frontend)..."
    (cd frontend && npm install --silent && npm run build) || warn "前端构建失败，跳过"
  fi
  info "go build ./cmd/server/ ..."
  go build -o smtp-lite ./cmd/server/ || err "编译失败"
  local version
  version=$(get_version)
  ok "编译完成 → ${version}"

  if prompt_yn "立即重启服务？" "y"; then
    restart_service
  fi
}

# ── autostart ──────────────────────────────────────────────────

cmd_autostart_enable() {
  echo -e "\n  ${W}── 开启开机自启 ──${N}"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
    if [ ! -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
      info "创建 Systemd 服务文件..."
      sudo tee "/etc/systemd/system/${SERVICE_NAME}.service" > /dev/null <<EOF
[Unit]
Description=SMTP Lite 个人邮箱聚合系统
After=network.target

[Service]
Type=simple
User=${USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/smtp-lite
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
      sudo systemctl daemon-reload
    fi
    sudo systemctl enable "$SERVICE_NAME"
    ok "开机自启已开启 (systemd)"
  elif [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    mkdir -p "$HOME/Library/LaunchAgents"
    cat > "$plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>             <string>com.smtp-lite</string>
  <key>ProgramArguments</key>  <array><string>${INSTALL_DIR}/smtp-lite</string></array>
  <key>WorkingDirectory</key>  <string>${INSTALL_DIR}</string>
  <key>RunAtLoad</key>         <true/>
  <key>KeepAlive</key>         <true/>
  <key>StandardOutPath</key>   <string>${INSTALL_DIR}/smtp-lite.log</string>
  <key>StandardErrorPath</key> <string>${INSTALL_DIR}/smtp-lite.log</string>
</dict>
</plist>
EOF
    launchctl load "$plist" 2>/dev/null || true
    ok "开机自启已开启 (launchd)"
  else
    err "不支持的系统"
  fi
}

cmd_autostart_disable() {
  echo -e "\n  ${W}── 关闭开机自启 ──${N}"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
    sudo systemctl disable "$SERVICE_NAME"
    ok "开机自启已关闭"
  elif [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    launchctl unload "$plist" 2>/dev/null || true
    rm -f "$plist"
    ok "开机自启已关闭"
  fi
}

cmd_autostart_status() {
  echo -e "\n  ${W}── 开机自启状态 ──${N}"
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
    local enabled
    enabled=$(systemctl is-enabled "$SERVICE_NAME" 2>/dev/null || echo "未配置")
    echo -e "  状态: ${C}${enabled}${N} (systemd)"
  elif [ "$os_type" = "darwin" ]; then
    if [ -f "$HOME/Library/LaunchAgents/com.smtp-lite.plist" ]; then
      echo -e "  状态: ${G}已开启${N} (launchd)"
    else
      echo -e "  状态: ${R}未配置${N}"
    fi
  fi
}

cmd_autostart() {
  local subcmd="${1:-status}"
  case "$subcmd" in
    enable)  cmd_autostart_enable ;;
    disable) cmd_autostart_disable ;;
    status)  cmd_autostart_status ;;
    *)       menu_autostart ;;
  esac
}

# ── uninstall ─────────────────────────────────────────────────
cmd_uninstall() {
  echo -e "\n  ${W}── 卸载 SMTP Lite ──${N}"
  warn "此操作将删除 SMTP Lite 及其所有数据"
  prompt_yn "是否先创建备份？" "y" && cmd_backup
  prompt_yn "确认卸载 SMTP Lite？此操作不可逆" "n" || return

  cmd_stop 2>/dev/null || true

  local os_type
  os_type=$(detect_os)

  if [ "$os_type" = "linux" ] && [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    sudo rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    sudo systemctl daemon-reload
    ok "Systemd 服务已移除"
  fi

  if [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    launchctl unload "$plist" 2>/dev/null || true
    rm -f "$plist"
    ok "LaunchAgent 已移除"
  fi

  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ -f "$nginx_conf" ]; then
    if prompt_yn "是否移除 Nginx 配置？" "y"; then
      sudo rm -f "$nginx_conf"
      sudo rm -f "/etc/nginx/sites-enabled/smtp-lite" 2>/dev/null || true
      command -v nginx &>/dev/null && (sudo nginx -t && sudo systemctl reload nginx 2>/dev/null || true)
      ok "Nginx 配置已移除"
    fi
  fi

  sudo rm -f /usr/local/bin/smtp-lite 2>/dev/null || true
  ok "CLI 软链接已移除"

  if crontab -l 2>/dev/null | grep -q "certbot.*smtp"; then
    crontab -l 2>/dev/null | grep -v "certbot.*smtp" | crontab - 2>/dev/null || true
    ok "证书续期 Cron 已移除"
  fi

  if prompt_yn "删除安装目录 ${INSTALL_DIR}？" "y"; then
    rm -rf "$INSTALL_DIR"
    ok "安装目录已删除"
  fi

  echo ""
  ok "SMTP Lite 已卸载"
}

# ── info ──────────────────────────────────────────────────────
cmd_info() {
  echo -e "\n  ${W}── 系统信息 ──${N}"
  divider
  echo -e "  ${DIM}操作系统${N}    → ${C}$(uname -s) $(uname -r)${N}"
  echo -e "  ${DIM}架构${N}        → ${C}$(uname -m)${N}"
  echo -e "  ${DIM}Go 版本${N}     → ${C}$(go version 2>/dev/null | awk '{print $3}' || echo '未安装')${N}"
  if command -v nginx &>/dev/null; then
    echo -e "  ${DIM}Nginx${N}       → ${C}$(nginx -v 2>&1 | grep -o '[0-9].*' || echo 'installed')${N}"
  fi
  if command -v certbot &>/dev/null; then
    echo -e "  ${DIM}Certbot${N}     → ${C}$(certbot --version 2>&1 | awk '{print $2}' || echo 'installed')${N}"
  fi
  echo -e "  ${DIM}安装目录${N}    → ${C}${INSTALL_DIR}${N}"
  echo -e "  ${DIM}配置文件${N}    → ${C}${CONFIG}${N}"
  echo -e "  ${DIM}SMTP Lite${N}   → ${C}$(get_version)${N}"
  divider
}

# ── reset ─────────────────────────────────────────────────────
cmd_reset() {
  echo -e "\n  ${W}── 重置 SMTP Lite ──${N}"
  warn "此操作将重置配置和数据库为初始状态"
  prompt_yn "确认重置？" "n" || return
  prompt_yn "是否先备份当前数据？" "y" && cmd_backup

  cmd_stop 2>/dev/null || true

  local jwt_secret enc_key
  jwt_secret=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
  enc_key=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)

  local new_user="admin" new_pass new_port="8090"
  prompt_input new_user "管理员用户名" "admin"
  while true; do
    prompt_secret new_pass "新密码（至少 8 位）"
    [ ${#new_pass} -ge 8 ] && break
    warn "密码至少 8 位"
  done
  prompt_input new_port "端口" "8090"

  cat > "$CONFIG" <<EOF
server:
  port: ${new_port}
  mode: release

auth:
  username: ${new_user}
  password: ${new_pass}

jwt:
  secret: ${jwt_secret}
  expire_hours: 168

encryption:
  key: ${enc_key}

database:
  driver: sqlite
  sqlite: smtp-lite.db
EOF
  ok "配置已重置"

  if prompt_yn "是否清空数据库？" "n"; then
    rm -f "$INSTALL_DIR/smtp-lite.db"
    ok "数据库已清空"
  fi

  restart_service
  ok "SMTP Lite 已重置"
}

cmd_version() {
  echo "SMTP Lite $(get_version)"
}

# ══════════════════════════════════════════════════════════════
#  交互式菜单系统
# ══════════════════════════════════════════════════════════════

show_banner() {
  clear
  local version
  version=$(get_version)
  local pid
  pid=$(get_pid)
  local status_str
  if [ -n "$pid" ]; then
    status_str="${G}● 运行中${N} (PID: ${pid})"
  else
    status_str="${R}● 已停止${N}"
  fi

  echo ""
  echo -e "  ${B}███████╗${N}${C}███╗   ███╗${N}${B}████████╗${N}${C}██████╗${N}  ${W}SMTP Lite${N} ${DIM}${version}${N}"
  echo -e "  ${B}██╔════╝${N}${C}████╗ ████║${N}${B}╚══██╔══╝${N}${C}██╔══██╗${N} ${status_str}"
  echo -e "  ${B}███████╗${N}${C}██╔████╔██║${N}${B}   ██║   ${N}${C}██████╔╝${N} ${DIM}$(get_current_domain)${N}"
  echo -e "  ${B}╚════██║${N}${C}██║╚██╔╝██║${N}${B}   ██║   ${N}${C}██╔═══╝${N}"
  echo -e "  ${B}███████║${N}${C}██║ ╚═╝ ██║${N}${B}   ██║   ${N}${C}██║${N}"
  echo -e "  ${B}╚══════╝${N}${C}╚═╝     ╚═╝${N}${B}   ╚═╝   ${N}${C}╚═╝${N}"
  echo ""
}

menu_main() {
  while true; do
    show_banner
    echo -e "  ${W}服务管理${N}"
    echo -e "    ${C} 1${N})  启动服务          ${C} 2${N})  停止服务"
    echo -e "    ${C} 3${N})  重启服务          ${C} 4${N})  查看状态"
    echo ""
    echo -e "  ${W}配置管理${N}"
    echo -e "    ${C} 5${N})  修改端口          ${C} 6${N})  换绑域名"
    echo -e "    ${C} 7${N})  修改密码          ${C} 8${N})  修改用户名"
    echo -e "    ${C} 9${N})  配置管理          ${C}10${N})  SSL 证书"
    echo ""
    echo -e "  ${W}系统维护${N}"
    echo -e "    ${C}11${N})  在线更新          ${C}12${N})  重新编译"
    echo -e "    ${C}13${N})  日志管理          ${C}14${N})  Nginx 管理"
    echo -e "    ${C}15${N})  开机自启          ${C}16${N})  备份数据"
    echo -e "    ${C}17${N})  恢复数据          ${C}18${N})  系统信息"
    echo ""
    echo -e "  ${W}高级操作${N}"
    echo -e "    ${C}19${N})  重置系统          ${C}20${N})  卸载"
    echo ""
    echo -e "    ${C} 0${N})  退出"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-20]: "
    read -r choice

    case "$choice" in
      1)  cmd_start;     press_enter ;;
      2)  cmd_stop;      press_enter ;;
      3)  cmd_restart;   press_enter ;;
      4)  cmd_status;    press_enter ;;
      5)  cmd_port;      press_enter ;;
      6)  cmd_domain;    press_enter ;;
      7)  cmd_password;  press_enter ;;
      8)  cmd_username;  press_enter ;;
      9)  menu_config ;;
      10) menu_ssl ;;
      11) cmd_update;    press_enter ;;
      12) cmd_build;     press_enter ;;
      13) menu_log ;;
      14) menu_nginx ;;
      15) menu_autostart ;;
      16) cmd_backup;    press_enter ;;
      17) cmd_restore;   press_enter ;;
      18) cmd_info;      press_enter ;;
      19) cmd_reset;     press_enter ;;
      20) cmd_uninstall; press_enter ;;
      0|q|Q|exit|quit)
        echo -e "\n  ${DIM}再见！${N}\n"
        exit 0
        ;;
      *)
        warn "无效选项，请重新选择"
        sleep 1
        ;;
    esac
  done
}

# ── 配置管理子菜单 ────────────────────────────────────────────
menu_config() {
  while true; do
    clear
    echo ""
    echo -e "  ${W}配置管理${N}"
    divider
    echo -e "    ${C}1${N})  查看当前配置"
    echo -e "    ${C}2${N})  编辑配置文件"
    echo -e "    ${C}3${N})  重置 JWT Secret"
    echo -e "    ${C}4${N})  重置加密密钥"
    echo ""
    echo -e "    ${C}0${N})  返回主菜单"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-4]: "
    read -r choice

    case "$choice" in
      1) cmd_config_show;      press_enter ;;
      2) cmd_config_edit;      press_enter ;;
      3) cmd_config_reset_jwt; press_enter ;;
      4) cmd_config_reset_enc; press_enter ;;
      0|q|b) return ;;
      *) warn "无效选项"; sleep 1 ;;
    esac
  done
}

# ── SSL 子菜单 ────────────────────────────────────────────────
menu_ssl() {
  while true; do
    clear
    echo ""
    echo -e "  ${W}SSL 证书管理${N}"
    divider
    echo -e "    ${C}1${N})  查看证书状态"
    echo -e "    ${C}2${N})  申请/重新申请 SSL"
    echo -e "    ${C}3${N})  手动续期证书"
    echo -e "    ${C}4${N})  移除 SSL（降级 HTTP）"
    echo ""
    echo -e "    ${C}0${N})  返回主菜单"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-4]: "
    read -r choice

    case "$choice" in
      1) cmd_ssl_status; press_enter ;;
      2) cmd_ssl_apply;  press_enter ;;
      3) cmd_ssl_renew;  press_enter ;;
      4) cmd_ssl_remove; press_enter ;;
      0|q|b) return ;;
      *) warn "无效选项"; sleep 1 ;;
    esac
  done
}

# ── 日志子菜单 ────────────────────────────────────────────────
menu_log() {
  while true; do
    clear
    echo ""
    echo -e "  ${W}日志管理${N}"
    divider
    echo -e "    ${C}1${N})  实时跟踪日志 (Ctrl+C 退出)"
    echo -e "    ${C}2${N})  查看最近 50 行"
    echo -e "    ${C}3${N})  查看最近 200 行"
    echo -e "    ${C}4${N})  清空日志"
    echo ""
    echo -e "    ${C}0${N})  返回主菜单"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-4]: "
    read -r choice

    case "$choice" in
      1) cmd_log_tail;      press_enter ;;
      2) cmd_log_show 50;   press_enter ;;
      3) cmd_log_show 200;  press_enter ;;
      4) cmd_log_clear;     press_enter ;;
      0|q|b) return ;;
      *) warn "无效选项"; sleep 1 ;;
    esac
  done
}

# ── Nginx 子菜单 ──────────────────────────────────────────────
menu_nginx() {
  while true; do
    clear
    echo ""
    echo -e "  ${W}Nginx 管理${N}"
    divider
    echo -e "    ${C}1${N})  查看 Nginx 状态/配置"
    echo -e "    ${C}2${N})  重载 Nginx"
    echo -e "    ${C}3${N})  编辑 Nginx 配置"
    echo ""
    echo -e "    ${C}0${N})  返回主菜单"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-3]: "
    read -r choice

    case "$choice" in
      1) cmd_nginx_status; press_enter ;;
      2) cmd_nginx_reload; press_enter ;;
      3) cmd_nginx_edit;   press_enter ;;
      0|q|b) return ;;
      *) warn "无效选项"; sleep 1 ;;
    esac
  done
}

# ── 开机自启子菜单 ────────────────────────────────────────────
menu_autostart() {
  while true; do
    clear
    echo ""
    echo -e "  ${W}开机自启管理${N}"
    divider
    echo -e "    ${C}1${N})  查看状态"
    echo -e "    ${C}2${N})  开启自启"
    echo -e "    ${C}3${N})  关闭自启"
    echo ""
    echo -e "    ${C}0${N})  返回主菜单"
    echo ""

    local choice
    echo -ne "  ${C}→${N} 请选择 [0-3]: "
    read -r choice

    case "$choice" in
      1) cmd_autostart_status;  press_enter ;;
      2) cmd_autostart_enable;  press_enter ;;
      3) cmd_autostart_disable; press_enter ;;
      0|q|b) return ;;
      *) warn "无效选项"; sleep 1 ;;
    esac
  done
}

# ══════════════════════════════════════════════════════════════
#  主入口：支持交互式菜单 + 直接命令兼容
# ══════════════════════════════════════════════════════════════
main() {
  # 无参数 → 交互式菜单
  if [ $# -eq 0 ]; then
    menu_main
    exit 0
  fi

  # 有参数 → 直接执行命令（向后兼容）
  local cmd="$1"
  shift

  case "$cmd" in
    start)             cmd_start "$@" ;;
    stop)              cmd_stop "$@" ;;
    restart)           cmd_restart "$@" ;;
    status)            cmd_status "$@" ;;
    port)              cmd_port "$@" ;;
    domain)            cmd_domain "$@" ;;
    password|passwd)   cmd_password "$@" ;;
    username|user)     cmd_username "$@" ;;
    ssl)               cmd_ssl "$@" ;;
    log|logs)          cmd_log "$@" ;;
    update)            cmd_update "$@" ;;
    config|cfg)        cmd_config "$@" ;;
    backup)            cmd_backup "$@" ;;
    restore)           cmd_restore "$@" ;;
    nginx)             cmd_nginx "$@" ;;
    build)             cmd_build "$@" ;;
    autostart)         cmd_autostart "$@" ;;
    uninstall)         cmd_uninstall "$@" ;;
    info)              cmd_info "$@" ;;
    reset)             cmd_reset "$@" ;;
    help|-h|--help)
      echo ""
      echo -e "  ${W}SMTP Lite 管理工具 v2.1${N}"
      echo ""
      echo -e "  ${W}用法:${N}"
      echo -e "    smtp-lite              进入交互式菜单"
      echo -e "    smtp-lite <命令>       直接执行命令"
      echo ""
      echo -e "  ${W}可用命令:${N}"
      echo -e "    start stop restart status port domain password username"
      echo -e "    ssl config log update build backup restore nginx autostart"
      echo -e "    info reset uninstall version help"
      echo ""
      echo -e "  提示：直接运行 ${W}smtp-lite${N} 进入交互式菜单更方便操作"
      echo ""
      ;;
    version|-v|--version) cmd_version ;;
    *)
      echo -e "  ${R}未知命令:${N} ${cmd}"
      echo -e "  运行 ${W}smtp-lite${N} 进入交互式菜单，或 ${W}smtp-lite help${N} 查看帮助"
      exit 1
      ;;
  esac
}

main "$@"
