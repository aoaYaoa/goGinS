package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/config"
	"github.com/aoaYaoa/go-gin-starter/internal/container"
	"github.com/aoaYaoa/go-gin-starter/internal/database"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/aoaYaoa/go-gin-starter/internal/publisher"
	"github.com/aoaYaoa/go-gin-starter/internal/routes"
	"github.com/aoaYaoa/go-gin-starter/pkg/notify"
	"github.com/aoaYaoa/go-gin-starter/pkg/storage"
	authzutil "github.com/aoaYaoa/go-gin-starter/pkg/utils/authz"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/captcha"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tracing"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	config.Init()

	// 2. 初始化日志
	logger.Init()

	// 3. 初始化 OpenTelemetry
	traceShutdown, err := tracing.Init(tracing.Config{
		Endpoint:    config.AppConfig.OTELEndpoint,
		ServiceName: config.AppConfig.OTELServiceName,
	})
	if err != nil {
		logger.Errorf("OpenTelemetry 初始化失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := traceShutdown(ctx); err != nil {
			logger.Errorf("OpenTelemetry 关闭失败: %v", err)
		}
	}()

	// 4. 设置 Gin 模式
	gin.SetMode(config.AppConfig.ServerMode)

	// 5. 创建 Gin 引擎
	r := gin.New()

	// 全局中间件管道
	r.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Trace(config.AppConfig.OTELServiceName),
		middleware.TraceContext(),
		middleware.Logger(),
		middleware.CORS(config.AppConfig.CORSOrigins),
		middleware.IPBlacklist(config.GetIPBlacklist()),
		middleware.IPWhitelist(config.GetIPWhitelist()),
		middleware.Metrics(),
	)

	// 6. 初始化数据库
	db, err := database.New(&config.AppConfig)
	if err != nil {
		logger.Errorf("数据库初始化失败: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// 5. 执行数据库迁移
	if config.AppConfig.MigrationsPath != "" {
		dsn := database.BuildMigrateDSN(&config.AppConfig)
		if err := database.RunMigrations(config.AppConfig.MigrationsPath, dsn); err != nil {
			logger.Errorf("数据库迁移失败: %v", err)
			os.Exit(1)
		}
	}

	// 6. JWT 初始化
	if err := jwt.SetDefaultSecret(config.AppConfig.JWTSecret); err != nil {
		logger.Errorf("JWT 初始化失败: %v", err)
		os.Exit(1)
	}

	// 7. 验证码存储初始化
	captcha.InitStore(captcha.StoreConfig{
		Addr:     config.AppConfig.RedisAddr,
		Username: config.AppConfig.RedisUsername,
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
		UseTLS:   config.AppConfig.RedisTLS,
	})

	// 8. Refresh Token 存储初始化（Redis 可选，未配置时退化为内存存储）
	var refreshStore jwt.RefreshTokenStore
	if config.AppConfig.RedisAddr != "" {
		rs, err := jwt.NewRefreshStore(
			config.AppConfig.RedisAddr,
			config.AppConfig.RedisUsername,
			config.AppConfig.RedisPassword,
			config.AppConfig.RedisDB,
			config.AppConfig.RedisTLS,
		)
		if err != nil {
			logger.Errorf("RefreshStore Redis 初始化失败: %v", err)
			os.Exit(1)
		}
		refreshStore = rs
		logger.Infof("RefreshStore: Redis (%s)", config.AppConfig.RedisAddr)
	} else {
		refreshStore = jwt.NewMemoryRefreshStore()
		logger.Warn("REDIS_ADDR 未配置，RefreshStore 退化为内存存储（不适合多实例部署）")
	}

	// 9. Kafka AsyncPublisher + OutboxPublisher 初始化
	var outbox *publisher.OutboxPublisher
	if config.AppConfig.KafkaBrokers != "" {
		asyncPub, err := publisher.NewAsyncPublisher(publisher.Config{
			Brokers:          config.AppConfig.KafkaBrokers,
			SecurityProtocol: config.AppConfig.KafkaSecurityProtocol,
			SSLCAFile:        config.AppConfig.KafkaSSLCAFile,
			SSLCertFile:      config.AppConfig.KafkaSSLCertFile,
			SSLKeyFile:       config.AppConfig.KafkaSSLKeyFile,
		})
		if err != nil {
			logger.Errorf("Kafka 初始化失败: %v", err)
			os.Exit(1)
		}
		outbox = publisher.NewOutboxPublisher(db.DB, asyncPub)
		logger.Infof("Kafka 已连接: %s", config.AppConfig.KafkaBrokers)
	} else {
		logger.Info("KAFKA_BROKERS 未配置，跳过 Kafka 初始化")
	}

	// 10. 权限引擎初始化
	authorizer, err := authzutil.New(db.DB, "configs/casbin_model.conf")
	if err != nil {
		logger.Errorf("Casbin 初始化失败: %v", err)
		os.Exit(1)
	}
	authzutil.SetDefault(authorizer)

	// 11. 文件存储初始化
	var objectStorage storage.Storage
	switch config.AppConfig.StorageType {
	case "s3":
		objectStorage, err = storage.NewS3(context.Background(), storage.S3Config{
			Bucket:       config.AppConfig.S3Bucket,
			Region:       config.AppConfig.S3Region,
			Endpoint:     config.AppConfig.S3Endpoint,
			AccessKey:    config.AppConfig.S3AccessKey,
			SecretKey:    config.AppConfig.S3SecretKey,
			UsePathStyle: config.AppConfig.S3UsePathStyle,
		})
		if err != nil {
			logger.Errorf("S3 存储初始化失败: %v", err)
			os.Exit(1)
		}
	default:
		objectStorage = storage.NewLocal(config.AppConfig.LocalStorageDir, config.AppConfig.StoragePublicBaseURL)
	}

	// 12. 通知器初始化
	var notifier notify.Notifier = notify.NewNoop()
	if config.AppConfig.SMTPHost != "" {
		notifier = notify.NewSMTP(notify.SMTPConfig{
			Host: config.AppConfig.SMTPHost,
			Port: config.AppConfig.SMTPPort,
			User: config.AppConfig.SMTPUser,
			Pass: config.AppConfig.SMTPPass,
			From: config.AppConfig.SMTPFrom,
		})
	}

	// 13. 依赖注入 & 路由注册
	if config.AppConfig.ServerMode == "release" && config.AppConfig.MetricsAllowedIPs == "" {
		logger.Warn("警告：生产模式下 METRICS_ALLOWED_IPS 未配置，/metrics 端点对所有 IP 开放")
	}
	c := container.New(db.DB, refreshStore, notifier, outbox, authorizer, objectStorage)
	middleware.SetAuditWriter(c.AuditRepo)
	healthHandler := handlers.NewHealthHandler(db)
	routes.Register(r, c, healthHandler, config.GetMetricsAllowedIPs())

	// 启动服务器
	addr := ":" + config.AppConfig.ServerPort
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Infof("服务器启动在 http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("服务器强制关闭: %v", err)
	}

	// 等待 Outbox Worker 完成
	if outbox != nil {
		logger.Info("等待 Outbox Worker 完成...")
		outbox.Close()
	}

	logger.Info("服务器已退出")
}
