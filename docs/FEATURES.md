# 功能特性详解

本文档详细介绍 go-gin-starter 的所有功能特性。

## 目录

- [认证与授权](#认证与授权)
- [数据库管理](#数据库管理)
- [API 文档](#api-文档)
- [分页支持](#分页支持)
- [消息队列](#消息队列)
- [可观测性](#可观测性)
- [多租户](#多租户)
- [文件存储与上传](#文件存储与上传)
- [通知](#通知)
- [审计日志](#审计日志)
- [安全特性](#安全特性)
- [开发工具](#开发工具)

## 认证与授权

### JWT 双 Token 机制

**Access Token**
- 有效期：15 分钟
- 用途：API 访问认证
- 存储：客户端内存（不持久化）

**Refresh Token**
- 有效期：7 天
- 用途：刷新 Access Token
- 存储：Redis（支持强制登出）；未配置时可退化为单实例内存存储

### Token Rotation

每次刷新 Token 时：
1. 验证旧 Refresh Token
2. 立即吊销旧 Token（从 Redis 删除）
3. 生成新的 Access Token 和 Refresh Token
4. 返回新 Token 给客户端

**优势**：防止 Token 被盗用后长期有效。

### 基于角色与资源的访问控制（RBAC + Authz）

支持的角色：
- `user`: 普通用户
- `admin`: 管理员

权限控制：
```go
// 需要认证
middleware.JWTAuth()

// 需要资源级权限
middleware.Authz("users", "write")
```

Casbin policy 存在数据库中，每次鉴权前会 reload policy，规则修改可立即生效。

### 验证码保护

- 注册和登录需要验证码
- 优先使用 Redis 存储（TTL 5 分钟）
- Redis 不可用时自动回退到内存存储
- 验证后自动删除（防止重复使用）

## 数据库管理

### GORM ORM

- 自动迁移支持
- 软删除
- 关联查询
- 事务支持

### golang-migrate 版本化迁移

```bash
# 迁移文件位置
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_tasks.up.sql
└── 000002_create_tasks.down.sql
```

### 连接池配置

可通过环境变量调整：

```bash
DB_MAX_OPEN_CONNS=25        # 最大连接数
DB_MAX_IDLE_CONNS=5         # 最大空闲连接数
DB_CONN_MAX_LIFETIME=300    # 连接最大生命周期（秒）
DB_CONN_MAX_IDLE_TIME=60    # 连接最大空闲时间（秒）
```

### Repository 模式

分层架构：
```
Handler → Service → Repository → Database
```

优势：
- 业务逻辑与数据访问分离
- 易于测试（可 Mock Repository）
- 易于切换数据库实现

## API 文档

### Swagger/OpenAPI

- 自动生成 API 文档
- 交互式测试界面
- 支持 Bearer Token 认证
- 非 release 模式自动启用

访问：`http://localhost:8080/swagger/index.html`

### 文档注解示例

```go
// GetProfile godoc
// @Summary 获取当前用户资料
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
    // ...
}
```

## 分页支持

### 统一分页参数

```go
type PageQuery struct {
    Page int `form:"page" binding:"omitempty,min=1"`
    Size int `form:"size" binding:"omitempty,min=1,max=100"`
}
```

### 使用示例

```bash
# 获取第 2 页，每页 20 条
GET /api/v1/tasks?page=2&size=20
```

响应：
```json
{
  "success": true,
  "data": {
    "total": 100,
    "page": 2,
    "size": 20,
    "items": [...]
  }
}
```

### 支持分页的 API

- `GET /api/v1/tasks` - 用户任务列表
- `GET /api/v1/admin/users` - 管理员用户列表
- `GET /api/v1/admin/tasks` - 管理员任务列表

## 消息队列

### Transactional Outbox 模式

**问题**：如何保证数据库操作和消息发送的一致性？

**解决方案**：
1. 业务操作和消息写入同一事务
2. 消息写入 `outbox_messages` 表
3. 后台 Worker 定时扫描并发送
4. 发送成功后标记为已发送

**优势**：
- 保证消息不丢失
- 支持消息重试
- 数据库和消息队列最终一致

### 使用示例

```go
// 在 Service 层
event := events.LoginEvent{
    EventID:    uuid.New().String(),
    UserID:     user.ID,
    Username:   user.Username,
    OccurredAt: time.Now(),
}

// 写入 outbox（与业务操作同事务）
err := s.pub.Save(ctx, tx, events.TopicUserLogin, event)
```

### 优雅关闭

服务关闭时：
1. 停止接收新请求
2. 等待 Outbox Worker 完成最后一次 flush
3. 关闭数据库连接

## 可观测性

### Prometheus 指标

访问：`http://localhost:8080/metrics`

内置指标：
- `http_requests_total` - 请求总数
- `http_request_duration_seconds` - 请求延迟
- `go_goroutines` - Goroutine 数量
- `go_memstats_alloc_bytes` - 内存使用

### OpenTelemetry

- Gin 请求自动创建 span
- GORM 查询通过官方 `gorm.io/plugin/opentelemetry/tracing` 插件埋点
- 默认 stdout exporter
- 配置 `OTEL_ENDPOINT` 后切换为 OTLP gRPC exporter

### 结构化日志

每个请求包含：
- `request_id` - 请求唯一标识
- `trace_id` - 分布式追踪 ID
- `method` - HTTP 方法
- `path` - 请求路径
- `status` - 响应状态码
- `latency` - 请求延迟
- `ip` - 客户端 IP

### 健康检查

**Liveness Probe**：`GET /healthz`
- 检查服务是否存活
- 用于 Kubernetes liveness probe

**Readiness Probe**：`GET /readyz`
- 检查服务是否就绪（数据库连接等）
- 用于 Kubernetes readiness probe

## 多租户

- JWT claims 支持 `tenant_id`
- `TenantResolver()` 支持从 claims、`X-Tenant-ID` 或 subdomain 解析租户
- Repository 层自动按 `tenant_id` 过滤
- 审计日志与上传文件路径都感知 tenant

## 文件存储与上传

### Storage 抽象

```go
type Storage interface {
    Upload(ctx context.Context, key string, reader io.Reader) (string, error)
    Delete(ctx context.Context, key string) error
}
```

### 支持实现

- `pkg/storage/local.go`：开发环境本地磁盘
- `pkg/storage/s3.go`：S3/OSS/MinIO 兼容实现

### 上传接口

- `POST /api/v1/upload`
- 使用 multipart form，字段名为 `file`
- 本地模式下文件会通过 `/uploads` 暴露

## 通知

### Notifier 抽象

- `pkg/notify/noop.go`：本地/测试环境
- `pkg/notify/smtp.go`：生产环境 SMTP

### 已接入场景

- 注册成功后发送欢迎邮件
- 发起密码重置后发送重置邮件

## 审计日志

- 关键操作异步写入 `audit_logs`
- 当前已覆盖：注册、登录、任务创建/更新/删除、管理员用户写操作、管理员删任务、文件上传
- 管理员可通过 `GET /api/v1/admin/audit-logs` 查询

## 安全特性

### 请求签名校验

启用后，所有 API 请求需要签名：

```bash
ENABLE_SIGNATURE=true
SIGNATURE_SECRET=your-secret
```

签名算法：
```
signature = HMAC-SHA256(request_body + timestamp, secret)
```

### IP 访问控制

**白名单模式**：
```bash
ENABLE_IP_WHITELIST=true
IP_WHITELIST=10.0.1.100,10.0.1.101
```

**黑名单模式**：
```bash
ENABLE_IP_BLACKLIST=true
IP_BLACKLIST=192.168.0.0/16
```

### 限流

基于令牌桶算法：
```go
// 认证路由限流：5 req/s
middleware.RateLimit(5)
```

已补充用户级限流：

```go
middleware.RateLimitByUser(20)
```

### CORS 配置

```bash
CORS_ORIGIN=https://your-frontend.com
```

支持多个域名（逗号分隔）。

### 安全头

自动添加：
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`

## 开发工具

### Air 热重载

代码变更自动重启：
```bash
air
```

配置文件：`.air.toml`

### testcontainers 集成测试

自动启动 PostgreSQL 容器：
```go
db, cleanup := testhelper.SetupPostgres(t)
defer cleanup()
```

### Makefile 命令

```bash
make help              # 显示所有命令
make dev               # 热重载开发
make test              # 运行测试
make lint              # 代码检查
make swagger           # 生成文档
make docker-build      # 构建镜像
```

### 代码检查

使用 golangci-lint：
```bash
make lint
```

检查项：
- errcheck - 未处理的错误
- gosimple - 代码简化
- govet - Go 官方检查
- gosec - 安全检查
- 等等...

### CI/CD

GitHub Actions 自动化：
- 运行测试
- 代码检查
- 构建 Docker 镜像
- 生成覆盖率报告

## 性能优化

### 数据库

- 连接池配置
- 索引优化
- 慢查询日志
- 预编译语句

### 缓存

- Redis 缓存
- 验证码缓存
- Refresh Token 缓存

### HTTP

- Keep-Alive 连接
- GZIP 压缩（Nginx）
- 静态资源缓存

### 并发

- Goroutine 池
- 异步消息发布
- 非阻塞 I/O

## 扩展性

### 水平扩容

- 无状态设计
- Session 存储在 Redis
- 支持多实例部署

### 垂直扩容

- 调整连接池大小
- 增加内存和 CPU
- 优化数据库查询

### 微服务化

- 清晰的分层架构
- Repository 模式
- 事件驱动（Outbox）
- 易于拆分为微服务

## 最佳实践

### 错误处理

统一错误类型：
```go
var (
    ErrUserNotFound = apperr.New(404, "用户不存在")
    ErrInvalidPassword = apperr.New(401, "密码错误")
)
```

### 输入验证

使用 Gin 的 binding 标签：
```go
type CreateTaskRequest struct {
    Title string `json:"title" binding:"required,min=1,max=200"`
}
```

### 事务管理

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 业务操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

### 测试

- 单元测试：测试业务逻辑
- 集成测试：测试数据库操作
- E2E 测试：测试完整流程

## 下一步

- 📖 查看 [快速开始](./QUICKSTART.md)
- 🚀 阅读 [部署指南](./DEPLOYMENT.md)
- 🧱 替换示例业务模块为你自己的领域模型与服务
