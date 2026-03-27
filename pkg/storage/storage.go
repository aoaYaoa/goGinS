// Package storage provides a Storage interface with local-disk and S3-compatible
// implementations. Use NewLocal for development and NewS3 for production.
package storage

import (
	"context"
	"io"
)

// Storage is the file-storage abstraction used by upload handlers.
// Upload stores the file at key and returns a public URL.
// Delete removes the file; a missing key is not treated as an error.
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
}
