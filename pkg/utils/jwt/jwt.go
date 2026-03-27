package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var defaultSecret []byte

const passwordResetSubject = "password_reset"

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id,omitempty"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// SetDefaultSecret 设置全局签名密钥（启动时调用）
func SetDefaultSecret(secret string) error {
	if len(secret) < 32 {
		return errors.New("JWT_SECRET 长度不得少于 32 字节")
	}
	defaultSecret = []byte(secret)
	return nil
}

// GenerateAccessToken 生成 access token（默认 15 分钟）
func GenerateAccessToken(userID uuid.UUID, username, role string) (string, error) {
	return generateToken(userID, uuid.Nil, username, role, 15*time.Minute)
}

// GenerateAccessTokenWithTenant 生成包含 tenant_id 的 access token。
func GenerateAccessTokenWithTenant(userID, tenantID uuid.UUID, username, role string) (string, error) {
	return generateToken(userID, tenantID, username, role, 15*time.Minute)
}

// GenerateRefreshToken 生成 refresh token（默认 7 天）
func GenerateRefreshToken(userID uuid.UUID, username, role string) (string, error) {
	return generateToken(userID, uuid.Nil, username, role, 7*24*time.Hour)
}

// GenerateRefreshTokenWithTenant 生成包含 tenant_id 的 refresh token。
func GenerateRefreshTokenWithTenant(userID, tenantID uuid.UUID, username, role string) (string, error) {
	return generateToken(userID, tenantID, username, role, 7*24*time.Hour)
}

func GeneratePasswordResetTokenWithTenant(userID, tenantID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			Subject:   passwordResetSubject,
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(defaultSecret)
}

func generateToken(userID, tenantID uuid.UUID, username, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(defaultSecret)
}

func ParsePasswordResetToken(tokenStr string) (*Claims, error) {
	claims, err := ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Subject != passwordResetSubject {
		return nil, errors.New("无效的 password reset token")
	}
	return claims, nil
}

// ParseToken 解析并校验 token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return defaultSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token 已过期")
		}
		return nil, errors.New("无效的 token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 token")
	}
	return claims, nil
}
