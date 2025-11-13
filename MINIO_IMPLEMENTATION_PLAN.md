# MinIO Implementation Plan for Buyer Application

## Overview

This document provides a step-by-step implementation plan for integrating MinIO object storage with the buyer application's document management system.

**Estimated Time**: 1-2 days
**Difficulty**: Medium
**Prerequisites**: Docker (for development), Go 1.21+

## Phase 1: Environment Setup (30 minutes)

### Step 1.1: Start MinIO Server

```bash
# Using Docker Compose (recommended)
docker-compose -f docker-compose.minio.yml up -d

# Verify MinIO is running
docker logs buyer-minio

# Access MinIO Console
open http://localhost:9001
# Login: minioadmin / minioadmin
```

### Step 1.2: Install MinIO Client (mc)

```bash
# macOS
brew install minio/stable/mc

# Linux
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/

# Configure
mc alias set buyerlocal http://localhost:9000 minioadmin minioadmin
mc admin info buyerlocal
```

### Step 1.3: Create Buckets

```bash
# Create buckets for each entity type
mc mb buyerlocal/vendor-docs
mc mb buyerlocal/product-docs
mc mb buyerlocal/quote-docs
mc mb buyerlocal/po-docs
mc mb buyerlocal/requisition-docs
mc mb buyerlocal/project-docs
mc mb buyerlocal/brand-docs

# Enable versioning (optional but recommended)
mc version enable buyerlocal/vendor-docs
mc version enable buyerlocal/quote-docs
mc version enable buyerlocal/po-docs

# List buckets
mc ls buyerlocal
```

### Step 1.4: Install Go Dependencies

```bash
# Install MinIO Go SDK
go get github.com/minio/minio-go/v7

# Update go.mod
go mod tidy
```

**Verification**:
- [ ] MinIO server accessible at http://localhost:9000
- [ ] MinIO console accessible at http://localhost:9001
- [ ] All 7 buckets created
- [ ] mc client configured and working
- [ ] Go dependencies installed

## Phase 2: Implement Storage Layer (2-3 hours)

### Step 2.1: Create Storage Interface

**File**: `internal/storage/storage.go`

```bash
mkdir -p internal/storage
```

Create the storage interface that abstracts MinIO and local filesystem:

```go
package storage

import (
    "context"
    "io"
    "time"
)

// StorageBackend defines the interface for document storage
type StorageBackend interface {
    Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
    Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, bucket, key string) error
    GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
    Exists(ctx context.Context, bucket, key string) (bool, error)
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

### Step 2.2: Implement MinIO Backend

**File**: `internal/storage/minio.go`

Refer to MINIO_INTEGRATION.md lines 1103-1283 for complete implementation.

Key methods to implement:
- `NewMinioBackend(endpoint, accessKey, secretKey string, useSSL bool) (*MinioBackend, error)`
- `Upload(ctx, bucket, key, reader, size, contentType) error`
- `Download(ctx, bucket, key) (io.ReadCloser, error)`
- `Delete(ctx, bucket, key) error`
- `GetPresignedURL(ctx, bucket, key, expiry) (string, error)`
- `Exists(ctx, bucket, key) (bool, error)`
- `GetMetadata(ctx, bucket, key) (*FileMetadata, error)`

### Step 2.3: Implement Local Backend

**File**: `internal/storage/local.go`

Refer to MINIO_INTEGRATION.md lines 1285-1383 for complete implementation.

This provides a fallback for local development without MinIO.

### Step 2.4: Create Storage Factory

**File**: `internal/storage/factory.go`

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

**Verification**:
- [ ] All storage files created
- [ ] Code compiles without errors: `go build ./internal/storage/...`
- [ ] Interface properly defined

## Phase 3: Update Configuration (30 minutes)

### Step 3.1: Update Config Struct

**File**: `internal/config/config.go`

Add these fields to the Config struct:

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
```

### Step 3.2: Update Config Loading

Add to the `Load()` function:

```go
func Load() (*Config, error) {
    // ... existing code ...

    cfg.MinioEndpoint = getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000")
    cfg.MinioAccessKey = getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin")
    cfg.MinioSecretKey = getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin")
    cfg.MinioUseSSL = getEnvOrDefault("MINIO_USE_SSL", "false") == "true"
    cfg.MinioRegion = getEnvOrDefault("MINIO_REGION", "us-east-1")
    cfg.StorageType = getEnvOrDefault("DOCUMENT_STORAGE_TYPE", "local")
    cfg.StorageLocalPath = getEnvOrDefault("DOCUMENT_STORAGE_LOCAL_PATH", filepath.Join(homeDir, ".buyer", "documents"))
    cfg.DocumentURLExpiry = 3600
    if expiry := getEnvOrDefault("DOCUMENT_URL_EXPIRY", "3600"); expiry != "" {
        if val, err := strconv.Atoi(expiry); err == nil {
            cfg.DocumentURLExpiry = val
        }
    }

    return cfg, nil
}
```

### Step 3.3: Update .env.example

**File**: `.env.example`

Add these lines:

```bash
# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_REGION=us-east-1

# Storage Configuration
DOCUMENT_STORAGE_TYPE=local  # Options: local, minio
DOCUMENT_STORAGE_LOCAL_PATH=/var/lib/buyer/documents
DOCUMENT_URL_EXPIRY=3600  # Presigned URL expiry in seconds (1 hour)
```

### Step 3.4: Create .env for Development

**File**: `.env`

```bash
# Development settings
BUYER_ENV=development

# MinIO Configuration (for development)
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Use MinIO for development
DOCUMENT_STORAGE_TYPE=minio
DOCUMENT_URL_EXPIRY=3600
```

**Verification**:
- [ ] Config struct updated
- [ ] Config loading updated
- [ ] .env.example created
- [ ] .env created for development
- [ ] Config compiles: `go build ./internal/config/...`

## Phase 4: Update Document Service (1-2 hours)

### Step 4.1: Modify DocumentService Constructor

**File**: `internal/services/document.go`

Update the service to accept a storage backend:

```go
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
```

### Step 4.2: Update Create Method

Change from file path to file content:

```go
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
        FilePath:    objectKey,  // Store the object key
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
```

### Step 4.3: Add Download Methods

```go
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
```

### Step 4.4: Update Delete Method

```go
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
```

**Verification**:
- [ ] DocumentService updated
- [ ] Code compiles: `go build ./internal/services/...`
- [ ] All methods use storage backend interface

## Phase 5: Update Application Initialization (30 minutes)

### Step 5.1: Update main.go

**File**: `cmd/buyer/main.go`

Initialize storage backend:

```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Initialize storage backend
storageBackend, err := storage.NewStorageBackend(cfg)
if err != nil {
    log.Fatalf("Failed to initialize storage: %v", err)
}

// Initialize services with storage backend
docSvc := services.NewDocumentService(cfg.DB, storageBackend)
```

### Step 5.2: Update web.go

**File**: `cmd/buyer/web.go`

Initialize storage backend in web command:

```go
var webCmd = &cobra.Command{
    Use:   "web",
    Short: "Start the web server",
    Run: func(cmd *cobra.Command, args []string) {
        // ... existing code ...

        // Initialize storage backend
        storageBackend, err := storage.NewStorageBackend(cfg)
        if err != nil {
            log.Fatalf("Failed to initialize storage: %v", err)
        }

        // Initialize services
        docSvc := services.NewDocumentService(db, storageBackend)
        // ... other services ...

        setupRoutes(app, specSvc, brandSvc, productSvc, vendorSvc,
            requisitionSvc, quoteSvc, forexSvc, dashboardSvc,
            projectSvc, projectReqSvc, poSvc, docSvc, ratingsSvc)
    },
}
```

### Step 5.3: Update CLI Commands

**File**: `cmd/buyer/add.go`

Update document add command to read file and upload:

```go
var addDocumentCmd = &cobra.Command{
    Use:   "document --entity-type [type] --entity-id [id] --file-path [path]",
    Short: "Add a new document to an entity",
    Run: func(cmd *cobra.Command, args []string) {
        // ... validation ...

        // Initialize storage backend
        storageBackend, err := storage.NewStorageBackend(cfg)
        if err != nil {
            log.Fatalf("Failed to initialize storage: %v", err)
        }

        svc := services.NewDocumentService(cfg.DB, storageBackend)

        // Open file
        file, err := os.Open(filePath)
        if err != nil {
            log.Fatalf("Failed to open file: %v", err)
        }
        defer file.Close()

        // Get file info
        fileInfo, err := file.Stat()
        if err != nil {
            log.Fatalf("Failed to stat file: %v", err)
        }

        // Create document
        doc, err := svc.Create(services.CreateDocumentInput{
            EntityType:  entityType,
            EntityID:    entityID,
            FileName:    filepath.Base(filePath),
            FileType:    fileType,
            FileContent: file,
            FileSize:    fileInfo.Size(),
            Description: description,
            UploadedBy:  uploadedBy,
        })

        if err != nil {
            log.Fatalf("Failed to create document: %v", err)
        }

        fmt.Printf("Document created successfully (ID: %d)\n", doc.ID)
    },
}
```

**Verification**:
- [ ] main.go updated
- [ ] web.go updated
- [ ] add.go updated
- [ ] Application compiles: `make build`

## Phase 6: Update Web Handlers (1 hour)

### Step 6.1: Update Document Upload Handler

**File**: `cmd/buyer/web.go`

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
    entityIDStr := c.FormValue("entity_id")
    entityID, _ := strconv.ParseUint(entityIDStr, 10, 32)
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
    fileType := file.Header.Get("Content-Type")
    if fileType == "" {
        ext := filepath.Ext(file.Filename)
        fileType = mime.TypeByExtension(ext)
        if fileType == "" {
            fileType = "application/octet-stream"
        }
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
```

### Step 6.2: Add Download Handler

```go
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
    if doc.FileSize > 0 {
        c.Set("Content-Length", strconv.FormatInt(doc.FileSize, 10))
    }

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

    return c.JSON(fiber.Map{"url": url, "expires_in": 3600})
})
```

### Step 6.3: Update HTML Template

**File**: `web/templates/documents.html`

Change from text input to file input:

```html
<label for="file">
    File
    <input type="file" id="file" name="file" required>
    <small>Select a file to upload</small>
</label>
```

**Verification**:
- [ ] Upload handler updated
- [ ] Download handlers added
- [ ] Template updated
- [ ] Application compiles: `make build`

## Phase 7: Testing (1-2 hours)

### Step 7.1: Unit Tests

Create test file: `internal/storage/minio_test.go`

```go
package storage

import (
    "bytes"
    "context"
    "testing"
    "time"
)

func TestMinioBackend(t *testing.T) {
    // Skip if MinIO not available
    backend, err := NewMinioBackend("localhost:9000", "minioadmin", "minioadmin", false)
    if err != nil {
        t.Skipf("MinIO not available: %v", err)
    }

    ctx := context.Background()
    bucket := "test-bucket"
    key := "test/file.txt"
    content := []byte("test content")

    // Test upload
    t.Run("Upload", func(t *testing.T) {
        err := backend.Upload(ctx, bucket, key, bytes.NewReader(content), int64(len(content)), "text/plain")
        if err != nil {
            t.Fatalf("Upload failed: %v", err)
        }
    })

    // Test exists
    t.Run("Exists", func(t *testing.T) {
        exists, err := backend.Exists(ctx, bucket, key)
        if err != nil {
            t.Fatalf("Exists check failed: %v", err)
        }
        if !exists {
            t.Fatal("File should exist")
        }
    })

    // Test download
    t.Run("Download", func(t *testing.T) {
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
    })

    // Test metadata
    t.Run("GetMetadata", func(t *testing.T) {
        meta, err := backend.GetMetadata(ctx, bucket, key)
        if err != nil {
            t.Fatalf("GetMetadata failed: %v", err)
        }
        if meta.Size != int64(len(content)) {
            t.Errorf("Expected size %d, got %d", len(content), meta.Size)
        }
    })

    // Test presigned URL
    t.Run("GetPresignedURL", func(t *testing.T) {
        url, err := backend.GetPresignedURL(ctx, bucket, key, time.Hour)
        if err != nil {
            t.Fatalf("GetPresignedURL failed: %v", err)
        }
        if url == "" {
            t.Fatal("URL should not be empty")
        }
    })

    // Test delete
    t.Run("Delete", func(t *testing.T) {
        err := backend.Delete(ctx, bucket, key)
        if err != nil {
            t.Fatalf("Delete failed: %v", err)
        }

        exists, _ := backend.Exists(ctx, bucket, key)
        if exists {
            t.Fatal("File should not exist after deletion")
        }
    })
}
```

Run tests:

```bash
# Start MinIO first
docker-compose -f docker-compose.minio.yml up -d

# Run tests
go test ./internal/storage/... -v

# Run all tests
make test
```

### Step 7.2: Integration Testing

Create test script: `scripts/test_minio_integration.sh`

```bash
#!/bin/bash

set -e

echo "=== MinIO Integration Test ==="

# 1. Check MinIO is running
echo "1. Checking MinIO connection..."
mc admin info buyerlocal || {
    echo "ERROR: MinIO not running. Start with: docker-compose -f docker-compose.minio.yml up -d"
    exit 1
}

# 2. Create test file
echo "2. Creating test file..."
echo "This is a test document" > /tmp/test-doc.txt

# 3. Upload via CLI
echo "3. Uploading document via CLI..."
./buyer add document \
    --entity-type vendor \
    --entity-id 1 \
    --file-path /tmp/test-doc.txt \
    --description "Test document" \
    --uploaded-by "test@example.com"

# 4. List documents
echo "4. Listing documents..."
./buyer list documents --entity-type vendor

# 5. Verify in MinIO
echo "5. Verifying in MinIO..."
mc ls buyerlocal/vendor-docs/vendor/1/ --recursive

# 6. Check via web (manual)
echo "6. Test web interface manually:"
echo "   - Start web server: ./buyer web"
echo "   - Open http://localhost:8080/documents"
echo "   - Upload a file"
echo "   - Download the file"

echo ""
echo "=== Integration test completed successfully ==="
```

Make executable:

```bash
chmod +x scripts/test_minio_integration.sh
```

### Step 7.3: Manual Testing Checklist

- [ ] Start MinIO: `docker-compose -f docker-compose.minio.yml up -d`
- [ ] Build application: `make build`
- [ ] Upload via CLI: `./buyer add document --entity-type vendor --entity-id 1 --file-path /path/to/file.pdf`
- [ ] List documents: `./buyer list documents`
- [ ] Start web: `./buyer web`
- [ ] Upload via web at http://localhost:8080/documents
- [ ] Download document via web
- [ ] Check MinIO console at http://localhost:9001
- [ ] Verify file exists in correct bucket
- [ ] Delete document
- [ ] Verify file removed from MinIO

## Phase 8: Migration from Local Storage (1 hour)

### Step 8.1: Run Migration Script

See `scripts/migrate_to_minio.go` for the complete migration script.

```bash
# Build migration tool
go build -o migrate-docs ./scripts/migrate_to_minio.go

# Dry run (preview what will be migrated)
./migrate-docs --dry-run

# Run actual migration
./migrate-docs

# Verify migration
./migrate-docs --verify
```

### Step 8.2: Backup Before Migration

```bash
# Backup database
cp ~/.buyer/buyer.db ~/.buyer/buyer.db.backup

# Backup local files
tar -czf ~/buyer-docs-backup.tar.gz ~/.buyer/documents/
```

### Step 8.3: Verify Migration

```bash
# Check all documents migrated
./buyer list documents

# Verify in MinIO
mc ls buyerlocal/vendor-docs --recursive
mc ls buyerlocal/product-docs --recursive
mc ls buyerlocal/quote-docs --recursive

# Test downloads
./buyer web
# Open http://localhost:8080/documents and test downloading
```

**Verification**:
- [ ] Migration script runs successfully
- [ ] All documents migrated to MinIO
- [ ] Document records updated in database
- [ ] Downloads work correctly
- [ ] Original files backed up

## Phase 9: Production Deployment (varies)

### Step 9.1: Update Production Configuration

**File**: `.env.production`

```bash
BUYER_ENV=production

# MinIO Production Configuration
MINIO_ENDPOINT=minio.yourdomain.com:9000
MINIO_ACCESS_KEY=${MINIO_PROD_ACCESS_KEY}
MINIO_SECRET_KEY=${MINIO_PROD_SECRET_KEY}
MINIO_USE_SSL=true
MINIO_REGION=us-east-1

# Use MinIO in production
DOCUMENT_STORAGE_TYPE=minio
DOCUMENT_URL_EXPIRY=3600
```

### Step 9.2: Deploy MinIO in Production

**Option A: Docker Compose** (single node)

```bash
docker-compose -f docker-compose.minio.prod.yml up -d
```

**Option B: Kubernetes** (distributed)

See MINIO_INTEGRATION.md lines 1738-1805 for Kubernetes StatefulSet configuration.

```bash
kubectl apply -f k8s/minio-secret.yaml
kubectl apply -f k8s/minio-statefulset.yaml
kubectl apply -f k8s/minio-service.yaml
```

**Option C: Cloud MinIO** (managed)

Use MinIO cloud service or run on AWS/GCP/Azure VMs.

### Step 9.3: Setup TLS/SSL

```bash
# Generate certificate (or use Let's Encrypt)
openssl req -new -x509 -days 365 -nodes \
    -out /etc/minio/certs/public.crt \
    -keyout /etc/minio/certs/private.key

# Configure MinIO to use TLS
# Set MINIO_USE_SSL=true in application config
```

### Step 9.4: Setup Backup and Monitoring

```bash
# Setup daily backup to S3
mc mirror --watch buyerlocal/vendor-docs s3/backup-bucket/vendor-docs

# Setup Prometheus monitoring
# Configure Prometheus to scrape MinIO metrics endpoint
```

**Verification**:
- [ ] Production MinIO deployed
- [ ] TLS/SSL configured
- [ ] Backup configured
- [ ] Monitoring configured
- [ ] Application deployed with MinIO config

## Rollback Plan

If issues occur, rollback to local storage:

```bash
# 1. Update .env
DOCUMENT_STORAGE_TYPE=local

# 2. Restore database backup
cp ~/.buyer/buyer.db.backup ~/.buyer/buyer.db

# 3. Restore local files
tar -xzf ~/buyer-docs-backup.tar.gz -C ~/

# 4. Restart application
./buyer web
```

## Troubleshooting

### MinIO Connection Issues

```bash
# Check MinIO is running
docker ps | grep minio

# Check logs
docker logs buyer-minio

# Test connection
mc admin info buyerlocal

# Verify network
curl http://localhost:9000/minio/health/live
```

### Upload Failures

```bash
# Check bucket permissions
mc stat buyerlocal/vendor-docs

# Check disk space
df -h

# Check application logs
tail -f /var/log/buyer/app.log
```

### Download Failures

```bash
# Verify file exists in MinIO
mc ls buyerlocal/vendor-docs/vendor/1/ --recursive

# Check presigned URL generation
# Enable debug logging in application

# Test direct MinIO download
mc cp buyerlocal/vendor-docs/vendor/1/[file] /tmp/test-download
```

## Success Criteria

- [ ] MinIO server running and accessible
- [ ] All 7 buckets created
- [ ] Storage layer implemented and tested
- [ ] DocumentService updated and working
- [ ] CLI commands upload and download working
- [ ] Web interface upload and download working
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Migration script successfully migrates existing documents
- [ ] Production deployment plan documented

## Next Steps

After successful implementation:

1. **Performance Optimization**
   - Enable MinIO caching
   - Configure CDN for public documents
   - Implement connection pooling

2. **Advanced Features**
   - Enable object versioning
   - Setup lifecycle policies (auto-archive old docs)
   - Implement document sharing via presigned URLs
   - Add document preview functionality

3. **Security Hardening**
   - Implement bucket policies
   - Enable server-side encryption
   - Setup IAM users for different access levels
   - Enable audit logging

4. **Monitoring and Alerts**
   - Setup Grafana dashboards
   - Configure alerts for storage capacity
   - Monitor upload/download latency
   - Track storage costs

## Resources

- MINIO_INTEGRATION.md - Complete integration guide
- docker-compose.minio.yml - Test MinIO setup
- scripts/migrate_to_minio.go - Migration script
- MinIO Documentation: https://min.io/docs
- MinIO Go SDK: https://github.com/minio/minio-go

## Support

For issues during implementation:
- Check MinIO logs: `docker logs buyer-minio`
- Check application logs
- Review MINIO_INTEGRATION.md for detailed examples
- MinIO Community: https://slack.min.io
