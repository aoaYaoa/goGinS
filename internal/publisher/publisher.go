package publisher

// Publisher 消息发布接口
type Publisher interface {
	Publish(topic string, payload []byte) error
	Close() error
}
