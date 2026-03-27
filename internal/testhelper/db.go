package testhelper

import (
	"context"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

// SetupPostgres 启动一个临时 PostgreSQL 容器，返回 *gorm.DB 和清理函数
func SetupPostgres(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("启动 postgres 容器失败: %v", err)
	}

	// 使用 ConnectionString 方法获取连接字符串
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("获取连接字符串失败: %v", err)
	}

	db, err := database.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("连接测试数据库失败: %v", err)
	}

	return db, func() {
		_ = ctr.Terminate(ctx)
	}
}
