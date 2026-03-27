#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 交互式安装脚本
#  支持: macOS (Intel/Apple Silicon) · Linux (x86_64/arm64)
#  用法: bash install.sh
# =============================================================
set -euo pipefail

# ── 全局常量 ──────────────────────────────────────────────────
REPO="https://github.com/DoBestone/smtp-lite.git"
DEFAULT_INSTALL_DIR="$HOME/smtp-lite"
MIN_GO_MAJOR=1; MIN_GO_MINOR=22; GO_INSTALL_VERSION="1.22.5"
SERVICE_NAME="smtp-lite"

# ── 颜色 ─────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; DIM='\033[2m'; N='\033[0m'

# ── 日志函数 ──────────────────────────────────────────────────
info()    { echo -e "  ${B}▸${N} $*"; }
ok()      { echo -e "  ${G}✓${N} $*"; }
warn()    { echo -e "  ${Y}⚠${N}  $*"; }
err()     { echo -e "  ${R}✗${N} $*"; exit 1; }
step()    { echo -e "\n${W}  ╔═══════════════════════════════════════╗${N}"; \
            printf "${W}  ║  %-38s║${N}\n" " $*"; \
            echo -e "${W}  ╚═══════════════════════════════════════╝${N}"; }
divider() { echo -e "  ${DIM}───────────────────────────────────────────${N}"; }

# ── 用户输入辅助 ──────────────────────────────────────────────
# prompt_input <变量名> <提示> [默认值]
prompt_input() {
  local __var="$1" __prompt="$2" __default="${3:-}"
  local __display_default=""
  [ -n "$__default" ] && __display_default=" ${DIM}[${__default}]${N}"
  echo -ne "  ${C}→${N} ${__prompt}${__display_default}: "
  read -r __val
  eval "$__var='${__val:-$__default}'"
}

# prompt_secret <变量名> <提示>
prompt_secret() {
  local __var="$1" __prompt="$2"
  echo -ne "  ${C}→${N} ${__prompt}: "
  read -rs __val; echo
  eval "$__var='$__val'"
}

# prompt_yn <提示> [默认 y/n]  → 返回 0=yes 1=no
prompt_yn() {
  local __prompt="$1" __default="${2:-n}"
  local __hint; [ "$__default" = "y" ] && __hint="[Y/n]" || __hint="[y/N]"
  echo -ne "  ${C}→${N} ${__prompt} ${DIM}${__hint}${N}: "
  read -r __ans
  __ans="${__ans:-$__default}"
  [[ "$__ans" =~ ^[Yy] ]]
}

# ── 用户输入配置 ──────────────────────────────────────────────
# shellcheck disable=SC2034
INSTALL_DIR="" PORT="" ADMIN_USER="" ADMIN_PASS=""
USE_DOMAIN=false DOMAIN="" USE_NGINX=false USE_SSL=false
NGINX_CONF="" CERTBOT_EMAIL=""

collect_config() {
  step "安装配置"

  # 安装目录
  prompt_input INSTALL_DIR "安装目录" "$DEFAULT_INSTALL_DIR"

  # 端口
  while true; do
    prompt_input PORT "监听端口" "8090"
    [[ "$PORT" =~ ^[0-9]+$ ]] && [ "$PORT" -ge 1 ] && [ "$PORT" -le 65535 ] && break
    warn "请输入有效端口号 (1-65535)"
  done

  divider
  echo -e "  ${W}管理员账号${N}"

  # 用户名
  prompt_input ADMIN_USER "管理员用户名" "admin"

  # 密码（带确认）
  while true; do
    prompt_secret ADMIN_PASS "管理员密码（至少 8 位）"
    [ ${#ADMIN_PASS} -ge 8 ] || { warn "密码至少 8 位，请重新输入"; continue; }
    local __confirm
    prompt_secret __confirm "再次确认密码"
    [ "$ADMIN_PASS" = "$__confirm" ] && break
    warn "两次密码不一致，请重新输入"
  done

  divider
  echo -e "  ${W}域名 & SSL${N}"

  # 域名绑定
  if prompt_yn "是否绑定自定义域名？"; then
    USE_DOMAIN=true

    while true; do
      prompt_input DOMAIN "域名（如 smtp.example.com）" ""
      # 简单格式校验
      [[ "$DOMAIN" =~ ^[a-zA-Z0-9]([a-zA-Z0-9\-]*\.)+[a-zA-Z]{2,}$ ]] && break
      warn "域名格式不正确，请重新输入"
    done

    # Nginx
    if prompt_yn "是否安装/配置 Nginx 反向代理？" "y"; then
      USE_NGINX=true

      # SSL
      if prompt_yn "是否申请 Let's Encrypt SSL 证书（需要域名已解析到本机）？" "y"; then
        USE_SSL=true
        prompt_input CERTBOT_EMAIL "用于 SSL 证书通知的邮箱" ""
      fi
    fi
  fi

  divider
  # 汇总预览
  echo ""
  echo -e "  ${W}安装预览${N}"
  echo -e "  ${DIM}安装目录 ${N}→ ${C}${INSTALL_DIR}${N}"
  echo -e "  ${DIM}监听端口 ${N}→ ${C}${PORT}${N}"
  echo -e "  ${DIM}管理账号 ${N}→ ${C}${ADMIN_USER}${N} / ${DIM}(已设置密码)${N}"
  if $USE_DOMAIN; then
    echo -e "  ${DIM}绑定域名 ${N}→ ${C}${DOMAIN}${N}"
    $USE_NGINX && echo -e "  ${DIM}Nginx    ${N}→ ${G}启用${N}"
    $USE_SSL   && echo -e "  ${DIM}SSL      ${N}→ ${G}Let's Encrypt${N}"
  fi
  echo ""

  prompt_yn "确认以上配置，开始安装？" "y" || err "已取消安装"
}

# ── 系统检测 ──────────────────────────────────────────────────
detect_system() {
  step "系统检测"
  case "$(uname -s)" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux"  ;;
    *)      err "不支持的操作系统: $(uname -s)" ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             err "不支持的架构: $(uname -m)" ;;
  esac
  if   command -v brew    &>/dev/null; then PKG="brew"
  elif command -v apt-get &>/dev/null; then PKG="apt"
  elif command -v dnf     &>/dev/null; then PKG="dnf"
  elif command -v yum     &>/dev/null; then PKG="yum"
  else                                      PKG="none"
  fi
  ok "系统: ${OS}/${ARCH}  包管理器: ${PKG}"
}

# ── 依赖检测 ──────────────────────────────────────────────────
check_git() {
  if command -v git &>/dev/null; then
    ok "Git $(git --version | awk '{print $3}')"
    return
  fi
  warn "Git 未安装，正在安装..."
  case "$PKG" in
    brew) brew install git ;;
    apt)  sudo apt-get install -y git ;;
    dnf)  sudo dnf install -y git ;;
    yum)  sudo yum install -y git ;;
    *)    err "请手动安装 Git: https://git-scm.com" ;;
  esac
  ok "Git 安装完成"
}

check_go() {
  if command -v go &>/dev/null; then
    local ver major minor
    ver=$(go version | grep -o '[0-9]\+\.[0-9]\+' | head -1)
    major=$(echo "$ver" | cut -d. -f1)
    minor=$(echo "$ver" | cut -d. -f2)
    if [ "$major" -gt "$MIN_GO_MAJOR" ] || \
       { [ "$major" -eq "$MIN_GO_MAJOR" ] && [ "$minor" -ge "$MIN_GO_MINOR" ]; }; then
      ok "Go ${ver}"; return
    fi
    warn "Go ${ver} 过低（需要 >= ${MIN_GO_MAJOR}.${MIN_GO_MINOR}），重新安装..."
  else
    warn "Go 未安装"
  fi
  install_go
}

install_go() {
  local pkg="go${GO_INSTALL_VERSION}.${OS}-${ARCH}.tar.gz"
  local url="https://go.dev/dl/${pkg}"
  info "下载 Go ${GO_INSTALL_VERSION} ..."
  local tmp; tmp=$(mktemp -d)
  curl -fsSL --progress-bar "$url" -o "$tmp/$pkg" \
    || err "Go 下载失败，请手动安装: https://go.dev/dl"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "$tmp/$pkg"
  rm -rf "$tmp"
  export PATH=$PATH:/usr/local/go/bin
  if [ "$OS" = "darwin" ]; then
    local prof="$HOME/.zshrc"
    [ -f "$HOME/.bash_profile" ] && prof="$HOME/.bash_profile"
    grep -q '/usr/local/go/bin' "$prof" 2>/dev/null \
      || echo 'export PATH=$PATH:/usr/local/go/bin' >> "$prof"
  else
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null
  fi
  ok "Go ${GO_INSTALL_VERSION} 安装完成"
}

check_nginx() {
  $USE_NGINX || return
  if command -v nginx &>/dev/null; then
    ok "Nginx $(nginx -v 2>&1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' || echo 'installed')"
    return
  fi
  warn "Nginx 未安装，正在安装..."
  case "$PKG" in
    brew) brew install nginx ;;
    apt)  sudo apt-get install -y nginx ;;
    dnf)  sudo dnf install -y nginx ;;
    yum)  sudo yum install -y nginx ;;
    *)    err "请手动安装 Nginx" ;;
  esac
  ok "Nginx 安装完成"
}

check_certbot() {
  $USE_SSL || return
  if command -v certbot &>/dev/null; then
    ok "Certbot $(certbot --version 2>&1 | awk '{print $2}')"
    return
  fi
  warn "Certbot 未安装，正在安装..."
  case "$PKG" in
    brew) brew install certbot ;;
    apt)  sudo apt-get install -y certbot python3-certbot-nginx ;;
    dnf)  sudo dnf install -y certbot python3-certbot-nginx ;;
    yum)  sudo yum install -y certbot python3-certbot-nginx ;;
    *)    err "请手动安装 Certbot: https://certbot.eff.org" ;;
  esac
  ok "Certbot 安装完成"
}

# ── 克隆/更新代码 ─────────────────────────────────────────────
setup_repo() {
  step "获取代码"
  if [ -d "$INSTALL_DIR/.git" ]; then
    info "已有安装目录，更新代码..."
    git -C "$INSTALL_DIR" pull
    ok "代码更新完成"
  else
    info "克隆仓库 → ${INSTALL_DIR}"
    git clone "$REPO" "$INSTALL_DIR"
    ok "克隆完成"
  fi
}

# ── 编译 ─────────────────────────────────────────────────────
build() {
  step "编译"
  cd "$INSTALL_DIR"
  info "go build ./cmd/server/ ..."
  go build -o smtp-lite ./cmd/server/
  ok "编译完成"
}

# ── 生成配置文件 ──────────────────────────────────────────────
write_config() {
  step "生成配置"
  cd "$INSTALL_DIR"

  # 随机密钥（兼容 macOS/Linux）
  local jwt_secret enc_key
  jwt_secret=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
  enc_key=$(LC_ALL=C    tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)

  cat > config.yaml <<EOF
server:
  port: ${PORT}
  mode: release

# 登录账号
auth:
  username: ${ADMIN_USER}
  password: ${ADMIN_PASS}

# JWT 配置
jwt:
  secret: ${jwt_secret}
  expire_hours: 168

# AES-256 加密密钥（32字节）
encryption:
  key: ${enc_key}
EOF
  ok "config.yaml 已生成（密钥已随机化）"
}

# ── Nginx 配置 ────────────────────────────────────────────────
setup_nginx() {
  $USE_NGINX || return
  step "配置 Nginx"

  if [ "$OS" = "darwin" ]; then
    NGINX_CONF_DIR="/usr/local/etc/nginx/servers"
    mkdir -p "$NGINX_CONF_DIR"
    NGINX_CONF="$NGINX_CONF_DIR/smtp-lite.conf"
  else
    NGINX_CONF_DIR="/etc/nginx/sites-available"
    sudo mkdir -p "$NGINX_CONF_DIR"
    NGINX_CONF="$NGINX_CONF_DIR/smtp-lite"
  fi

  if $USE_SSL; then
    _write_nginx_ssl
  else
    _write_nginx_http
  fi

  # 启用配置（Linux）— SSL 路径内部已处理 ln 和 reload，此处确保兜底
  if [ "$OS" = "linux" ]; then
    sudo ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/smtp-lite"
    sudo nginx -t && sudo systemctl reload nginx || true
  else
    nginx -t && { brew services restart nginx 2>/dev/null || nginx -s reload; } || true
  fi

  ok "Nginx 配置完成"
}

_write_nginx_http() {
  info "生成 HTTP 反向代理配置..."
  local tee_cmd; [ "$OS" = "darwin" ] && tee_cmd="tee" || tee_cmd="sudo tee"
  $tee_cmd "$NGINX_CONF" > /dev/null <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

    # 安全头
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;

    location / {
        proxy_pass         http://127.0.0.1:${PORT};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF
  ok "HTTP 配置 → ${NGINX_CONF}"
  warn "SSL 未配置，建议后续运行: sudo certbot --nginx -d ${DOMAIN}"
}

_write_nginx_ssl() {
  # 使用 webroot 文件验证方式申请证书（兼容 Cloudflare 代理）
  step "申请 SSL 证书（文件验证）"
  info "使用 webroot 文件验证方式（兼容 Cloudflare CDN 代理）"
  info "请确保 ${DOMAIN} 已解析到本机，且 80 端口已开放"
  echo ""
  prompt_yn "DNS 已解析，80 端口已开放？继续申请？" "y" || {
    warn "跳过 SSL 申请，已配置 HTTP 反向代理"
    USE_SSL=false; _write_nginx_http; return
  }

  # 1) 创建 webroot 验证目录
  local webroot="/var/www/certbot"
  sudo mkdir -p "${webroot}/.well-known/acme-challenge"
  sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
    || sudo chown -R "$(whoami)" "$webroot"

  # 2) 写入临时 Nginx 配置，仅提供 HTTP 并开放验证路径
  info "配置临时 Nginx 用于文件验证..."
  local tee_cmd; [ "$OS" = "darwin" ] && tee_cmd="tee" || tee_cmd="sudo tee"
  $tee_cmd "$NGINX_CONF" > /dev/null <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

    # Let's Encrypt 文件验证
    location /.well-known/acme-challenge/ {
        root ${webroot};
        allow all;
    }

    location / {
        proxy_pass         http://127.0.0.1:${PORT};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }
}
EOF

  # 启用临时配置并重载 Nginx
  if [ "$OS" = "linux" ]; then
    sudo ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/smtp-lite"
    sudo nginx -t && sudo systemctl reload nginx
  else
    nginx -t && (brew services restart nginx 2>/dev/null || nginx -s reload)
  fi
  ok "临时 HTTP 配置已应用，准备验证"

  # 3) 使用 webroot 方式申请证书
  local certbot_args=("certonly" "--webroot" "-w" "$webroot" "-d" "$DOMAIN" "--non-interactive" "--agree-tos")
  [ -n "$CERTBOT_EMAIL" ] && certbot_args+=("--email" "$CERTBOT_EMAIL") \
                          || certbot_args+=("--register-unsafely-without-email")

  if sudo certbot "${certbot_args[@]}"; then
    ok "SSL 证书申请成功"
    _write_nginx_ssl_conf
    # 4) 自动续期 cron（使用 webroot 方式续期）
    if [ "$OS" = "linux" ]; then
      ( crontab -l 2>/dev/null || true; echo "0 3 * * * certbot renew --quiet --webroot -w ${webroot} --deploy-hook 'systemctl reload nginx'" ) \
        | sort -u | crontab -
      ok "已添加证书自动续期 Cron（每天凌晨 3:00，webroot 文件验证）"
    fi
  else
    warn "SSL 申请失败（请检查 DNS/防火墙），已降级为 HTTP 配置"
    USE_SSL=false; _write_nginx_http
  fi
}

_write_nginx_ssl_conf() {
  local cert_dir="/etc/letsencrypt/live/${DOMAIN}"
  local webroot="/var/www/certbot"
  local tee_cmd; [ "$OS" = "darwin" ] && tee_cmd="tee" || tee_cmd="sudo tee"
  $tee_cmd "$NGINX_CONF" > /dev/null <<EOF
# HTTP → HTTPS 重定向（保留证书续期验证路径）
server {
    listen 80;
    server_name ${DOMAIN};

    # Let's Encrypt 文件验证（续期用）
    location /.well-known/acme-challenge/ {
        root ${webroot};
        allow all;
    }

    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

# HTTPS
server {
    listen 443 ssl http2;
    server_name ${DOMAIN};

    ssl_certificate     ${cert_dir}/fullchain.pem;
    ssl_certificate_key ${cert_dir}/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache   shared:SSL:10m;
    ssl_session_timeout 1d;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;

    location / {
        proxy_pass         http://127.0.0.1:${PORT};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF
  ok "HTTPS 配置 → ${NGINX_CONF}"

  # 重载 Nginx 使 SSL 配置生效
  if [ "$OS" = "linux" ]; then
    sudo nginx -t && sudo systemctl reload nginx
  else
    nginx -t && (brew services restart nginx 2>/dev/null || nginx -s reload)
  fi
}

# ── 系统服务 ──────────────────────────────────────────────────
setup_service() {
  step "系统服务"
  if ! prompt_yn "是否配置为开机自启服务？" "y"; then
    info "跳过，手动启动: cd ${INSTALL_DIR} && ./smtp-lite"
    return
  fi

  if [ "$OS" = "linux" ] && command -v systemctl &>/dev/null; then
    _setup_systemd
  elif [ "$OS" = "darwin" ]; then
    _setup_launchd
  else
    warn "未找到服务管理器，手动启动: cd ${INSTALL_DIR} && ./smtp-lite"
  fi
}

# ── 安装 CLI 管理工具 ────────────────────────────────────────
setup_cli() {
  step "安装 CLI 管理工具"
  cd "$INSTALL_DIR"
  chmod +x cli.sh
  sudo ln -sf "$INSTALL_DIR/cli.sh" /usr/local/bin/smtp-lite
  ok "CLI 管理工具已安装 → /usr/local/bin/smtp-lite"
  info "用法: smtp-lite help"
}

_setup_systemd() {
  local svc="/etc/systemd/system/${SERVICE_NAME}.service"
  sudo tee "$svc" > /dev/null <<EOF
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
  sudo systemctl enable  "$SERVICE_NAME"
  sudo systemctl restart "$SERVICE_NAME"
  ok "Systemd 服务已启动 (systemctl status ${SERVICE_NAME})"
}

_setup_launchd() {
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
  launchctl unload "$plist" 2>/dev/null || true
  launchctl load   "$plist"
  ok "LaunchAgent 已加载（开机自启）"
}

# ── 完成提示 ──────────────────────────────────────────────────
print_done() {
  # 计算访问地址
  local access_url
  if $USE_SSL; then
    access_url="https://${DOMAIN}"
  elif $USE_DOMAIN; then
    access_url="http://${DOMAIN}"
  else
    access_url="http://localhost:${PORT}"
  fi

  echo ""
  echo -e "${G}  ╔══════════════════════════════════════════════╗${N}"
  echo -e "${G}  ║${N}           ${W}🎉  安装完成！${N}                    ${G}║${N}"
  echo -e "${G}  ╠══════════════════════════════════════════════╣${N}"
  echo -e "${G}  ║${N}  访问地址  ${C}${access_url}${N}"
  echo -e "${G}  ║${N}  管理账号  ${W}${ADMIN_USER}${N} / ${DIM}(你设置的密码)${N}"
  echo -e "${G}  ║${N}  安装目录  ${INSTALL_DIR}"
  echo -e "${G}  ║${N}  运行日志  ${INSTALL_DIR}/smtp-lite.log"
  if $USE_SSL; then
    echo -e "${G}  ║${N}  SSL 证书  ${G}Let's Encrypt（90天自动续期）${N}"
  fi
  echo -e "${G}  ╠══════════════════════════════════════════════╣${N}"
  echo -e "${G}  ║${N}  管理工具  ${W}smtp-lite help${N}"
  echo -e "${G}  ║${N}  一键更新  ${W}smtp-lite update${N}"
  if [ "$OS" = "linux" ] && command -v systemctl &>/dev/null; then
    echo -e "${G}  ║${N}  查看日志  ${W}journalctl -u ${SERVICE_NAME} -f${N}"
    echo -e "${G}  ║${N}  重启服务  ${W}systemctl restart ${SERVICE_NAME}${N}"
  fi
  echo -e "${G}  ╚══════════════════════════════════════════════╝${N}"
  echo ""
  warn "首次登录后请前往「设置」修改密码"
}

# ── 主流程 ────────────────────────────────────────────────────
banner() {
  clear 2>/dev/null || true
  echo ""
  echo -e "${B}   ███████╗███╗   ███╗████████╗██████╗     ██╗     ██╗████████╗███████╗${N}"
  echo -e "${B}   ██╔════╝████╗ ████║╚══██╔══╝██╔══██╗    ██║     ██║╚══██╔══╝██╔════╝${N}"
  echo -e "${B}   ███████╗██╔████╔██║   ██║   ██████╔╝    ██║     ██║   ██║   █████╗  ${N}"
  echo -e "${B}   ╚════██║██║╚██╔╝██║   ██║   ██╔═══╝     ██║     ██║   ██║   ██╔══╝  ${N}"
  echo -e "${B}   ███████║██║ ╚═╝ ██║   ██║   ██║         ███████╗██║   ██║   ███████╗${N}"
  echo -e "${B}   ╚══════╝╚═╝     ╚═╝   ╚═╝   ╚═╝         ╚══════╝╚═╝   ╚═╝   ╚══════╝${N}"
  echo ""
  echo -e "   ${W}个人邮箱聚合发送系统${N}  ${DIM}·  交互式安装程序${N}"
  echo ""
}

main() {
  banner
  collect_config      # 交互式收集所有配置
  detect_system       # 检测 OS/架构
  step "依赖检查"
  check_git
  check_go
  check_nginx
  check_certbot
  setup_repo          # 克隆/更新仓库
  build               # 编译
  write_config        # 写入 config.yaml
  setup_nginx         # Nginx 配置（如需要）
  setup_service       # 系统服务（如需要）
  setup_cli           # 安装 CLI 管理工具
  print_done
}

main "$@"
