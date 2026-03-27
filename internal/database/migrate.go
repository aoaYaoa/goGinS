package database

import (
	"errors"
	"fmt"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations 执行版本化数据库迁移
func RunMigrations(migrationsPath, dsn string) error {
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("初始化 migrate 失败: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("数据库迁移：无新变更")
			return nil
		}
		return fmt.Errorf("执行迁移失败: %w", err)
	}

	version, _, _ := m.Version()
	logger.Infof("数据库迁移完成，当前版本: %d", version)
	return nil
}
