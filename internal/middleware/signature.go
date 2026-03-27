package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

const (
	signatureMaxAge = 5 * time.Minute
)

// SignatureVerify 校验请求签名
// 请求头：X-Timestamp（Unix 秒）、X-Nonce、X-Signature
// 签名规则：HMAC-SHA256(secret, "timestamp=T&nonce=N&path=P")
func SignatureVerify(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		timestampStr := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		sig := c.GetHeader("X-Signature")

		if timestampStr == "" || nonce == "" || sig == "" {
			response.FailWithCode(c, http.StatusBadRequest, "缺少签名头")
			c.Abort()
			return
		}

		ts, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.FailWithCode(c, http.StatusBadRequest, "时间戳格式错误")
			c.Abort()
			return
		}

		if time.Since(time.Unix(ts, 0)).Abs() > signatureMaxAge {
			response.FailWithCode(c, http.StatusUnauthorized, "请求已过期")
			c.Abort()
			return
		}

		expected := computeSignature(secret, timestampStr, nonce, c.Request.URL.Path)
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			response.FailWithCode(c, http.StatusUnauthorized, "签名校验失败")
			c.Abort()
			return
		}

		c.Next()
	}
}

func computeSignature(secret, timestamp, nonce, path string) string {
	msg := fmt.Sprintf("timestamp=%s&nonce=%s&path=%s", timestamp, nonce, path)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}
