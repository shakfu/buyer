package services

import (
	"testing"
	"time"
)

func TestQuoteService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, err := brandSvc.Create("Canon")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("EOS R5", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("B&H Photo", "USD", "SAVE10")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	// Create EUR vendor and forex rate
	eurVendor, err := vendorSvc.Create("European Camera", "EUR", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("EUR", "USD", 1.20, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	tests := []struct {
		name      string
		vendorID  uint
		productID uint
		price     float64
		currency  string
		wantErr   bool
	}{
		{
			name:      "valid quote in USD",
			vendorID:  vendor.ID,
			productID: product.ID,
			price:     3899.99,
			currency:  "USD",
			wantErr:   false,
		},
		{
			name:      "valid quote with conversion",
			vendorID:  eurVendor.ID,
			productID: product.ID,
			price:     3500.00,
			currency:  "EUR",
			wantErr:   false,
		},
		{
			name:      "invalid vendor",
			vendorID:  999,
			productID: product.ID,
			price:     100,
			currency:  "USD",
			wantErr:   true,
		},
		{
			name:      "invalid product",
			vendorID:  vendor.ID,
			productID: 999,
			price:     100,
			currency:  "USD",
			wantErr:   true,
		},
		{
			name:      "negative price",
			vendorID:  vendor.ID,
			productID: product.ID,
			price:     -100,
			currency:  "USD",
			wantErr:   true,
		},
		{
			name:      "zero price",
			vendorID:  vendor.ID,
			productID: product.ID,
			price:     0,
			currency:  "USD",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, err := quoteSvc.Create(CreateQuoteInput{
				VendorID:  tt.vendorID,
				ProductID: tt.productID,
				Price:     tt.price,
				Currency:  tt.currency,
				QuoteDate: time.Now(),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if quote.Price != tt.price {
					t.Errorf("Expected price %f, got %f", tt.price, quote.Price)
				}
				if quote.ConvertedPrice <= 0 {
					t.Error("Expected positive converted price")
				}
				if quote.Vendor == nil {
					t.Error("Expected vendor to be preloaded")
				}
				if quote.Product == nil {
					t.Error("Expected product to be preloaded")
				}
			}
		})
	}
}

func TestQuoteService_GetBestQuote(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, err := brandSvc.Create("Nikon")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("Z9", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor1, err := vendorSvc.Create("Adorama", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	vendor2, err := vendorSvc.Create("Amazon", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create quotes
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     5499.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     5299.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// Get best quote
	bestQuote, err := quoteSvc.GetBestQuote(product.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if bestQuote.Price != 5299.99 {
		t.Errorf("Expected best price 5299.99, got %f", bestQuote.Price)
	}

	// Test non-existent product
	_, err = quoteSvc.GetBestQuote(999)
	if err == nil {
		t.Error("Expected error for non-existent product")
	}
}

func TestQuoteService_ListByProduct(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, err := brandSvc.Create("Sony")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product1, err := productSvc.Create("A7 IV", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	product2, err := productSvc.Create("A7R V", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Focus Camera", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create quotes
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     2499.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     2399.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     3899.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// List quotes for product1
	quotes, err := quoteSvc.ListByProduct(product1.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(quotes) != 2 {
		t.Errorf("Expected 2 quotes for product1, got %d", len(quotes))
	}

	// List quotes for product2
	quotes2, err := quoteSvc.ListByProduct(product2.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(quotes2) != 1 {
		t.Errorf("Expected 1 quote for product2, got %d", len(quotes2))
	}
}

func TestQuoteService_ListByVendor(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, err := brandSvc.Create("Fujifilm")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("X-T5", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor1, err := vendorSvc.Create("KEH Camera", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	vendor2, err := vendorSvc.Create("MPB", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create quotes
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     1699.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     1649.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// List quotes for vendor1
	quotes, err := quoteSvc.ListByVendor(vendor1.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(quotes) != 1 {
		t.Errorf("Expected 1 quote for vendor1, got %d", len(quotes))
	}

	if quotes[0].Vendor.ID != vendor1.ID {
		t.Errorf("Expected vendor ID %d, got %d", vendor1.ID, quotes[0].Vendor.ID)
	}
}

func TestQuoteService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Panasonic")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("Lumix S5", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Camera Store", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	quote, err := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     1999.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing quote",
			id:      quote.ID,
			wantErr: false,
		},
		{
			name:    "non-existent quote",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := quoteSvc.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() error = nil, wantErr true")
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
			if result.Vendor == nil {
				t.Error("Expected vendor to be preloaded")
			}
			if result.Product == nil {
				t.Error("Expected product to be preloaded")
			}
		})
	}
}

func TestQuoteService_List(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Olympus")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("OM-1", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Photo Shop", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     2199.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     2099.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     1999.99,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all quotes",
			limit:     0,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited quotes",
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
			quotes, err := quoteSvc.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(quotes) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(quotes), tt.wantCount)
			}
		})
	}
}

func TestQuoteService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Leica")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("M11", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Leica Store", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	quote, err := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     8995.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing quote",
			id:      quote.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent quote",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := quoteSvc.Delete(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}

			// Verify it's actually deleted
			_, err = quoteSvc.GetByID(tt.id)
			if err == nil {
				t.Error("Quote should be deleted but still exists")
			}
		})
	}
}

func TestQuoteService_Count(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Hasselblad")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("X2D 100C", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Hasselblad Store", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     8199.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     7999.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	count, err := quoteSvc.Count()
	if err != nil {
		t.Errorf("Count() error = %v", err)
		return
	}

	if count != 2 {
		t.Errorf("Count() = %v, want 2", count)
	}
}

func TestQuoteService_ListActiveQuotes(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Sigma")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("fp L", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Sigma Store", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create active quote (no expiry)
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     2499.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// Create active quote (future expiry)
	futureDate := time.Now().AddDate(0, 0, 30)
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      2399.00,
		Currency:   "USD",
		ValidUntil: &futureDate,
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// Create expired quote
	pastDate := time.Now().AddDate(0, 0, -1)
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      2299.00,
		Currency:   "USD",
		ValidUntil: &pastDate,
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// List active quotes
	activeQuotes, err := quoteSvc.ListActiveQuotes(0, 0)
	if err != nil {
		t.Errorf("ListActiveQuotes() error = %v", err)
		return
	}

	if len(activeQuotes) != 2 {
		t.Errorf("ListActiveQuotes() count = %v, want 2", len(activeQuotes))
	}
}

func TestQuoteService_CompareQuotesForProduct(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	brand, err := brandSvc.Create("Pentax")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productSvc.Create("K-3 III", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor1, err := vendorSvc.Create("Vendor A", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	vendor2, err := vendorSvc.Create("Vendor B", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     1999.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     1899.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	quotes, err := quoteSvc.CompareQuotesForProduct(product.ID)
	if err != nil {
		t.Errorf("CompareQuotesForProduct() error = %v", err)
		return
	}

	if len(quotes) != 2 {
		t.Errorf("CompareQuotesForProduct() count = %v, want 2", len(quotes))
	}

	// Should be ordered by price ascending
	if quotes[0].ConvertedPrice > quotes[1].ConvertedPrice {
		t.Error("CompareQuotesForProduct() should return quotes ordered by price ascending")
	}
}

func TestQuoteService_CompareQuotesForSpecification(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	spec, err := specSvc.Create("Camera", "Digital camera")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	brand1, err := brandSvc.Create("Brand X")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	brand2, err := brandSvc.Create("Brand Y")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product1, err := productSvc.Create("Camera X1", brand1.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	product2, err := productSvc.Create("Camera Y1", brand2.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Vendor C", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     1500.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     1300.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	quotes, err := quoteSvc.CompareQuotesForSpecification(spec.ID)
	if err != nil {
		t.Errorf("CompareQuotesForSpecification() error = %v", err)
		return
	}

	if len(quotes) != 2 {
		t.Errorf("CompareQuotesForSpecification() count = %v, want 2", len(quotes))
	}
}

func TestQuoteService_GetBestQuoteForSpecification(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	spec, err := specSvc.Create("Lens", "Camera lens")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	brand, err := brandSvc.Create("Brand Z")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product1, err := productSvc.Create("50mm f/1.8", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	product2, err := productSvc.Create("50mm f/1.4", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorSvc.Create("Vendor D", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     199.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     399.00,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	bestQuote, err := quoteSvc.GetBestQuoteForSpecification(spec.ID)
	if err != nil {
		t.Errorf("GetBestQuoteForSpecification() error = %v", err)
		return
	}

	if bestQuote.Price != 199.00 {
		t.Errorf("GetBestQuoteForSpecification() price = %v, want 199.00", bestQuote.Price)
	}

	// Test non-existent specification
	_, err = quoteSvc.GetBestQuoteForSpecification(9999)
	if err == nil {
		t.Error("Expected error for non-existent specification")
	}
}
