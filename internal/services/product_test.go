package services

import (
	"testing"
)

func TestProductService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	// Create a brand first
	brand, _ := brandSvc.Create("Intel")

	tests := []struct {
		name    string
		pName   string
		brandID uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid product",
			pName:   "Core i7",
			brandID: brand.ID,
			wantErr: false,
		},
		{
			name:    "empty name",
			pName:   "",
			brandID: brand.ID,
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:    "invalid brand",
			pName:   "Core i9",
			brandID: 999,
			wantErr: true,
			errType: &NotFoundError{},
		},
		{
			name:    "duplicate product",
			pName:   "Core i7",
			brandID: brand.ID,
			wantErr: true,
			errType: &DuplicateError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := productSvc.Create(tt.pName, tt.brandID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if product.Name != tt.pName {
					t.Errorf("Expected name %s, got %s", tt.pName, product.Name)
				}
				if product.Brand == nil {
					t.Error("Expected brand to be preloaded")
				}
			}
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, _ := brandSvc.Create("AMD")
	product, _ := productSvc.Create("Ryzen 9", brand.ID)

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing product",
			id:      product.ID,
			wantErr: false,
		},
		{
			name:    "non-existent product",
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := productSvc.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result.ID != tt.id {
					t.Errorf("Expected ID %d, got %d", tt.id, result.ID)
				}
				if result.Brand == nil {
					t.Error("Expected brand to be preloaded")
				}
			}
		})
	}
}

func TestProductService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, _ := brandSvc.Create("NVIDIA")
	product, _ := productSvc.Create("RTX 3080", brand.ID)

	tests := []struct {
		name    string
		id      uint
		newName string
		wantErr bool
	}{
		{
			name:    "valid update",
			id:      product.ID,
			newName: "RTX 4080",
			wantErr: false,
		},
		{
			name:    "empty name",
			id:      product.ID,
			newName: "",
			wantErr: true,
		},
		{
			name:    "non-existent product",
			id:      999,
			newName: "NewName",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := productSvc.Update(tt.id, tt.newName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
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

func TestProductService_ListByBrand(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, _ := brandSvc.Create("Microsoft")
	productSvc.Create("Surface Pro", brand.ID)
	productSvc.Create("Surface Laptop", brand.ID)

	products, err := productSvc.ListByBrand(brand.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}
}
