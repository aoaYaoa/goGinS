package services

import (
	"context"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/events"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/internal/publisher"
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/notify"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	// Admin
	ListUsers(ctx context.Context, page, size int) (*dto.UserListResponse, error)
	SetUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	SetUserRole(ctx context.Context, userID uuid.UUID, role string) error
}

type userService struct {
	userRepo     repositories.UserRepository
	refreshStore jwt.RefreshTokenStore
	notifier     notify.Notifier
	pub          *publisher.OutboxPublisher
	db           *gorm.DB
}

func NewUserService(userRepo repositories.UserRepository, refreshStore jwt.RefreshTokenStore, notifier notify.Notifier, pub *publisher.OutboxPublisher, db *gorm.DB) UserService {
	return &userService{
		userRepo:     userRepo,
		refreshStore: refreshStore,
		notifier:     notifier,
		pub:          pub,
		db:           db,
	}
}

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(ctx, req.Username); err == nil {
		return nil, apperr.ErrUserExists
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		if _, err := s.userRepo.FindByEmail(ctx, req.Email); err == nil {
			return nil, apperr.ErrUserExists
		}
	}

	// 哈希密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
		Role:     "user",
		Status:   "active",
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}

	if s.notifier != nil && created.Email != "" {
		if err := s.notifier.Send(ctx, created.Email, "欢迎注册 go-gin-starter", "欢迎加入 go-gin-starter。"); err != nil {
			logger.Warnf("[notify] 欢迎邮件发送失败: %v", err)
		}
	}

	return s.generateAuthResponse(ctx, created)
}

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperr.ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperr.ErrInvalidPassword
	}

	if user.Status != "active" {
		return nil, apperr.ErrAccountDisabled
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_, _ = s.userRepo.Update(ctx, user)

	// 发布登录事件（写入 outbox 表，保证事务一致性）
	if s.pub != nil && s.db != nil {
		event := events.LoginEvent{
			EventID:    uuid.New().String(),
			UserID:     user.ID,
			Username:   user.Username,
			OccurredAt: now,
		}
		_ = s.pub.Save(ctx, s.db, events.TopicUserLogin, event)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// 校验 refresh token 签名与有效期
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return nil, apperr.ErrRefreshToken
	}
	if claims.TenantID != uuid.Nil {
		ctx = tenantctx.WithTenantID(ctx, claims.TenantID)
	}

	// 校验 Redis 中是否存在（未被吊销）
	exists, err := s.refreshStore.Exists(ctx, refreshToken)
	if err != nil || !exists {
		return nil, apperr.ErrRefreshToken
	}

	// 吊销旧 refresh token（token rotation）
	_ = s.refreshStore.Delete(ctx, refreshToken)

	// 查询用户（确保用户仍存在且未被禁用）
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}

	if user.Status != "active" {
		_ = s.refreshStore.Delete(ctx, refreshToken)
		return nil, apperr.ErrAccountDisabled
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshStore.Delete(ctx, refreshToken)
}

func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if s.notifier == nil || user.Email == "" {
		return nil
	}

	token, err := jwt.GeneratePasswordResetTokenWithTenant(user.ID, user.TenantID)
	if err != nil {
		return apperr.ErrInternalServer
	}

	if err := s.notifier.Send(ctx, user.Email, "go-gin-starter 密码重置", token); err != nil {
		logger.Warnf("[notify] 密码重置邮件发送失败: %v", err)
	}
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := jwt.ParsePasswordResetToken(token)
	if err != nil {
		return apperr.ErrInvalidToken
	}
	if claims.TenantID != uuid.Nil {
		ctx = tenantctx.WithTenantID(ctx, claims.TenantID)
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return apperr.ErrUserNotFound
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperr.ErrInternalServer
	}
	user.Password = string(hashed)
	if _, err := s.userRepo.Update(ctx, user); err != nil {
		return apperr.ErrInternalServer
	}
	return nil
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	if req.FullName != nil {
		user.FullName = req.FullName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	resp := toUserResponse(updated)
	return &resp, nil
}

func (s *userService) generateAuthResponse(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := jwt.GenerateAccessTokenWithTenant(user.ID, user.TenantID, user.Username, user.Role)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	refreshToken, err := jwt.GenerateRefreshTokenWithTenant(user.ID, user.TenantID, user.Username, user.Role)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	if err := s.refreshStore.Save(ctx, refreshToken, refreshTokenTTL); err != nil {
		return nil, apperr.ErrInternalServer
	}
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *userService) ListUsers(ctx context.Context, page, size int) (*dto.UserListResponse, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.List(ctx, offset, size)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	result := make([]*dto.UserResponse, 0, len(users))
	for _, u := range users {
		r := toUserResponse(u)
		result = append(result, &r)
	}
	return &dto.UserListResponse{Total: total, Page: page, Size: size, Items: result}, nil
}

func (s *userService) SetUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrUserNotFound
	}
	user.Status = status
	_, err = s.userRepo.Update(ctx, user)
	return err
}

func (s *userService) SetUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrUserNotFound
	}
	user.Role = role
	_, err = s.userRepo.Update(ctx, user)
	return err
}

func toUserResponse(u *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		Status:   u.Status,
	}
}
