package captcha

import (
	"context"
	"crypto/tls"
	"encoding/json"
"time"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

const captchaTTL = 5 * time.Minute

// Store 验证码存储接口
type Store interface {
	Set(id string, value string) error
	Get(id string, clear bool) string
	Verify(id, answer string, clear bool) bool
}

// RedisStore Redis 验证码存储
type RedisStore struct {
	client *redis.Client
	prefix string
}

// MemoryStore 内存验证码存储（Redis 不可用时回退）
type MemoryStore struct {
	inner base64Captcha.Store
}

var defaultStore Store

// StoreConfig 验证码存储配置
type StoreConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
	UseTLS   bool
	Prefix   string
}

// InitStore 初始化验证码存储，优先 Redis，失败回退内存
func InitStore(cfg StoreConfig) {
	if cfg.Addr == "" {
		logger.Info("[captcha] REDIS_ADDR 未配置，使用内存存储")
		defaultStore = &MemoryStore{inner: base64Captcha.DefaultMemStore}
		return
	}

	opt := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.UseTLS {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warnf("[captcha] Redis 初始化失败，回退内存存储: %v", err)
		defaultStore = &MemoryStore{inner: base64Captcha.DefaultMemStore}
		return
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "captcha:"
	}
	defaultStore = &RedisStore{client: client, prefix: prefix}
	logger.Infof("[captcha] 使用 Redis 存储验证码: %s", cfg.Addr)
}

// GetStore 获取当前存储实例
func GetStore() Store {
	if defaultStore == nil {
		defaultStore = &MemoryStore{inner: base64Captcha.DefaultMemStore}
	}
	return defaultStore
}

// --- RedisStore 实现 ---

func (s *RedisStore) Set(id, value string) error {
	ctx := context.Background()
	data, _ := json.Marshal(value)
	return s.client.Set(ctx, s.prefix+id, data, captchaTTL).Err()
}

func (s *RedisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	val, err := s.client.Get(ctx, s.prefix+id).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.client.Del(ctx, s.prefix+id)
	}
	var answer string
	_ = json.Unmarshal([]byte(val), &answer)
	return answer
}

func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v != "" && v == answer
}

// --- MemoryStore 实现 ---

func (s *MemoryStore) Set(id, value string) error {
	return s.inner.Set(id, value)
}

func (s *MemoryStore) Get(id string, clear bool) string {
	return s.inner.Get(id, clear)
}

func (s *MemoryStore) Verify(id, answer string, clear bool) bool {
	return s.inner.Verify(id, answer, clear)
}
