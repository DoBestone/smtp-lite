# SMTP Lite - 个人邮箱聚合系统

轻量级 SMTP 聚合系统，适合个人使用。

## 特性

- 🔐 单用户登录（配置文件定义账号密码）
- 📧 SMTP 账号管理（添加、删除、启用/禁用）
- 🔑 API Key 管理
- 📤 统一 API 发件接口
- 📊 发送日志与统计
- 🔄 账号轮询、故障转移
- 📦 SQLite 数据库（无需额外安装）
- 🐳 Docker 支持

## 快速开始

### 方式一：直接运行

```bash
# 1. 修改配置
cp config.yaml.example config.yaml
vim config.yaml  # 修改用户名密码

# 2. 编译运行
go mod tidy
go build -o smtp-lite ./cmd/server/main.go
./smtp-lite

# 3. 访问
open http://localhost:8090
```

### 方式二：Docker

```bash
# 1. 修改配置
cp config.yaml.example config.yaml
vim config.yaml

# 2. 构建运行
docker-compose up -d

# 3. 访问
open http://localhost:8090
```

### 方式三：环境变量

```bash
export SMTP_USERNAME=admin
export SMTP_PASSWORD=your-password
export SMTP_JWT_SECRET=random-32-chars
export SMTP_ENCRYPTION_KEY=32-byte-encryption-key!

./smtp-lite
```

## 配置说明

### config.yaml

```yaml
server:
  port: 8090
  mode: release  # debug/release

auth:
  username: admin
  password: your-strong-password

jwt:
  secret: random-32-byte-string
  expire_hours: 168

encryption:
  key: smtp-lite-encryption-key-32b!  # 必须32字节
```

### 环境变量

| 变量 | 说明 |
|------|------|
| SMTP_PORT | 服务端口 |
| SMTP_USERNAME | 登录用户名 |
| SMTP_PASSWORD | 登录密码 |
| SMTP_JWT_SECRET | JWT 密钥 |
| SMTP_ENCRYPTION_KEY | 加密密钥(32字节) |

## API 文档

### 认证

```bash
# 登录
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "your-password"
}

# 响应
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "username": "admin"
}
```

### SMTP 账号

```bash
# 列表
GET /api/v1/smtp-accounts
Authorization: Bearer <token>

# 添加
POST /api/v1/smtp-accounts
Authorization: Bearer <token>
{
  "email": "user@gmail.com",
  "password": "app_password",
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "daily_limit": 500
}

# 测试连接
POST /api/v1/smtp-accounts/:id/test

# 启用/禁用
POST /api/v1/smtp-accounts/:id/toggle

# 删除
DELETE /api/v1/smtp-accounts/:id
```

### API Key

```bash
# 列表
GET /api/v1/api-keys

# 创建
POST /api/v1/api-keys
{
  "name": "My App"
}

# 响应（Key 只显示一次）
{
  "id": "...",
  "key": "sk_xxxxx...",
  "key_prefix": "sk_xxxxx",
  "warning": "Save this key! It won't be shown again."
}

# 删除
DELETE /api/v1/api-keys/:id
```

### 发送邮件

```bash
# 方式1: API Key 认证
POST /api/v1/send
X-API-Key: sk_xxxxx...

# 方式2: Token 认证
POST /api/v1/send
Authorization: Bearer <token>

# 请求体
{
  "to": "recipient@example.com",
  "subject": "Hello",
  "body": "Email content",
  "from_name": "Sender",  // 可选
  "is_html": false        // 可选
}

# 响应
{
  "success": true,
  "message": "Email sent successfully",
  "used_smtp": "us***@gmail.com"
}
```

## 项目结构

```
smtp-lite/
├── cmd/server/main.go      # 入口
├── internal/
│   ├── config/config.go    # 配置管理
│   ├── handler/            # HTTP handlers
│   ├── service/            # 业务逻辑
│   ├── model/model.go      # 数据模型
│   └── middleware/auth.go  # 认证中间件
├── frontend/               # Vue 3 前端
├── config.yaml             # 配置文件
├── docker-compose.yml      # Docker 编排
├── Dockerfile
└── README.md
```

## 常见 SMTP 配置

| 邮箱 | SMTP 服务器 | 端口 | 备注 |
|------|-----------|------|------|
| Gmail | smtp.gmail.com | 587 | 需应用专用密码 |
| QQ邮箱 | smtp.qq.com | 587 | 需授权码 |
| 163邮箱 | smtp.163.com | 465 | 需授权码 |
| Outlook | smtp.office365.com | 587 | 需应用密码 |

## Nginx 配置

参考 `nginx.conf.example` 文件配置反向代理和 SSL。

## License

MIT