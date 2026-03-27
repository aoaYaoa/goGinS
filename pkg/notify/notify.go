// Package notify provides a Notifier interface and implementations for
// sending notifications (e.g. email). Use NewSMTP for production and
// NewNoop for testing or when notifications are disabled.
package notify

import "context"

// Notifier sends a message to a single recipient.
type Notifier interface {
	Send(ctx context.Context, to, subject, body string) error
}
