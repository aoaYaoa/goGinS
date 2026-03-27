package notify

import "context"

type NoopNotifier struct{}

// NewNoop returns a NoopNotifier that silently discards all messages.
// Use this when notifications are disabled or in tests.
func NewNoop() *NoopNotifier {
	return &NoopNotifier{}
}

func (n *NoopNotifier) Send(_ context.Context, _, _, _ string) error {
	return nil
}
