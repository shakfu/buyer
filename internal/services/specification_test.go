package services

import (
	"testing"
)

func TestSpecificationService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	tests := []struct {
		name        string
		specName    string
		description string
		wantErr     bool
		errType     interface{}
	}{
		{
			name:        "valid specification with description",
			specName:    "Laptop",
			description: "Portable computer device",
			wantErr:     false,
		},
		{
			name:        "valid specification without description",
			specName:    "Smartphone",
			description: "",
			wantErr:     false,
		},
		{
			name:        "empty name",
			specName:    "",
			description: "Test",
			wantErr:     true,
			errType:     &ValidationError{},
		},
		{
			name:        "whitespace name",
			specName:    "   ",
			description: "Test",
			wantErr:     true,
			errType:     &ValidationError{},
		},
		{
			name:        "duplicate specification",
			specName:    "Laptop", // Already created in first test
			description: "Different description",
			wantErr:     true,
			errType:     &DuplicateError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := service.Create(tt.specName, tt.description)

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

			if spec.Name != tt.specName {
				t.Errorf("Create() name = %v, want %v", spec.Name, tt.specName)
			}
			if spec.Description != tt.description {
				t.Errorf("Create() description = %v, want %v", spec.Description, tt.description)
			}
			if spec.ID == 0 {
				t.Error("Create() ID should be set")
			}
		})
	}
}

func TestSpecificationService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	// Create a test specification
	spec, err := service.Create("Test Spec", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create test specification: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing specification",
			id:      spec.ID,
			wantErr: false,
		},
		{
			name:    "non-existent specification",
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

func TestSpecificationService_GetByName(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	// Create a test specification
	_, err := service.Create("Monitor", "Display device")
	if err != nil {
		t.Fatalf("Failed to create test specification: %v", err)
	}

	tests := []struct {
		name     string
		specName string
		wantErr  bool
	}{
		{
			name:     "existing specification",
			specName: "Monitor",
			wantErr:  false,
		},
		{
			name:     "non-existent specification",
			specName: "NonExistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByName(tt.specName)

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

			if result.Name != tt.specName {
				t.Errorf("GetByName() Name = %v, want %v", result.Name, tt.specName)
			}
		})
	}
}

func TestSpecificationService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	// Create test specifications
	spec1, err := service.Create("Original", "Original description")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	_, err = service.Create("Other", "Other description")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	tests := []struct {
		name           string
		id             uint
		newName        string
		newDescription string
		wantErr        bool
		errType        interface{}
	}{
		{
			name:           "valid update",
			id:             spec1.ID,
			newName:        "Updated",
			newDescription: "Updated description",
			wantErr:        false,
		},
		{
			name:           "empty name",
			id:             spec1.ID,
			newName:        "",
			newDescription: "Test",
			wantErr:        true,
			errType:        &ValidationError{},
		},
		{
			name:           "duplicate name",
			id:             spec1.ID,
			newName:        "Other", // Already exists as spec2
			newDescription: "Test",
			wantErr:        true,
			errType:        &DuplicateError{},
		},
		{
			name:           "non-existent specification",
			id:             9999,
			newName:        "Test",
			newDescription: "Test",
			wantErr:        true,
			errType:        &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Update(tt.id, tt.newName, tt.newDescription)

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
			if result.Description != tt.newDescription {
				t.Errorf("Update() description = %v, want %v", result.Description, tt.newDescription)
			}
		})
	}
}

func TestSpecificationService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	// Create a test specification
	spec, err := service.Create("ToDelete", "Will be deleted")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "delete existing specification",
			id:      spec.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent specification",
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
				t.Error("Specification should be deleted but still exists")
			}
		})
	}
}

func TestSpecificationService_List(t *testing.T) {
	cfg := setupTestDB(t)
	service := NewSpecificationService(cfg.DB)

	// Create test specifications
	_, err := service.Create("Spec A", "Description A")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	_, err = service.Create("Spec B", "Description B")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	_, err = service.Create("Spec C", "Description C")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all specifications",
			limit:     0,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited specifications",
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
			specs, err := service.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(specs) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(specs), tt.wantCount)
			}
		})
	}
}

func TestSpecificationService_WithProducts(t *testing.T) {
	cfg := setupTestDB(t)
	specService := NewSpecificationService(cfg.DB)
	brandService := NewBrandService(cfg.DB)
	productService := NewProductService(cfg.DB)

	// Create a specification
	spec, err := specService.Create("Electronics", "Electronic devices")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Create a brand
	brand, err := brandService.Create("TestBrand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	// Create a product with this specification
	_, err = productService.Create("TestProduct", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Get specification with preloaded products
	result, err := specService.GetByID(spec.ID)
	if err != nil {
		t.Fatalf("Failed to get specification: %v", err)
	}

	if len(result.Products) != 1 {
		t.Errorf("GetByID() should preload products, got %d products", len(result.Products))
	}
}
