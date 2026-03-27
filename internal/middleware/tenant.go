package middleware

import (
	"net"
	"strings"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderTenantID = "X-Tenant-ID"

// TenantResolver resolves tenant_id from JWT claims first, then from the
// X-Tenant-ID header, and finally from a UUID-shaped subdomain.
func TenantResolver() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := tenantFromClaims(c)
		if !ok {
			tenantID, ok = tenantFromHeader(c)
		}
		if !ok {
			tenantID, ok = tenantFromSubdomain(c.Request.Host)
		}
		if ok {
			c.Set("tenant_id", tenantID)
			c.Request = c.Request.WithContext(tenantctx.WithTenantID(c.Request.Context(), tenantID))
		}
		c.Next()
	}
}

func tenantFromClaims(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("claims")
	if !exists {
		return uuid.Nil, false
	}
	claims, ok := value.(*jwt.Claims)
	if !ok || claims.TenantID == uuid.Nil {
		return uuid.Nil, false
	}
	return claims.TenantID, true
}

func tenantFromHeader(c *gin.Context) (uuid.UUID, bool) {
	return parseTenantID(c.GetHeader(HeaderTenantID))
}

func tenantFromSubdomain(host string) (uuid.UUID, bool) {
	host = strings.TrimSpace(host)
	if host == "" {
		return uuid.Nil, false
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return uuid.Nil, false
	}
	return parseTenantID(parts[0])
}

func parseTenantID(raw string) (uuid.UUID, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil, false
	}
	tenantID, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return tenantID, true
}
