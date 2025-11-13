package services

import (
	"testing"
)

func TestProductService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	// Create a brand first
	brand, err := brandSvc.Create("Intel")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

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
			product, err := productSvc.Create(tt.pName, tt.brandID, nil)

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
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("AMD")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("Ryzen 9", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

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
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("NVIDIA")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("RTX 3080", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

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
			result, err := productSvc.Update(tt.id, tt.newName, nil)

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

func TestProductService_GetByName(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("Apple")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	_, err = productSvc.Create("MacBook Pro", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	tests := []struct {
		name        string
		productName string
		wantErr     bool
	}{
		{
			name:        "existing product",
			productName: "MacBook Pro",
			wantErr:     false,
		},
		{
			name:        "non-existent product",
			productName: "NonExistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := productSvc.GetByName(tt.productName)

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

			if result.Name != tt.productName {
				t.Errorf("GetByName() Name = %v, want %v", result.Name, tt.productName)
			}
		})
	}
}

func TestProductService_List(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("Sony")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	_, err = productSvc.Create("PlayStation 5", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = productSvc.Create("PlayStation 4", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = productSvc.Create("PlayStation VR", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all products",
			limit:     0,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited products",
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
			products, err := productSvc.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(products) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(products), tt.wantCount)
			}
		})
	}
}

func TestProductService_ListByBrand(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("Microsoft")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	_, err = productSvc.Create("Surface Pro", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = productSvc.Create("Surface Laptop", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	products, err := productSvc.ListByBrand(brand.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}
}

func TestProductService_ListBySpecification(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)

	brand, err := brandSvc.Create("Samsung")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	spec, err := specSvc.Create("Smartphone", "Mobile phone device")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	_, err = productSvc.Create("Galaxy S21", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = productSvc.Create("Galaxy S22", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	products, err := productSvc.ListBySpecification(spec.ID)
	if err != nil {
		t.Errorf("ListBySpecification() error = %v", err)
		return
	}

	if len(products) != 2 {
		t.Errorf("ListBySpecification() count = %v, want 2", len(products))
	}
}

func TestProductService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("LG")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("OLED TV", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		errType interface{}
	}{
		{
			name:    "delete existing product",
			id:      product.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent product",
			id:      9999,
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := productSvc.Delete(tt.id)

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
			_, err = productSvc.GetByID(tt.id)
			if err == nil {
				t.Error("Product should be deleted but still exists")
			}
		})
	}
}

func TestProductService_Count(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)

	brand, err := brandSvc.Create("HP")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	_, err = productSvc.Create("EliteBook", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = productSvc.Create("ProBook", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	count, err := productSvc.Count()
	if err != nil {
		t.Errorf("Count() error = %v", err)
		return
	}

	if count != 2 {
		t.Errorf("Count() = %v, want 2", count)
	}
}
