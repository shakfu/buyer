package services

import (
	"testing"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
)

func setupTestDB(t *testing.T) *config.Config {
	cfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run migrations for all models
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.Product{},
		&models.Quote{},
		&models.Forex{},
		&models.Requisition{},
		&models.RequisitionItem{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return cfg
}

func TestBrandService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewBrandService(cfg.DB)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid brand",
			input:   "Apple",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:    "whitespace name",
			input:   "   ",
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:    "duplicate brand",
			input:   "Apple",
			wantErr: true,
			errType: &DuplicateError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brand, err := svc.Create(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				// Type assertion to check error type
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T", err)
					}
				case *DuplicateError:
					if _, ok := err.(*DuplicateError); !ok {
						t.Errorf("Expected DuplicateError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if brand == nil {
					t.Error("Expected brand but got nil")
					return
				}
				if brand.Name != tt.input {
					t.Errorf("Expected name %s, got %s", tt.input, brand.Name)
				}
			}
		})
	}
}

func TestBrandService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewBrandService(cfg.DB)

	// Create a brand
	brand, err := svc.Create("Samsung")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing brand",
			id:      brand.ID,
			wantErr: false,
		},
		{
			name:    "non-existent brand",
			id:      999,
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("Expected NotFoundError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected brand but got nil")
					return
				}
				if result.ID != tt.id {
					t.Errorf("Expected ID %d, got %d", tt.id, result.ID)
				}
			}
		})
	}
}

func TestBrandService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewBrandService(cfg.DB)

	// Create brands
	brand1, _ := svc.Create("Sony")
	svc.Create("LG")

	tests := []struct {
		name    string
		id      uint
		newName string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid update",
			id:      brand1.ID,
			newName: "Sony Corporation",
			wantErr: false,
		},
		{
			name:    "empty name",
			id:      brand1.ID,
			newName: "",
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:    "duplicate name",
			id:      brand1.ID,
			newName: "LG",
			wantErr: true,
			errType: &DuplicateError{},
		},
		{
			name:    "non-existent brand",
			id:      999,
			newName: "NewName",
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Update(tt.id, tt.newName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result.Name != tt.newName {
					t.Errorf("Expected name %s, got %s", tt.newName, result.Name)
				}
			}
		})
	}
}

func TestBrandService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewBrandService(cfg.DB)

	// Create a brand
	brand, _ := svc.Create("Dell")

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "delete existing brand",
			id:      brand.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent brand",
			id:      999,
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Delete(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBrandService_List(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewBrandService(cfg.DB)

	// Create some brands
	svc.Create("HP")
	svc.Create("Lenovo")
	svc.Create("Asus")

	tests := []struct {
		name     string
		limit    int
		offset   int
		wantMin  int
		wantMax  int
	}{
		{
			name:    "all brands",
			limit:   0,
			offset:  0,
			wantMin: 3,
			wantMax: 3,
		},
		{
			name:    "limited brands",
			limit:   2,
			offset:  0,
			wantMin: 2,
			wantMax: 2,
		},
		{
			name:    "with offset",
			limit:   2,
			offset:  1,
			wantMin: 2,
			wantMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brands, err := svc.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(brands) < tt.wantMin || len(brands) > tt.wantMax {
				t.Errorf("Expected %d-%d brands, got %d", tt.wantMin, tt.wantMax, len(brands))
			}
		})
	}
}
