Visit main page: [https://github.com/go-web-services](https://github.com/go-web-services)

# Go Web Services - go-integration-minio

`github.com/go-web-services/go-integration-minio`

HTTP integration service that fronts [MinIO](https://min.io/) (S3-compatible) object storage. It provides a stable, JSON-first interface over the raw S3 API — upload, delete, read content, presigned URLs, and direct download. Intended to sit behind your API gateway or to be called directly by internal services.

---

## Responsibilities

- Accept multipart file uploads; stored objects are prefixed with a timestamp. Base filename is optional.
- Delete objects by bucket and key.
- Return file body as a JSON string (`content` field) for programmatic access.
- Stream binary downloads with appropriate `Content-Type`; images are inlined, other types served as attachments.
- Issue time-limited presigned GET URLs for client-side access.

---

## Configuration

| Variable | Purpose | Default |
|----------|---------|---------|
| `APP_PORT` | HTTP listen port | — |
| `APP_ENV` | Environment (`dev` / `prod`) | — |
| `MINIO_ENDPOINT` | MinIO host and port | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` |
| `MINIO_USE_SSL` | Enable TLS for MinIO connection | `false` |

---

## Run locally

```bash
git clone git@github.com:go-web-services/go-integration-minio.git
cd go-integration-minio
cp .env.sample .env
# Set MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY
docker compose up -d
```

MinIO itself must be running and reachable at `MINIO_ENDPOINT`. The MinIO console is typically available at port `9001`.

---

## Docker

- **Dev** (hot reload via `debug/Dockerfile`):
  ```bash
  docker compose up -d
  ```
- **Prod**:
  ```bash
  docker compose -f docker-compose-prod.yml up --build
  ```

---

## API surface

Swagger UI is available at `/swagger`.

### Delete file

`POST /api/v1/minio/delete`

```json
{ "fileName": "20260101120000-report.pdf", "bucketName": "my-bucket" }
```

Response: `{ "message": "File deleted successfully" }`

### Get file content

`POST /api/v1/minio/content`

```json
{ "fileName": "20260101120000-report.pdf", "bucketName": "my-bucket" }
```

Response: `{ "content": "<file bytes as string>" }`

### Presigned URL

`POST /api/v1/minio/url`

```json
{ "fileName": "20260101120000-photo.jpg", "bucketName": "my-bucket" }
```

Response: `{ "url": "https://minio.example/bucket/object?X-Amz-Algorithm=..." }`

### Upload file

`POST /api/v1/minio/upload` — multipart form

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `file` | file | Yes | File to upload |
| `bucket_name` | string | Yes | Target bucket |
| `filename` | string | No | Override base name (without forcing extension) |

Response: `{ "fileName": "20260101120000-report.pdf", "message": "File uploaded successfully" }`

### Direct download

`GET /api/v1/minio/file/{filename}?bucket_name=my-bucket`

Returns the object body. Images are inlined (`Content-Disposition: inline`); other types use `Content-Disposition: attachment`.

---

## Client module (`pkg/client`)

Other services import `github.com/go-web-services/go-integration-minio/pkg/client` for typed DTOs and the `MinioAPIService` HTTP client.

```go
import (
    clientapi "github.com/go-web-services/go-integration-minio/pkg/client/service"
    "github.com/go-web-services/go-integration-minio/pkg/client/dto"
)

svc := clientapi.NewMinioAPIService("http://localhost:8030")
result, err := svc.UploadV1(ctx, fileBytes, "my-bucket", "report.pdf")
```

For local development in a consuming service:

```bash
go mod edit -replace github.com/go-web-services/go-integration-minio=/path/to/go-integration-minio
```

---

## Private dependencies

```bash
export GOPRIVATE='github.com/go-web-services/*'
```

This service depends on `go-web-platform`.

---

## Author

[Lomank](https://lomank.com)
