package container

import (
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/internal/publisher"
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/internal/services"
	"github.com/aoaYaoa/go-gin-starter/pkg/notify"
	"github.com/aoaYaoa/go-gin-starter/pkg/storage"
	authzutil "github.com/aoaYaoa/go-gin-starter/pkg/utils/authz"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"gorm.io/gorm"
)

// Container 持有所有依赖
type Container struct {
	UserHandler   *handlers.UserHandler
	TaskHandler   *handlers.TaskHandler
	UploadHandler *handlers.UploadHandler
	AdminHandler  *handlers.AdminHandler
	AuditHandler  *handlers.AuditHandler
	Outbox        *publisher.OutboxPublisher
	UserRepo      repositories.UserRepository
	AuditRepo     repositories.AuditRepository
	Authorizer    authzutil.Authorizer
}

// New 构建依赖树
func New(db *gorm.DB, refreshStore jwt.RefreshTokenStore, notifier notify.Notifier, outbox *publisher.OutboxPublisher, authorizer authzutil.Authorizer, objectStorage storage.Storage) *Container {
	// Repositories
	userRepo := repositories.NewUserRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	// Services
	userSvc := services.NewUserService(userRepo, refreshStore, notifier, outbox, db)
	taskSvc := services.NewTaskService(taskRepo)
	auditSvc := services.NewAuditService(auditRepo)

	// Handlers
	return &Container{
		UserHandler:   handlers.NewUserHandler(userSvc),
		TaskHandler:   handlers.NewTaskHandler(taskSvc),
		UploadHandler: handlers.NewUploadHandler(objectStorage),
		AdminHandler:  handlers.NewAdminHandler(userSvc, taskSvc),
		AuditHandler:  handlers.NewAuditHandler(auditSvc),
		Outbox:        outbox,
		UserRepo:      userRepo,
		AuditRepo:     auditRepo,
		Authorizer:    authorizer,
	}
}
