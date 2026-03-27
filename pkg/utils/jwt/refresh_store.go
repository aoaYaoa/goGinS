package jwt

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshKeyPrefix = "refresh:"

// RefreshTokenStore is the interface for refresh token persistence.
// Redis-backed store is recommended for production; the in-memory store
// is suitable for single-process / local use only — tokens are lost on restart
// and are not shared across multiple instances.
type RefreshTokenStore interface {
	Save(ctx context.Context, token string, ttl time.Duration) error
	Exists(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
	Ping(ctx context.Context) error
}

// RefreshStore is the Redis-backed implementation of RefreshTokenStore.
type RefreshStore struct {
	client *redis.Client
}

// NewRefreshStore creates a Redis-backed RefreshStore.
func NewRefreshStore(addr, username, password string, db int, useTLS bool) (*RefreshStore, error) {
	opt := &redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	}
	if useTLS {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("RefreshStore Redis 连接失败: %w", err)
	}
	return &RefreshStore{client: client}, nil
}

func key(token string) string {
	return refreshKeyPrefix + token
}

func (s *RefreshStore) Save(ctx context.Context, token string, ttl time.Duration) error {
	return s.client.Set(ctx, key(token), "1", ttl).Err()
}

func (s *RefreshStore) Exists(ctx context.Context, token string) (bool, error) {
	n, err := s.client.Exists(ctx, key(token)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *RefreshStore) Delete(ctx context.Context, token string) error {
	return s.client.Del(ctx, key(token)).Err()
}

func (s *RefreshStore) Ping(ctx context.Context) error {
	if s == nil || s.client == nil {
		return errors.New("refresh store 未初始化")
	}
	return s.client.Ping(ctx).Err()
}

// MemoryRefreshStore is an in-memory RefreshTokenStore.
// Suitable for single-process / local use only. Tokens are not shared across
// multiple instances and are lost on process restart.
type MemoryRefreshStore struct {
	mu      sync.Mutex
	entries map[string]time.Time // token -> expiry
}

func NewMemoryRefreshStore() *MemoryRefreshStore {
	return &MemoryRefreshStore{entries: make(map[string]time.Time)}
}

func (m *MemoryRefreshStore) Save(_ context.Context, token string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[token] = time.Now().Add(ttl)
	return nil
}

func (m *MemoryRefreshStore) Exists(_ context.Context, token string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.entries[token]
	if !ok {
		return false, nil
	}
	if time.Now().After(exp) {
		delete(m.entries, token)
		return false, nil
	}
	return true, nil
}

func (m *MemoryRefreshStore) Delete(_ context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, token)
	return nil
}

func (m *MemoryRefreshStore) Ping(_ context.Context) error { return nil }
