package services

import (
	"context"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
)

type AuditService interface {
	List(ctx context.Context, page, size int) (*dto.AuditLogListResponse, error)
}

type auditService struct {
	auditRepo repositories.AuditRepository
}

func NewAuditService(auditRepo repositories.AuditRepository) AuditService {
	return &auditService{auditRepo: auditRepo}
}

func (s *auditService) List(ctx context.Context, page, size int) (*dto.AuditLogListResponse, error) {
	offset := (page - 1) * size
	entries, total, err := s.auditRepo.List(ctx, offset, size)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}

	items := make([]dto.AuditLogResponse, 0, len(entries))
	for _, entry := range entries {
		item := dto.AuditLogResponse{
			ID:           entry.ID,
			TenantID:     entry.TenantID,
			Action:       entry.Action,
			ResourceType: entry.ResourceType,
			ResourceID:   entry.ResourceID,
			IP:           entry.IP,
			UserAgent:    entry.UserAgent,
			CreatedAt:    entry.CreatedAt,
		}
		if entry.UserID != nil {
			item.UserID = *entry.UserID
		}
		items = append(items, item)
	}

	return &dto.AuditLogListResponse{
		Total: total,
		Page:  page,
		Size:  size,
		Items: items,
	}, nil
}
