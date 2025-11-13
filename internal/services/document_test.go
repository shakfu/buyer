package services

import (
	"testing"
)

func TestDocumentService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	// Create a vendor to attach documents to
	vendor, _ := vendorService.Create("Test Vendor", "USD", "")

	tests := []struct {
		name    string
		input   CreateDocumentInput
		wantErr bool
		errType string
	}{
		{
			name: "valid document with all fields",
			input: CreateDocumentInput{
				EntityType:  "vendor",
				EntityID:    vendor.ID,
				FileName:    "contract.pdf",
				FileType:    "pdf",
				FileSize:    1024000,
				FilePath:    "/docs/vendors/contract.pdf",
				Description: "Vendor contract",
				UploadedBy:  "admin@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid document with minimal fields",
			input: CreateDocumentInput{
				EntityType: "vendor",
				EntityID:   vendor.ID,
				FileName:   "invoice.pdf",
				FilePath:   "/docs/invoice.pdf",
			},
			wantErr: false,
		},
		{
			name: "missing entity type",
			input: CreateDocumentInput{
				EntityID: vendor.ID,
				FileName: "test.pdf",
				FilePath: "/docs/test.pdf",
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "missing file name",
			input: CreateDocumentInput{
				EntityType: "vendor",
				EntityID:   vendor.ID,
				FilePath:   "/docs/test.pdf",
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "missing file path",
			input: CreateDocumentInput{
				EntityType: "vendor",
				EntityID:   vendor.ID,
				FileName:   "test.pdf",
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "invalid entity type",
			input: CreateDocumentInput{
				EntityType: "invalid_type",
				EntityID:   vendor.ID,
				FileName:   "test.pdf",
				FilePath:   "/docs/test.pdf",
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "zero entity ID",
			input: CreateDocumentInput{
				EntityType: "vendor",
				EntityID:   0,
				FileName:   "test.pdf",
				FilePath:   "/docs/test.pdf",
			},
			wantErr: true,
			errType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := service.Create(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}
				if tt.errType != "" {
					// Basic type check - could be enhanced
					_ = err
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error: %v", err)
				return
			}

			if doc == nil {
				t.Error("Create() returned nil document")
				return
			}

			if doc.EntityType != tt.input.EntityType {
				t.Errorf("EntityType = %v, want %v", doc.EntityType, tt.input.EntityType)
			}
			if doc.EntityID != tt.input.EntityID {
				t.Errorf("EntityID = %v, want %v", doc.EntityID, tt.input.EntityID)
			}
			if doc.FileName != tt.input.FileName {
				t.Errorf("FileName = %v, want %v", doc.FileName, tt.input.FileName)
			}
			if doc.FilePath != tt.input.FilePath {
				t.Errorf("FilePath = %v, want %v", doc.FilePath, tt.input.FilePath)
			}
		})
	}
}

func TestDocumentService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	doc, _ := service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "test.pdf",
		FilePath:   "/docs/test.pdf",
	})

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing document",
			id:      doc.ID,
			wantErr: false,
		},
		{
			name:    "non-existent document",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error: %v", err)
				return
			}

			if result.ID != tt.id {
				t.Errorf("ID = %v, want %v", result.ID, tt.id)
			}
		})
	}
}

func TestDocumentService_ListByEntity(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor1, _ := vendorService.Create("Vendor 1", "USD", "")
	vendor2, _ := vendorService.Create("Vendor 2", "USD", "")

	// Create documents for vendor 1
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor1.ID,
		FileName:   "doc1.pdf",
		FilePath:   "/docs/doc1.pdf",
	})
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor1.ID,
		FileName:   "doc2.pdf",
		FilePath:   "/docs/doc2.pdf",
	})

	// Create document for vendor 2
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor2.ID,
		FileName:   "doc3.pdf",
		FilePath:   "/docs/doc3.pdf",
	})

	tests := []struct {
		name       string
		entityType string
		entityID   uint
		wantCount  int
	}{
		{
			name:       "vendor with 2 documents",
			entityType: "vendor",
			entityID:   vendor1.ID,
			wantCount:  2,
		},
		{
			name:       "vendor with 1 document",
			entityType: "vendor",
			entityID:   vendor2.ID,
			wantCount:  1,
		},
		{
			name:       "vendor with no documents",
			entityType: "vendor",
			entityID:   99999,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := service.ListByEntity(tt.entityType, tt.entityID)
			if err != nil {
				t.Errorf("ListByEntity() error: %v", err)
				return
			}

			if len(docs) != tt.wantCount {
				t.Errorf("ListByEntity() count = %v, want %v", len(docs), tt.wantCount)
			}
		})
	}
}

func TestDocumentService_ListByEntityType(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)
	brandService := NewBrandService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	brand, _ := brandService.Create("Test Brand")

	// Create vendor documents
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "vendor_doc1.pdf",
		FilePath:   "/docs/vendor_doc1.pdf",
	})
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "vendor_doc2.pdf",
		FilePath:   "/docs/vendor_doc2.pdf",
	})

	// Create brand document
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "brand",
		EntityID:   brand.ID,
		FileName:   "brand_doc1.pdf",
		FilePath:   "/docs/brand_doc1.pdf",
	})

	tests := []struct {
		name       string
		entityType string
		wantCount  int
	}{
		{
			name:       "vendor documents",
			entityType: "vendor",
			wantCount:  2,
		},
		{
			name:       "brand documents",
			entityType: "brand",
			wantCount:  1,
		},
		{
			name:       "no documents for type",
			entityType: "project",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := service.ListByEntityType(tt.entityType, 0, 0)
			if err != nil {
				t.Errorf("ListByEntityType() error: %v", err)
				return
			}

			if len(docs) != tt.wantCount {
				t.Errorf("ListByEntityType() count = %v, want %v", len(docs), tt.wantCount)
			}
		})
	}
}

func TestDocumentService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	doc, _ := service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "original.pdf",
		FilePath:   "/docs/original.pdf",
	})

	tests := []struct {
		name    string
		id      uint
		input   CreateDocumentInput
		wantErr bool
	}{
		{
			name: "valid update",
			id:   doc.ID,
			input: CreateDocumentInput{
				FileName:    "updated.pdf",
				FilePath:    "/docs/updated.pdf",
				Description: "Updated document",
			},
			wantErr: false,
		},
		{
			name: "update non-existent document",
			id:   99999,
			input: CreateDocumentInput{
				FileName: "test.pdf",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := service.Update(tt.id, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
				return
			}

			if tt.input.FileName != "" && updated.FileName != tt.input.FileName {
				t.Errorf("FileName = %v, want %v", updated.FileName, tt.input.FileName)
			}
		})
	}
}

func TestDocumentService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	doc, _ := service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "test.pdf",
		FilePath:   "/docs/test.pdf",
	})

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing document",
			id:      doc.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent document",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error: %v", err)
			}

			// Verify deletion
			if !tt.wantErr {
				_, err := service.GetByID(tt.id)
				if err == nil {
					t.Error("Document still exists after deletion")
				}
			}
		})
	}
}

func TestDocumentService_DeleteByEntity(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")

	// Create multiple documents for the vendor
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "doc1.pdf",
		FilePath:   "/docs/doc1.pdf",
	})
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "doc2.pdf",
		FilePath:   "/docs/doc2.pdf",
	})

	err := service.DeleteByEntity("vendor", vendor.ID)
	if err != nil {
		t.Errorf("DeleteByEntity() error: %v", err)
	}

	// Verify all documents were deleted
	docs, _ := service.ListByEntity("vendor", vendor.ID)
	if len(docs) != 0 {
		t.Errorf("DeleteByEntity() did not delete all documents, found %d", len(docs))
	}
}

func TestDocumentService_Count(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")

	// Initially should be 0
	count, _ := service.Count()
	if count != 0 {
		t.Errorf("Initial count = %v, want 0", count)
	}

	// Create documents
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "doc1.pdf",
		FilePath:   "/docs/doc1.pdf",
	})
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor.ID,
		FileName:   "doc2.pdf",
		FilePath:   "/docs/doc2.pdf",
	})

	count, _ = service.Count()
	if count != 2 {
		t.Errorf("Count = %v, want 2", count)
	}
}

func TestDocumentService_CountByEntity(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()
	service := NewDocumentService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor1, _ := vendorService.Create("Vendor 1", "USD", "")
	vendor2, _ := vendorService.Create("Vendor 2", "USD", "")

	// Create documents for vendor 1
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor1.ID,
		FileName:   "doc1.pdf",
		FilePath:   "/docs/doc1.pdf",
	})
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor1.ID,
		FileName:   "doc2.pdf",
		FilePath:   "/docs/doc2.pdf",
	})

	// Create document for vendor 2
	_, _ = 	service.Create(CreateDocumentInput{
		EntityType: "vendor",
		EntityID:   vendor2.ID,
		FileName:   "doc3.pdf",
		FilePath:   "/docs/doc3.pdf",
	})

	count1, _ := service.CountByEntity("vendor", vendor1.ID)
	if count1 != 2 {
		t.Errorf("CountByEntity(vendor1) = %v, want 2", count1)
	}

	count2, _ := service.CountByEntity("vendor", vendor2.ID)
	if count2 != 1 {
		t.Errorf("CountByEntity(vendor2) = %v, want 1", count2)
	}
}
