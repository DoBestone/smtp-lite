#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 交互式安装脚本 v2.1
#  支持: macOS (Intel/Apple Silicon) · Linux (x86_64/arm64/armv7/386/mips64le/riscv64)
#  特性: 优先下载预编译二进制，失败则源码编译
#        支持 SQLite / MySQL 数据库选择
#  用法: bash install.sh
# =============================================================
set -euo pipefail

# ── 全局常量 ──────────────────────────────────────────────────
REPO="https://github.com/DoBestone/smtp-lite.git"
GITHUB_REPO="DoBestone/smtp-lite"
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

prompt_choice() {
  local __var="$1" __prompt="$2"
  shift 2
  local __options=("$@")
  echo -e "  ${W}${__prompt}${N}"
  local i=1
  for opt in "${__options[@]}"; do
    echo -e "    ${C}${i})${N} ${opt}"
    i=$((i+1))
  done
  local __choice
  while true; do
    echo -ne "  ${C}→${N} 请选择 [1-${#__options[@]}]: "
    read -r __choice
    if [[ "$__choice" =~ ^[0-9]+$ ]] && [ "$__choice" -ge 1 ] && [ "$__choice" -le "${#__options[@]}" ]; then
      eval "$__var='$__choice'"
      return
    fi
    warn "无效选择，请重新输入"
  done
}

# ── 全局变量 ──────────────────────────────────────────────────
INSTALL_DIR="" PORT="" ADMIN_USER="" ADMIN_PASS=""
USE_DOMAIN=false DOMAIN="" USE_NGINX=false USE_SSL=false
NGINX_CONF="" CERTBOT_EMAIL=""
# 数据库相关
DB_DRIVER="sqlite" DB_HOST="" DB_PORT="" DB_NAME="" DB_USER="" DB_PASS=""
MYSQL_AUTO_INSTALL=false
# 安装方式
INSTALL_MODE="" # binary 或 source
OS="" ARCH="" PKG=""

# ── 用户配置收集 ──────────────────────────────────────────────
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
  prompt_input ADMIN_USER "管理员用户名" "admin"

  while true; do
    prompt_secret ADMIN_PASS "管理员密码（至少 8 位）"
    [ ${#ADMIN_PASS} -ge 8 ] || { warn "密码至少 8 位，请重新输入"; continue; }
    local __confirm
    prompt_secret __confirm "再次确认密码"
    [ "$ADMIN_PASS" = "$__confirm" ] && break
    warn "两次密码不一致，请重新输入"
  done

  divider
  echo -e "  ${W}数据库配置${N}"
  local db_choice
  prompt_choice db_choice "选择数据库类型" "SQLite（轻量级，无需额外安装）" "MySQL（生产推荐，支持高并发）"

  if [ "$db_choice" = "2" ]; then
    DB_DRIVER="mysql"
    local mysql_choice
    prompt_choice mysql_choice "MySQL 配置方式" "自动安装 MySQL 并创建数据库" "手动输入已有 MySQL 连接信息"

    if [ "$mysql_choice" = "1" ]; then
      MYSQL_AUTO_INSTALL=true
      DB_HOST="127.0.0.1"
      DB_PORT="3306"
      DB_NAME="smtp_lite"
      DB_USER="smtp_lite"
      info "将自动安装 MySQL 并创建数据库"
      prompt_secret DB_PASS "设置 MySQL smtp_lite 用户密码（至少 8 位）"
      while [ ${#DB_PASS} -lt 8 ]; do
        warn "密码至少 8 位"
        prompt_secret DB_PASS "设置 MySQL smtp_lite 用户密码（至少 8 位）"
      done
    else
      MYSQL_AUTO_INSTALL=false
      prompt_input DB_HOST "MySQL 主机地址" "127.0.0.1"
      prompt_input DB_PORT "MySQL 端口" "3306"
      prompt_input DB_NAME "数据库名" "smtp_lite"
      prompt_input DB_USER "数据库用户名" "smtp_lite"
      prompt_secret DB_PASS "数据库密码"
    fi
  fi

  divider
  echo -e "  ${W}域名 & SSL${N}"

  if prompt_yn "是否绑定自定义域名？"; then
    USE_DOMAIN=true
    while true; do
      prompt_input DOMAIN "域名（如 smtp.example.com）" ""
      [[ "$DOMAIN" =~ ^[a-zA-Z0-9]([a-zA-Z0-9\-]*\.)+[a-zA-Z]{2,}$ ]] && break
      warn "域名格式不正确，请重新输入"
    done
    if prompt_yn "是否安装/配置 Nginx 反向代理？" "y"; then
      USE_NGINX=true
      if prompt_yn "是否申请 Let's Encrypt SSL 证书？" "y"; then
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
  if [ "$DB_DRIVER" = "mysql" ]; then
    echo -e "  ${DIM}数据库   ${N}→ ${C}MySQL${N} (${DB_HOST}:${DB_PORT}/${DB_NAME})"
    $MYSQL_AUTO_INSTALL && echo -e "  ${DIM}MySQL    ${N}→ ${G}自动安装${N}"
  else
    echo -e "  ${DIM}数据库   ${N}→ ${C}SQLite${N}"
  fi
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
    x86_64|amd64)    ARCH="amd64" ;;
    aarch64|arm64)   ARCH="arm64" ;;
    armv7*|armhf)    ARCH="armv7" ;;
    i686|i386)       ARCH="386" ;;
    mips64el|mips64) ARCH="mips64le" ;;
    riscv64)         ARCH="riscv64" ;;
    *)               err "不支持的架构: $(uname -m)" ;;
  esac
  if   command -v brew    &>/dev/null; then PKG="brew"
  elif command -v apt-get &>/dev/null; then PKG="apt"
  elif command -v dnf     &>/dev/null; then PKG="dnf"
  elif command -v yum     &>/dev/null; then PKG="yum"
  else                                      PKG="none"
  fi
  ok "系统: ${OS}/${ARCH}  包管理器: ${PKG}"
}

# ── 依赖安装 ──────────────────────────────────────────────────
check_curl() {
  if command -v curl &>/dev/null; then
    ok "curl $(curl --version | head -1 | awk '{print $2}')"
    return
  fi
  warn "curl 未安装，正在安装..."
  case "$PKG" in
    brew) brew install curl ;;
    apt)  sudo apt-get install -y curl ;;
    dnf)  sudo dnf install -y curl ;;
    yum)  sudo yum install -y curl ;;
    *)    err "请手动安装 curl" ;;
  esac
  ok "curl 安装完成"
}

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
  if ! command -v go &>/dev/null && [ -x /usr/local/go/bin/go ]; then
    export PATH=$PATH:/usr/local/go/bin
  fi
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

check_node() {
  if command -v node &>/dev/null; then
    ok "Node.js $(node -v)"
    return
  fi
  warn "Node.js 未安装，正在安装..."
  case "$PKG" in
    brew) brew install node ;;
    apt)
      curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
      sudo apt-get install -y nodejs
      ;;
    dnf)  sudo dnf install -y nodejs npm ;;
    yum)  sudo yum install -y nodejs npm ;;
    *)    err "请手动安装 Node.js: https://nodejs.org" ;;
  esac
  ok "Node.js $(node -v) 安装完成"
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

# ── MySQL 安装与配置 ─────────────────────────────────────────
install_mysql() {
  $MYSQL_AUTO_INSTALL || return
  step "安装 MySQL"

  if command -v mysql &>/dev/null; then
    ok "MySQL 已安装: $(mysql --version | head -1)"
  else
    info "安装 MySQL Server..."
    case "$PKG" in
      brew)
        brew install mysql
        brew services start mysql
        ;;
      apt)
        sudo DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server
        sudo systemctl start mysql
        sudo systemctl enable mysql
        ;;
      dnf)
        sudo dnf install -y mysql-server
        sudo systemctl start mysqld
        sudo systemctl enable mysqld
        ;;
      yum)
        sudo yum install -y mysql-server
        sudo systemctl start mysqld
        sudo systemctl enable mysqld
        ;;
      *)
        err "请手动安装 MySQL Server"
        ;;
    esac
    ok "MySQL 安装完成"
  fi

  # 等待 MySQL 就绪
  info "等待 MySQL 启动..."
  local retries=0
  while ! mysqladmin ping -h 127.0.0.1 --silent 2>/dev/null; do
    retries=$((retries+1))
    [ $retries -ge 30 ] && err "MySQL 启动超时"
    sleep 1
  done
  ok "MySQL 已就绪"

  # 生成安全随机 root 密码（仅限自动安装模式）
  local MYSQL_ROOT_PASS
  MYSQL_ROOT_PASS=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 20 || true)

  # 创建数据库和用户
  info "创建数据库和用户..."
  # 尝试无密码连接（新装的 MySQL 在某些系统上允许）
  local mysql_cmd="mysql"
  if sudo mysql -e "SELECT 1" &>/dev/null; then
    mysql_cmd="sudo mysql"
  elif mysql -u root -e "SELECT 1" &>/dev/null; then
    mysql_cmd="mysql -u root"
  else
    warn "无法以 root 无密码连接 MySQL"
    local root_pass=""
    prompt_secret root_pass "请输入 MySQL root 密码"
    mysql_cmd="mysql -u root -p${root_pass}"
  fi

  $mysql_cmd <<EOF
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PASS}';
CREATE USER IF NOT EXISTS '${DB_USER}'@'127.0.0.1' IDENTIFIED BY '${DB_PASS}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'127.0.0.1';
FLUSH PRIVILEGES;
EOF
  ok "数据库 ${DB_NAME} 和用户 ${DB_USER} 已创建"
}

verify_mysql_connection() {
  [ "$DB_DRIVER" = "mysql" ] || return
  $MYSQL_AUTO_INSTALL && return  # 自动安装的已在上面验证过
  info "验证 MySQL 连接..."
  if command -v mysql &>/dev/null; then
    if mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" -p"${DB_PASS}" -e "USE \`${DB_NAME}\`" 2>/dev/null; then
      ok "MySQL 连接验证成功"
    else
      warn "MySQL 连接失败，请确认数据库配置正确"
      if ! prompt_yn "继续安装？"; then
        err "已取消安装"
      fi
    fi
  else
    warn "未安装 mysql 客户端，跳过连接验证"
  fi
}

# ── 获取二进制（优先策略） ───────────────────────────────────
try_download_binary() {
  step "获取程序"
  mkdir -p "$INSTALL_DIR"

  local asset_name="smtp-lite-${OS}-${ARCH}"
  info "尝试下载预编译二进制文件..."

  # 获取最新 Release
  local release_json=""
  release_json=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null || true)
  local download_url=""
  if [ -n "$release_json" ]; then
    download_url=$(echo "$release_json" \
      | grep '"browser_download_url"' \
      | grep "${asset_name}" \
      | grep -o '"https://[^"]*"' | tr -d '"' | head -1 || true)
  fi

  if [ -n "$download_url" ]; then
    info "下载: ${download_url}"
    local tmp_binary
    tmp_binary=$(mktemp "${INSTALL_DIR}/.smtp-lite.tmp.XXXXXX")
    if curl -fL --progress-bar "$download_url" -o "$tmp_binary"; then
      chmod +x "$tmp_binary"
      # 验证可执行
      if "$tmp_binary" --version &>/dev/null; then
        mv "$tmp_binary" "$INSTALL_DIR/smtp-lite"
        INSTALL_MODE="binary"
        ok "预编译二进制下载成功"
        # 还需要克隆仓库获取脚本文件
        clone_scripts_only
        return
      else
        warn "下载的二进制文件无法执行"
        rm -f "$tmp_binary"
      fi
    else
      warn "下载失败"
      rm -f "$tmp_binary"
    fi
  else
    warn "未找到预编译二进制 (${asset_name})"
  fi

  # 回退到源码编译
  info "回退到源码编译模式..."
  INSTALL_MODE="source"
  setup_source_deps
  setup_repo
  build_frontend
  build_from_source
}

# 仅克隆脚本和配置（不编译）
clone_scripts_only() {
  info "获取项目脚本文件..."
  if [ -d "$INSTALL_DIR/.git" ]; then
    git -C "$INSTALL_DIR" checkout -- . 2>/dev/null || true
    git -C "$INSTALL_DIR" pull 2>/dev/null || true
  else
    # 浅克隆
    local tmp_clone
    tmp_clone=$(mktemp -d)
    if git clone --depth 1 "$REPO" "$tmp_clone" 2>/dev/null; then
      # 复制脚本文件
      for f in cli.sh install.sh update.sh; do
        [ -f "$tmp_clone/$f" ] && cp "$tmp_clone/$f" "$INSTALL_DIR/$f"
      done
      # 复制前端 dist（如果二进制嵌入了则不需要，但保险起见）
      [ -d "$tmp_clone/web/dist" ] && cp -r "$tmp_clone/web" "$INSTALL_DIR/"
      rm -rf "$tmp_clone"
      ok "脚本文件已复制"
    else
      warn "无法克隆仓库获取脚本，部分管理功能可能不可用"
    fi
  fi
}

# 源码模式需要的额外依赖
setup_source_deps() {
  step "依赖检查（源码编译）"
  check_git
  check_go
  check_node
}

setup_repo() {
  if [ -d "$INSTALL_DIR/.git" ]; then
    info "已有安装目录，更新代码..."
    git -C "$INSTALL_DIR" checkout -- . 2>/dev/null || true
    git -C "$INSTALL_DIR" pull
    ok "代码更新完成"
  else
    info "克隆仓库 → ${INSTALL_DIR}"
    git clone "$REPO" "$INSTALL_DIR"
    ok "克隆完成"
  fi
}

build_frontend() {
  step "构建前端"
  cd "$INSTALL_DIR/frontend"
  info "npm install ..."
  npm install --silent
  info "npm run build ..."
  npm run build
  ok "前端构建完成"
}

build_from_source() {
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

  local jwt_secret enc_key
  jwt_secret=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)
  enc_key=$(LC_ALL=C    tr -dc 'A-Za-z0-9!@#%^&*_-' </dev/urandom | head -c 32 || true)

  if [ "$DB_DRIVER" = "mysql" ]; then
    cat > config.yaml <<EOF
server:
  port: ${PORT}
  mode: release

database:
  driver: mysql
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DB_NAME}
  user: ${DB_USER}
  password: "${DB_PASS}"

auth:
  username: ${ADMIN_USER}
  password: ${ADMIN_PASS}

jwt:
  secret: ${jwt_secret}
  expire_hours: 168

encryption:
  key: ${enc_key}
EOF
  else
    cat > config.yaml <<EOF
server:
  port: ${PORT}
  mode: release

database:
  driver: sqlite
  sqlite: smtp-lite.db

auth:
  username: ${ADMIN_USER}
  password: ${ADMIN_PASS}

jwt:
  secret: ${jwt_secret}
  expire_hours: 168

encryption:
  key: ${enc_key}
EOF
  fi
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
  step "申请 SSL 证书（文件验证）"
  info "使用 webroot 文件验证方式（兼容 Cloudflare CDN 代理）"
  prompt_yn "DNS 已解析，80 端口已开放？继续申请？" "y" || {
    warn "跳过 SSL 申请，已配置 HTTP 反向代理"
    USE_SSL=false; _write_nginx_http; return
  }

  local webroot="/var/www/certbot"
  sudo mkdir -p "${webroot}/.well-known/acme-challenge"
  sudo chown -R www-data:www-data "$webroot" 2>/dev/null \
    || sudo chown -R "$(whoami)" "$webroot"

  info "配置临时 Nginx 用于文件验证..."
  local tee_cmd; [ "$OS" = "darwin" ] && tee_cmd="tee" || tee_cmd="sudo tee"
  $tee_cmd "$NGINX_CONF" > /dev/null <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

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

  if [ "$OS" = "linux" ]; then
    sudo ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/smtp-lite"
    sudo nginx -t && sudo systemctl reload nginx
  else
    nginx -t && (brew services restart nginx 2>/dev/null || nginx -s reload)
  fi
  ok "临时 HTTP 配置已应用"

  local certbot_args=("certonly" "--webroot" "-w" "$webroot" "-d" "$DOMAIN" "--non-interactive" "--agree-tos")
  [ -n "$CERTBOT_EMAIL" ] && certbot_args+=("--email" "$CERTBOT_EMAIL") \
                          || certbot_args+=("--register-unsafely-without-email")

  if sudo certbot "${certbot_args[@]}"; then
    ok "SSL 证书申请成功"
    _write_nginx_ssl_conf
    if [ "$OS" = "linux" ]; then
      ( crontab -l 2>/dev/null || true; echo "0 3 * * * certbot renew --quiet --webroot -w ${webroot} --deploy-hook 'systemctl reload nginx'" ) \
        | sort -u | crontab -
      ok "已添加证书自动续期 Cron"
    fi
  else
    warn "SSL 申请失败，已降级为 HTTP 配置"
    USE_SSL=false; _write_nginx_http
  fi
}

_write_nginx_ssl_conf() {
  local cert_dir="/etc/letsencrypt/live/${DOMAIN}"
  local webroot="/var/www/certbot"
  local tee_cmd; [ "$OS" = "darwin" ] && tee_cmd="tee" || tee_cmd="sudo tee"
  $tee_cmd "$NGINX_CONF" > /dev/null <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

    location /.well-known/acme-challenge/ {
        root ${webroot};
        allow all;
    }

    location / {
        if (\$http_x_forwarded_proto = "https") {
            proxy_pass http://127.0.0.1:${PORT};
            break;
        }
        return 301 https://\$server_name\$request_uri;
    }

    proxy_http_version 1.1;
    proxy_set_header   Host              \$host;
    proxy_set_header   X-Real-IP         \$remote_addr;
    proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
    proxy_set_header   X-Forwarded-Proto \$scheme;
}

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

# ── 安装 CLI 管理工具 ────────────────────────────────────────
setup_cli() {
  step "安装 CLI 管理工具"
  cd "$INSTALL_DIR"
  chmod +x cli.sh
  sudo ln -sf "$INSTALL_DIR/cli.sh" /usr/local/bin/smtp-lite
  ok "CLI 管理工具已安装 → /usr/local/bin/smtp-lite"
  info "用法: smtp-lite (交互式菜单)"
}

# ── 完成提示 ──────────────────────────────────────────────────
print_done() {
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
  echo -e "${G}  ║${N}  安装方式  ${W}${INSTALL_MODE}${N}"
  if [ "$DB_DRIVER" = "mysql" ]; then
    echo -e "${G}  ║${N}  数据库    ${W}MySQL${N} (${DB_HOST}:${DB_PORT}/${DB_NAME})"
  else
    echo -e "${G}  ║${N}  数据库    ${W}SQLite${N}"
  fi
  echo -e "${G}  ║${N}  运行日志  ${INSTALL_DIR}/smtp-lite.log"
  if $USE_SSL; then
    echo -e "${G}  ║${N}  SSL 证书  ${G}Let's Encrypt（90天自动续期）${N}"
  fi
  echo -e "${G}  ╠══════════════════════════════════════════════╣${N}"
  echo -e "${G}  ║${N}  管理工具  ${W}smtp-lite${N}  (交互式菜单)"
  echo -e "${G}  ║${N}  一键更新  ${W}smtp-lite${N}  → 选择「在线更新」"
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
  echo -e "   ${W}个人邮箱聚合发送系统${N}  ${DIM}·  交互式安装程序 v2.1${N}"
  echo ""
}

main() {
  banner
  collect_config
  detect_system
  step "依赖检查"
  check_curl
  check_nginx
  check_certbot
  install_mysql
  verify_mysql_connection
  try_download_binary
  write_config
  setup_nginx
  setup_service
  setup_cli
  print_done
}

main "$@"
