package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/config"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tracing"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库实例
type DB struct {
	*gorm.DB
}

// New 初始化 PostgreSQL 连接
func New(cfg *config.Config) (*DB, error) {
	dsn := BuildDSN(cfg)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}
	if cfg.ServerMode != "release" {
		gormCfg.Logger = gormlogger.Default.LogMode(gormlogger.Warn)
	}

	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
		if err == nil {
			break
		}
		logger.Warnf("数据库连接失败（%d/5）: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	if err := db.Use(tracing.NewGormPlugin()); err != nil {
		return nil, fmt.Errorf("注册数据库 tracing 插件失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleTime) * time.Second)

	logger.Infof("数据库连接成功: %s:%s/%s", cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	return &DB{db}, nil
}

// NewFromDSN 直接用 DSN 字符串初始化（测试用）
func NewFromDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	if err := db.Use(tracing.NewGormPlugin()); err != nil {
		return nil, fmt.Errorf("注册数据库 tracing 插件失败: %w", err)
	}
	return db, nil
}

// BuildDSN 构建 PostgreSQL DSN（GORM key=value 格式）
func BuildDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePass,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)
}

// BuildMigrateDSN 构建 golang-migrate 所需的 URL 格式 DSN
func BuildMigrateDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DatabaseUser,
		cfg.DatabasePass,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)
}

// Ping 健康检查
func (d *DB) Ping(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close 关闭连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
