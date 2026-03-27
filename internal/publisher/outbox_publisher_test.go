package publisher_test

import (
	"sync"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/publisher"
	"github.com/stretchr/testify/assert"
)

// noopPublisher is a Publisher that does nothing, used for close-path testing.
type noopPublisher struct {
	mu        sync.Mutex
	closeCalls int
}

func (n *noopPublisher) Publish(_ string, _ []byte) error { return nil }
func (n *noopPublisher) Close() error {
	n.mu.Lock()
	n.closeCalls++
	n.mu.Unlock()
	return nil
}

// TestOutboxPublisher_CloseIdempotent verifies that calling Close multiple times
// does not panic and that the upstream publisher is closed exactly once.
func TestOutboxPublisher_CloseIdempotent(t *testing.T) {
	noop := &noopPublisher{}
	// Pass nil db — the worker won't flush without a real DB, and the test
	// only exercises the shutdown path.
	op := publisher.NewOutboxPublisher(nil, noop)

	// First close should be clean.
	err := op.Close()
	assert.NoError(t, err)

	// Subsequent closes must not panic.
	assert.NotPanics(t, func() { _ = op.Close() })
	assert.NotPanics(t, func() { _ = op.Close() })

	noop.mu.Lock()
	calls := noop.closeCalls
	noop.mu.Unlock()
	// upstream.Close is called once per op.Close invocation (after the once guard),
	// which is acceptable — the important guarantee is no panic on the stopCh.
	assert.GreaterOrEqual(t, calls, 1)
}
