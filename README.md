<div align="center">

# SMTP Lite

**个人邮箱聚合发送系统**

轻量级 · 多账号轮询 · 自动故障切换 · 一键安装

[![Release](https://img.shields.io/github/v/release/DoBestone/smtp-lite?style=flat-square&color=blue)](https://github.com/DoBestone/smtp-lite/releases)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey?style=flat-square)]()

</div>

---

## 功能特性

| 功能 | 说明 |
|------|------|
| 🔐 单用户认证 | JWT Token + API Key 双认证方式 |
| 📧 多账号管理 | 支持 Gmail / QQ / 163 / Outlook 及任意 SMTP |
| 🔄 智能轮询 | 自动选择可用账号，失败时自动切换，最多重试 3 次 |
| 📤 统一发送 | 支持 HTML / 纯文本 / CC / BCC |
| 📊 日志统计 | 发送日志分页查询、成功率、每日发送量 |
| 🔑 API Key | 应用级密钥，适合程序调用 |
| 🔒 加密存储 | SMTP 密码 AES-256 加密，API Key SHA3-256 哈希存储 |
| 📦 零依赖部署 | SQLite 数据库，单二进制文件运行 |
| 🐳 Docker 支持 | 提供 Dockerfile + docker-compose |
| 🔁 一键更新 | 网页端检测新版本，一键完成 git pull + 重编译 + 重启 |
| 🛠️ CLI 管理 | 终端命令 `smtp-lite` 管理服务、配置、SSL、备份等 20+ 功能 |
| 🌐 Cloudflare 兼容 | SSL 证书使用 webroot 文件验证，兼容 CF 代理模式 |

---

## 快速安装

### 方式一：交互式安装脚本（推荐）

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/DoBestone/smtp-lite/master/install.sh)
```

脚本会引导完成所有配置：

```
→ 安装目录        ~/smtp-lite
→ 监听端口        8090
→ 管理员账号      用户名 + 密码（隐藏输入 + 二次确认，最少 8 位）
→ 域名绑定        可选，含格式校验
→ Nginx 反代      自动安装并生成配置
→ SSL 证书        Let's Encrypt 自动申请（webroot 文件验证，兼容 Cloudflare）+ 自动续期
→ 开机自启        systemd（Linux）/ launchd（macOS）自动识别
→ CLI 工具        安装后可通过 smtp-lite 命令管理所有功能
```

自动安装缺失依赖（Git / Go / Nginx / Certbot），支持：

- macOS Intel · Apple Silicon
- Linux x86\_64 · arm64（Ubuntu / Debian / CentOS / RHEL）

---

### 方式二：手动安装

```bash
git clone https://github.com/DoBestone/smtp-lite.git
cd smtp-lite

cp config.yaml.example config.yaml
vim config.yaml        # 修改账号密码

go mod tidy
go build -o smtp-lite ./cmd/server/
./smtp-lite

# 访问
open http://localhost:8090
```

---

### 方式三：Docker

```bash
cp config.yaml.example config.yaml
vim config.yaml

docker-compose up -d
```

---

### 方式四：环境变量

```bash
export SMTP_PORT=8090
export SMTP_USERNAME=admin
export SMTP_PASSWORD=your-strong-password
export SMTP_JWT_SECRET=random-32-chars-string-here
export SMTP_ENCRYPTION_KEY=32-byte-encryption-key-here!

./smtp-lite
```

---

## 配置说明

### config.yaml

```yaml
server:
  port: 8090
  mode: release      # debug / release

auth:
  username: admin
  password: change-me   # 强烈建议修改

jwt:
  secret: random-32-byte-string   # 建议随机生成
  expire_hours: 168               # Token 有效期（默认 7 天）

encryption:
  key: smtp-lite-encryption-key-32b!   # 必须恰好 32 字节
```

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `SMTP_PORT` | 监听端口（默认 8090） |
| `SMTP_USERNAME` | 管理员用户名 |
| `SMTP_PASSWORD` | 管理员密码 |
| `SMTP_JWT_SECRET` | JWT 签名密钥 |
| `SMTP_ENCRYPTION_KEY` | AES 加密密钥（32 字节） |

---

## API 文档

> 完整交互式文档内置在 Web 界面的「**API 文档**」标签页中，支持一键复制代码、真实 Base URL 展示。

### 认证

```bash
POST /api/v1/auth/login
Content-Type: application/json

{"username": "admin", "password": "your-password"}
```

```json
{"token": "eyJhbGci...", "username": "admin"}
```

后续请求携带：`Authorization: Bearer <token>` 或 `X-API-Key: sk_xxxxxxxx`

---

### 发送邮件

```bash
POST /api/v1/send
X-API-Key: sk_xxxxxxxxxx
Content-Type: application/json
```

```json
{
  "to":        "recipient@example.com",
  "subject":   "Hello",
  "body":      "<h1>Hi</h1>",
  "is_html":   true,
  "from_name": "My Service",
  "cc":        ["cc@example.com"],
  "bcc":       ["bcc@example.com"]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `to` | string | ✅ | 收件人邮箱 |
| `subject` | string | ✅ | 邮件主题 |
| `body` | string | ✅ | 邮件正文 |
| `is_html` | bool | — | `true` 时以 HTML 发送（默认 false） |
| `from_name` | string | — | 发件人显示名称 |
| `cc` | string[] | — | 抄送列表（收件人可见） |
| `bcc` | string[] | — | 密送列表（收件人不可见） |

```json
{"success": true, "message": "Email sent successfully", "used_smtp": "us***@gmail.com"}
```

---

### SMTP 账号管理（需要 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET`    | `/api/v1/smtp-accounts` | 账号列表 |
| `POST`   | `/api/v1/smtp-accounts` | 添加账号 |
| `PUT`    | `/api/v1/smtp-accounts/:id` | 更新账号 |
| `DELETE` | `/api/v1/smtp-accounts/:id` | 删除账号 |
| `POST`   | `/api/v1/smtp-accounts/:id/test` | 测试 SMTP 连通性 |
| `POST`   | `/api/v1/smtp-accounts/:id/test-send` | 发送测试邮件 |
| `POST`   | `/api/v1/smtp-accounts/:id/toggle` | 启用 / 禁用 |

---

### API Key 管理（需要 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET`    | `/api/v1/api-keys` | Key 列表 |
| `POST`   | `/api/v1/api-keys` | 创建 Key（完整 Key 仅返回一次）|
| `POST`   | `/api/v1/api-keys/:id/reset` | 重置 Key |
| `DELETE` | `/api/v1/api-keys/:id` | 删除 Key |

---

### 日志 · 统计 · 系统（需要 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET`  | `/api/v1/send/logs?page=1&page_size=50` | 分页发送日志 |
| `GET`  | `/api/v1/stats` | 统计（总量 / 成功率 / 今日） |
| `GET`  | `/api/v1/version` | 当前版本号（公开） |
| `POST` | `/api/v1/system/update` | 触发一键更新 |

---

## 版本更新

### 方式一：网页端一键更新

登录 → 进入「API 文档」页 → 点击「**检测更新**」→ 发现新版本时点击「**立即更新**」。

前端实时展示进度（git pull → go build → 重启），完成后自动刷新页面。

### 方式二：CLI 命令

```bash
smtp-lite update              # 检测到新版本才更新
smtp-lite update --force      # 强制重新编译并重启
```

### 方式三：命令行脚本

```bash
bash ~/smtp-lite/update.sh           # 检测到新版本才更新
bash ~/smtp-lite/update.sh --force   # 强制重新编译并重启
```

自动识别服务管理方式：systemd → launchd → 直接进程重启。

---

## Nginx 反向代理

**HTTP**

```nginx
server {
    listen 80;
    server_name smtp.example.com;

    location / {
        proxy_pass         http://127.0.0.1:8090;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

**HTTPS（Let's Encrypt · webroot 文件验证）**

```bash
# 方式一：安装脚本自动处理（推荐）
# install.sh 安装时选择 SSL 即可，使用 webroot 验证，兼容 Cloudflare CDN 代理

# 方式二：CLI 管理工具
smtp-lite ssl apply           # 交互式申请证书（webroot 文件验证）
smtp-lite ssl renew           # 手动续期
smtp-lite ssl status          # 查看证书状态

# 方式三：手动申请
sudo mkdir -p /var/www/certbot/.well-known/acme-challenge
sudo certbot certonly --webroot -w /var/www/certbot -d smtp.example.com
```

> **为什么使用 webroot 文件验证？** 传统的 `certbot --nginx` 依赖 HTTP-01 直连验证，在开启 Cloudflare CDN 代理时会失败。webroot 方式通过在服务器放置验证文件完成验证，完全兼容 CF 代理模式。

---

## 常用 SMTP 配置

| 邮箱 | SMTP 服务器 | 端口 | 认证说明 |
|------|------------|------|----------|
| Gmail | `smtp.gmail.com` | `587` | 需开启两步验证并创建[应用专用密码](https://myaccount.google.com/apppasswords) |
| QQ 邮箱 | `smtp.qq.com` | `587` | 在邮箱设置中开启 SMTP 并获取授权码 |
| 163 邮箱 | `smtp.163.com` | `465` | 在邮箱设置中开启 SMTP 并获取授权码 |
| Outlook | `smtp.office365.com` | `587` | 使用账号密码或应用密码 |
| 自定义 | 任意 SMTP 服务器 | 自定义 | 支持隐式 TLS（465）及 STARTTLS（587） |

---

## CLI 管理工具

安装脚本会自动将 `smtp-lite` 命令注册到系统 PATH，可在任意目录使用。

```bash
smtp-lite help       # 查看所有命令
```

### 服务管理

```bash
smtp-lite start             # 启动服务
smtp-lite stop              # 停止服务
smtp-lite restart           # 重启服务
smtp-lite status            # 查看详细状态（版本/端口/域名/SSL/日志大小等）
smtp-lite autostart enable  # 开启开机自启
smtp-lite autostart disable # 关闭开机自启
```

### 配置修改

```bash
smtp-lite port              # 交互式修改端口（自动同步 Nginx 配置）
smtp-lite domain            # 换绑域名（可同时申请 SSL）
smtp-lite password          # 修改管理员密码
smtp-lite username          # 修改管理员用户名
smtp-lite config show       # 查看配置（敏感信息脱敏）
smtp-lite config edit       # 用编辑器打开配置文件
smtp-lite config reset-jwt  # 重置 JWT Secret（使所有登录失效）
```

### SSL 证书

```bash
smtp-lite ssl status        # 查看证书状态及到期时间
smtp-lite ssl apply         # 申请/重新申请证书（webroot 文件验证）
smtp-lite ssl renew         # 手动续期
smtp-lite ssl remove        # 移除 SSL，降级为 HTTP
```

### 维护操作

```bash
smtp-lite update            # 在线更新（git pull + 编译 + 重启）
smtp-lite update --force    # 强制重新编译
smtp-lite build             # 仅重新编译
smtp-lite log tail          # 实时跟踪日志
smtp-lite log show 100      # 查看最近 100 行日志
smtp-lite log clear         # 清空日志
smtp-lite backup            # 备份配置 + 数据库 + Nginx 配置
smtp-lite restore           # 从备份恢复
smtp-lite reset             # 重置为初始状态
smtp-lite info              # 查看系统环境信息
smtp-lite uninstall         # 完整卸载
```

---

## 项目结构

```
smtp-lite/
├── cmd/server/main.go          # 程序入口 & 路由注册
├── internal/
│   ├── config/config.go        # 配置加载（YAML + 环境变量）
│   ├── handler/                # HTTP 处理器
│   │   ├── auth.go             # 登录、修改密码
│   │   ├── smtp.go             # SMTP 账号 CRUD
│   │   ├── api_key.go          # API Key 管理
│   │   ├── send.go             # 发送、日志、统计
│   │   └── system.go           # 版本查询、一键更新
│   ├── service/                # 业务逻辑
│   │   ├── auth.go             # JWT 签发与验证
│   │   ├── smtp.go             # SMTP 操作、AES 加解密、轮询
│   │   ├── api_key.go          # Key 生成与 SHA3 哈希
│   │   └── send.go             # 发送调度、故障切换、日志记录
│   ├── model/model.go          # 数据模型（GORM + SQLite）
│   ├── middleware/auth.go      # Token / API Key 认证中间件
│   └── version/version.go     # 版本常量
├── frontend/src/App.vue        # Vue 3 前端（单文件）
├── install.sh                  # 交互式安装脚本
├── cli.sh                      # CLI 管理工具（smtp-lite 命令）
├── update.sh                   # 一键更新脚本
├── config.yaml.example         # 配置文件模板
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## License

[MIT](LICENSE)
