package publisher

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"gorm.io/gorm"
)

// OutboxPublisher 将消息写入 outbox 表（与业务同事务），由 Worker 轮询发送
type OutboxPublisher struct {
	db       *gorm.DB
	upstream Publisher // 真实 Kafka 发布器
	stopCh   chan struct{}
	doneCh   chan struct{}
	once     sync.Once
}

// NewOutboxPublisher 创建 OutboxPublisher，并启动后台 Worker
func NewOutboxPublisher(db *gorm.DB, upstream Publisher) *OutboxPublisher {
	op := &OutboxPublisher{
		db:       db,
		upstream: upstream,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	go op.worker()
	return op
}

// Save 在同一事务中写入 outbox 消息（供 Service 层调用）
func (op *OutboxPublisher) Save(ctx context.Context, tx *gorm.DB, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := &models.OutboxMessage{
		Topic:   topic,
		Payload: string(data),
		Status:  models.OutboxStatusPending,
	}
	return tx.WithContext(ctx).Create(msg).Error
}

// Publish 直接发布（非事务场景，降级使用）
func (op *OutboxPublisher) Publish(topic string, payload []byte) error {
	return op.upstream.Publish(topic, payload)
}

// Close signals the worker to stop and waits for it to finish.
// Safe to call multiple times.
func (op *OutboxPublisher) Close() error {
	op.once.Do(func() {
		close(op.stopCh)
	})
	<-op.doneCh
	return op.upstream.Close()
}

// worker 定时扫描 pending 消息并通过 upstream 发送
func (op *OutboxPublisher) worker() {
	defer close(op.doneCh)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-op.stopCh:
			logger.Info("[outbox] Worker 收到停止信号，执行最后一次 flush")
			op.flush()
			logger.Info("[outbox] Worker 已退出")
			return
		case <-ticker.C:
			op.flush()
		}
	}
}

func (op *OutboxPublisher) flush() {
	if op.db == nil {
		return
	}
	ctx := context.Background()
	var msgs []models.OutboxMessage
	if err := op.db.WithContext(ctx).
		Where("status = ?", models.OutboxStatusPending).
		Order("id asc").
		Limit(100).
		Find(&msgs).Error; err != nil {
		logger.Warnf("[outbox] 查询失败: %v", err)
		return
	}
	for _, msg := range msgs {
		err := op.upstream.Publish(msg.Topic, []byte(msg.Payload))
		now := time.Now()
		update := map[string]any{"processed_at": &now}
		if err != nil {
			logger.Warnf("[outbox] 发送失败 id=%d: %v", msg.ID, err)
			update["status"] = models.OutboxStatusFailed
		} else {
			update["status"] = models.OutboxStatusSent
		}
		op.db.WithContext(ctx).Model(&msg).Updates(update)
	}
}
