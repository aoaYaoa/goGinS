package middleware

import (
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// JWTAuth 验证 Access Token，将 claims 注入 context
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Fail(c, apperr.ErrInvalidToken)
			c.Abort()
			return
		}

		if claims.TenantID != uuid.Nil {
			c.Set("tenant_id", claims.TenantID)
			c.Request = c.Request.WithContext(tenantctx.WithTenantID(c.Request.Context(), claims.TenantID))
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// RequireRole 角色权限检查，在 JWTAuth 之后使用
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("claims")
		if !exists {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		claims := v.(*jwt.Claims)
		for _, r := range roles {
			if claims.Role == r {
				c.Next()
				return
			}
		}
		response.Fail(c, apperr.ErrForbidden)
		c.Abort()
	}
}

// JWTAuthWithStatusCheck validates the Access Token and additionally checks that
// the user account is still active. Use this on routes where a banned user must
// be rejected immediately even if their token has not yet expired.
func JWTAuthWithStatusCheck(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Fail(c, apperr.ErrInvalidToken)
			c.Abort()
			return
		}

		reqCtx := c.Request.Context()
		if claims.TenantID != uuid.Nil {
			c.Set("tenant_id", claims.TenantID)
			reqCtx = tenantctx.WithTenantID(reqCtx, claims.TenantID)
			c.Request = c.Request.WithContext(reqCtx)
		}

		user, err := userRepo.FindByID(reqCtx, claims.UserID)
		if err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		if user.Status != "active" {
			response.Fail(c, apperr.ErrAccountDisabled)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
