package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/storage"
	"gorm.io/gorm"
)

// MigrationStats tracks migration progress
type MigrationStats struct {
	Total     int
	Migrated  int
	Failed    int
	Skipped   int
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	// Command line flags
	dryRun := flag.Bool("dry-run", false, "Preview migration without making changes")
	verify := flag.Bool("verify", false, "Verify migration integrity")
	entityType := flag.String("entity-type", "", "Migrate only specific entity type (optional)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	log.Println("=== MinIO Migration Tool ===")
	log.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Verify MinIO is configured
	if cfg.StorageType != "minio" {
		log.Println("WARNING: DOCUMENT_STORAGE_TYPE is not set to 'minio'")
		log.Println("Please update your .env file:")
		log.Println("  DOCUMENT_STORAGE_TYPE=minio")
		log.Println()
	}

	// Initialize storage backends
	log.Println("Initializing storage backends...")

	// Source: Local storage
	localBackend, err := storage.NewLocalBackend(cfg.StorageLocalPath)
	if err != nil {
		log.Fatalf("Failed to initialize local storage: %v", err)
	}

	// Destination: MinIO storage
	minioBackend, err := storage.NewMinioBackend(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}

	log.Printf("  Local storage: %s\n", cfg.StorageLocalPath)
	log.Printf("  MinIO endpoint: %s\n", cfg.MinioEndpoint)
	log.Println()

	// Get documents from database
	var documents []models.Document
	query := cfg.DB
	if *entityType != "" {
		query = query.Where("entity_type = ?", *entityType)
	}

	if err := query.Find(&documents).Error; err != nil {
		log.Fatalf("Failed to query documents: %v", err)
	}

	log.Printf("Found %d document(s) to process\n", len(documents))

	if len(documents) == 0 {
		log.Println("No documents to migrate. Exiting.")
		return
	}

	// Handle verify mode
	if *verify {
		verifyMigration(cfg.DB, minioBackend, documents, *verbose)
		return
	}

	// Handle dry run
	if *dryRun {
		dryRunMigration(localBackend, minioBackend, documents, *verbose)
		return
	}

	// Confirm migration
	log.Println()
	log.Println("This will migrate documents from local filesystem to MinIO.")
	log.Println("Database records will be updated with new file paths.")
	log.Print("Continue? (yes/no): ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		log.Println("Migration cancelled.")
		return
	}

	// Perform migration
	stats := performMigration(cfg.DB, localBackend, minioBackend, documents, *verbose)

	// Print summary
	printSummary(stats)
}

func dryRunMigration(local, minio storage.StorageBackend, documents []models.Document, verbose bool) {
	log.Println()
	log.Println("=== DRY RUN MODE ===")
	log.Println("No changes will be made.")
	log.Println()

	ctx := context.Background()

	for i, doc := range documents {
		log.Printf("[%d/%d] Processing: %s\n", i+1, len(documents), doc.FileName)

		bucket := storage.BucketForEntityType(doc.EntityType)
		oldKey := doc.FilePath
		newKey := generateNewKey(doc)

		if verbose {
			log.Printf("  Entity: %s/%d\n", doc.EntityType, doc.EntityID)
			log.Printf("  Old path: %s\n", oldKey)
			log.Printf("  New path: %s/%s\n", bucket, newKey)
		}

		// Check if file exists in local storage
		exists, err := local.Exists(ctx, "", oldKey)
		if err != nil || !exists {
			log.Printf("  ⚠️  WARNING: File not found in local storage\n")
			continue
		}

		// Check if already exists in MinIO
		exists, err = minio.Exists(ctx, bucket, newKey)
		if err == nil && exists {
			log.Printf("  ℹ️  Already exists in MinIO (will skip)\n")
			continue
		}

		log.Printf("  ✓ Will migrate to MinIO\n")
	}

	log.Println()
	log.Println("Dry run completed. Run without --dry-run to perform actual migration.")
}

func performMigration(db *gorm.DB, local, minio storage.StorageBackend, documents []models.Document, verbose bool) MigrationStats {
	stats := MigrationStats{
		Total:     len(documents),
		StartTime: time.Now(),
	}

	ctx := context.Background()

	log.Println()
	log.Println("=== Starting Migration ===")
	log.Println()

	for i, doc := range documents {
		log.Printf("[%d/%d] Migrating: %s\n", i+1, stats.Total, doc.FileName)

		bucket := storage.BucketForEntityType(doc.EntityType)
		oldKey := doc.FilePath
		newKey := generateNewKey(doc)

		if verbose {
			log.Printf("  Entity: %s/%d\n", doc.EntityType, doc.EntityID)
			log.Printf("  Bucket: %s\n", bucket)
			log.Printf("  Old key: %s\n", oldKey)
			log.Printf("  New key: %s\n", newKey)
		}

		// Check if already migrated
		exists, err := minio.Exists(ctx, bucket, newKey)
		if err == nil && exists {
			log.Printf("  ℹ️  Already exists in MinIO, skipping...\n")
			stats.Skipped++
			continue
		}

		// Download from local storage
		reader, err := local.Download(ctx, "", oldKey)
		if err != nil {
			log.Printf("  ❌ Failed to read from local storage: %v\n", err)
			stats.Failed++
			continue
		}

		// Get file size
		var fileSize int64 = doc.FileSize
		if fileSize == 0 {
			// Try to get size from metadata
			meta, err := local.GetMetadata(ctx, "", oldKey)
			if err == nil {
				fileSize = meta.Size
			}
		}

		// Upload to MinIO
		err = minio.Upload(ctx, bucket, newKey, reader, fileSize, doc.FileType)
		reader.Close()

		if err != nil {
			log.Printf("  ❌ Failed to upload to MinIO: %v\n", err)
			stats.Failed++
			continue
		}

		// Update database record
		doc.FilePath = newKey
		if err := db.Save(&doc).Error; err != nil {
			log.Printf("  ❌ Failed to update database: %v\n", err)
			// Try to rollback - delete from MinIO
			_ = minio.Delete(ctx, bucket, newKey)
			stats.Failed++
			continue
		}

		log.Printf("  ✓ Successfully migrated\n")
		stats.Migrated++

		if verbose {
			// Verify upload
			meta, err := minio.GetMetadata(ctx, bucket, newKey)
			if err == nil {
				log.Printf("  Verified in MinIO: %d bytes\n", meta.Size)
			}
		}
	}

	stats.EndTime = time.Now()
	return stats
}

func verifyMigration(db *gorm.DB, minio storage.StorageBackend, documents []models.Document, verbose bool) {
	log.Println()
	log.Println("=== Verifying Migration ===")
	log.Println()

	ctx := context.Background()
	verified := 0
	missing := 0
	errors := 0

	for i, doc := range documents {
		if verbose {
			log.Printf("[%d/%d] Verifying: %s\n", i+1, len(documents), doc.FileName)
		}

		bucket := storage.BucketForEntityType(doc.EntityType)
		key := doc.FilePath

		// Check if exists in MinIO
		exists, err := minio.Exists(ctx, bucket, key)
		if err != nil {
			log.Printf("  ❌ Error checking %s: %v\n", doc.FileName, err)
			errors++
			continue
		}

		if !exists {
			log.Printf("  ⚠️  Missing: %s (ID: %d, Path: %s/%s)\n",
				doc.FileName, doc.ID, bucket, key)
			missing++
			continue
		}

		// Get metadata and verify
		meta, err := minio.GetMetadata(ctx, bucket, key)
		if err != nil {
			log.Printf("  ⚠️  Cannot get metadata for %s: %v\n", doc.FileName, err)
			errors++
			continue
		}

		// Check size mismatch
		if doc.FileSize > 0 && meta.Size != doc.FileSize {
			log.Printf("  ⚠️  Size mismatch for %s: DB=%d, MinIO=%d\n",
				doc.FileName, doc.FileSize, meta.Size)
		}

		if verbose {
			log.Printf("  ✓ %s: %d bytes\n", doc.FileName, meta.Size)
		}

		verified++
	}

	log.Println()
	log.Println("=== Verification Results ===")
	log.Printf("Total documents: %d\n", len(documents))
	log.Printf("Verified:        %d ✓\n", verified)
	log.Printf("Missing:         %d ⚠️\n", missing)
	log.Printf("Errors:          %d ❌\n", errors)

	if missing > 0 || errors > 0 {
		log.Println()
		log.Println("⚠️  Verification failed! Some documents are missing or have errors.")
		os.Exit(1)
	}

	log.Println()
	log.Println("✓ All documents verified successfully!")
}

func generateNewKey(doc models.Document) string {
	// Generate key in format: entity-type/entity-id/year/month/uuid-filename
	now := time.Now()
	if !doc.CreatedAt.IsZero() {
		now = doc.CreatedAt
	}

	// Use existing filename from path
	filename := filepath.Base(doc.FilePath)

	return fmt.Sprintf(
		"%s/%d/%d/%02d/%s",
		doc.EntityType,
		doc.EntityID,
		now.Year(),
		now.Month(),
		filename,
	)
}

func printSummary(stats MigrationStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Println()
	log.Println("=== Migration Summary ===")
	log.Printf("Total documents: %d\n", stats.Total)
	log.Printf("Migrated:        %d ✓\n", stats.Migrated)
	log.Printf("Skipped:         %d ℹ️\n", stats.Skipped)
	log.Printf("Failed:          %d ❌\n", stats.Failed)
	log.Printf("Duration:        %s\n", duration.Round(time.Second))

	if stats.Failed > 0 {
		log.Println()
		log.Printf("⚠️  Migration completed with %d failures.\n", stats.Failed)
		log.Println("Review the errors above and re-run the migration for failed documents.")
		os.Exit(1)
	}

	if stats.Migrated == 0 && stats.Skipped == stats.Total {
		log.Println()
		log.Println("ℹ️  All documents were already migrated.")
	} else {
		log.Println()
		log.Println("✓ Migration completed successfully!")
	}

	log.Println()
	log.Println("Next steps:")
	log.Println("  1. Verify migration: ./migrate-docs --verify")
	log.Println("  2. Test document downloads via CLI and web")
	log.Println("  3. Update production config to use MinIO")
	log.Println("  4. Backup original files before cleanup")
}
