package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	rootDir       string
	publicBaseURL string
}

// NewLocal creates a LocalStorage that saves files under rootDir and
// serves them via publicBaseURL (defaults to "/uploads").
func NewLocal(rootDir, publicBaseURL string) *LocalStorage {
	if publicBaseURL == "" {
		publicBaseURL = "/uploads"
	}
	return &LocalStorage{
		rootDir:       rootDir,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (s *LocalStorage) Upload(_ context.Context, key string, reader io.Reader) (string, error) {
	fullPath := filepath.Join(s.rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", err
	}

	return s.publicBaseURL + "/" + strings.TrimLeft(filepath.ToSlash(key), "/"), nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.rootDir, filepath.FromSlash(key))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
