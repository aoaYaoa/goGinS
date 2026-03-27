package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aoaYaoa/go-gin-starter/pkg/storage"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	storage storage.Storage
}

func NewUploadHandler(storage storage.Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// Upload godoc
// @Summary 上传文件
// @Tags upload
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "上传文件"
// @Success 200 {object} map[string]string
// @Router /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "缺少上传文件")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无法读取上传文件")
		return
	}
	defer src.Close()

	key := buildUploadKey(c, file.Filename)
	url, err := h.storage.Upload(c.Request.Context(), key, src)
	if err != nil {
		response.FailWithCode(c, http.StatusInternalServerError, "文件上传失败")
		return
	}

	response.OK(c, gin.H{
		"url": url,
		"key": key,
	})
}

func buildUploadKey(c *gin.Context, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	id := uuid.New().String()
	if tenantID, ok := tenantctx.FromContext(c.Request.Context()); ok {
		return tenantID.String() + "/" + id + ext
	}
	return "public/" + id + ext
}
