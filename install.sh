#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 快速安装脚本
#  支持: macOS (Intel/Apple Silicon) / Linux (x86_64/arm64)
#  用法: bash install.sh
# =============================================================
set -e

REPO="https://github.com/DoBestone/smtp-lite.git"
DEFAULT_INSTALL_DIR="$HOME/smtp-lite"
MIN_GO_MAJOR=1
MIN_GO_MINOR=22
GO_INSTALL_VERSION="1.22.5"
SERVICE_NAME="smtp-lite"
PORT=8090

# ---- 颜色 ----
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; N='\033[0m'

banner() {
  echo -e "${B}"
  echo "   ███████╗███╗   ███╗████████╗██████╗     ██╗     ██╗████████╗███████╗"
  echo "   ██╔════╝████╗ ████║╚══██╔══╝██╔══██╗    ██║     ██║╚══██╔══╝██╔════╝"
  echo "   ███████╗██╔████╔██║   ██║   ██████╔╝    ██║     ██║   ██║   █████╗  "
  echo "   ╚════██║██║╚██╔╝██║   ██║   ██╔═══╝     ██║     ██║   ██║   ██╔══╝  "
  echo "   ███████║██║ ╚═╝ ██║   ██║   ██║         ███████╗██║   ██║   ███████╗"
  echo "   ╚══════╝╚═╝     ╚═╝   ╚═╝   ╚═╝         ╚══════╝╚═╝   ╚═╝   ╚══════╝"
  echo -e "${N}   ${W}个人邮箱聚合发送系统${N}  ·  快速安装脚本"
  echo ""
}

info()    { echo -e "  ${B}▸${N} $1"; }
ok()      { echo -e "  ${G}✓${N} $1"; }
warn()    { echo -e "  ${Y}⚠${N}  $1"; }
err()     { echo -e "  ${R}✗${N} $1"; exit 1; }
step()    { echo -e "\n  ${W}── $1 ──${N}"; }
ask()     { echo -ne "  ${C}?${N}  $1 "; }

# ---- 检测系统 ----
detect_system() {
  step "检测系统环境"

  case "$(uname -s)" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux"  ;;
    *)      err "不支持的操作系统: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)   ARCH="amd64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    *)              err "不支持的架构: $(uname -m)" ;;
  esac

  # 包管理器
  if   command -v brew      &>/dev/null; then PKG="brew"
  elif command -v apt-get   &>/dev/null; then PKG="apt"
  elif command -v dnf       &>/dev/null; then PKG="dnf"
  elif command -v yum       &>/dev/null; then PKG="yum"
  else                                        PKG="none"
  fi

  ok "系统: ${OS}/${ARCH}，包管理器: ${PKG}"
}

# ---- 检测/安装 Git ----
check_git() {
  step "检测 Git"
  if command -v git &>/dev/null; then
    ok "Git $(git --version | awk '{print $3}')"
    return
  fi

  warn "Git 未安装，尝试安装..."
  case "$PKG" in
    brew) brew install git ;;
    apt)  sudo apt-get install -y git ;;
    dnf)  sudo dnf install -y git ;;
    yum)  sudo yum install -y git ;;
    *)    err "请手动安装 Git：https://git-scm.com" ;;
  esac
  ok "Git 安装完成"
}

# ---- 检测/安装 Go ----
check_go() {
  step "检测 Go"
  if command -v go &>/dev/null; then
    GO_VER=$(go version | grep -o '[0-9]\+\.[0-9]\+' | head -1)
    MAJOR=$(echo "$GO_VER" | cut -d. -f1)
    MINOR=$(echo "$GO_VER" | cut -d. -f2)
    if [ "$MAJOR" -gt "$MIN_GO_MAJOR" ] || \
       { [ "$MAJOR" -eq "$MIN_GO_MAJOR" ] && [ "$MINOR" -ge "$MIN_GO_MINOR" ]; }; then
      ok "Go ${GO_VER} (满足 >= ${MIN_GO_MAJOR}.${MIN_GO_MINOR})"
      return
    fi
    warn "Go ${GO_VER} 版本过低，需要 >= ${MIN_GO_MAJOR}.${MIN_GO_MINOR}"
  else
    warn "Go 未安装"
  fi
  install_go
}

install_go() {
  info "安装 Go ${GO_INSTALL_VERSION} (${OS}/${ARCH})..."
  TMP=$(mktemp -d)
  GO_PKG="go${GO_INSTALL_VERSION}.${OS}-${ARCH}.tar.gz"
  GO_URL="https://go.dev/dl/${GO_PKG}"

  curl -fsSL --progress-bar "$GO_URL" -o "$TMP/$GO_PKG" \
    || err "下载 Go 失败，请检查网络或手动安装: https://go.dev/dl"

  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "$TMP/$GO_PKG"
  rm -rf "$TMP"

  # 写入 PATH
  export PATH=$PATH:/usr/local/go/bin
  if [ "$OS" = "darwin" ]; then
    for f in "$HOME/.zshrc" "$HOME/.bash_profile" "$HOME/.bashrc"; do
      [ -f "$f" ] && grep -q '/usr/local/go/bin' "$f" 2>/dev/null && break
      [ -f "$f" ] && echo 'export PATH=$PATH:/usr/local/go/bin' >> "$f" && break
    done
  else
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null
  fi

  ok "Go ${GO_INSTALL_VERSION} 安装完成"
}

# ---- 克隆/更新仓库 ----
setup_repo() {
  step "获取代码"

  ask "安装目录 [${DEFAULT_INSTALL_DIR}]: "
  read -r INPUT_DIR
  INSTALL_DIR="${INPUT_DIR:-$DEFAULT_INSTALL_DIR}"

  if [ -d "$INSTALL_DIR/.git" ]; then
    info "检测到已有安装，更新代码..."
    git -C "$INSTALL_DIR" pull
    ok "代码更新完成"
  else
    info "克隆仓库到 $INSTALL_DIR ..."
    git clone "$REPO" "$INSTALL_DIR"
    ok "代码克隆完成"
  fi
}

# ---- 编译 ----
build() {
  step "编译"
  info "go build ./cmd/server/ ..."
  cd "$INSTALL_DIR"
  go build -o smtp-lite ./cmd/server/
  ok "编译完成 → $INSTALL_DIR/smtp-lite"
}

# ---- 初始化配置 ----
configure() {
  step "初始化配置"
  cd "$INSTALL_DIR"

  if [ -f config.yaml ]; then
    ok "config.yaml 已存在，跳过"
    return
  fi

  cp config.yaml.example config.yaml

  # 生成随机密钥（兼容 macOS/Linux）
  rand32() { LC_ALL=C tr -dc 'A-Za-z0-9!@#$%^&*' < /dev/urandom | head -c 32; }
  JWT_SECRET=$(rand32)
  ENC_KEY=$(rand32)

  if [ "$OS" = "darwin" ]; then
    sed -i '' "s|change-this-to-random-32-byte-string|${JWT_SECRET}|g" config.yaml
    sed -i '' "s|smtp-lite-encryption-key-32b!|${ENC_KEY}|g"          config.yaml
  else
    sed -i  "s|change-this-to-random-32-byte-string|${JWT_SECRET}|g"  config.yaml
    sed -i  "s|smtp-lite-encryption-key-32b!|${ENC_KEY}|g"            config.yaml
  fi

  ok "config.yaml 已生成（JWT 和加密密钥已随机化）"
  warn "请修改 ${INSTALL_DIR}/config.yaml 中的登录密码（auth.password）"
}

# ---- 设置系统服务 ----
setup_service() {
  step "系统服务（可选）"
  ask "是否配置为开机自启服务？[y/N]: "
  read -r CHOICE

  case "$CHOICE" in
    y|Y) ;;
    *)   info "跳过，手动启动命令：cd ${INSTALL_DIR} && ./smtp-lite"; return ;;
  esac

  if [ "$OS" = "linux" ] && command -v systemctl &>/dev/null; then
    _setup_systemd
  elif [ "$OS" = "darwin" ]; then
    _setup_launchd
  else
    warn "未检测到支持的服务管理器，请手动配置"
  fi
}

_setup_systemd() {
  SVCFILE="/etc/systemd/system/${SERVICE_NAME}.service"
  sudo tee "$SVCFILE" > /dev/null <<EOF
[Unit]
Description=SMTP Lite 个人邮箱聚合系统
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/smtp-lite
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
  PLIST="$HOME/Library/LaunchAgents/com.smtp-lite.plist"
  mkdir -p "$HOME/Library/LaunchAgents"
  cat > "$PLIST" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>             <string>com.smtp-lite</string>
  <key>ProgramArguments</key>  <array><string>$INSTALL_DIR/smtp-lite</string></array>
  <key>WorkingDirectory</key>  <string>$INSTALL_DIR</string>
  <key>RunAtLoad</key>         <true/>
  <key>KeepAlive</key>         <true/>
  <key>StandardOutPath</key>   <string>$INSTALL_DIR/smtp-lite.log</string>
  <key>StandardErrorPath</key> <string>$INSTALL_DIR/smtp-lite.log</string>
</dict>
</plist>
EOF
  launchctl unload "$PLIST" 2>/dev/null || true
  launchctl load   "$PLIST"
  ok "LaunchAgent 已加载（开机自启）"
}

# ---- 完成提示 ----
print_done() {
  echo ""
  echo -e "  ${G}┌─────────────────────────────────────────┐${N}"
  echo -e "  ${G}│${N}         ${W}🎉  安装完成！${N}                  ${G}│${N}"
  echo -e "  ${G}├─────────────────────────────────────────┤${N}"
  echo -e "  ${G}│${N}  访问地址  ${C}http://localhost:${PORT}${N}          ${G}│${N}"
  echo -e "  ${G}│${N}  默认账号  ${W}admin / change-me${N}             ${G}│${N}"
  echo -e "  ${G}│${N}  安装目录  ${INSTALL_DIR}${N}"
  echo -e "  ${G}├─────────────────────────────────────────┤${N}"
  echo -e "  ${G}│${N}  一键更新  ${W}bash ${INSTALL_DIR}/update.sh${N}"
  echo -e "  ${G}└─────────────────────────────────────────┘${N}"
  echo ""
}

# ---- 主流程 ----
main() {
  banner
  detect_system
  check_git
  check_go
  setup_repo
  build
  configure
  setup_service
  print_done
}

main "$@"
