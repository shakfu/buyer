package services

import (
	"testing"
	"time"
)

func TestQuoteService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, _ := brandSvc.Create("Canon")
	product, _ := productSvc.Create("EOS R5", brand.ID, nil)
	vendor, _ := vendorSvc.Create("B&H Photo", "USD", "SAVE10")

	// Create EUR vendor and forex rate
	eurVendor, _ := vendorSvc.Create("European Camera", "EUR", "")
	forexSvc.Create("EUR", "USD", 1.20, time.Now())

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
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, _ := brandSvc.Create("Nikon")
	product, _ := productSvc.Create("Z9", brand.ID, nil)
	vendor1, _ := vendorSvc.Create("Adorama", "USD", "")
	vendor2, _ := vendorSvc.Create("Amazon", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quotes
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     5499.99,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     5299.99,
		Currency:  "USD",
	})

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
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, _ := brandSvc.Create("Sony")
	product1, _ := productSvc.Create("A7 IV", brand.ID, nil)
	product2, _ := productSvc.Create("A7R V", brand.ID, nil)
	vendor, _ := vendorSvc.Create("Focus Camera", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quotes
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     2499.99,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     2399.99,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     3899.99,
		Currency:  "USD",
	})

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
	defer cfg.Close()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Setup test data
	brand, _ := brandSvc.Create("Fujifilm")
	product, _ := productSvc.Create("X-T5", brand.ID, nil)
	vendor1, _ := vendorSvc.Create("KEH Camera", "USD", "")
	vendor2, _ := vendorSvc.Create("MPB", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quotes
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     1699.99,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     1649.99,
		Currency:  "USD",
	})

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
