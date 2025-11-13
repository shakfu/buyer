# Buyer Scripts

This directory contains utility scripts for the buyer application.

## MinIO Migration Tools

### migrate_to_minio.go

Migration tool for moving documents from local filesystem storage to MinIO object storage.

**Build:**
```bash
go build -o migrate-docs ./scripts/migrate_to_minio.go
```

**Usage:**

```bash
# Dry run - preview migration without making changes
./migrate-docs --dry-run

# Verbose dry run - see detailed information
./migrate-docs --dry-run --verbose

# Migrate all documents
./migrate-docs

# Migrate only specific entity type
./migrate-docs --entity-type vendor

# Verify migration integrity
./migrate-docs --verify

# Verify with verbose output
./migrate-docs --verify --verbose
```

**Features:**
- Dry run mode for safe preview
- Automatic bucket selection based on entity type
- Progress tracking with statistics
- Error handling and rollback
- Verification mode to check migration integrity
- Database record updates
- Skip already-migrated documents

**Before Migration:**
1. Backup your database: `cp ~/.buyer/buyer.db ~/.buyer/buyer.db.backup`
2. Backup local documents: `tar -czf ~/buyer-docs-backup.tar.gz ~/.buyer/documents/`
3. Ensure MinIO is running: `docker-compose -f docker-compose.minio.yml up -d`
4. Set environment: `DOCUMENT_STORAGE_TYPE=minio` in `.env`

**After Migration:**
1. Verify: `./migrate-docs --verify`
2. Test downloads: `./buyer web` and test document downloads
3. Keep backups for a few days before cleanup

### test_minio_integration.sh

Automated integration test script for MinIO setup.

**Usage:**
```bash
./scripts/test_minio_integration.sh
```

**Tests:**
- MinIO connection and availability
- Docker container status
- API endpoint accessibility
- Console accessibility
- Bucket creation
- Document upload via CLI
- Document listing
- Document verification in MinIO
- Web interface availability

**Prerequisites:**
- Docker and docker-compose installed
- MinIO running: `docker-compose -f docker-compose.minio.yml up -d`
- MinIO client (mc) installed (optional but recommended)
- Buyer application built: `make build`

**Example Output:**
```
=== MinIO Integration Test ===

→ Checking MinIO connection...
✓ MinIO connection successful
→ Checking MinIO Docker container...
✓ MinIO container is running
→ Checking MinIO API endpoint...
✓ MinIO API is accessible
...
✓ All automated tests passed!
```

## Adding New Scripts

When adding new scripts to this directory:

1. **Make executable:**
   ```bash
   chmod +x scripts/your-script.sh
   ```

2. **Add shebang:**
   ```bash
   #!/bin/bash
   ```

3. **Document in this README:**
   - Brief description
   - Usage examples
   - Prerequisites
   - Expected output

4. **Follow conventions:**
   - Use descriptive names
   - Include help messages
   - Handle errors gracefully
   - Provide verbose modes
   - Add exit codes

## Script Conventions

### Exit Codes
- `0` - Success
- `1` - General error
- `2` - Invalid usage/arguments
- `3` - Missing prerequisites

### Output Colors
Use colors for better visibility:
- Green (✓) - Success
- Red (✗) - Error
- Yellow (⚠) - Warning
- Blue (→) - Info

### Error Handling
Always use `set -e` for Bash scripts to exit on errors:
```bash
#!/bin/bash
set -e
```

## Troubleshooting

### Migration Script Issues

**Error: "Failed to initialize MinIO storage"**
- Check MinIO is running: `docker ps | grep minio`
- Verify credentials in `.env` file
- Test connection: `mc admin info buyerlocal`

**Error: "Failed to read from local storage"**
- Check file paths in database are correct
- Verify local storage path in config
- Ensure files exist: `ls -la ~/.buyer/documents/`

**Error: "Failed to upload to MinIO"**
- Check MinIO disk space: `mc admin info buyerlocal`
- Verify bucket exists: `mc ls buyerlocal`
- Check MinIO logs: `docker logs buyer-minio`

### Test Script Issues

**Error: "MinIO container not running"**
```bash
# Start MinIO
docker-compose -f docker-compose.minio.yml up -d

# Check status
docker-compose -f docker-compose.minio.yml ps
```

**Error: "Buyer binary not found"**
```bash
# Build application
make build

# Or specify path
./buyer add document ...
```

**Error: "MinIO client (mc) not installed"**
```bash
# macOS
brew install minio/stable/mc

# Linux
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/
```

## Resources

- [MinIO Documentation](https://min.io/docs)
- [MinIO Go SDK](https://github.com/minio/minio-go)
- [MINIO_INTEGRATION.md](../MINIO_INTEGRATION.md) - Complete integration guide
- [MINIO_IMPLEMENTATION_PLAN.md](../MINIO_IMPLEMENTATION_PLAN.md) - Step-by-step implementation
