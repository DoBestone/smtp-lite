#!/usr/bin/env bash
# =============================================================
#  SMTP Lite - 自动更新脚本 v2.2
#  默认仅使用 GitHub Release 预编译二进制（含内嵌前端），
#  确保线上一致性；源码编译需显式 --source 才会启用。
#  用法:
#    bash update.sh           检查并下载预编译二进制
#    bash update.sh --force   强制重新下载（即使版本相同）
#    bash update.sh --source  允许在预编译不可用时回退到源码编译
#    bash update.sh --frontend-only  只刷新前端（保留后端二进制不变）
# =============================================================
set -euo pipefail

# 识别执行方式：
# - 普通 bash update.sh → BASH_SOURCE 指向真实路径
# - bash <(curl ...)    → BASH_SOURCE 是 /dev/fd/63 等虚拟 fd，不能写文件
# - curl ... | bash     → BASH_SOURCE 为空
# 后两种情况下用 $PWD 作为安装目录（用户应在项目目录里执行）
_resolve_script_dir() {
  local src="${BASH_SOURCE[0]:-}"
  if [ -z "$src" ] || [[ "$src" == /dev/fd/* ]] || [[ "$src" == /proc/self/fd/* ]] || [[ "$src" == pipe:* ]]; then
    echo "$PWD"
  else
    (cd "$(dirname "$src")" && pwd)
  fi
}
SCRIPT_DIR="$(_resolve_script_dir)"
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
ALLOW_SOURCE=false
FRONTEND_ONLY=false
for arg in "$@"; do
  case "$arg" in
    --force) FORCE=true ;;
    --source) ALLOW_SOURCE=true ;;
    --frontend-only) FRONTEND_ONLY=true ;;
    *) warn "未知参数: $arg" ;;
  esac
done

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
info "安装目录: ${SCRIPT_DIR}"

# 目录必须可写（避免 /dev/fd 等场景下失败在下载中途）
if [ ! -d "$SCRIPT_DIR" ] || [ ! -w "$SCRIPT_DIR" ]; then
  err "安装目录不可写: ${SCRIPT_DIR}\n请 cd 到 smtp-lite 安装目录后再运行 update.sh"
fi

# 简单合理性检查：目录看起来不像项目目录就警告
if [ ! -f "$SCRIPT_DIR/smtp-lite" ] \
   && [ ! -f "$SCRIPT_DIR/config.yaml" ] \
   && [ ! -f "$SCRIPT_DIR/smtp-lite.db" ]; then
  warn "当前目录 ${SCRIPT_DIR} 不像 smtp-lite 安装目录（未找到二进制/config/db），继续执行将在此新建文件"
fi

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

# 提取所有下载链接供按名称匹配
ASSET_URLS=$(echo "$RELEASE_JSON" \
  | grep '"browser_download_url"' \
  | grep -o '"https://[^"]*"' | tr -d '"' || true)

# 精确匹配二进制 (文件名末尾要么行尾要么是 .sha256)
DOWNLOAD_URL=$(echo "$ASSET_URLS" | awk -v name="$ASSET_NAME" '
  $0 ~ "/" name "$" { print; exit }
' || true)

# 独立前端 tarball (如果 CI 上传了)
FRONTEND_URL=$(echo "$ASSET_URLS" | grep -E '/smtp-lite-frontend\.tar\.gz$' | head -1 || true)

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
  TMP_BINARY=$(mktemp "${SCRIPT_DIR}/.smtp-lite.tmp.XXXXXX" 2>/dev/null \
    || mktemp -t smtp-lite.tmp.XXXXXX)
  trap 'rm -f "$TMP_BINARY"' EXIT

  if ! curl -fL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BINARY"; then
    warn "下载失败"
    rm -f "$TMP_BINARY"
    return 1
  fi

  chmod +x "$TMP_BINARY"

  # SHA256 校验
  CHECKSUM_URL="${DOWNLOAD_URL}.sha256"
  EXPECTED_SHA=$(curl -fsSL "$CHECKSUM_URL" 2>/dev/null | awk '{print $1}')
  if [ -n "$EXPECTED_SHA" ]; then
    if command -v sha256sum &>/dev/null; then
      ACTUAL_SHA=$(sha256sum "$TMP_BINARY" | awk '{print $1}')
    elif command -v shasum &>/dev/null; then
      ACTUAL_SHA=$(shasum -a 256 "$TMP_BINARY" | awk '{print $1}')
    else
      warn "sha256sum/shasum 不可用，跳过校验"
      ACTUAL_SHA="$EXPECTED_SHA"
    fi
    if [ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]; then
      rm -f "$TMP_BINARY"
      err "SHA256 校验失败！预期: ${EXPECTED_SHA}，实际: ${ACTUAL_SHA}"
      return 1
    fi
    ok "SHA256 校验通过"
  else
    warn "未找到 .sha256 校验文件，跳过校验"
  fi

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
    git fetch origin || err "git fetch 失败"
    local branch
    branch=$(git symbolic-ref --short HEAD 2>/dev/null || echo "master")
    git reset --hard "origin/${branch}" || err "git reset 失败"
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

# 收集占用目标端口的 PID 列表（空格分隔）
_port_pids() {
  local port="$1"
  local pids=""
  if command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -ti:"${port}" -sTCP:LISTEN 2>/dev/null | tr '\n' ' ' || true)
  fi
  if [ -z "$pids" ] && command -v ss >/dev/null 2>&1; then
    pids=$(ss -tlnp 2>/dev/null \
      | awk -v p=":${port}" '$4 ~ p { print }' \
      | grep -oE 'pid=[0-9]+' | cut -d= -f2 | sort -u | tr '\n' ' ' || true)
  fi
  echo "$pids"
}

# 轮询直到端口释放或超时（秒）
_wait_port_free() {
  local port="$1"
  local timeout="${2:-15}"
  local i=0
  while [ $i -lt "$timeout" ]; do
    local pids
    pids=$(_port_pids "$port")
    [ -z "${pids// }" ] && return 0
    sleep 1
    i=$((i + 1))
  done
  return 1
}

_restart_process() {
  # 1) 按端口精确找到占用者（比 pgrep 靠谱），再补一次 pgrep 作 fallback
  local port_pids pgrep_pids all_pids
  port_pids=$(_port_pids "$PORT")
  pgrep_pids=$(pgrep -f "${BINARY}" 2>/dev/null | tr '\n' ' ' || true)
  # 合并去重，排除当前 shell 及其父进程
  all_pids=$(echo "$port_pids $pgrep_pids" | tr ' ' '\n' \
    | awk -v s=$$ -v p=$PPID 'NF && $0!=s && $0!=p' \
    | sort -u | tr '\n' ' ')

  if [ -n "${all_pids// }" ]; then
    info "停止旧进程: ${all_pids}"
    # SIGTERM
    for pid in $all_pids; do
      kill "$pid" 2>/dev/null || true
    done
    # 最多等 10s 优雅退出
    for _i in 1 2 3 4 5 6 7 8 9 10; do
      local alive=""
      for pid in $all_pids; do
        kill -0 "$pid" 2>/dev/null && alive="$alive $pid"
      done
      [ -z "${alive// }" ] && break
      sleep 1
    done
    # 还活着就 SIGKILL
    for pid in $all_pids; do
      if kill -0 "$pid" 2>/dev/null; then
        warn "PID ${pid} 未响应 SIGTERM，强制 SIGKILL"
        kill -9 "$pid" 2>/dev/null || true
      fi
    done
  fi

  # 2) 等端口真正释放（内核可能还占着几秒）
  if ! _wait_port_free "$PORT" 15; then
    local leftover
    leftover=$(_port_pids "$PORT")
    err "端口 ${PORT} 仍被占用: ${leftover}\n请手动处理: kill -9 ${leftover}"
  fi

  info "启动新进程..."
  cd "$SCRIPT_DIR"
  nohup "$BINARY" >> "$SCRIPT_DIR/smtp-lite.log" 2>&1 &
  NEW_PID=$!
  # 给足服务启动时间（sqlite 迁移/路由注册），并二次确认进程真的在监听端口
  sleep 3
  if ! kill -0 "$NEW_PID" 2>/dev/null; then
    err "服务启动失败（PID ${NEW_PID} 已退出），请检查日志: $SCRIPT_DIR/smtp-lite.log"
  fi
  if ! _port_pids "$PORT" | grep -q .; then
    warn "进程存活但尚未监听端口 ${PORT}，再等 5s"
    sleep 5
    if ! _port_pids "$PORT" | grep -q .; then
      err "服务未能绑定端口 ${PORT}，请检查日志: $SCRIPT_DIR/smtp-lite.log"
    fi
  fi
  ok "服务已启动 (PID: ${NEW_PID}, 监听 :${PORT})"
}

# ── 仅更新前端 ───────────────────────────────────────────────
update_frontend_only() {
  if [ -z "$FRONTEND_URL" ]; then
    err "Release ${LATEST} 未提供独立前端包 (smtp-lite-frontend.tar.gz)"
  fi
  info "下载前端 tarball..."
  local tmp_tar
  tmp_tar=$(mktemp "${SCRIPT_DIR}/.frontend.tmp.XXXXXX" 2>/dev/null \
    || mktemp -t smtp-lite-frontend.tmp.XXXXXX)
  trap 'rm -f "$tmp_tar"' EXIT
  curl -fL --progress-bar "$FRONTEND_URL" -o "$tmp_tar" || err "前端下载失败"

  # 校验 sha256（可选）
  local sha_url expected actual
  sha_url="${FRONTEND_URL}.sha256"
  expected=$(curl -fsSL "$sha_url" 2>/dev/null | awk '{print $1}' || true)
  if [ -n "$expected" ]; then
    if command -v sha256sum &>/dev/null; then
      actual=$(sha256sum "$tmp_tar" | awk '{print $1}')
    else
      actual=$(shasum -a 256 "$tmp_tar" | awk '{print $1}')
    fi
    [ "$expected" = "$actual" ] || err "前端包 SHA256 校验失败"
    ok "前端 SHA256 校验通过"
  fi

  mkdir -p "$SCRIPT_DIR/web/dist"
  rm -rf "$SCRIPT_DIR/web/dist"/*
  tar -xzf "$tmp_tar" -C "$SCRIPT_DIR/web/dist"
  rm -f "$tmp_tar"
  ok "前端已更新到 ${LATEST}"
}

# ── 主流程 ────────────────────────────────────────────────────
echo ""
echo -e "  ${W}SMTP Lite 更新${N}"
echo ""

if [ "$FRONTEND_ONLY" = true ]; then
  update_frontend_only
  restart_service
  echo ""
  echo -e "  ${G}前端更新完成 → ${W}${LATEST}${N}"
  echo ""
  exit 0
fi

if update_binary; then
  info "使用: 预编译二进制"
elif [ "$ALLOW_SOURCE" = true ]; then
  warn "预编译二进制不可用，按 --source 要求回退到源码编译..."
  update_source
  info "使用: 源码编译"
else
  err "预编译二进制不可用。如确实需要源码编译，请加 --source 参数重试"
fi

restart_service

echo ""
echo -e "  ${G}更新完成: ${CURRENT} → ${W}${LATEST}${N}"
echo ""
