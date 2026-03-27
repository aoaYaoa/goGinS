package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

func (b *bucket) allow(rate float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.lastTime = now
	b.tokens += elapsed * rate
	if b.tokens > rate {
		b.tokens = rate
	}
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

type ipLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
}

func newIPLimiter(rate float64) *ipLimiter {
	return &ipLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
	}
}

func (l *ipLimiter) getBucket(ip string) *bucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.buckets[ip]; ok {
		return b
	}
	b := &bucket{tokens: l.rate, lastTime: time.Now()}
	l.buckets[ip] = b
	return b
}

// RateLimit 基于令牌桶的 IP 级限流，rate 为每秒允许请求数
func RateLimit(rate float64) gin.HandlerFunc {
	limiter := newIPLimiter(rate)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.getBucket(ip).allow(limiter.rate) {
			response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitByUser 基于令牌桶的用户级限流，需在 JWTAuth 之后使用。
// rate 为每秒允许请求数；未登录时退化为 IP 级限流。
func RateLimitByUser(rate float64) gin.HandlerFunc {
	limiter := newIPLimiter(rate)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if v, exists := c.Get("claims"); exists {
			if claims, ok := v.(*jwt.Claims); ok {
				key = claims.UserID.String()
			}
		}
		if !limiter.getBucket(key).allow(limiter.rate) {
			response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}
		c.Next()
	}
}
