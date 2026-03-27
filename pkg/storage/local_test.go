package storage_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_UploadAndDelete(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocal(dir, "/uploads")

	url, err := store.Upload(context.Background(), "avatars/test.txt", bytes.NewBufferString("hello"))
	require.NoError(t, err)
	require.Equal(t, "/uploads/avatars/test.txt", url)

	content, err := os.ReadFile(filepath.Join(dir, "avatars", "test.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(content))

	require.NoError(t, store.Delete(context.Background(), "avatars/test.txt"))

	_, err = os.Stat(filepath.Join(dir, "avatars", "test.txt"))
	require.Error(t, err)
}
