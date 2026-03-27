package repositories

import (
	"context"
	"errors"

	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) (*models.Task, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Task, int64, error)
	Update(ctx context.Context, task *models.Task) (*models.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, offset, limit int) ([]*models.Task, int64, error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) (*models.Task, error) {
	if tenantID, ok := tenantIDFromContext(ctx); ok && task.TenantID == uuid.Nil {
		task.TenantID = tenantID
	}
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := applyTenantScope(ctx, r.db.WithContext(ctx)).First(&task, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Task, int64, error) {
	var tasks []*models.Task
	var total int64
	db := applyTenantScope(ctx, r.db.WithContext(ctx)).Model(&models.Task{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) (*models.Task, error) {
	if err := applyTenantScope(ctx, r.db.WithContext(ctx)).Save(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := applyTenantScope(ctx, r.db.WithContext(ctx)).Delete(&models.Task{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("任务不存在")
	}
	return nil
}

func (r *taskRepository) ListAll(ctx context.Context, offset, limit int) ([]*models.Task, int64, error) {
	var tasks []*models.Task
	var total int64
	db := applyTenantScope(ctx, r.db.WithContext(ctx)).Model(&models.Task{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
