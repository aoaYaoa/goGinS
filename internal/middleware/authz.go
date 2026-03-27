package middleware

import (
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	authzutil "github.com/aoaYaoa/go-gin-starter/pkg/utils/authz"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

// Authz performs resource-level authorization using the configured authorizer.
func Authz(obj, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("claims")
		if !exists {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, ok := v.(*jwt.Claims)
		if !ok {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		allowed, err := authzutil.Check(claims.Role, obj, act)
		if err != nil {
			response.Fail(c, apperr.ErrInternalServer)
			c.Abort()
			return
		}
		if !allowed {
			response.Fail(c, apperr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
