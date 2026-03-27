package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	kafkago "github.com/segmentio/kafka-go"
)

// AsyncPublisher 异步 Kafka 发布器，带本地队列和重试
type AsyncPublisher struct {
	writer *kafkago.Writer
	queue  chan message
	done   chan struct{}
	once   sync.Once
}

type message struct {
	topic   string
	payload []byte
}

type Config struct {
	Brokers          string
	SecurityProtocol string
	SSLCAFile        string
	SSLCertFile      string
	SSLKeyFile       string
}

// NewAsyncPublisher 创建 Kafka 异步发布器
func NewAsyncPublisher(cfg Config) (*AsyncPublisher, error) {
	brokers := strings.Split(cfg.Brokers, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	wCfg := kafkago.WriterConfig{
		Brokers:      brokers,
		Balancer:     &kafkago.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	if strings.ToUpper(cfg.SecurityProtocol) == "SSL" {
		tlsCfg, err := buildTLS(cfg)
		if err != nil {
			return nil, fmt.Errorf("kafka TLS 配置失败: %w", err)
		}
		wCfg.Dialer = &kafkago.Dialer{
			Timeout:   10 * time.Second,
			TLS: tlsCfg,
		}
	}

	p := &AsyncPublisher{
		writer: kafkago.NewWriter(wCfg),
		queue:  make(chan message, 1024),
		done:   make(chan struct{}),
	}
	go p.run()
	return p, nil
}

func (p *AsyncPublisher) Publish(topic string, payload []byte) error {
	select {
	case p.queue <- message{topic: topic, payload: payload}:
		return nil
	default:
		return fmt.Errorf("kafka 队列已满，消息丢弃")
	}
}

// Close shuts down the async publisher. Safe to call multiple times.
func (p *AsyncPublisher) Close() error {
	p.once.Do(func() {
		close(p.done)
	})
	return p.writer.Close()
}

func (p *AsyncPublisher) run() {
	for {
		select {
		case <-p.done:
			return
		case msg := <-p.queue:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := p.writer.WriteMessages(ctx, kafkago.Message{
				Topic: msg.topic,
				Value: msg.payload,
			})
			cancel()
			if err != nil {
				logger.Warnf("[kafka] 消息发送失败: %v", err)
			}
		}
	}
}

func buildTLS(cfg Config) (*tls.Config, error) {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.SSLCAFile != "" {
		ca, err := os.ReadFile(cfg.SSLCAFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(ca)
		tlsCfg.RootCAs = pool
	}
	if cfg.SSLCertFile != "" && cfg.SSLKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.SSLCertFile, cfg.SSLKeyFile)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return tlsCfg, nil
}
