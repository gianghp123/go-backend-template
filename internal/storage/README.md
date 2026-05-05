# internal/storage/

**Purpose:** File/object storage abstraction — upload, download, delete.

**What it can contain:**
- `IStorageProvider` interface
- Provider implementations (S3, GCS, MinIO, local filesystem)
- `Init()` factory that returns the chosen implementation
- Presigned URL generation, multipart upload, CDN integration

**Pattern:**

```go
type IStorageProvider interface {
    Upload(ctx context.Context, path string, data io.Reader) (string, error)
    Delete(ctx context.Context, path string) error
    GetURL(ctx context.Context, path string) (string, error)
}
```

**Example: AWS S3 implementation using `github.com/aws/aws-sdk-go-v2`:**

```go
type S3Storage struct {
    client *s3.Client
    bucket string
}

func (s *S3Storage) Upload(ctx context.Context, path string, data io.Reader) (string, error) {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
        Body:   data,
    })
    return path, err
}
```

**Recommendations:**
- `aws/aws-sdk-go-v2/service/s3` — AWS S3
- `minio/minio-go` — MinIO (self-hosted S3-compatible)
- `googleapis/google-cloud-go/storage` — Google Cloud Storage
- For local dev: filesystem implementation using `os` package

Extend with presigned URLs, multipart uploads, or CDN purging as needed.
