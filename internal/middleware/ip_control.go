package middleware

import (
	"net"
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

// matchesAny reports whether ip matches any entry in list.
// Each entry may be a plain IP address or a CIDR range.
func matchesAny(ip string, list []string) bool {
	parsed := net.ParseIP(ip)
	for _, entry := range list {
		if _, network, err := net.ParseCIDR(entry); err == nil {
			if parsed != nil && network.Contains(parsed) {
				return true
			}
		} else if entry == ip {
			return true
		}
	}
	return false
}

// IPWhitelist allows only listed IPs/CIDRs. Empty list passes all.
func IPWhitelist(whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(whitelist) == 0 {
			c.Next()
			return
		}
		if !matchesAny(c.ClientIP(), whitelist) {
			response.FailWithCode(c, http.StatusForbidden, "IP 不在白名单")
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPBlacklist blocks listed IPs/CIDRs.
func IPBlacklist(blacklist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(blacklist) > 0 && matchesAny(c.ClientIP(), blacklist) {
			response.FailWithCode(c, http.StatusForbidden, "IP 已被封禁")
			c.Abort()
			return
		}
		c.Next()
	}
}
