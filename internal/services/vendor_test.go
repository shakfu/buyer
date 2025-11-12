package services

import (
	"testing"
)

func TestVendorService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	tests := []struct {
		name         string
		vendorName   string
		currency     string
		discountCode string
		wantErr      bool
		errType      interface{}
		wantCurrency string
	}{
		{
			name:         "valid vendor with USD",
			vendorName:   "Vendor A",
			currency:     "USD",
			discountCode: "DISCOUNT10",
			wantErr:      false,
			wantCurrency: "USD",
		},
		{
			name:         "valid vendor with EUR",
			vendorName:   "Vendor B",
			currency:     "EUR",
			discountCode: "",
			wantErr:      false,
			wantCurrency: "EUR",
		},
		{
			name:         "valid vendor with lowercase currency",
			vendorName:   "Vendor C",
			currency:     "gbp",
			discountCode: "",
			wantErr:      false,
			wantCurrency: "GBP",
		},
		{
			name:         "valid vendor with empty currency defaults to USD",
			vendorName:   "Vendor D",
			currency:     "",
			discountCode: "",
			wantErr:      false,
			wantCurrency: "USD",
		},
		{
			name:         "empty name",
			vendorName:   "",
			currency:     "USD",
			discountCode: "",
			wantErr:      true,
			errType:      &ValidationError{},
		},
		{
			name:         "whitespace name",
			vendorName:   "   ",
			currency:     "USD",
			discountCode: "",
			wantErr:      true,
			errType:      &ValidationError{},
		},
		{
			name:         "invalid currency length",
			vendorName:   "Vendor E",
			currency:     "US",
			discountCode: "",
			wantErr:      true,
			errType:      &ValidationError{},
		},
		{
			name:         "duplicate vendor",
			vendorName:   "Vendor A", // Already created in first test
			currency:     "USD",
			discountCode: "",
			wantErr:      true,
			errType:      &DuplicateError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor, err := service.Create(tt.vendorName, tt.currency, tt.discountCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() error = nil, wantErr %v", tt.wantErr)
					return
				}
				// Check error type
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Create() error type = %T, want ValidationError", err)
					}
				case *DuplicateError:
					if _, ok := err.(*DuplicateError); !ok {
						t.Errorf("Create() error type = %T, want DuplicateError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if vendor.Name != tt.vendorName {
				t.Errorf("Create() name = %v, want %v", vendor.Name, tt.vendorName)
			}
			if vendor.Currency != tt.wantCurrency {
				t.Errorf("Create() currency = %v, want %v", vendor.Currency, tt.wantCurrency)
			}
			if vendor.DiscountCode != tt.discountCode {
				t.Errorf("Create() discountCode = %v, want %v", vendor.DiscountCode, tt.discountCode)
			}
			if vendor.ID == 0 {
				t.Error("Create() ID should be set")
			}
		})
	}
}

func TestVendorService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create a test vendor
	vendor, err := service.Create("Test Vendor", "USD", "CODE123")
	if err != nil {
		t.Fatalf("Failed to create test vendor: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing vendor",
			id:      vendor.ID,
			wantErr: false,
		},
		{
			name:    "non-existent vendor",
			id:      9999,
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() error = nil, wantErr true")
					return
				}
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("GetByID() error type = %T, want NotFoundError", err)
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error = %v", err)
				return
			}

			if result.ID != tt.id {
				t.Errorf("GetByID() ID = %v, want %v", result.ID, tt.id)
			}
		})
	}
}

func TestVendorService_GetByName(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create a test vendor
	_, err := service.Create("Named Vendor", "EUR", "")
	if err != nil {
		t.Fatalf("Failed to create test vendor: %v", err)
	}

	tests := []struct {
		name       string
		vendorName string
		wantErr    bool
	}{
		{
			name:       "existing vendor",
			vendorName: "Named Vendor",
			wantErr:    false,
		},
		{
			name:       "non-existent vendor",
			vendorName: "NonExistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByName(tt.vendorName)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByName() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("GetByName() unexpected error = %v", err)
				return
			}

			if result.Name != tt.vendorName {
				t.Errorf("GetByName() Name = %v, want %v", result.Name, tt.vendorName)
			}
		})
	}
}

func TestVendorService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create test vendors
	vendor1, err := service.Create("Original Vendor", "USD", "CODE1")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = service.Create("Other Vendor", "EUR", "CODE2")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		newName string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid update",
			id:      vendor1.ID,
			newName: "Updated Vendor",
			wantErr: false,
		},
		{
			name:    "empty name",
			id:      vendor1.ID,
			newName: "",
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:    "duplicate name",
			id:      vendor1.ID,
			newName: "Other Vendor", // Already exists as vendor2
			wantErr: true,
			errType: &DuplicateError{},
		},
		{
			name:    "non-existent vendor",
			id:      9999,
			newName: "Test",
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Update(tt.id, tt.newName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.wantErr)
					return
				}
				// Check error type
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Update() error type = %T, want ValidationError", err)
					}
				case *DuplicateError:
					if _, ok := err.(*DuplicateError); !ok {
						t.Errorf("Update() error type = %T, want DuplicateError", err)
					}
				case *NotFoundError:
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("Update() error type = %T, want NotFoundError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error = %v", err)
				return
			}

			if result.Name != tt.newName {
				t.Errorf("Update() name = %v, want %v", result.Name, tt.newName)
			}
		})
	}
}

func TestVendorService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create a test vendor
	vendor, err := service.Create("ToDelete", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "delete existing vendor",
			id:      vendor.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent vendor",
			id:      9999,
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() error = nil, wantErr true")
					return
				}
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("Delete() error type = %T, want NotFoundError", err)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}

			// Verify it's actually deleted
			_, err = service.GetByID(tt.id)
			if err == nil {
				t.Error("Vendor should be deleted but still exists")
			}
		})
	}
}

func TestVendorService_List(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create test vendors
	_, err := service.Create("Vendor 1", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = service.Create("Vendor 2", "EUR", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = service.Create("Vendor 3", "GBP", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all vendors",
			limit:     0,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited vendors",
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "with offset",
			limit:     0,
			offset:    1,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendors, err := service.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(vendors) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(vendors), tt.wantCount)
			}
		})
	}
}

func TestVendorService_AddBrand(t *testing.T) {
	cfg := setupTestDB(t)
	vendorService := NewVendorService(cfg.DB)
	brandService := NewBrandService(cfg.DB)

	// Create test vendor and brand
	vendor, err := vendorService.Create("Test Vendor", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	brand, err := brandService.Create("Test Brand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	tests := []struct {
		name     string
		vendorID uint
		brandID  uint
		wantErr  bool
		errType  interface{}
	}{
		{
			name:     "add brand to vendor",
			vendorID: vendor.ID,
			brandID:  brand.ID,
			wantErr:  false,
		},
		{
			name:     "non-existent vendor",
			vendorID: 9999,
			brandID:  brand.ID,
			wantErr:  true,
			errType:  &NotFoundError{},
		},
		{
			name:     "non-existent brand",
			vendorID: vendor.ID,
			brandID:  9999,
			wantErr:  true,
			errType:  &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vendorService.AddBrand(tt.vendorID, tt.brandID)

			if tt.wantErr {
				if err == nil {
					t.Error("AddBrand() error = nil, wantErr true")
					return
				}
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("AddBrand() error type = %T, want NotFoundError", err)
				}
				return
			}

			if err != nil {
				t.Errorf("AddBrand() unexpected error = %v", err)
			}

			// Verify brand was added
			result, err := vendorService.GetByID(tt.vendorID)
			if err != nil {
				t.Fatalf("Failed to get vendor: %v", err)
			}

			if len(result.Brands) == 0 {
				t.Error("AddBrand() brand not added to vendor")
			}
		})
	}
}

func TestVendorService_RemoveBrand(t *testing.T) {
	cfg := setupTestDB(t)
	vendorService := NewVendorService(cfg.DB)
	brandService := NewBrandService(cfg.DB)

	// Create test vendor and brand
	vendor, err := vendorService.Create("Test Vendor 2", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	brand, err := brandService.Create("Test Brand 2")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	// Add brand to vendor first
	err = vendorService.AddBrand(vendor.ID, brand.ID)
	if err != nil {
		t.Fatalf("Failed to add brand: %v", err)
	}

	tests := []struct {
		name     string
		vendorID uint
		brandID  uint
		wantErr  bool
	}{
		{
			name:     "remove brand from vendor",
			vendorID: vendor.ID,
			brandID:  brand.ID,
			wantErr:  false,
		},
		{
			name:     "non-existent vendor",
			vendorID: 9999,
			brandID:  brand.ID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vendorService.RemoveBrand(tt.vendorID, tt.brandID)

			if tt.wantErr {
				if err == nil {
					t.Error("RemoveBrand() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("RemoveBrand() unexpected error = %v", err)
			}
		})
	}
}

func TestVendorService_Count(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewVendorService(cfg.DB)

	// Create test vendors
	_, err := service.Create("Count Vendor 1", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = service.Create("Count Vendor 2", "EUR", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	count, err := service.Count()
	if err != nil {
		t.Errorf("Count() error = %v", err)
		return
	}

	if count != 2 {
		t.Errorf("Count() = %v, want 2", count)
	}
}
