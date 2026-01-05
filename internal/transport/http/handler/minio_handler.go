package http

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/Lomank123/go-web-platform/logger"

	"github.com/gin-gonic/gin"

	"github.com/Lomank123/go-integration-minio/internal/service"

	clientDTO "github.com/Lomank123/go-integration-minio/pkg/client/dto"
	platformError "github.com/Lomank123/go-web-platform/error"
)

type MinioHandler interface {
	UploadFileV1(c *gin.Context)
	DeleteFileV1(c *gin.Context)
	GetFileContentV1(c *gin.Context)
	GetFileURLV1(c *gin.Context)
	GetFile(c *gin.Context)
}

type minioHandler struct {
	service service.MinioService
	log     logger.Logger
}

func NewMinioHandler(service service.MinioService, log logger.Logger) MinioHandler {
	return &minioHandler{service: service, log: log}
}

// UploadFileV1 handles file upload requests
// @Summary Upload a file
// @Description Upload a file to MinIO storage
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param bucket_name formData string true "Bucket name"
// @Param filename formData string false "Custom filename (optional)"
// @Success 200 {object} clientDTO.FileUploadV1ResponseDTO
// @Failure 400 {object} map[string]string "Bad Request - Invalid file or file type"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/minio/upload [post]
// @Tags minio
func (eh *minioHandler) UploadFileV1(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	bucketName := c.PostForm("bucket_name")
	if bucketName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bucket name is required"})
		return
	}

	// Get optional filename from form
	filename := c.PostForm("filename")

	// Validate file size (e.g., max 100MB)
	if file.Size > 100<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 100MB limit"})
		return
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".svg":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".webp": true,
	}
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	objectName, err := eh.service.UploadFile(c, file, bucketName, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file", "error_message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clientDTO.FileUploadV1ResponseDTO{
		FileName: objectName,
		Message:  "File uploaded successfully",
	})
}

// DeleteFileV1 handles file deletion requests
// @Summary Delete a file
// @Description Delete a file from MinIO storage
// @Accept json
// @Produce json
// @Param request body clientDTO.FileDeleteV1RequestDTO true "Delete file request"
// @Success 200 {object} clientDTO.FileDeleteV1ResponseDTO
// @Router /v1/minio/delete [post]
// @Tags minio
func (eh *minioHandler) DeleteFileV1(c *gin.Context) {
	var payload clientDTO.FileDeleteV1RequestDTO
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	err := eh.service.DeleteFile(c, payload.FileName, payload.BucketName)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		clientDTO.FileDeleteV1ResponseDTO{
			Message: "File deleted successfully",
		},
	)
}

// GetFileContentV1 handles file content retrieval requests
// @Summary Get file content
// @Description Get the content of a file from MinIO storage
// @Accept json
// @Produce json
// @Param request body clientDTO.FileContentV1RequestDTO true "Get file content request"
// @Success 200 {object} clientDTO.FileContentV1ResponseDTO
// @Router /v1/minio/get-file-content [post]
// @Tags minio
func (eh *minioHandler) GetFileContentV1(c *gin.Context) {
	var payload clientDTO.FileContentV1RequestDTO
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	reader, err := eh.service.GetFile(c, payload.FileName, payload.BucketName)
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		clientDTO.FileContentV1ResponseDTO{
			Content: string(content),
		},
	)
}

// GetFileURLV1 handles file URL generation requests
// @Summary Get file URL
// @Description Get a presigned URL for a file from MinIO storage
// @Accept json
// @Produce json
// @Param request body clientDTO.FileURLV1RequestDTO true "Get file URL request"
// @Success 200 {object} clientDTO.FileURLV1ResponseDTO
// @Router /v1/minio/get-file-url [post]
// @Tags minio
func (eh *minioHandler) GetFileURLV1(c *gin.Context) {
	var payload clientDTO.FileURLV1RequestDTO
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	url, err := eh.service.GetFileURL(c, payload.FileName, payload.BucketName)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		clientDTO.FileURLV1ResponseDTO{
			URL: url,
		},
	)
}

// GetFile handles direct file download requests
// @Summary Download a file
// @Description Download a file directly from MinIO storage
// @Accept json
// @Produce octet-stream
// @Param filename path string true "Filename to download"
// @Param bucket_name query string true "Bucket name"
// @Success 200 {file} binary "File content"
// @Failure 400 {object} map[string]string "Bad Request - Missing filename or bucket name"
// @Failure 404 {object} map[string]string "Not Found - File not found"
// @Router /v1/minio/file/{filename} [get]
// @Tags minio
func (eh *minioHandler) GetFile(c *gin.Context) {
	objectName := c.Param("filename")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	bucketName := c.Query("bucket_name")
	if bucketName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bucket name is required"})
		return
	}

	reader, err := eh.service.GetFile(c, objectName, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	// Set appropriate headers based on file type
	ext := filepath.Ext(objectName)
	contentType := "application/octet-stream"
	isImage := false

	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
		isImage = true
	case ".png":
		contentType = "image/png"
		isImage = true
	case ".svg":
		contentType = "image/svg+xml"
		isImage = true
	case ".webp":
		contentType = "image/webp"
		isImage = true
	case ".pdf":
		contentType = "application/pdf"
	case ".doc":
		contentType = "application/msword"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	c.Header("Content-Type", contentType)

	// Only set Content-Disposition to attachment for non-image files
	if !isImage {
		c.Header("Content-Disposition", "attachment; filename="+objectName)
	}

	// Stream the file to the response
	io.Copy(c.Writer, reader)
}
