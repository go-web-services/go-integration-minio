package http

import (
	"github.com/go-web-services/go-integration-minio/internal/service"
	http "github.com/go-web-services/go-integration-minio/internal/transport/http/handler"

	"github.com/go-web-services/go-web-platform/logger"

	"github.com/gin-gonic/gin"
)

// SetupRouter setups gin router with all routes
func SetupRouter(router *gin.Engine, minioService service.MinioService, log logger.Logger) *gin.Engine {

	minioHandler := http.NewMinioHandler(minioService, log)
	v1 := router.Group("/api/v1")
	{
		v1.POST("/minio/upload", minioHandler.UploadFileV1)
		v1.POST("/minio/delete", minioHandler.DeleteFileV1)
		v1.POST("/minio/get-file-content", minioHandler.GetFileContentV1)
		v1.POST("/minio/get-file-url", minioHandler.GetFileURLV1)
		v1.GET("/minio/file/:filename", minioHandler.GetFile)
	}

	return router
}
