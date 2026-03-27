# 部署指南

本文档提供生产环境部署的最佳实践和步骤。

## 目录

- [环境准备](#环境准备)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [Render 部署](#render-部署)
- [性能优化](#性能优化)
- [监控告警](#监控告警)
- [文件存储与通知](#文件存储与通知)
- [故障排查](#故障排查)

## 环境准备

### 最低要求

- **CPU**: 2 核
- **内存**: 2GB
- **磁盘**: 20GB
- **Go**: 1.26+
- **PostgreSQL**: 14+
- **Redis**: 6+（可选；未配置 `REDIS_ADDR` 时，Refresh Token 和验证码自动使用内存存储，不适合多实例部署）
- **对象存储**: S3/OSS/MinIO（推荐生产启用）
- **OTLP Collector**: 推荐生产启用

### 推荐配置

- **CPU**: 4 核
- **内存**: 4GB
- **磁盘**: 50GB SSD
- **PostgreSQL**: 16+
- **Redis**: 7+

## Docker 部署

### 1. 准备配置文件

```bash
# 复制生产环境配置
cp .env.production.example .env.production

# 编辑配置
vim .env.production

# 生成强随机密钥
openssl rand -base64 64
```

### 2. 构建镜像

```bash
# 构建
docker build -t go-gin-starter:latest .

# 验证镜像大小（应该 < 30MB）
docker images go-gin-starter
```

### 3. 运行容器

```bash
# 使用 docker-compose
docker compose -f docker-compose.prod.yml up -d

# 或单独运行
docker run -d \
  --name go-gin-starter \
  --env-file .env.production \
  -p 8080:8080 \
  --restart unless-stopped \
  go-gin-starter:latest
```

### 4. 健康检查

```bash
# 检查服务状态
./scripts/health-check.sh localhost:8080

# 查看日志
docker logs -f go-gin-starter
```

## Kubernetes 部署

### 1. 创建 ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  SERVER_MODE: "release"
  DATABASE_HOST: "postgres-service"
  REDIS_ADDR: "redis-service:6379"
```

### 2. 创建 Secret

```bash
kubectl create secret generic app-secrets \
  --from-literal=JWT_SECRET=$(openssl rand -base64 64) \
  --from-literal=DATABASE_PASS=your-db-password \
  --from-literal=REDIS_PASSWORD=your-redis-password
```

### 3. 部署应用

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-gin-starter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-gin-starter
  template:
    metadata:
      labels:
        app: go-gin-starter
    spec:
      containers:
      - name: app
        image: go-gin-starter:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: app-secrets
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 4. 创建 Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-gin-starter
spec:
  selector:
    app: go-gin-starter
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Render 部署

### 1. 连接 GitHub 仓库

1. 登录 [Render](https://render.com)
2. 点击 "New +" → "Blueprint"
3. 连接 GitHub 仓库

### 2. 配置环境变量

在 Render Dashboard 中设置：

- `JWT_SECRET`: 使用 Render 的 "Generate" 功能
- `DATABASE_HOST` / `DATABASE_NAME` / `DATABASE_USER` / `DATABASE_PASS`
- `REDIS_ADDR` / `REDIS_PASSWORD`（可选）
- `OTEL_ENDPOINT`（可选，建议）
- `STORAGE_TYPE`, `LOCAL_STORAGE_DIR`, `STORAGE_PUBLIC_BASE_URL`，或完整 `S3_*`
- `SMTP_*`（可选）
- `SERVER_MODE=release`

### 3. 自动部署

推送到 `main` 分支会自动触发部署。

## 性能优化

### 数据库连接池

根据负载调整：

```bash
# 低负载（< 100 QPS）
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# 中负载（100-1000 QPS）
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

# 高负载（> 1000 QPS）
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=20
```

### Redis 优化

```bash
# 启用 TLS
REDIS_TLS=true

# 使用连接池
REDIS_POOL_SIZE=10
```

### 应用层优化

1. **启用 release 模式**
   ```bash
   SERVER_MODE=release
   ```

2. **配置 CORS**
   ```bash
   CORS_ORIGIN=https://your-domain.com
   ```

3. **启用请求签名**
   ```bash
   ENABLE_SIGNATURE=true
   SIGNATURE_SECRET=your-secret
   ```

4. **启用 OpenTelemetry**
   ```bash
   OTEL_ENDPOINT=otel-collector:4317
   OTEL_SERVICE_NAME=go-gin-starter
   ```

5. **启用对象存储**
   ```bash
   STORAGE_TYPE=s3
   S3_BUCKET=your-bucket
   S3_REGION=ap-southeast-1
   ```

## 监控告警

### Prometheus 指标

访问 `/metrics` 端点获取指标：

```bash
# 限制访问 IP
METRICS_ALLOWED_IPS=10.0.1.100,10.0.1.101
```

### 关键指标

- `http_requests_total`: 请求总数
- `http_request_duration_seconds`: 请求延迟
- `go_goroutines`: Goroutine 数量
- `go_memstats_alloc_bytes`: 内存使用

### Grafana Dashboard

导入预配置的 Dashboard（待创建）。

### 日志聚合

推荐使用：
- ELK Stack
- Loki + Grafana
- CloudWatch Logs

## 文件存储与通知

### 文件存储

开发环境默认：

```bash
STORAGE_TYPE=local
LOCAL_STORAGE_DIR=uploads
STORAGE_PUBLIC_BASE_URL=/uploads
```

生产环境建议：

```bash
STORAGE_TYPE=s3
S3_BUCKET=your-bucket
S3_REGION=ap-southeast-1
S3_ENDPOINT=https://s3.your-provider.com
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
```

### 通知

未配置 SMTP 时，系统会使用 noop notifier，不会真正发信。

```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=no-reply@example.com
SMTP_PASS=...
SMTP_FROM=no-reply@example.com
```

## 故障排查

### 服务无法启动

```bash
# 检查日志
docker logs go-gin-starter

# 常见问题：
# 1. 数据库连接失败 → 检查 DATABASE_HOST 和凭证
# 2. Redis 连接失败 → 检查 REDIS_ADDR
# 3. OTLP 连接失败 → 检查 OTEL_ENDPOINT
# 4. S3 上传失败 → 检查 S3_* / STORAGE_TYPE
# 5. 端口被占用 → 修改 SERVER_PORT
```

### 数据库连接池耗尽

```bash
# 增加连接数
DB_MAX_OPEN_CONNS=100

# 减少连接生命周期
DB_CONN_MAX_LIFETIME=300
```

### 内存泄漏

```bash
# 启用 pprof（仅开发环境）
go tool pprof http://localhost:8080/debug/pprof/heap
```

### 性能问题

```bash
# 查看慢查询
# PostgreSQL: 启用 log_min_duration_statement

# 查看 Goroutine 泄漏
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## 安全检查清单

- [ ] 修改默认 JWT_SECRET
- [ ] 启用 DATABASE_SSL_MODE=require
- [ ] 启用 REDIS_TLS=true
- [ ] 配置 METRICS_ALLOWED_IPS
- [ ] 启用 ENABLE_SIGNATURE（API 调用）
- [ ] 配置 IP_BLACKLIST（阻止内网）
- [ ] 配置 OTEL_ENDPOINT
- [ ] 生产环境切换到 STORAGE_TYPE=s3
- [ ] 配置 SMTP_* 启用欢迎邮件与密码重置
- [ ] 定期更新依赖（go get -u）
- [ ] 定期备份数据库
- [ ] 配置防火墙规则
- [ ] 启用 HTTPS（使用 Nginx/Caddy）

## 回滚策略

### Docker

```bash
# 标记版本
docker tag go-gin-starter:latest go-gin-starter:v1.0.0

# 回滚
docker stop go-gin-starter
docker run -d --name go-gin-starter go-gin-starter:v1.0.0
```

### Kubernetes

```bash
# 回滚到上一个版本
kubectl rollout undo deployment/go-gin-starter

# 回滚到指定版本
kubectl rollout undo deployment/go-gin-starter --to-revision=2
```

## 扩容指南

### 水平扩容

```bash
# Docker Swarm
docker service scale go-gin-starter=5

# Kubernetes
kubectl scale deployment go-gin-starter --replicas=5
```

### 垂直扩容

修改资源限制：

```yaml
resources:
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

## 备份恢复

### 数据库备份

```bash
# 备份
pg_dump -h $DATABASE_HOST -U $DATABASE_USER $DATABASE_NAME > backup.sql

# 恢复
psql -h $DATABASE_HOST -U $DATABASE_USER $DATABASE_NAME < backup.sql
```

### Redis 备份

```bash
# 备份
redis-cli --rdb /path/to/backup.rdb

# 恢复
cp backup.rdb /var/lib/redis/dump.rdb
systemctl restart redis
```

## 支持

如有问题，请：
1. 查看 [FAQ](./FAQ.md)
2. 提交 [Issue](https://github.com/your-repo/issues)
3. 联系技术支持
