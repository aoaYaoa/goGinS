# go-gin-starter

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/your-username/go-gin-starter/workflows/CI/badge.svg)](https://github.com/your-username/go-gin-starter/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/go-gin-starter)](https://goreportcard.com/report/github.com/your-username/go-gin-starter)
[![codecov](https://codecov.io/gh/your-username/go-gin-starter/branch/main/graph/badge.svg)](https://codecov.io/gh/your-username/go-gin-starter)

生产就绪的 Go + Gin 后端脚手架，包含认证、任务管理、细粒度权限、多租户、可观测性、文件存储、通知与自动部署。

> 🚀 开箱即用的 Go Web 应用启动模板，集成最佳实践和完整的 DevOps 工具链。
>
> **模板说明**：本仓库分为两层。**基础层**（配置、数据库、中间件、认证、健康检查、日志、指标、发布者、启动生命周期）是可复用的脚手架。**示例层**（`auth`、`user`、`tasks`、`admin` 模块）是可运行的参考实现，展示如何在基础层之上构建业务逻辑，替换为自己的业务模块是正常的使用方式。

## ✨ 特点

- 🏗️ **生产就绪**：完整的错误处理、日志、监控、健康检查
- 🔐 **安全第一**：JWT 认证、Token Rotation、Casbin 权限、请求签名、IP 控制
- 📦 **开箱即用**：一键初始化脚本，5 分钟快速启动
- 🧪 **测试完备**：testcontainers 集成测试，CI/CD 自动化
- 📚 **文档齐全**：Swagger API 文档，部署指南，快速开始
- 🔥 **开发友好**：Air 热重载，完整的 Makefile，代码检查
- 🐳 **容器化**：Docker 多阶段构建，docker-compose 一键部署
- 📊 **可观测性**：Prometheus 指标，结构化日志，OpenTelemetry 链路追踪

## 技术栈

- **运行时**：Go 1.26 + Gin
- **数据库**：PostgreSQL（GORM + golang-migrate 版本化迁移）
- **缓存**：Redis（验证码存储 + Refresh Token，可选；未配置时退化为内存存储）
- **消息队列**：Kafka（登录事件，Transactional Outbox 模式，可选）
- **认证**：JWT（15min Access Token + 7d Refresh Token，Token Rotation）
- **授权**：Casbin RBAC（资源级权限）
- **多租户**：应用层行级隔离（tenant_id）
- **链路追踪**：OpenTelemetry + OTLP gRPC + Gin/GORM instrumentation
- **存储**：本地磁盘 / S3 兼容对象存储
- **通知**：SMTP / Noop Notifier
- **部署**：Docker 多阶段构建 + Render
- **测试**：testcontainers-go 集成测试
- **文档**：Swagger/OpenAPI 自动生成
- **开发**：Air 热重载

## 核心特性

✅ **认证与授权**
- JWT 双 Token 机制（Access + Refresh）
- Token Rotation 防止重放攻击
- Casbin 资源级权限控制（Authz）
- 验证码防护（支持 Redis 或内存存储）
- 密码重置与欢迎邮件通知

✅ **数据库**
- GORM ORM 框架
- golang-migrate 版本化迁移
- 连接池配置外置（支持环境变量）
- Repository 模式分层架构
- 多租户行级隔离（tenant_id）

✅ **分页支持**
- 统一分页查询参数（page, size）
- 任务列表分页
- 管理员用户/任务列表分页

✅ **消息队列**
- Transactional Outbox 模式保证消息可靠性
- 异步 Kafka 发布器
- 优雅关闭等待 Worker 完成

✅ **可观测性**
- Prometheus 指标暴露
- 结构化日志（RequestID + TraceID）
- 健康检查端点（liveness + readiness）
- OpenTelemetry trace 导出（stdout / OTLP gRPC）

✅ **安全**
- 请求签名校验（可选）
- IP 黑白名单
- CORS 跨域配置
- 限流中间件（令牌桶算法）
- 审计日志（登录/注册/管理操作/任务操作/上传）

✅ **扩展能力**
- 本地文件存储与 S3 兼容对象存储
- 管理员审计日志查询接口
- 多租户感知的 JWT / Repository / Audit 流程

✅ **开发体验**
- Swagger UI 文档（非 release 模式）
- Air 热重载
- testcontainers 集成测试
- 输入验证错误友好提示

## 项目结构

```
.
├── cmd/server/          # 入口，依赖注入与启动
├── internal/
│   ├── config/          # 环境变量配置
│   ├── container/       # 依赖注入容器
│   ├── database/        # GORM 连接 + 迁移
│   ├── dto/             # 请求/响应 DTO
│   ├── events/          # 领域事件定义
│   ├── handlers/        # HTTP 处理器
│   ├── middleware/      # 中间件管道
│   ├── models/          # GORM 模型
│   ├── publisher/       # Kafka AsyncPublisher + Outbox
│   ├── repositories/    # 数据访问层
│   ├── routes/          # 路由注册
│   └── services/        # 业务逻辑
├── migrations/          # SQL 迁移文件
├── configs/             # Casbin 等配置模型
├── pkg/
│   ├── apperr/          # 统一错误类型
│   ├── notify/          # SMTP / noop 通知器
│   ├── storage/         # local / s3 文件存储
│   └── utils/           # jwt / captcha / logger / response / authz / tracing / tenantctx
├── Dockerfile
├── docker-compose.yml
└── render.yaml
```

## 快速开始

### ⚡ 一键启动（推荐）

```bash
# 克隆项目
git clone https://github.com/your-username/go-gin-starter.git
cd go-gin-starter

# 运行初始化脚本
./scripts/init.sh

# 启动服务
make dev  # 或 air（热重载）
```

访问：
- 🌐 API: http://localhost:8080
- 📖 Swagger 文档: http://localhost:8080/swagger/index.html
- 📊 Metrics: http://localhost:8080/metrics

详细步骤请查看 [快速开始指南](./docs/QUICKSTART.md)。

### 本地开发（docker-compose）

```bash
cp .env.example .env
# 编辑 .env，至少设置 JWT_SECRET（32 字节以上）

docker compose up -d
```

服务启动后访问 `http://localhost:8080`。

### 手动运行

```bash
# 依赖：PostgreSQL（Redis 可选）
cp .env.example .env

go run ./cmd/server
```

### 热重载开发（推荐）

使用 [Air](https://github.com/air-verse/air) 实现代码变更自动重启：

```bash
# 安装 Air
go install github.com/air-verse/air@latest

# 启动热重载
air
```

### API 文档

开发模式下访问 Swagger UI：`http://localhost:8080/swagger/index.html`

生产模式（`SERVER_MODE=release`）下 Swagger 自动禁用。

## API 路由

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/healthz` | 存活探针 | 无 |
| GET | `/readyz` | 就绪探针（含 DB ping） | 无 |
| GET | `/metrics` | Prometheus 指标 | IP 白名单 |
| GET | `/api/v1/auth/captcha` | 获取验证码 | 无 |
| POST | `/api/v1/auth/register` | 注册 | 无 |
| POST | `/api/v1/auth/login` | 登录 | 无 |
| POST | `/api/v1/auth/refresh` | 刷新 Token | 无 |
| POST | `/api/v1/auth/logout` | 登出 | 无 |
| POST | `/api/v1/auth/forgot-password` | 发起密码重置 | 无 |
| POST | `/api/v1/auth/reset-password` | 重置密码 | 无 |
| GET | `/api/v1/user/profile` | 获取个人资料 | JWT |
| PUT | `/api/v1/user/profile` | 更新个人资料 | JWT |
| POST | `/api/v1/tasks` | 创建任务 | JWT |
| GET | `/api/v1/tasks` | 任务列表 | JWT |
| GET | `/api/v1/tasks/:id` | 任务详情 | JWT |
| PUT | `/api/v1/tasks/:id` | 更新任务 | JWT |
| DELETE | `/api/v1/tasks/:id` | 删除任务 | JWT |
| GET | `/api/v1/admin/users` | 用户列表 | JWT + Authz |
| GET | `/api/v1/admin/users/:id` | 用户详情 | JWT + Authz |
| PUT | `/api/v1/admin/users/:id/ban` | 封禁用户 | JWT + Authz |
| PUT | `/api/v1/admin/users/:id/unban` | 解封用户 | JWT + Authz |
| PUT | `/api/v1/admin/users/:id/promote` | 提升为 admin | JWT + Authz |
| PUT | `/api/v1/admin/users/:id/demote` | 降级为 user | JWT + Authz |
| GET | `/api/v1/admin/tasks` | 全部任务 | JWT + Authz |
| DELETE | `/api/v1/admin/tasks/:id` | 强制删除任务 | JWT + Authz |
| POST | `/api/v1/upload` | 上传文件 | JWT |
| GET | `/api/v1/admin/audit-logs` | 查询审计日志 | JWT + Authz |

认证路由限流：5 req/s（令牌桶，IP 级）。

## 中间件管道

```
Recovery → RequestID → OTel Trace → TraceContext → Logger → CORS
→ IPBlacklist → IPWhitelist → Metrics
→ [SignatureVerify]  // ENABLE_SIGNATURE=true 时启用
```

## 环境变量

完整配置见 [`.env.example`](.env.example)。关键变量：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | 监听端口 | `8080` |
| `JWT_SECRET` | JWT 签名密钥（≥32 字节） | — |
| `DATABASE_HOST` | PostgreSQL 主机 | `localhost` |
| `DB_MAX_OPEN_CONNS` | 数据库最大连接数 | `25` |
| `DB_MAX_IDLE_CONNS` | 数据库最大空闲连接数 | `5` |
| `DB_CONN_MAX_LIFETIME` | 连接最大生命周期（秒） | `300` |
| `DB_CONN_MAX_IDLE_TIME` | 连接最大空闲时间（秒） | `60` |
| `REDIS_ADDR` | Redis 地址（留空使用内存存储） | — |
| `KAFKA_BROKERS` | Kafka broker 列表（留空跳过） | — |
| `OTEL_ENDPOINT` | OTLP gRPC endpoint，留空使用 stdout exporter | — |
| `OTEL_SERVICE_NAME` | OTel 服务名 | `go-gin-starter` |
| `STORAGE_TYPE` | 文件存储类型（`local` / `s3`） | `local` |
| `SMTP_HOST` | SMTP 主机，留空时使用 noop notifier | — |
| `MIGRATIONS_PATH` | 迁移文件目录（留空跳过） | `migrations` |
| `ENABLE_SIGNATURE` | 启用请求签名校验 | `false` |
| `METRICS_ALLOWED_IPS` | `/metrics` 允许的 IP（逗号分隔，留空不限制） | — |

## 部署

### Docker

```bash
docker build -t go-gin-starter .
docker run -p 8080:8080 --env-file .env go-gin-starter
```

### Render

项目根目录已包含 `render.yaml`，推送到 GitHub 后在 Render 控制台导入仓库即可自动部署。默认配置为本地文件存储；生产场景建议切到 S3 兼容存储并配置 `OTEL_ENDPOINT`、`SMTP_*`。

## 测试

### 运行集成测试

项目使用 [testcontainers-go](https://golang.testcontainers.org/) 进行集成测试，自动启动 PostgreSQL 容器：

```bash
# 运行所有测试
go test ./...

# 运行 repository 层测试
go test -v ./internal/repositories/...

# 运行测试并显示覆盖率
go test -cover ./...
```

### 生成 Swagger 文档

```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/server/main.go -o docs
```

## 📚 文档

- [快速开始指南](./docs/QUICKSTART.md) - 5 分钟快速上手
- [部署指南](./docs/DEPLOYMENT.md) - 生产环境部署最佳实践
- [API 文档](http://localhost:8080/swagger/index.html) - Swagger UI（开发模式）
- [功能特性详解](./docs/FEATURES.md) - 企业级能力与实现说明

## 🛠️ 常用命令

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

# 工具
make install-tools    # 安装开发工具
make clean            # 清理构建产物
```

完整命令列表：`make help`

## 架构说明

### JWT 认证与多租户

- Access Token 有效期 15 分钟，Refresh Token 有效期 7 天
- 每次刷新执行 Token Rotation（旧 Refresh Token 立即吊销）
- Refresh Token 存储在 Redis，支持强制登出
- JWT Claims 可携带 `tenant_id`，Repository 层按租户自动隔离

### Transactional Outbox

登录事件通过 Outbox 模式发布：写入 `outbox_messages` 表与业务操作同事务，后台 Worker 每 5 秒扫描并通过 Kafka AsyncPublisher 发送，保证消息不丢失。

### 授权与审计

- 管理员接口使用 Casbin 资源级权限校验
- 每次鉴权前 reload policy，规则修改后可立即生效
- 关键操作会异步写入 `audit_logs` 表并支持管理员查询

### 验证码、存储与通知

优先使用 Redis 存储（TTL 5 分钟），Redis 不可用时自动回退到内存存储。
- 文件上传支持本地磁盘与 S3 兼容对象存储
- 通知支持 `noop` 和 SMTP，两条示例业务已接入：欢迎邮件、密码重置

## 📄 许可证

本项目采用 [MIT](./LICENSE) 许可证。

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [testcontainers-go](https://golang.testcontainers.org/) - 集成测试
- [swag](https://github.com/swaggo/swag) - Swagger 文档生成

## 📞 支持

- 📧 Email: support@example.com
- 💬 [Discussions](https://github.com/your-username/go-gin-starter/discussions)
- 🐛 [Issues](https://github.com/your-username/go-gin-starter/issues)

---

⭐ 如果这个项目对你有帮助，请给个 Star！
