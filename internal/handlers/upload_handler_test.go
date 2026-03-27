package handlers_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	uploadFn func(ctx context.Context, key string, reader io.Reader) (string, error)
	deleteFn func(ctx context.Context, key string) error
}

func (m *mockStorage) Upload(ctx context.Context, key string, reader io.Reader) (string, error) {
	return m.uploadFn(ctx, key, reader)
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

var _ storage.Storage = (*mockStorage)(nil)

func TestUploadHandlerUpload_Success(t *testing.T) {
	store := &mockStorage{
		uploadFn: func(_ context.Context, key string, reader io.Reader) (string, error) {
			assert.NotEmpty(t, key)
			body, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, "hello", string(body))
			return "/uploads/test.txt", nil
		},
	}

	h := handlers.NewUploadHandler(store)
	r := gin.New()
	r.POST("/upload", h.Upload)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	_, err = part.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/uploads/test.txt")
}

func TestUploadHandlerUpload_MissingFile(t *testing.T) {
	h := handlers.NewUploadHandler(&mockStorage{})
	r := gin.New()
	r.POST("/upload", h.Upload)

	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
