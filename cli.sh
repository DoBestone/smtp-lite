#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 终端管理工具
#  用法: smtp-lite <命令> [参数]
#        smtp-lite help
# =============================================================
set -euo pipefail

# ── 自动检测安装目录 ──────────────────────────────────────────
# 优先使用脚本所在目录，否则回退到 $HOME/smtp-lite
if [ -f "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/config.yaml" ]; then
  INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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

# 读取 YAML 值 (简单解析，无需 yq 依赖)
yaml_get() {
  local key="$1" file="${2:-$CONFIG}"
  grep -E "^\s*${key}:" "$file" 2>/dev/null | head -1 | sed "s/.*${key}:\s*//" | sed 's/#.*//' | xargs
}

# 写入 YAML 值 (替换已有 key)
yaml_set() {
  local key="$1" value="$2" file="${3:-$CONFIG}"
  if grep -qE "^\s*${key}:" "$file" 2>/dev/null; then
    sed -i.bak "s|^\(\s*${key}:\).*|\1 ${value}|" "$file"
    rm -f "${file}.bak"
  else
    echo "  ${key}: ${value}" >> "$file"
  fi
}

# 检测操作系统
detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux"  ;;
    *)      echo "unknown" ;;
  esac
}

# 获取服务 PID
get_pid() {
  pgrep -f "${INSTALL_DIR}/smtp-lite" 2>/dev/null | head -1 || true
}

# 获取当前版本
get_version() {
  grep 'Version' "$INSTALL_DIR/internal/version/version.go" 2>/dev/null \
    | grep -o '"v[^"]*"' | tr -d '"' || echo "unknown"
}

# 获取 Nginx 配置路径
get_nginx_conf() {
  local os_type
  os_type=$(detect_os)
  if [ "$os_type" = "darwin" ]; then
    echo "/usr/local/etc/nginx/servers/smtp-lite.conf"
  else
    echo "/etc/nginx/sites-available/smtp-lite"
  fi
}

# 读取当前域名 (从 Nginx 配置)
get_current_domain() {
  local conf
  conf=$(get_nginx_conf)
  if [ -f "$conf" ]; then
    grep "server_name" "$conf" 2>/dev/null | head -1 | awk '{print $2}' | tr -d ';'
  else
    echo "(未配置域名)"
  fi
}

# 交互式输入
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

# 重启后端服务
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

# 重新加载 Nginx
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

# ── start ─────────────────────────────────────────────────────
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

# ── stop ──────────────────────────────────────────────────────
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

# ── restart ───────────────────────────────────────────────────
cmd_restart() {
  echo -e "\n  ${W}── 重启服务 ──${N}"
  restart_service
}

# ── status ────────────────────────────────────────────────────
cmd_status() {
  echo ""
  echo -e "  ${W}SMTP Lite 状态${N}"
  divider

  # 版本
  local version
  version=$(get_version)
  echo -e "  ${DIM}版本        ${N}→ ${C}${version}${N}"

  # 安装目录
  echo -e "  ${DIM}安装目录    ${N}→ ${C}${INSTALL_DIR}${N}"

  # 端口
  local port
  port=$(yaml_get "port")
  echo -e "  ${DIM}监听端口    ${N}→ ${C}${port}${N}"

  # 管理账号
  local username
  username=$(yaml_get "username")
  echo -e "  ${DIM}管理账号    ${N}→ ${C}${username}${N}"

  # 服务状态
  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    echo -e "  ${DIM}服务状态    ${N}→ ${G}运行中${N} (PID: ${pid})"
  else
    echo -e "  ${DIM}服务状态    ${N}→ ${R}已停止${N}"
  fi

  # 域名
  local domain
  domain=$(get_current_domain)
  echo -e "  ${DIM}绑定域名    ${N}→ ${C}${domain}${N}"

  # SSL 证书
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

  # Nginx
  if command -v nginx &>/dev/null; then
    if [ -f "$nginx_conf" ]; then
      echo -e "  ${DIM}Nginx       ${N}→ ${G}已配置${N}"
    else
      echo -e "  ${DIM}Nginx       ${N}→ ${Y}已安装但未配置${N}"
    fi
  else
    echo -e "  ${DIM}Nginx       ${N}→ ${DIM}未安装${N}"
  fi

  # 数据库大小
  local db_file="$INSTALL_DIR/smtp-lite.db"
  if [ -f "$db_file" ]; then
    local db_size
    db_size=$(du -h "$db_file" | awk '{print $1}')
    echo -e "  ${DIM}数据库大小  ${N}→ ${C}${db_size}${N}"
  fi

  # 日志大小
  local log_file="$INSTALL_DIR/smtp-lite.log"
  if [ -f "$log_file" ]; then
    local log_size
    log_size=$(du -h "$log_file" | awk '{print $1}')
    echo -e "  ${DIM}日志大小    ${N}→ ${C}${log_size}${N}"
  fi

  divider

  # 服务管理类型
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

# ── port ──────────────────────────────────────────────────────
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

  # 同步更新 Nginx 配置
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
    warn "请手动重启服务: smtp-lite restart"
  fi
}

# ── domain ────────────────────────────────────────────────────
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

  # 询问是否配置 SSL
  local use_ssl=false
  if prompt_yn "是否为新域名申请 SSL 证书？" "y"; then
    use_ssl=true
    if ! command -v certbot &>/dev/null; then
      err "未安装 certbot，请先安装: sudo apt install certbot"
    fi
  fi

  if $use_ssl; then
    # webroot 方式申请证书
    local webroot="/var/www/certbot"
    sudo mkdir -p "${webroot}/.well-known/acme-challenge"
    sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
      || sudo chown -R "$(whoami)" "$webroot"

    # 先写临时 HTTP 配置
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

    # 申请证书
    info "申请 SSL 证书..."
    local certbot_email=""
    prompt_input certbot_email "证书通知邮箱（可留空）" ""
    local certbot_args=("certonly" "--webroot" "-w" "$webroot" "-d" "$new_domain" "--non-interactive" "--agree-tos")
    [ -n "$certbot_email" ] && certbot_args+=("--email" "$certbot_email") \
                            || certbot_args+=("--register-unsafely-without-email")

    if sudo certbot "${certbot_args[@]}"; then
      ok "SSL 证书申请成功"
      # 写入正式 SSL 配置
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
    # 仅 HTTP 配置
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

# ── password ──────────────────────────────────────────────────
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
  else
    warn "请手动重启服务: smtp-lite restart"
  fi
}

# ── username ──────────────────────────────────────────────────
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
  else
    warn "请手动重启服务: smtp-lite restart"
  fi
}

# ── ssl ───────────────────────────────────────────────────────
cmd_ssl() {
  local subcmd="${1:-}"

  case "$subcmd" in
    renew)  cmd_ssl_renew ;;
    status) cmd_ssl_status ;;
    apply)  cmd_ssl_apply ;;
    remove) cmd_ssl_remove ;;
    *)
      echo -e "\n  ${W}SSL 证书管理${N}"
      echo ""
      echo -e "  用法: smtp-lite ssl <子命令>"
      echo ""
      echo -e "  ${C}status${N}   查看 SSL 证书状态"
      echo -e "  ${C}apply${N}    申请/重新申请 SSL 证书"
      echo -e "  ${C}renew${N}    手动续期 SSL 证书"
      echo -e "  ${C}remove${N}   移除 SSL 配置（降级为 HTTP）"
      echo ""
      ;;
  esac
}

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
    warn "请先配置域名: smtp-lite domain"
    return
  fi

  if ! command -v certbot &>/dev/null; then
    err "未安装 certbot，请先安装"
  fi

  local webroot="/var/www/certbot"
  sudo mkdir -p "${webroot}/.well-known/acme-challenge"
  sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
    || sudo chown -R "$(whoami)" "$webroot"

  # 确保 nginx 配置包含 acme-challenge location
  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if ! grep -q "acme-challenge" "$nginx_conf" 2>/dev/null; then
    info "在 Nginx 配置中添加验证路径..."
    # 在第一个 location 前插入
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
    # 更新 nginx 配置为 SSL 版本
    cmd_domain  # 重新走域名配置流程以更新 nginx
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

# ── log ───────────────────────────────────────────────────────
cmd_log() {
  local subcmd="${1:-tail}"
  local log_file="$INSTALL_DIR/smtp-lite.log"
  local os_type
  os_type=$(detect_os)

  case "$subcmd" in
    tail|follow|-f)
      echo -e "\n  ${W}── 实时日志 ──${N} ${DIM}(Ctrl+C 退出)${N}"
      if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        sudo journalctl -u "$SERVICE_NAME" -f --no-pager
      elif [ -f "$log_file" ]; then
        tail -f "$log_file"
      else
        warn "未找到日志文件"
      fi
      ;;
    show|view)
      local lines="${2:-50}"
      echo -e "\n  ${W}── 最近 ${lines} 行日志 ──${N}"
      if [ "$os_type" = "linux" ] && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        sudo journalctl -u "$SERVICE_NAME" -n "$lines" --no-pager
      elif [ -f "$log_file" ]; then
        tail -n "$lines" "$log_file"
      else
        warn "未找到日志文件"
      fi
      ;;
    clear)
      echo -e "\n  ${W}── 清空日志 ──${N}"
      prompt_yn "确认清空日志文件？" "n" || return
      if [ -f "$log_file" ]; then
        > "$log_file"
        ok "日志已清空"
      fi
      if [ "$os_type" = "linux" ] && command -v journalctl &>/dev/null; then
        sudo journalctl --vacuum-time=1s -u "$SERVICE_NAME" &>/dev/null || true
        ok "Journald 日志已清理"
      fi
      ;;
    *)
      echo -e "\n  ${W}日志管理${N}"
      echo ""
      echo -e "  用法: smtp-lite log <子命令>"
      echo ""
      echo -e "  ${C}tail${N}     实时跟踪日志 (默认)"
      echo -e "  ${C}show${N}     查看最近日志 (可加行数，如: smtp-lite log show 100)"
      echo -e "  ${C}clear${N}    清空日志"
      echo ""
      ;;
  esac
}

# ── update ────────────────────────────────────────────────────
cmd_update() {
  echo -e "\n  ${W}── 检查更新 ──${N}"
  local force=false
  [ "${1:-}" = "--force" ] && force=true

  cd "$INSTALL_DIR"

  local current
  current=$(get_version)
  info "当前版本: ${W}${current}${N}"

  # 查询最新版本
  local latest
  latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | grep -o '"v[^"]*"' | tr -d '"' || echo "")

  if [ -n "$latest" ]; then
    info "最新版本: ${W}${latest}${N}"
    if [ "$latest" = "$current" ] && [ "$force" = false ]; then
      ok "已是最新版本"
      echo -e "  提示：使用 ${W}smtp-lite update --force${N} 强制更新"
      return
    fi
  else
    warn "无法获取最新版本号，继续本地更新"
  fi

  info "拉取代码..."
  git pull || err "git pull 失败"
  ok "代码更新完成"

  info "重新编译..."
  go build -o smtp-lite ./cmd/server/ || err "编译失败"
  local new_ver
  new_ver=$(get_version)
  ok "编译完成 → ${new_ver}"

  restart_service
  echo ""
  echo -e "  ${G}更新完成: ${current} → ${W}${new_ver}${N}"
  echo ""
}

# ── config ────────────────────────────────────────────────────
cmd_config() {
  local subcmd="${1:-show}"

  case "$subcmd" in
    show|view)
      echo -e "\n  ${W}── 当前配置 ──${N}"
      echo ""
      # 安全显示配置（隐藏敏感信息）
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
      ;;
    edit)
      local editor="${EDITOR:-vi}"
      echo -e "\n  ${W}── 编辑配置 ──${N}"
      info "使用编辑器: ${editor}"
      warn "修改后需重启服务: smtp-lite restart"
      $editor "$CONFIG"
      ;;
    reset-jwt)
      echo -e "\n  ${W}── 重置 JWT Secret ──${N}"
      prompt_yn "重置 JWT Secret 将使所有现有登录失效，确认？" "n" || return
      local new_secret
      new_secret=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
      yaml_set "secret" "$new_secret"
      ok "JWT Secret 已重置"
      restart_service
      ;;
    reset-encryption-key)
      echo -e "\n  ${W}── 重置加密密钥 ──${N}"
      warn "重置加密密钥后，已加密的 SMTP 密码将无法解密！"
      warn "建议先导出 SMTP 账号信息，重置后重新添加"
      prompt_yn "确认重置加密密钥？此操作不可逆！" "n" || return
      local new_key
      new_key=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
      yaml_set "key" "$new_key"
      ok "加密密钥已重置"
      restart_service
      ;;
    *)
      echo -e "\n  ${W}配置管理${N}"
      echo ""
      echo -e "  用法: smtp-lite config <子命令>"
      echo ""
      echo -e "  ${C}show${N}                  查看当前配置 (默认)"
      echo -e "  ${C}edit${N}                  使用编辑器打开配置文件"
      echo -e "  ${C}reset-jwt${N}             重置 JWT Secret"
      echo -e "  ${C}reset-encryption-key${N}  重置加密密钥"
      echo ""
      ;;
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

  # 备份配置文件
  cp "$CONFIG" "$backup_path/config.yaml"
  ok "配置文件已备份"

  # 备份数据库
  local db_file="$INSTALL_DIR/smtp-lite.db"
  if [ -f "$db_file" ]; then
    cp "$db_file" "$backup_path/smtp-lite.db"
    ok "数据库已备份"
  fi

  # 备份 Nginx 配置
  local nginx_conf
  nginx_conf=$(get_nginx_conf)
  if [ -f "$nginx_conf" ]; then
    cp "$nginx_conf" "$backup_path/nginx-smtp-lite.conf"
    ok "Nginx 配置已备份"
  fi

  # 打包
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
    # 列出所有备份
    if [ ! -d "$backup_dir" ] || [ -z "$(ls -A "$backup_dir"/*.tar.gz 2>/dev/null)" ]; then
      warn "未找到备份文件"
      info "用法: smtp-lite restore <备份文件路径>"
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
      echo -e "  ${C}${i}${N}) ${fname} (${size})"
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

  # 停止服务
  local pid
  pid=$(get_pid)
  if [ -n "$pid" ]; then
    info "停止服务..."
    cmd_stop
  fi

  # 解压
  local tmp_dir
  tmp_dir=$(mktemp -d)
  tar -xzf "$archive" -C "$tmp_dir"
  local extracted
  extracted=$(ls -d "$tmp_dir"/smtp-lite_* | head -1)

  # 恢复配置
  if [ -f "$extracted/config.yaml" ]; then
    cp "$extracted/config.yaml" "$CONFIG"
    ok "配置文件已恢复"
  fi

  # 恢复数据库
  if [ -f "$extracted/smtp-lite.db" ]; then
    cp "$extracted/smtp-lite.db" "$INSTALL_DIR/smtp-lite.db"
    ok "数据库已恢复"
  fi

  # 恢复 Nginx 配置
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

  # 重启服务
  restart_service
  ok "数据恢复完成"
}

# ── nginx ─────────────────────────────────────────────────────
cmd_nginx() {
  local subcmd="${1:-status}"

  case "$subcmd" in
    status)
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
      ;;
    reload)
      echo -e "\n  ${W}── 重载 Nginx ──${N}"
      reload_nginx && ok "Nginx 已重载"
      ;;
    edit)
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
      ;;
    *)
      echo -e "\n  ${W}Nginx 管理${N}"
      echo ""
      echo -e "  用法: smtp-lite nginx <子命令>"
      echo ""
      echo -e "  ${C}status${N}   查看 Nginx 配置 (默认)"
      echo -e "  ${C}reload${N}   重载 Nginx"
      echo -e "  ${C}edit${N}     编辑 Nginx 配置"
      echo ""
      ;;
  esac
}

# ── build ─────────────────────────────────────────────────────
cmd_build() {
  echo -e "\n  ${W}── 重新编译 ──${N}"
  cd "$INSTALL_DIR"
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
cmd_autostart() {
  local subcmd="${1:-status}"
  local os_type
  os_type=$(detect_os)

  case "$subcmd" in
    enable)
      echo -e "\n  ${W}── 开启开机自启 ──${N}"
      if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
        # 确保 service 文件存在
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
      ;;
    disable)
      echo -e "\n  ${W}── 关闭开机自启 ──${N}"
      if [ "$os_type" = "linux" ] && command -v systemctl &>/dev/null; then
        sudo systemctl disable "$SERVICE_NAME"
        ok "开机自启已关闭"
      elif [ "$os_type" = "darwin" ]; then
        local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
        launchctl unload "$plist" 2>/dev/null || true
        rm -f "$plist"
        ok "开机自启已关闭"
      fi
      ;;
    status|*)
      echo -e "\n  ${W}── 开机自启状态 ──${N}"
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
      if [ "$subcmd" != "status" ]; then
        echo ""
        echo -e "  用法: smtp-lite autostart <enable|disable|status>"
      fi
      ;;
  esac
}

# ── uninstall ─────────────────────────────────────────────────
cmd_uninstall() {
  echo -e "\n  ${W}── 卸载 SMTP Lite ──${N}"
  warn "此操作将删除 SMTP Lite 及其所有数据"
  prompt_yn "是否先创建备份？" "y" && cmd_backup
  prompt_yn "确认卸载 SMTP Lite？此操作不可逆" "n" || return

  # 停止服务
  cmd_stop 2>/dev/null || true

  local os_type
  os_type=$(detect_os)

  # 移除 systemd 服务
  if [ "$os_type" = "linux" ] && [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    sudo rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    sudo systemctl daemon-reload
    ok "Systemd 服务已移除"
  fi

  # 移除 LaunchAgent
  if [ "$os_type" = "darwin" ]; then
    local plist="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
    launchctl unload "$plist" 2>/dev/null || true
    rm -f "$plist"
    ok "LaunchAgent 已移除"
  fi

  # 移除 Nginx 配置
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

  # 移除 CLI 软链接
  sudo rm -f /usr/local/bin/smtp-lite 2>/dev/null || true
  ok "CLI 软链接已移除"

  # 移除 cron
  if crontab -l 2>/dev/null | grep -q "certbot.*smtp"; then
    crontab -l 2>/dev/null | grep -v "certbot.*smtp" | crontab - 2>/dev/null || true
    ok "证书续期 Cron 已移除"
  fi

  # 移除安装目录
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

  # 停止服务
  cmd_stop 2>/dev/null || true

  # 重新生成 config.yaml
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
EOF
  ok "配置已重置"

  # 删除数据库
  if prompt_yn "是否清空数据库？" "n"; then
    rm -f "$INSTALL_DIR/smtp-lite.db"
    ok "数据库已清空"
  fi

  restart_service
  ok "SMTP Lite 已重置"
}

# ── help ──────────────────────────────────────────────────────
cmd_help() {
  echo ""
  echo -e "  ${B}███████${N} ${W}SMTP Lite 管理工具${N}"
  echo ""
  echo -e "  ${W}用法:${N} smtp-lite <命令> [参数]"
  echo ""
  echo -e "  ${W}服务管理${N}"
  echo -e "    ${C}start${N}       启动服务"
  echo -e "    ${C}stop${N}        停止服务"
  echo -e "    ${C}restart${N}     重启服务"
  echo -e "    ${C}status${N}      查看服务状态"
  echo -e "    ${C}autostart${N}   管理开机自启 (enable/disable/status)"
  echo ""
  echo -e "  ${W}配置修改${N}"
  echo -e "    ${C}port${N}        修改监听端口"
  echo -e "    ${C}domain${N}      换绑域名"
  echo -e "    ${C}password${N}    修改管理员密码"
  echo -e "    ${C}username${N}    修改管理员用户名"
  echo -e "    ${C}config${N}      配置管理 (show/edit/reset-jwt/reset-encryption-key)"
  echo ""
  echo -e "  ${W}SSL 证书${N}"
  echo -e "    ${C}ssl${N}         SSL 管理 (status/apply/renew/remove)"
  echo ""
  echo -e "  ${W}Nginx${N}"
  echo -e "    ${C}nginx${N}       Nginx 管理 (status/reload/edit)"
  echo ""
  echo -e "  ${W}维护操作${N}"
  echo -e "    ${C}update${N}      在线更新 (--force 强制更新)"
  echo -e "    ${C}build${N}       重新编译"
  echo -e "    ${C}log${N}         日志管理 (tail/show/clear)"
  echo -e "    ${C}backup${N}      备份数据"
  echo -e "    ${C}restore${N}     恢复数据"
  echo -e "    ${C}reset${N}       重置为初始状态"
  echo -e "    ${C}info${N}        查看系统信息"
  echo -e "    ${C}uninstall${N}   卸载 SMTP Lite"
  echo ""
  echo -e "  ${W}其他${N}"
  echo -e "    ${C}help${N}        显示此帮助信息"
  echo -e "    ${C}version${N}     显示版本号"
  echo ""
  echo -e "  ${W}示例:${N}"
  echo -e "    smtp-lite status"
  echo -e "    smtp-lite port"
  echo -e "    smtp-lite domain"
  echo -e "    smtp-lite password"
  echo -e "    smtp-lite ssl apply"
  echo -e "    smtp-lite log tail"
  echo -e "    smtp-lite backup"
  echo -e "    smtp-lite update --force"
  echo ""
}

# ── version ───────────────────────────────────────────────────
cmd_version() {
  echo "SMTP Lite $(get_version)"
}

# ══════════════════════════════════════════════════════════════
#  主入口
# ══════════════════════════════════════════════════════════════
main() {
  local cmd="${1:-help}"
  shift 2>/dev/null || true

  case "$cmd" in
    start)      cmd_start "$@" ;;
    stop)       cmd_stop "$@" ;;
    restart)    cmd_restart "$@" ;;
    status)     cmd_status "$@" ;;
    port)       cmd_port "$@" ;;
    domain)     cmd_domain "$@" ;;
    password|passwd) cmd_password "$@" ;;
    username|user)   cmd_username "$@" ;;
    ssl)        cmd_ssl "$@" ;;
    log|logs)   cmd_log "$@" ;;
    update)     cmd_update "$@" ;;
    config|cfg) cmd_config "$@" ;;
    backup)     cmd_backup "$@" ;;
    restore)    cmd_restore "$@" ;;
    nginx)      cmd_nginx "$@" ;;
    build)      cmd_build "$@" ;;
    autostart)  cmd_autostart "$@" ;;
    uninstall)  cmd_uninstall "$@" ;;
    info)       cmd_info "$@" ;;
    reset)      cmd_reset "$@" ;;
    help|-h|--help) cmd_help ;;
    version|-v|--version) cmd_version ;;
    *)
      echo -e "  ${R}未知命令:${N} ${cmd}"
      echo -e "  运行 ${W}smtp-lite help${N} 查看帮助"
      exit 1
      ;;
  esac
}

main "$@"
