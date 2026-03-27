# 快速开始指南

5 分钟快速启动 go-gin-starter 项目。

## 前置要求

- Go 1.26+
- Docker & Docker Compose（推荐）
- 或 PostgreSQL 14+（必需）与 Redis 6+（可选）

## 方式一：一键启动（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/your-username/go-gin-starter.git
cd go-gin-starter

# 2. 运行初始化脚本
./scripts/init.sh

# 3. 启动服务（选择一种）
make dev              # 热重载开发模式
# 或
make run              # 普通模式
# 或
docker compose up -d  # Docker 模式
```

访问：
- API: http://localhost:8080
- Swagger 文档: http://localhost:8080/swagger/index.html
- 本地文件上传目录: http://localhost:8080/uploads/...

## 方式二：手动启动

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

```bash
cp .env.example .env

# 编辑 .env，至少修改：
# - JWT_SECRET（生成随机密钥）
# - DATABASE_PASS（数据库密码）
# - STORAGE_TYPE（默认 local；生产建议 s3）
```

生成随机 JWT_SECRET：
```bash
openssl rand -base64 32
```

### 3. 启动数据库

```bash
# 使用 Docker（Redis 可选，不启动时应用自动退化为内存存储）
docker compose up -d postgres

# 或同时启动 Redis（生产推荐）
docker compose up -d postgres redis
```

### 4. 运行服务

```bash
go run ./cmd/server
```

### 5. 可选能力

```bash
# OpenTelemetry（留空则用 stdout exporter）
export OTEL_ENDPOINT=localhost:4317

# 本地上传文件目录
export LOCAL_STORAGE_DIR=uploads

# SMTP（不配置则使用 noop notifier）
export SMTP_HOST=smtp.example.com
```

## 测试 API

### 1. 获取验证码

```bash
curl http://localhost:8080/api/v1/auth/captcha
```

响应：
```json
{
  "success": true,
  "code": 200,
  "data": {
    "captcha_id": "xxx",
    "captcha_image": "data:image/png;base64,..."
  }
}
```

### 2. 注册用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "captcha_id": "xxx",
    "captcha_code": "123456"
  }'
```

### 3. 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "captcha_id": "xxx",
    "captcha_code": "123456"
  }'
```

响应：
```json
{
  "success": true,
  "code": 200,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "status": "active"
    }
  }
}
```

### 4. 访问受保护的 API

```bash
# 使用 access_token
curl http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. 创建任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的第一个任务",
    "description": "任务描述"
  }'
```

### 6. 获取任务列表（分页）

```bash
curl "http://localhost:8080/api/v1/tasks?page=1&size=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 7. 上传文件

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@./README.md"
```

### 8. 发起密码重置

```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

## 开发工具

### 热重载

```bash
# 安装 Air
go install github.com/air-verse/air@latest

# 启动热重载
air
```

### 生成 Swagger 文档

```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/server/main.go -o docs
```

### 运行测试

```bash
# 所有测试
make test

# 集成测试
make test-integration

# 覆盖率
make test-coverage
```

## 常用命令

```bash
# 开发
make dev              # 热重载开发
make run              # 普通运行
make build            # 编译

# 测试
make test             # 运行测试
make test-integration # 集成测试
make lint             # 代码检查

# 文档
make swagger          # 生成 Swagger

# Docker
make docker-build     # 构建镜像
make docker-run       # 运行容器

# 清理
make clean            # 清理构建产物
```

## 项目结构

```
.
├── cmd/server/          # 入口文件
├── internal/            # 内部代码
│   ├── config/          # 配置
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑
│   ├── repositories/    # 数据访问
│   ├── models/          # 数据模型
│   ├── dto/             # 数据传输对象
│   └── middleware/      # 中间件
├── pkg/                 # 公共包
├── migrations/          # 数据库迁移
├── docs/                # Swagger 文档
└── scripts/             # 脚本工具
```

## 下一步

- 📖 阅读 [API 文档](http://localhost:8080/swagger/index.html)
- 🚀 查看 [部署指南](./DEPLOYMENT.md)
- 🔧 了解 [配置选项](../README.md#环境变量)
- 🧪 编写测试用例
- 🎨 自定义业务逻辑

## 常见问题

### 端口被占用

```bash
# 修改 .env 中的 SERVER_PORT
SERVER_PORT=8081
```

### 数据库连接失败

```bash
# 检查数据库是否启动
docker compose ps

# 查看日志
docker compose logs postgres
```

### Swagger 无法访问

确保 `SERVER_MODE=debug`（开发模式），release 模式会禁用 Swagger。

### 验证码无法显示

如果没有 Redis，会自动使用内存存储。检查 `REDIS_ADDR` 配置。

### 上传文件无法访问

检查 `STORAGE_TYPE`、`LOCAL_STORAGE_DIR`、`STORAGE_PUBLIC_BASE_URL` 配置；本地模式下文件由应用直接通过 `/uploads` 暴露。

## 获取帮助

- 📚 [完整文档](../README.md)
- 🐛 [提交 Issue](https://github.com/your-repo/issues)
- 💬 [讨论区](https://github.com/your-repo/discussions)
