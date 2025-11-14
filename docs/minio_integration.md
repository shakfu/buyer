# MinIO Integration for Document Storage

## Overview

This document describes how to integrate MinIO, an open-source high-performance object storage system, with the buyer application's document management system. MinIO provides S3-compatible storage that can be deployed on-premises or in the cloud, offering a scalable and reliable solution for document storage.

## What is MinIO?

MinIO is a high-performance, S3-compatible object storage system released under GNU AGPLv3. It is designed for large-scale AI/ML, data lake, and database workloads but works equally well for storing documents of any size.

**Key Features:**
- **S3 Compatible**: Uses the same API as Amazon S3, making it a drop-in replacement
- **High Performance**: Can handle millions of operations per second
- **Kubernetes Native**: First-class support for containerized deployments
- **Erasure Coding**: Built-in data protection and high availability
- **Encryption**: Server-side and client-side encryption support
- **Versioning**: Object versioning for compliance and data protection
- **Open Source**: Free to use and modify under AGPLv3

## Why MinIO for Document Storage?

### Current Limitations

The current `Document` model stores file paths as strings in the database:

```go
type Document struct {
    ID          uint      `gorm:"primaryKey"`
    EntityType  string    `gorm:"index:idx_document_entity"`
    EntityID    uint      `gorm:"index:idx_document_entity"`
    FileName    string    `gorm:"not null"`
    FilePath    string    `gorm:"not null"` // Local filesystem path
    FileType    string
    FileSize    int64
    Description string
    UploadedBy  string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Limitations:**
- Files stored on local filesystem are not easily scalable
- No built-in redundancy or backup
- Difficult to share across multiple application instances
- No versioning or access control
- Limited by server disk space
- No CDN integration for fast global access

### Benefits of MinIO Integration

1. **Scalability**: Store unlimited documents across multiple drives/servers
2. **Reliability**: Built-in erasure coding provides data protection
3. **High Availability**: Deploy in distributed mode for zero-downtime
4. **Security**: Encryption at rest and in transit, fine-grained access control
5. **Versioning**: Track document changes over time
6. **Cost Effective**: Run on commodity hardware or cloud storage
7. **S3 Compatibility**: Easy migration to AWS S3 if needed
8. **Multi-tenancy**: Bucket policies for isolating different entity types
9. **Lifecycle Management**: Automatic archival and deletion policies
10. **Audit Trail**: Track all object operations for compliance

## Architecture

### Storage Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Buyer Application                        │
│  ┌──────────────┐        ┌─────────────────────────────┐   │
│  │  Document    │───────▶│  DocumentStorageService     │   │
│  │  Service     │        │  (MinIO Client)             │   │
│  └──────────────┘        └────────────┬────────────────┘   │
│                                        │                     │
└────────────────────────────────────────┼─────────────────────┘
                                         │ S3 API
                                         ▼
                        ┌────────────────────────────────┐
                        │         MinIO Server           │
                        │  ┌──────────────────────────┐  │
                        │  │   Bucket: vendor-docs    │  │
                        │  │   Bucket: product-docs   │  │
                        │  │   Bucket: quote-docs     │  │
                        │  │   Bucket: po-docs        │  │
                        │  │   Bucket: project-docs   │  │
                        │  └──────────────────────────┘  │
                        └────────────────────────────────┘
```

### Bucket Organization

Each entity type gets its own bucket for better organization and access control:

- `vendor-docs` - Vendor-related documents
- `product-docs` - Product specifications, datasheets
- `quote-docs` - Quote PDFs, vendor quotes
- `po-docs` - Purchase order documents, invoices
- `requisition-docs` - Requisition attachments
- `project-docs` - Project documentation
- `brand-docs` - Brand information, catalogs

### Object Key Structure

Use a hierarchical key structure for organization:

```
{entity-type}/{entity-id}/{year}/{month}/{uuid}-{filename}

Examples:
vendor/123/2025/01/550e8400-e29b-41d4-a716-446655440000-contract.pdf
product/456/2025/01/6ba7b810-9dad-11d1-80b4-00c04fd430c8-datasheet.pdf
quote/789/2025/01/3f2504e0-4f89-11d3-9a0c-0305e82c3301-vendor-quote.pdf
```

**Benefits:**
- Chronological organization for easier management
- UUID prefix prevents filename collisions
- Easy to browse and filter by date
- Supports S3 lifecycle policies by prefix

## Installation and Setup

### 1. Install MinIO Server

**Using Docker (Recommended for Development):**

```bash
# Run MinIO in a Docker container
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  -v /data/minio:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

**Using Docker Compose:**

Create `docker-compose.minio.yml`:

```yaml
version: '3.8'

services:
  minio:
    image: quay.io/minio/minio:latest
    container_name: buyer-minio
    ports:
      - "9000:9000"   # API
      - "9001:9001"   # Console
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:-minioadmin}
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  minio_data:
    driver: local
```

Start with:
```bash
docker-compose -f docker-compose.minio.yml up -d
```

**Native Installation:**

```bash
# Linux
wget https://dl.min.io/server/minio/release/linux-amd64/minio
chmod +x minio
./minio server /data --console-address ":9001"

# macOS
brew install minio/stable/minio
minio server /data --console-address ":9001"
```

### 2. Access MinIO Console

Open http://localhost:9001 in your browser and login with:
- Username: `minioadmin`
- Password: `minioadmin`

### 3. Install MinIO Client (mc)

```bash
# Linux
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/

# macOS
brew install minio/stable/mc

# Configure mc
mc alias set local http://localhost:9000 minioadmin minioadmin
```

### 4. Create Buckets

```bash
# Create buckets for each entity type
mc mb local/vendor-docs
mc mb local/product-docs
mc mb local/quote-docs
mc mb local/po-docs
mc mb local/requisition-docs
mc mb local/project-docs
mc mb local/brand-docs

# Set bucket policies (optional - for public read access)
mc anonymous set download local/product-docs

# Enable versioning
mc version enable local/vendor-docs
mc version enable local/quote-docs
mc version enable local/po-docs
```

## Configuration

### Environment Variables

Add to `.env`:

```bash
# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_REGION=us-east-1

# Storage Configuration
DOCUMENT_STORAGE_TYPE=minio  # Options: local, minio, s3
DOCUMENT_STORAGE_LOCAL_PATH=/var/lib/buyer/documents
DOCUMENT_URL_EXPIRY=3600  # Presigned URL expiry in seconds (1 hour)
```

### Configuration Struct

Update `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields ...

    // MinIO Configuration
    MinioEndpoint   string
    MinioAccessKey  string
    MinioSecretKey  string
    MinioUseSSL     bool
    MinioRegion     string

    // Storage Configuration
    StorageType         string // "local" or "minio"
    StorageLocalPath    string
    DocumentURLExpiry   int    // seconds
}

func Load() (*Config, error) {
    // ... existing code ...

    cfg.MinioEndpoint = getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000")
    cfg.MinioAccessKey = getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin")
    cfg.MinioSecretKey = getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin")
    cfg.MinioUseSSL = getEnvOrDefault("MINIO_USE_SSL", "false") == "true"
    cfg.MinioRegion = getEnvOrDefault("MINIO_REGION", "us-east-1")
    cfg.StorageType = getEnvOrDefault("DOCUMENT_STORAGE_TYPE", "local")
    cfg.StorageLocalPath = getEnvOrDefault("DOCUMENT_STORAGE_LOCAL_PATH", "/var/lib/buyer/documents")
    cfg.DocumentURLExpiry = 3600
    if expiry := getEnvOrDefault("DOCUMENT_URL_EXPIRY", "3600"); expiry != "" {
        if val, err := strconv.Atoi(expiry); err == nil {
            cfg.DocumentURLExpiry = val
        }
    }

    return cfg, nil
}
```

## MinIO Go SDK Deep Dive

### Overview

The MinIO Go Client SDK (`minio-go`) provides a comprehensive, idiomatic Go interface for interacting with MinIO and any S3-compatible object storage. The SDK is actively maintained, feature-complete, and used in production by thousands of organizations.

**Repository**: https://github.com/minio/minio-go
**Documentation**: https://pkg.go.dev/github.com/minio/minio-go/v7
**License**: Apache License 2.0

### Installation

```bash
# Install the latest version
go get github.com/minio/minio-go/v7

# Import in your code
import (
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)
```

### Client Initialization

#### Basic Initialization

```go
package main

import (
    "log"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
    endpoint := "localhost:9000"
    accessKeyID := "minioadmin"
    secretAccessKey := "minioadmin"
    useSSL := false

    // Initialize MinIO client
    minioClient, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        log.Fatalln(err)
    }

    log.Printf("%#v\n", minioClient)
}
```

#### Credential Options

The SDK supports multiple credential providers:

```go
// 1. Static credentials (as shown above)
creds := credentials.NewStaticV4(accessKey, secretKey, "")

// 2. IAM credentials (for AWS EC2/ECS)
creds := credentials.NewIAM("")

// 3. Environment variables
// Reads from AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
creds := credentials.NewEnvAWS()

// 4. MinIO environment variables
// Reads from MINIO_ACCESS_KEY and MINIO_SECRET_KEY
creds := credentials.NewEnvMinio()

// 5. Chained credentials (try multiple providers)
creds := credentials.NewChainCredentials([]credentials.Provider{
    &credentials.EnvAWS{},
    &credentials.EnvMinio{},
    &credentials.IAM{},
})
```

#### Client with Custom Transport

```go
import (
    "crypto/tls"
    "net/http"
    "time"
)

// Custom HTTP transport for production
transport := &http.Transport{
    MaxIdleConns:       100,
    IdleConnTimeout:    90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    TLSClientConfig: &tls.Config{
        // Skip certificate verification (not recommended for production)
        InsecureSkipVerify: false,
    },
}

minioClient, err := minio.New(endpoint, &minio.Options{
    Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
    Secure:    true,
    Transport: transport,
    Region:    "us-east-1",
})
```

### Core API Operations

#### Bucket Operations

```go
import (
    "context"
    "log"
)

ctx := context.Background()

// 1. Check if bucket exists
exists, err := minioClient.BucketExists(ctx, "vendor-docs")
if err != nil {
    log.Fatalln(err)
}
if !exists {
    log.Println("Bucket does not exist")
}

// 2. Create a bucket
err = minioClient.MakeBucket(ctx, "vendor-docs", minio.MakeBucketOptions{
    Region: "us-east-1",
})
if err != nil {
    log.Fatalln(err)
}

// 3. List all buckets
buckets, err := minioClient.ListBuckets(ctx)
if err != nil {
    log.Fatalln(err)
}
for _, bucket := range buckets {
    log.Println(bucket.Name, bucket.CreationDate)
}

// 4. Remove a bucket (must be empty)
err = minioClient.RemoveBucket(ctx, "vendor-docs")
if err != nil {
    log.Fatalln(err)
}
```

#### Object Upload Operations

```go
import (
    "bytes"
    "os"
)

// 1. Upload from memory (PutObject)
data := []byte("Hello MinIO!")
reader := bytes.NewReader(data)

info, err := minioClient.PutObject(ctx, "vendor-docs", "hello.txt", reader, int64(len(data)), minio.PutObjectOptions{
    ContentType: "text/plain",
    UserMetadata: map[string]string{
        "uploaded-by": "buyer-app",
        "entity-type": "vendor",
    },
})
if err != nil {
    log.Fatalln(err)
}
log.Printf("Successfully uploaded %s of size %d\n", info.Key, info.Size)

// 2. Upload from file (FPutObject)
info, err = minioClient.FPutObject(ctx, "vendor-docs", "contract.pdf", "/path/to/contract.pdf", minio.PutObjectOptions{
    ContentType: "application/pdf",
})
if err != nil {
    log.Fatalln(err)
}

// 3. Upload with progress tracking
progress := make(chan int64)
go func() {
    for p := range progress {
        log.Printf("Uploaded: %d bytes\n", p)
    }
}()

info, err = minioClient.PutObject(ctx, "vendor-docs", "large-file.zip", reader, fileSize, minio.PutObjectOptions{
    ContentType: "application/zip",
    Progress:    progress,
})

// 4. Upload with server-side encryption
info, err = minioClient.PutObject(ctx, "vendor-docs", "secure.txt", reader, size, minio.PutObjectOptions{
    ServerSideEncryption: encrypt.NewSSE(),
})
```

#### Object Download Operations

```go
// 1. Download to memory (GetObject)
object, err := minioClient.GetObject(ctx, "vendor-docs", "hello.txt", minio.GetObjectOptions{})
if err != nil {
    log.Fatalln(err)
}
defer object.Close()

// Read into buffer
buf := new(bytes.Buffer)
_, err = buf.ReadFrom(object)
if err != nil {
    log.Fatalln(err)
}
log.Println(buf.String())

// 2. Download to file (FGetObject)
err = minioClient.FGetObject(ctx, "vendor-docs", "contract.pdf", "/tmp/contract.pdf", minio.GetObjectOptions{})
if err != nil {
    log.Fatalln(err)
}

// 3. Download with byte range (partial download)
options := minio.GetObjectOptions{}
options.SetRange(0, 1023)  // First 1KB
object, err = minioClient.GetObject(ctx, "vendor-docs", "large-file.zip", options)
if err != nil {
    log.Fatalln(err)
}

// 4. Stream download with progress
object, err = minioClient.GetObject(ctx, "vendor-docs", "video.mp4", minio.GetObjectOptions{})
if err != nil {
    log.Fatalln(err)
}
defer object.Close()

file, err := os.Create("/tmp/video.mp4")
if err != nil {
    log.Fatalln(err)
}
defer file.Close()

n, err := io.Copy(file, object)
log.Printf("Downloaded %d bytes\n", n)
```

#### Object Information and Metadata

```go
// 1. Get object statistics
stat, err := minioClient.StatObject(ctx, "vendor-docs", "hello.txt", minio.StatObjectOptions{})
if err != nil {
    log.Fatalln(err)
}
log.Printf("Object: %s, Size: %d, ContentType: %s, LastModified: %s\n",
    stat.Key, stat.Size, stat.ContentType, stat.LastModified)

// Access user metadata
for key, val := range stat.UserMetadata {
    log.Printf("Metadata: %s = %s\n", key, val)
}

// 2. Check if object exists
_, err = minioClient.StatObject(ctx, "vendor-docs", "hello.txt", minio.StatObjectOptions{})
if err != nil {
    errResponse := minio.ToErrorResponse(err)
    if errResponse.Code == "NoSuchKey" {
        log.Println("Object does not exist")
    }
}
```

#### Object Deletion

```go
// 1. Delete single object
err = minioClient.RemoveObject(ctx, "vendor-docs", "hello.txt", minio.RemoveObjectOptions{})
if err != nil {
    log.Fatalln(err)
}

// 2. Delete multiple objects
objectsCh := make(chan minio.ObjectInfo)

// Send objects to delete
go func() {
    defer close(objectsCh)
    objectsCh <- minio.ObjectInfo{Key: "file1.txt"}
    objectsCh <- minio.ObjectInfo{Key: "file2.txt"}
    objectsCh <- minio.ObjectInfo{Key: "file3.txt"}
}()

// Delete objects
for rErr := range minioClient.RemoveObjects(ctx, "vendor-docs", objectsCh, minio.RemoveObjectsOptions{}) {
    log.Printf("Error deleting %s: %s\n", rErr.ObjectName, rErr.Err)
}

// 3. Delete with versioning
opts := minio.RemoveObjectOptions{
    VersionID: "version-id-here",
}
err = minioClient.RemoveObject(ctx, "vendor-docs", "hello.txt", opts)
```

#### List Objects

```go
// 1. List all objects in a bucket
for object := range minioClient.ListObjects(ctx, "vendor-docs", minio.ListObjectsOptions{}) {
    if object.Err != nil {
        log.Println(object.Err)
        return
    }
    log.Println(object.Key, object.Size, object.LastModified)
}

// 2. List with prefix (like a directory)
opts := minio.ListObjectsOptions{
    Prefix:    "vendor/123/",
    Recursive: true,
}
for object := range minioClient.ListObjects(ctx, "vendor-docs", opts) {
    if object.Err != nil {
        log.Println(object.Err)
        return
    }
    log.Println(object.Key)
}

// 3. List with pagination
opts = minio.ListObjectsOptions{
    Prefix:    "vendor/",
    MaxKeys:   100,
    Recursive: false,  // Don't recurse into subdirectories
}
for object := range minioClient.ListObjects(ctx, "vendor-docs", opts) {
    log.Println(object.Key)
}
```

#### Presigned URLs

```go
import "net/url"

// 1. Generate presigned GET URL (for downloads)
presignedURL, err := minioClient.PresignedGetObject(ctx, "vendor-docs", "contract.pdf", time.Hour, nil)
if err != nil {
    log.Fatalln(err)
}
log.Println("Download URL:", presignedURL)

// 2. Generate presigned PUT URL (for uploads)
presignedURL, err = minioClient.PresignedPutObject(ctx, "vendor-docs", "upload.pdf", time.Hour*24)
if err != nil {
    log.Fatalln(err)
}
log.Println("Upload URL:", presignedURL)

// 3. Generate presigned URL with custom query parameters
reqParams := make(url.Values)
reqParams.Set("response-content-disposition", "attachment; filename=\"download.pdf\"")
reqParams.Set("response-content-type", "application/pdf")

presignedURL, err = minioClient.PresignedGetObject(ctx, "vendor-docs", "contract.pdf", time.Hour, reqParams)
if err != nil {
    log.Fatalln(err)
}

// 4. Generate POST policy for browser uploads
policy := minio.NewPostPolicy()
policy.SetBucket("vendor-docs")
policy.SetKey("upload/")
policy.SetExpires(time.Now().UTC().Add(24 * time.Hour))
policy.SetContentType("image/jpeg")
policy.SetContentLengthRange(1024, 1024*1024*10)  // 1KB to 10MB

url, formData, err := minioClient.PresignedPostPolicy(ctx, policy)
if err != nil {
    log.Fatalln(err)
}
log.Println("POST URL:", url)
for k, v := range formData {
    log.Printf("Form field: %s = %s\n", k, v)
}
```

### Advanced Features

#### Multipart Upload

For large files, MinIO automatically uses multipart uploads:

```go
// Automatic multipart upload (SDK handles it internally)
info, err := minioClient.FPutObject(ctx, "vendor-docs", "large-video.mp4", "/path/to/large-video.mp4", minio.PutObjectOptions{
    ContentType: "video/mp4",
    PartSize:    10 * 1024 * 1024,  // 10MB parts
})
```

Manual multipart upload for more control:

```go
// 1. Initiate multipart upload
uploadID, err := minioClient.NewMultipartUpload(ctx, "vendor-docs", "manual-upload.bin", minio.PutObjectOptions{})
if err != nil {
    log.Fatalln(err)
}

// 2. Upload parts
var completeParts []minio.CompletePart
partNumber := 1
partSize := int64(5 * 1024 * 1024)  // 5MB

for {
    // Read part data
    partData := make([]byte, partSize)
    n, err := file.Read(partData)
    if n == 0 {
        break
    }

    // Upload part
    part, err := minioClient.PutObjectPart(ctx, "vendor-docs", "manual-upload.bin", uploadID, partNumber,
        bytes.NewReader(partData[:n]), int64(n), minio.PutObjectPartOptions{})
    if err != nil {
        log.Fatalln(err)
    }

    completeParts = append(completeParts, minio.CompletePart{
        PartNumber: partNumber,
        ETag:       part.ETag,
    })
    partNumber++
}

// 3. Complete multipart upload
_, err = minioClient.CompleteMultipartUpload(ctx, "vendor-docs", "manual-upload.bin", uploadID,
    completeParts, minio.PutObjectOptions{})
if err != nil {
    log.Fatalln(err)
}
```

#### Object Copying

```go
// 1. Copy object within same bucket
src := minio.CopySrcOptions{
    Bucket: "vendor-docs",
    Object: "original.pdf",
}
dst := minio.CopyDestOptions{
    Bucket: "vendor-docs",
    Object: "copy.pdf",
}

_, err = minioClient.CopyObject(ctx, dst, src)
if err != nil {
    log.Fatalln(err)
}

// 2. Copy with metadata changes
dst = minio.CopyDestOptions{
    Bucket:          "vendor-docs",
    Object:          "copy-with-metadata.pdf",
    ReplaceMetadata: true,
    UserMetadata: map[string]string{
        "copied-from": "original.pdf",
        "copy-date":   time.Now().Format(time.RFC3339),
    },
}

_, err = minioClient.CopyObject(ctx, dst, src)
```

#### Bucket Policies

```go
import "encoding/json"

// 1. Get bucket policy
policy, err := minioClient.GetBucketPolicy(ctx, "vendor-docs")
if err != nil {
    log.Fatalln(err)
}
log.Println(policy)

// 2. Set bucket policy (make bucket publicly readable)
policyJSON := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::vendor-docs/*"]
    }
  ]
}`

err = minioClient.SetBucketPolicy(ctx, "vendor-docs", policyJSON)
if err != nil {
    log.Fatalln(err)
}

// 3. Delete bucket policy
err = minioClient.DeleteBucketPolicy(ctx, "vendor-docs")
```

#### Object Versioning

```go
// 1. Enable versioning
err = minioClient.EnableVersioning(ctx, "vendor-docs")
if err != nil {
    log.Fatalln(err)
}

// 2. Check versioning status
status, err := minioClient.GetBucketVersioning(ctx, "vendor-docs")
if err != nil {
    log.Fatalln(err)
}
log.Println("Versioning status:", status.Status)

// 3. List object versions
opts := minio.ListObjectsOptions{
    Prefix:       "document.pdf",
    WithVersions: true,
}
for object := range minioClient.ListObjects(ctx, "vendor-docs", opts) {
    log.Printf("Version: %s, IsLatest: %v, LastModified: %s\n",
        object.VersionID, object.IsLatest, object.LastModified)
}

// 4. Download specific version
opts := minio.GetObjectOptions{
    VersionID: "version-id-here",
}
object, err := minioClient.GetObject(ctx, "vendor-docs", "document.pdf", opts)
```

### Error Handling

#### Checking Specific Errors

```go
import (
    "errors"
    "github.com/minio/minio-go/v7"
)

info, err := minioClient.StatObject(ctx, "vendor-docs", "file.txt", minio.StatObjectOptions{})
if err != nil {
    // Convert to MinIO error response
    errResponse := minio.ToErrorResponse(err)

    switch errResponse.Code {
    case "NoSuchKey":
        log.Println("Object does not exist")
    case "NoSuchBucket":
        log.Println("Bucket does not exist")
    case "AccessDenied":
        log.Println("Access denied")
    case "InvalidBucketName":
        log.Println("Invalid bucket name")
    default:
        log.Printf("Error: %s - %s\n", errResponse.Code, errResponse.Message)
    }
    return
}
```

#### Retry Logic

```go
import "time"

func uploadWithRetry(ctx context.Context, client *minio.Client, bucket, key string, reader io.Reader, size int64, maxRetries int) error {
    var err error
    for attempt := 0; attempt < maxRetries; attempt++ {
        _, err = client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{})
        if err == nil {
            return nil
        }

        // Check if error is retryable
        errResponse := minio.ToErrorResponse(err)
        if errResponse.StatusCode >= 500 {
            // Server error, retry
            waitTime := time.Duration(attempt+1) * 2 * time.Second
            log.Printf("Upload failed (attempt %d/%d), retrying in %s: %v\n",
                attempt+1, maxRetries, waitTime, err)
            time.Sleep(waitTime)
            continue
        }

        // Client error, don't retry
        return err
    }
    return fmt.Errorf("upload failed after %d attempts: %w", maxRetries, err)
}
```

### Best Practices

#### 1. Use Context for Cancellation

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// All operations respect context cancellation
_, err := minioClient.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{})
```

#### 2. Reuse Client Instances

```go
// Bad: Creating new client for each operation
func uploadFile(endpoint, key string) error {
    client, _ := minio.New(endpoint, &minio.Options{...})
    return client.FPutObject(ctx, "bucket", key, "/path", minio.PutObjectOptions{})
}

// Good: Reuse client instance
type DocumentService struct {
    minioClient *minio.Client
}

func (s *DocumentService) uploadFile(key string) error {
    return s.minioClient.FPutObject(ctx, "bucket", key, "/path", minio.PutObjectOptions{})
}
```

#### 3. Close Readers Properly

```go
object, err := minioClient.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
if err != nil {
    return err
}
defer object.Close()  // Always close

// Use the object
data, err := io.ReadAll(object)
```

#### 4. Set Appropriate Content Types

```go
import "mime"

contentType := mime.TypeByExtension(filepath.Ext(filename))
if contentType == "" {
    contentType = "application/octet-stream"
}

_, err := minioClient.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
    ContentType: contentType,
})
```

#### 5. Use User Metadata for Tracking

```go
_, err := minioClient.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
    UserMetadata: map[string]string{
        "entity-type":   "vendor",
        "entity-id":     "123",
        "uploaded-by":   "user@example.com",
        "upload-time":   time.Now().Format(time.RFC3339),
        "app-version":   "1.0.0",
    },
})
```

#### 6. Implement Health Checks

```go
func (s *DocumentService) HealthCheck(ctx context.Context) error {
    // Try to list buckets as health check
    _, err := s.minioClient.ListBuckets(ctx)
    if err != nil {
        return fmt.Errorf("MinIO health check failed: %w", err)
    }
    return nil
}
```

### Performance Optimization

#### Concurrent Uploads

```go
import "sync"

func uploadFilesParallel(files []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))

    // Limit concurrent uploads
    semaphore := make(chan struct{}, 10)

    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release

            err := uploadFile(f)
            if err != nil {
                errChan <- err
            }
        }(file)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}
```

#### Connection Pooling

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 100,
    IdleConnTimeout:     90 * time.Second,
}

minioClient, err := minio.New(endpoint, &minio.Options{
    Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
    Secure:    true,
    Transport: transport,
})
```

### Testing

#### Mock Client for Unit Tests

```go
// Define interface
type ObjectStorage interface {
    Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64) error
    Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
}

// Real implementation
type MinioStorage struct {
    client *minio.Client
}

// Mock implementation for tests
type MockStorage struct {
    storage map[string][]byte
}

func (m *MockStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
    data, _ := io.ReadAll(reader)
    m.storage[bucket+"/"+key] = data
    return nil
}

// Use in tests
func TestDocumentService(t *testing.T) {
    mockStorage := &MockStorage{storage: make(map[string][]byte)}
    service := NewDocumentService(mockStorage)
    // Test service...
}
```

### Additional Resources

- **Official Documentation**: https://min.io/docs/minio/linux/developers/go/minio-go.html
- **API Reference**: https://pkg.go.dev/github.com/minio/minio-go/v7
- **GitHub Repository**: https://github.com/minio/minio-go
- **Examples**: https://github.com/minio/minio-go/tree/master/examples
- **Community**: https://slack.min.io

## Implementation

### 1. Install Go SDK

```bash
go get github.com/minio/minio-go/v7
```

### 2. Create Storage Interface

Create `internal/storage/storage.go`:

```go
package storage

import (
    "context"
    "io"
    "time"
)

// StorageBackend defines the interface for document storage
type StorageBackend interface {
    // Upload uploads a file to storage
    Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error

    // Download downloads a file from storage
    Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)

    // Delete deletes a file from storage
    Delete(ctx context.Context, bucket, key string) error

    // GetPresignedURL generates a temporary URL for accessing a file
    GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)

    // Exists checks if a file exists
    Exists(ctx context.Context, bucket, key string) (bool, error)

    // GetMetadata retrieves file metadata
    GetMetadata(ctx context.Context, bucket, key string) (*FileMetadata, error)
}

type FileMetadata struct {
    Size         int64
    ContentType  string
    LastModified time.Time
    ETag         string
    Metadata     map[string]string
}

// BucketForEntityType returns the appropriate bucket name for an entity type
func BucketForEntityType(entityType string) string {
    switch entityType {
    case "vendor":
        return "vendor-docs"
    case "product":
        return "product-docs"
    case "quote":
        return "quote-docs"
    case "purchase_order":
        return "po-docs"
    case "requisition":
        return "requisition-docs"
    case "project":
        return "project-docs"
    case "brand":
        return "brand-docs"
    default:
        return "buyer-docs"
    }
}
```

### 3. Implement MinIO Backend

Create `internal/storage/minio.go`:

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioBackend struct {
    client *minio.Client
}

// NewMinioBackend creates a new MinIO storage backend
func NewMinioBackend(endpoint, accessKey, secretKey string, useSSL bool) (*MinioBackend, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create MinIO client: %w", err)
    }

    return &MinioBackend{client: client}, nil
}

// Upload uploads a file to MinIO
func (m *MinioBackend) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
    // Ensure bucket exists
    exists, err := m.client.BucketExists(ctx, bucket)
    if err != nil {
        return fmt.Errorf("failed to check bucket existence: %w", err)
    }
    if !exists {
        if err := m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return fmt.Errorf("failed to create bucket: %w", err)
        }
    }

    // Upload file
    opts := minio.PutObjectOptions{
        ContentType: contentType,
        UserMetadata: map[string]string{
            "uploaded-by": "buyer-app",
            "upload-time": time.Now().Format(time.RFC3339),
        },
    }

    _, err = m.client.PutObject(ctx, bucket, key, reader, size, opts)
    if err != nil {
        return fmt.Errorf("failed to upload object: %w", err)
    }

    return nil
}

// Download downloads a file from MinIO
func (m *MinioBackend) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
    obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get object: %w", err)
    }
    return obj, nil
}

// Delete deletes a file from MinIO
func (m *MinioBackend) Delete(ctx context.Context, bucket, key string) error {
    err := m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete object: %w", err)
    }
    return nil
}

// GetPresignedURL generates a temporary URL for accessing a file
func (m *MinioBackend) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
    url, err := m.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }
    return url.String(), nil
}

// Exists checks if a file exists in MinIO
func (m *MinioBackend) Exists(ctx context.Context, bucket, key string) (bool, error) {
    _, err := m.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
    if err != nil {
        errResponse := minio.ToErrorResponse(err)
        if errResponse.Code == "NoSuchKey" {
            return false, nil
        }
        return false, fmt.Errorf("failed to stat object: %w", err)
    }
    return true, nil
}

// GetMetadata retrieves file metadata from MinIO
func (m *MinioBackend) GetMetadata(ctx context.Context, bucket, key string) (*FileMetadata, error) {
    stat, err := m.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to stat object: %w", err)
    }

    return &FileMetadata{
        Size:         stat.Size,
        ContentType:  stat.ContentType,
        LastModified: stat.LastModified,
        ETag:         stat.ETag,
        Metadata:     stat.UserMetadata,
    }, nil
}
```

### 4. Implement Local Filesystem Backend

Create `internal/storage/local.go`:

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

type LocalBackend struct {
    basePath string
}

// NewLocalBackend creates a new local filesystem storage backend
func NewLocalBackend(basePath string) (*LocalBackend, error) {
    // Ensure base path exists
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, fmt.Errorf("failed to create base path: %w", err)
    }
    return &LocalBackend{basePath: basePath}, nil
}

// Upload uploads a file to local filesystem
func (l *LocalBackend) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
    filePath := filepath.Join(l.basePath, bucket, key)

    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Create file
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    // Copy data
    _, err = io.Copy(file, reader)
    if err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }

    return nil
}

// Download downloads a file from local filesystem
func (l *LocalBackend) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
    filePath := filepath.Join(l.basePath, bucket, key)
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    return file, nil
}

// Delete deletes a file from local filesystem
func (l *LocalBackend) Delete(ctx context.Context, bucket, key string) error {
    filePath := filepath.Join(l.basePath, bucket, key)
    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    return nil
}

// GetPresignedURL is not supported for local backend
func (l *LocalBackend) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
    // For local storage, return a direct file path or web path
    return fmt.Sprintf("/api/documents/download/%s/%s", bucket, key), nil
}

// Exists checks if a file exists in local filesystem
func (l *LocalBackend) Exists(ctx context.Context, bucket, key string) (bool, error) {
    filePath := filepath.Join(l.basePath, bucket, key)
    _, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("failed to stat file: %w", err)
    }
    return true, nil
}

// GetMetadata retrieves file metadata from local filesystem
func (l *LocalBackend) GetMetadata(ctx context.Context, bucket, key string) (*FileMetadata, error) {
    filePath := filepath.Join(l.basePath, bucket, key)
    stat, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to stat file: %w", err)
    }

    return &FileMetadata{
        Size:         stat.Size(),
        LastModified: stat.ModTime(),
    }, nil
}
```

### 5. Create Storage Factory

Create `internal/storage/factory.go`:

```go
package storage

import (
    "fmt"

    "github.com/shakfu/buyer/internal/config"
)

// NewStorageBackend creates a storage backend based on configuration
func NewStorageBackend(cfg *config.Config) (StorageBackend, error) {
    switch cfg.StorageType {
    case "minio", "s3":
        return NewMinioBackend(
            cfg.MinioEndpoint,
            cfg.MinioAccessKey,
            cfg.MinioSecretKey,
            cfg.MinioUseSSL,
        )
    case "local":
        return NewLocalBackend(cfg.StorageLocalPath)
    default:
        return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
    }
}
```

### 6. Update Document Service

Update `internal/services/document.go`:

```go
package services

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "path/filepath"
    "time"

    "github.com/google/uuid"
    "github.com/shakfu/buyer/internal/models"
    "github.com/shakfu/buyer/internal/storage"
    "gorm.io/gorm"
)

type DocumentService struct {
    db      *gorm.DB
    storage storage.StorageBackend
}

func NewDocumentService(db *gorm.DB, storageBackend storage.StorageBackend) *DocumentService {
    return &DocumentService{
        db:      db,
        storage: storageBackend,
    }
}

type CreateDocumentInput struct {
    EntityType  string
    EntityID    uint
    FileName    string
    FileType    string
    FileContent io.Reader  // Changed from FilePath
    FileSize    int64
    Description string
    UploadedBy  string
}

func (s *DocumentService) Create(input CreateDocumentInput) (*models.Document, error) {
    if input.EntityType == "" || input.EntityID == 0 || input.FileName == "" {
        return nil, ErrInvalidInput
    }

    ctx := context.Background()

    // Generate storage key
    now := time.Now()
    objectKey := fmt.Sprintf(
        "%s/%d/%d/%02d/%s-%s",
        input.EntityType,
        input.EntityID,
        now.Year(),
        now.Month(),
        uuid.New().String(),
        input.FileName,
    )

    // Get bucket name
    bucket := storage.BucketForEntityType(input.EntityType)

    // Upload to storage
    err := s.storage.Upload(ctx, bucket, objectKey, input.FileContent, input.FileSize, input.FileType)
    if err != nil {
        return nil, fmt.Errorf("failed to upload file: %w", err)
    }

    // Create database record
    doc := &models.Document{
        EntityType:  input.EntityType,
        EntityID:    input.EntityID,
        FileName:    input.FileName,
        FilePath:    objectKey,  // Store the object key, not local path
        FileType:    input.FileType,
        FileSize:    input.FileSize,
        Description: input.Description,
        UploadedBy:  input.UploadedBy,
    }

    if err := s.db.Create(doc).Error; err != nil {
        // Rollback - delete uploaded file
        _ = s.storage.Delete(ctx, bucket, objectKey)
        return nil, err
    }

    return doc, nil
}

func (s *DocumentService) GetDownloadURL(id uint, expiry time.Duration) (string, error) {
    var doc models.Document
    if err := s.db.First(&doc, id).Error; err != nil {
        return "", err
    }

    bucket := storage.BucketForEntityType(doc.EntityType)
    ctx := context.Background()

    url, err := s.storage.GetPresignedURL(ctx, bucket, doc.FilePath, expiry)
    if err != nil {
        return "", fmt.Errorf("failed to generate download URL: %w", err)
    }

    return url, nil
}

func (s *DocumentService) Download(id uint) (io.ReadCloser, *models.Document, error) {
    var doc models.Document
    if err := s.db.First(&doc, id).Error; err != nil {
        return nil, nil, err
    }

    bucket := storage.BucketForEntityType(doc.EntityType)
    ctx := context.Background()

    reader, err := s.storage.Download(ctx, bucket, doc.FilePath)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to download file: %w", err)
    }

    return reader, &doc, nil
}

func (s *DocumentService) Delete(id uint) error {
    var doc models.Document
    if err := s.db.First(&doc, id).Error; err != nil {
        return err
    }

    bucket := storage.BucketForEntityType(doc.EntityType)
    ctx := context.Background()

    // Delete from storage
    if err := s.storage.Delete(ctx, bucket, doc.FilePath); err != nil {
        return fmt.Errorf("failed to delete file from storage: %w", err)
    }

    // Delete from database
    if err := s.db.Delete(&doc).Error; err != nil {
        return err
    }

    return nil
}

// ... rest of the service methods remain similar ...
```

### 7. Update Web Handlers

Update `cmd/buyer/web.go` to handle file uploads:

```go
// Document upload handler
app.Post("/documents", func(c *fiber.Ctx) error {
    // Parse multipart form
    form, err := c.MultipartForm()
    if err != nil {
        return c.Status(fiber.StatusBadRequest).SendString("Failed to parse form")
    }

    // Get form fields
    entityType := c.FormValue("entity_type")
    entityID, _ := strconv.ParseUint(c.FormValue("entity_id"), 10, 32)
    description := c.FormValue("description")
    uploadedBy := c.FormValue("uploaded_by")

    // Get uploaded file
    files := form.File["file"]
    if len(files) == 0 {
        return c.Status(fiber.StatusBadRequest).SendString("No file uploaded")
    }

    file := files[0]

    // Open file
    fileContent, err := file.Open()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to open file")
    }
    defer fileContent.Close()

    // Detect file type
    fileType := filepath.Ext(file.Filename)
    if fileType != "" && fileType[0] == '.' {
        fileType = fileType[1:]
    }

    // Create document
    doc, err := docSvc.Create(services.CreateDocumentInput{
        EntityType:  entityType,
        EntityID:    uint(entityID),
        FileName:    file.Filename,
        FileType:    fileType,
        FileContent: fileContent,
        FileSize:    file.Size,
        Description: description,
        UploadedBy:  uploadedBy,
    })

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
    }

    // Return HTML row for HTMX
    html, err := RenderDocumentRow(doc)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
    }

    return c.Status(fiber.StatusCreated).SendString(string(html))
})

// Document download handler
app.Get("/documents/:id/download", func(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
    }

    reader, doc, err := docSvc.Download(uint(id))
    if err != nil {
        return c.Status(fiber.StatusNotFound).SendString("Document not found")
    }
    defer reader.Close()

    // Set headers
    c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.FileName))
    c.Set("Content-Type", doc.FileType)

    // Stream file
    return c.SendStream(reader)
})

// Get presigned download URL
app.Get("/documents/:id/url", func(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
    }

    url, err := docSvc.GetDownloadURL(uint(id), time.Hour)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found"})
    }

    return c.JSON(fiber.Map{"url": url})
})
```

### 8. Update HTML Template

Update `web/templates/documents.html` to use file upload:

```html
<form hx-post="/documents" hx-target="#documents-table tbody" hx-swap="beforeend"
      hx-encoding="multipart/form-data" hx-on::after-request="if(event.detail.successful) this.reset()">
    <div class="grid">
        <label for="entity_type">
            Entity Type
            <select id="entity_type" name="entity_type" required>
                <option value="">Select entity type...</option>
                <option value="vendor">Vendor</option>
                <option value="brand">Brand</option>
                <option value="product">Product</option>
                <option value="quote">Quote</option>
                <option value="purchase_order">Purchase Order</option>
                <option value="requisition">Requisition</option>
                <option value="project">Project</option>
            </select>
        </label>
        <label for="entity_id">
            Entity ID
            <input type="number" id="entity_id" name="entity_id" placeholder="Enter entity ID" required>
        </label>
    </div>

    <label for="file">
        File
        <input type="file" id="file" name="file" required>
    </label>

    <label for="description">
        Description
        <textarea id="description" name="description" placeholder="Document description (optional)" rows="3"></textarea>
    </label>

    <label for="uploaded_by">
        Uploaded By
        <input type="text" id="uploaded_by" name="uploaded_by" placeholder="user@example.com">
    </label>

    <button type="submit">Upload Document</button>
    <button type="button" onclick="toggleForm('add-document-form')" class="secondary">Cancel</button>
</form>
```

## Production Deployment

### Distributed MinIO Setup

For production, deploy MinIO in distributed mode for high availability:

```bash
# 4-node distributed setup
minio server http://minio{1...4}/data{1...4}
```

**Using Kubernetes:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: minio
spec:
  ports:
    - port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app: minio
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: minio
spec:
  serviceName: minio
  replicas: 4
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: quay.io/minio/minio:latest
        args:
        - server
        - http://minio-{0...3}.minio.default.svc.cluster.local/data
        env:
        - name: MINIO_ROOT_USER
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: root-user
        - name: MINIO_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: root-password
        ports:
        - containerPort: 9000
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 100Gi
```

### Backup and Disaster Recovery

```bash
# Backup bucket to S3
mc mirror local/vendor-docs s3/backup/vendor-docs

# Schedule with cron
0 2 * * * mc mirror local/vendor-docs s3/backup/vendor-docs
```

### Monitoring

Enable Prometheus metrics:

```bash
# Start MinIO with metrics
minio server /data --address :9000 --console-address :9001

# Scrape metrics
curl http://localhost:9000/minio/v2/metrics/cluster
```

## Security Best Practices

### 1. Access Keys

Never use default credentials in production:

```bash
# Generate strong credentials
export MINIO_ROOT_USER=$(openssl rand -hex 16)
export MINIO_ROOT_PASSWORD=$(openssl rand -hex 32)
```

### 2. TLS/SSL

Always use TLS in production:

```bash
# Generate self-signed certificate (or use Let's Encrypt)
openssl req -new -x509 -days 365 -nodes \
  -out ~/.minio/certs/public.crt \
  -keyout ~/.minio/certs/private.key

# Start with TLS
minio server --address :9000 --certs-dir ~/.minio/certs /data
```

Update config:
```bash
MINIO_USE_SSL=true
MINIO_ENDPOINT=minio.example.com:9000
```

### 3. Bucket Policies

Restrict access per entity type:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["arn:aws:iam::*:user/buyer-app"]},
      "Action": ["s3:GetObject", "s3:PutObject"],
      "Resource": ["arn:aws:s3:::vendor-docs/*"]
    }
  ]
}
```

### 4. Encryption

Enable server-side encryption:

```bash
# Enable SSE-S3
mc encrypt set sse-s3 local/vendor-docs
```

## Testing

### Unit Tests

Create `internal/storage/minio_test.go`:

```go
package storage

import (
    "bytes"
    "context"
    "testing"
    "time"
)

func TestMinioBackend(t *testing.T) {
    // Setup test MinIO instance
    backend, err := NewMinioBackend("localhost:9000", "minioadmin", "minioadmin", false)
    if err != nil {
        t.Skipf("MinIO not available: %v", err)
    }

    ctx := context.Background()
    bucket := "test-bucket"
    key := "test/file.txt"
    content := []byte("test content")

    // Test upload
    err = backend.Upload(ctx, bucket, key, bytes.NewReader(content), int64(len(content)), "text/plain")
    if err != nil {
        t.Fatalf("Upload failed: %v", err)
    }

    // Test exists
    exists, err := backend.Exists(ctx, bucket, key)
    if err != nil {
        t.Fatalf("Exists check failed: %v", err)
    }
    if !exists {
        t.Fatal("File should exist")
    }

    // Test download
    reader, err := backend.Download(ctx, bucket, key)
    if err != nil {
        t.Fatalf("Download failed: %v", err)
    }
    defer reader.Close()

    downloaded := new(bytes.Buffer)
    downloaded.ReadFrom(reader)
    if !bytes.Equal(downloaded.Bytes(), content) {
        t.Fatal("Downloaded content doesn't match")
    }

    // Test presigned URL
    url, err := backend.GetPresignedURL(ctx, bucket, key, time.Hour)
    if err != nil {
        t.Fatalf("GetPresignedURL failed: %v", err)
    }
    if url == "" {
        t.Fatal("URL should not be empty")
    }

    // Test delete
    err = backend.Delete(ctx, bucket, key)
    if err != nil {
        t.Fatalf("Delete failed: %v", err)
    }

    exists, _ = backend.Exists(ctx, bucket, key)
    if exists {
        t.Fatal("File should not exist after deletion")
    }
}
```

## Migration from Local Storage

Script to migrate existing documents from local filesystem to MinIO:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/shakfu/buyer/internal/config"
    "github.com/shakfu/buyer/internal/models"
    "github.com/shakfu/buyer/internal/storage"
    "gorm.io/gorm"
)

func migrateDocuments(db *gorm.DB, localStorage, minioBackend storage.StorageBackend) error {
    var documents []models.Document
    if err := db.Find(&documents).Error; err != nil {
        return err
    }

    ctx := context.Background()

    for _, doc := range documents {
        fmt.Printf("Migrating document %d: %s\n", doc.ID, doc.FileName)

        // Read from local storage
        reader, err := localStorage.Download(ctx, "", doc.FilePath)
        if err != nil {
            fmt.Printf("  Warning: Failed to read %s: %v\n", doc.FilePath, err)
            continue
        }

        // Generate new key
        bucket := storage.BucketForEntityType(doc.EntityType)
        newKey := fmt.Sprintf("%s/%d/%s", doc.EntityType, doc.EntityID, filepath.Base(doc.FilePath))

        // Upload to MinIO
        err = minioBackend.Upload(ctx, bucket, newKey, reader, doc.FileSize, doc.FileType)
        reader.Close()

        if err != nil {
            fmt.Printf("  Warning: Failed to upload to MinIO: %v\n", err)
            continue
        }

        // Update database record
        doc.FilePath = newKey
        if err := db.Save(&doc).Error; err != nil {
            fmt.Printf("  Warning: Failed to update database: %v\n", err)
            continue
        }

        fmt.Printf("  Success: Migrated to %s/%s\n", bucket, newKey)
    }

    return nil
}

func main() {
    cfg, _ := config.Load()

    // Create backends
    localStorage, _ := storage.NewLocalBackend(cfg.StorageLocalPath)
    minioBackend, _ := storage.NewMinioBackend(
        cfg.MinioEndpoint,
        cfg.MinioAccessKey,
        cfg.MinioSecretKey,
        cfg.MinioUseSSL,
    )

    // Migrate
    if err := migrateDocuments(cfg.DB, localStorage, minioBackend); err != nil {
        fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Migration completed successfully!")
}
```

## Benefits Summary

| Feature | Local Storage | MinIO |
|---------|--------------|-------|
| Scalability | Limited by disk | Unlimited (distributed) |
| Redundancy | None | Erasure coding |
| Versioning | Manual | Built-in |
| Access Control | File permissions | IAM policies |
| Multi-instance | Shared filesystem needed | Native support |
| CDN Integration | Complex | Direct S3 compatibility |
| Disaster Recovery | Manual backups | Automated replication |
| Cost | Server disk space | Commodity hardware |
| Performance | Disk I/O limited | Distributed, parallel |
| Compliance | Limited | Audit logs, retention |

## Conclusion

Integrating MinIO with the buyer application provides a robust, scalable, and production-ready document storage solution. The abstraction through the `StorageBackend` interface allows seamless switching between local and MinIO storage, making it easy to start with local development and scale to distributed MinIO in production.

## Resources

- MinIO Documentation: https://min.io/docs/minio/linux/index.html
- MinIO Go SDK: https://min.io/docs/minio/linux/developers/go/minio-go.html
- MinIO Kubernetes: https://min.io/docs/minio/kubernetes/upstream/index.html
- S3 API Reference: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html
