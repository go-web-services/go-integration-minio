# go-integration-minio

HTTP integration service that fronts [MinIO](https://min.io/) object storage: upload, delete, read content, presigned URLs, and direct download. It is intended to sit behind your API gateway or to be called by other internal services that need a stable, JSON-first interface over S3-compatible storage.

## Responsibilities

- Accept multipart uploads with optional custom base filename; stored objects are prefixed with a timestamp.
- Delete objects by bucket and key.
- Return file body as JSON (string) or stream binary with sensible `Content-Type` for downloads.
- Issue time-limited presigned GET URLs.

Configuration is environment-driven (see `.env.sample`): MinIO endpoint, credentials, TLS, app port, and environment name for logging.

## API examples (JSON)

**Delete file**

```json
{
  "fileName": "20260101120000-report.pdf",
  "bucketName": "my-bucket"
}
```

**Delete response**

```json
{
  "message": "File deleted successfully"
}
```

**Get file content (request)**

```json
{
  "fileName": "20260101120000-report.pdf",
  "bucketName": "my-bucket"
}
```

**Get file content (response)**

```json
{
  "content": "<file bytes as string>"
}
```

**Presigned URL (request)**

```json
{
  "fileName": "20260101120000-photo.jpg",
  "bucketName": "my-bucket"
}
```

**Presigned URL (response)**

```json
{
  "url": "https://minio.example/bucket/object?X-Amz-Algorithm=..."
}
```

**Upload (multipart)** — form fields: `file` (file), `bucket_name` (string, required), `filename` (string, optional base name without forcing extension).

**Upload (response)**

```json
{
  "fileName": "20260101120000-report.pdf",
  "message": "File uploaded successfully"
}
```

**Download** — `GET /api/v1/minio/file/{filename}?bucket_name=my-bucket` returns the object body; images are typically inlined, other types use `Content-Disposition: attachment`.

## Author

[Lomank](https://lomank.com)
