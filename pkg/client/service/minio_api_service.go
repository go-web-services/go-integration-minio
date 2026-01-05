package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/Lomank123/go-integration-minio/pkg/client/dto"
	platformUtils "github.com/Lomank123/go-web-platform/utils"
	"github.com/andybalholm/brotli"
)

// MultipartFormData represents the structure for multipart form data
type MultipartFormData struct {
	Files       map[string]*multipart.FileHeader
	FormFields  map[string]string
	ContentType string
}

// HTTPClient interface for better testability
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type MinioAPIService interface {
	UploadFileV1(context *gin.Context, file *multipart.FileHeader, bucketName string, fileName string) (dto.FileUploadV1ResponseDTO, error)
	DeleteFileV1(context *gin.Context, payload dto.FileDeleteV1RequestDTO) (dto.FileDeleteV1ResponseDTO, error)
	GetFileContentV1(context *gin.Context, payload dto.FileContentV1RequestDTO) (dto.FileContentV1ResponseDTO, error)
	GetFileURLV1(context *gin.Context, payload dto.FileURLV1RequestDTO) (dto.FileURLV1ResponseDTO, error)
	GetFileV1(context *gin.Context, payload dto.FileGetV1RequestDTO) (io.ReadCloser, error)
}

type minioAPIService struct {
	apiURL     string
	httpClient HTTPClient
}

func NewMinioAPIService(host string) MinioAPIService {
	return &minioAPIService{
		apiURL:     fmt.Sprintf("%s/api/v1", host),
		httpClient: &http.Client{},
	}
}

func (s *minioAPIService) UploadFileV1(context *gin.Context, file *multipart.FileHeader, bucketName string, fileName string) (dto.FileUploadV1ResponseDTO, error) {
	url := fmt.Sprintf("%s/minio/upload", s.apiURL)

	formData := &MultipartFormData{
		Files: map[string]*multipart.FileHeader{
			"file": file,
		},
		FormFields: map[string]string{
			"bucket_name": bucketName,
		},
	}

	// Add filename to form fields if provided
	if fileName != "" {
		formData.FormFields["filename"] = fileName
	}

	var responseBody dto.FileUploadV1ResponseDTO
	err := SendMultipartRequest(s.httpClient, "POST", url, formData, context, &responseBody)
	return responseBody, err
}

func (s *minioAPIService) DeleteFileV1(context *gin.Context, payload dto.FileDeleteV1RequestDTO) (dto.FileDeleteV1ResponseDTO, error) {
	url := fmt.Sprintf("%s/minio/delete", s.apiURL)
	var responseBody dto.FileDeleteV1ResponseDTO

	err := platformUtils.SendRequest("POST", url, payload, &responseBody, context)
	return responseBody, err
}

func (s *minioAPIService) GetFileContentV1(context *gin.Context, payload dto.FileContentV1RequestDTO) (dto.FileContentV1ResponseDTO, error) {
	url := fmt.Sprintf("%s/minio/get-file-content", s.apiURL)
	var responseBody dto.FileContentV1ResponseDTO

	err := platformUtils.SendRequest("POST", url, payload, &responseBody, context)
	return responseBody, err
}

func (s *minioAPIService) GetFileURLV1(context *gin.Context, payload dto.FileURLV1RequestDTO) (dto.FileURLV1ResponseDTO, error) {
	url := fmt.Sprintf("%s/minio/get-file-url", s.apiURL)
	var responseBody dto.FileURLV1ResponseDTO

	err := platformUtils.SendRequest("POST", url, payload, &responseBody, context)
	return responseBody, err
}

func (s *minioAPIService) GetFileV1(context *gin.Context, payload dto.FileGetV1RequestDTO) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/minio/file/%s?bucket_name=%s", s.apiURL, payload.FileName, payload.BucketName)
	context.Request.Header.Del("Accept-Encoding")

	// Create request
	req, err := http.NewRequestWithContext(context.Request.Context(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers from the original context
	for key, values := range context.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// If request was not successful
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		var errPayload map[string]any
		err = json.NewDecoder(resp.Body).Decode(&errPayload)
		if err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}

		return nil, fmt.Errorf("request failed with status %d: %v", resp.StatusCode, errPayload)
	}

	// Return the response body as a ReadCloser
	return resp.Body, nil
}

// SendMultipartRequest is a generic function to send multipart form data requests
func SendMultipartRequest(client HTTPClient, method, url string, formData *MultipartFormData, context *gin.Context, response any) error {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files to the form
	for fieldName, file := range formData.Files {
		part, err := writer.CreateFormFile(fieldName, filepath.Base(file.Filename))
		if err != nil {
			return fmt.Errorf("failed to create form file: %w", err)
		}

		fileContent, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fileContent.Close()

		if _, err := io.Copy(part, fileContent); err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	// Add form fields
	for key, value := range formData.FormFields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("failed to write form field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(context.Request.Context(), method, url, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Copy headers from the original context
	for key, values := range context.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// If request was not successful
	if resp.StatusCode >= http.StatusBadRequest {
		var errPayload map[string]any
		err = json.NewDecoder(resp.Body).Decode(&errPayload)
		if err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}

		return fmt.Errorf("request failed with status %d: %v", resp.StatusCode, errPayload)
	}

	// Handle Brotli compression
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "br" {
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	}

	// Decode the response
	err = json.NewDecoder(reader).Decode(response)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
