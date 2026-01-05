package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lomank123/go-integration-minio/config"

	"github.com/Lomank123/go-web-platform/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, bucketName string, fileName string) (string, error)
	DeleteFile(ctx context.Context, fileName string, bucketName string) error
	GetFile(ctx context.Context, fileName string, bucketName string) (io.ReadCloser, error)
	GetFileURL(ctx context.Context, fileName string, bucketName string) (string, error)
}

type minioService struct {
	log    logger.Logger
	client *minio.Client
}

func NewMinioService(log logger.Logger) (MinioService, error) {
	// Get environment variables
	endpoint := config.Cfg.App.MinioEndpoint

	accessKeyID := config.Cfg.App.MinioAccessKey

	secretAccessKey := config.Cfg.App.MinioSecretKey

	useSSL := config.Cfg.App.MinioUseSSL

	log.Info("Initializing MinIO client",
		"endpoint", endpoint,
		"accessKeyID", accessKeyID,
		"useSSL", useSSL)

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Error("Failed to create MinIO client", "error", err)
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &minioService{
		log:    log,
		client: minioClient,
	}, nil
}

func (s *minioService) UploadFile(ctx context.Context, file *multipart.FileHeader, bucketName string, fileName string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate timestamp prefix
	timestamp := time.Now().Format("20060102150405")

	// Get the extension from the original filename
	originalFilename := filepath.Base(file.Filename)
	extension := filepath.Ext(originalFilename)

	// Determine the base name (without extension)
	var baseName string
	if fileName != "" {
		// If custom filename is provided, use it without extension
		baseName = fileName
	} else {
		// If no custom filename, use original filename without extension
		baseName = strings.TrimSuffix(originalFilename, extension)
	}

	// Always include timestamp in the object name and use the original extension
	objectName := timestamp + "-" + baseName + extension

	// Upload the file
	_, err = s.client.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	return objectName, nil
}

func (s *minioService) DeleteFile(ctx context.Context, fileName string, bucketName string) error {
	err := s.client.RemoveObject(ctx, bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *minioService) GetFile(ctx context.Context, fileName string, bucketName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *minioService) GetFileURL(ctx context.Context, fileName string, bucketName string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucketName, fileName, time.Hour*24, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
