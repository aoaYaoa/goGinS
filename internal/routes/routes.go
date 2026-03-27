package routes

import (
	"github.com/aoaYaoa/go-gin-starter/docs"
	"github.com/aoaYaoa/go-gin-starter/internal/config"
	"github.com/aoaYaoa/go-gin-starter/internal/container"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Register 注册所有路由
func Register(r *gin.Engine, c *container.Container, health *handlers.HealthHandler, metricsAllowedIPs []string) {
	cfg := &config.AppConfig

	// Swagger UI（非 release 模式）
	if cfg.ServerMode != "release" {
		docs.SwaggerInfo.Host = "localhost:" + cfg.ServerPort
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 健康检查端点
	r.GET("/healthz", health.Liveness)
	r.GET("/readyz", health.Readiness)

	// Prometheus 指标端点（IP 访问控制）
	r.GET("/metrics", middleware.MetricsAllowedIPs(metricsAllowedIPs), gin.WrapH(promhttp.Handler()))

	if cfg.StorageType == "local" && cfg.LocalStorageDir != "" && cfg.StoragePublicBaseURL != "" {
		r.Static(cfg.StoragePublicBaseURL, cfg.LocalStorageDir)
	}

	v1Middlewares := []gin.HandlerFunc{}
	if cfg.EnableSignature && cfg.SignatureSecret != "" {
		v1Middlewares = append(v1Middlewares, middleware.SignatureVerify(cfg.SignatureSecret))
	}
	v1 := r.Group("/api/v1", v1Middlewares...)

	// 认证路由（公开，IP 级限流 5 req/s）
	auth := v1.Group("/auth", middleware.TenantResolver(), middleware.RateLimit(5))
	{
		auth.GET("/captcha", c.UserHandler.GetCaptcha)
		auth.POST("/register", middleware.Audit("auth.register", "user"), c.UserHandler.Register)
		auth.POST("/login", middleware.Audit("auth.login", "user"), c.UserHandler.Login)
		auth.POST("/refresh", c.UserHandler.RefreshToken)
		auth.POST("/logout", c.UserHandler.Logout)
		auth.POST("/forgot-password", c.UserHandler.RequestPasswordReset)
		auth.POST("/reset-password", c.UserHandler.ResetPassword)
	}

	// 用户路由（需要认证，用户级限流 20 req/s）
	user := v1.Group("/user", middleware.JWTAuthWithStatusCheck(c.UserRepo), middleware.TenantResolver(), middleware.RateLimitByUser(20))
	{
		user.GET("/profile", c.UserHandler.GetProfile)
		user.PUT("/profile", c.UserHandler.UpdateProfile)
	}

	// 任务路由（需要认证，用户级限流 20 req/s）
	tasks := v1.Group("/tasks", middleware.JWTAuthWithStatusCheck(c.UserRepo), middleware.TenantResolver(), middleware.RateLimitByUser(20))
	{
		tasks.POST("", middleware.Audit("task.create", "task"), c.TaskHandler.Create)
		tasks.GET("", c.TaskHandler.List)
		tasks.GET("/:id", c.TaskHandler.GetByID)
		tasks.PUT("/:id", middleware.Audit("task.update", "task"), c.TaskHandler.Update)
		tasks.DELETE("/:id", middleware.Audit("task.delete", "task"), c.TaskHandler.Delete)
	}

	v1.POST("/upload",
		middleware.JWTAuthWithStatusCheck(c.UserRepo),
		middleware.TenantResolver(),
		middleware.RateLimitByUser(20),
		middleware.Audit("file.upload", "file"),
		c.UploadHandler.Upload,
	)

	// 管理员路由（需要细粒度权限）
	admin := v1.Group("/admin", middleware.JWTAuthWithStatusCheck(c.UserRepo), middleware.TenantResolver())
	{
		admin.GET("/users", middleware.Authz("users", "read"), c.AdminHandler.ListUsers)
		admin.GET("/users/:id", middleware.Authz("users", "read"), c.AdminHandler.GetUser)
		admin.PUT("/users/:id/ban", middleware.Authz("users", "write"), middleware.Audit("admin.user.ban", "user"), c.AdminHandler.BanUser)
		admin.PUT("/users/:id/unban", middleware.Authz("users", "write"), middleware.Audit("admin.user.unban", "user"), c.AdminHandler.UnbanUser)
		admin.PUT("/users/:id/promote", middleware.Authz("users", "write"), middleware.Audit("admin.user.promote", "user"), c.AdminHandler.PromoteUser)
		admin.PUT("/users/:id/demote", middleware.Authz("users", "write"), middleware.Audit("admin.user.demote", "user"), c.AdminHandler.DemoteUser)
		admin.GET("/tasks", middleware.Authz("tasks", "read"), c.AdminHandler.ListAllTasks)
		admin.DELETE("/tasks/:id", middleware.Authz("tasks", "write"), middleware.Audit("admin.task.delete", "task"), c.AdminHandler.DeleteTask)
		admin.GET("/audit-logs", middleware.Authz("audit_logs", "read"), c.AuditHandler.List)
	}
}
