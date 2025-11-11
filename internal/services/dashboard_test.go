package services

import (
	"testing"
	"time"
)

func TestDashboardService_GetStats(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	dashboardSvc := NewDashboardService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	reqSvc := NewRequisitionService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Test Brand")
	spec, _ := specSvc.Create("Test Spec", "Description")
	product, _ := productSvc.Create("Test Product", brand.ID, &spec.ID)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	reqSvc.Create("Test Req", "Justification", 1000.0, []RequisitionItemInput{})

	// Get stats
	stats, err := dashboardSvc.GetStats()
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
		return
	}

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	if stats.TotalQuotes != 1 {
		t.Errorf("GetStats() TotalQuotes = %v, want 1", stats.TotalQuotes)
	}

	if stats.TotalRequisitions != 1 {
		t.Errorf("GetStats() TotalRequisitions = %v, want 1", stats.TotalRequisitions)
	}

	if stats.TotalVendors != 1 {
		t.Errorf("GetStats() TotalVendors = %v, want 1", stats.TotalVendors)
	}

	if stats.TotalProducts != 1 {
		t.Errorf("GetStats() TotalProducts = %v, want 1", stats.TotalProducts)
	}

	if stats.TotalBrands != 1 {
		t.Errorf("GetStats() TotalBrands = %v, want 1", stats.TotalBrands)
	}

	if stats.TotalSpecifications != 1 {
		t.Errorf("GetStats() TotalSpecifications = %v, want 1", stats.TotalSpecifications)
	}

	if stats.ActiveQuotes != 1 {
		t.Errorf("GetStats() ActiveQuotes = %v, want 1", stats.ActiveQuotes)
	}
}

func TestDashboardService_GetVendorSpending(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	dashboardSvc := NewDashboardService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Vendor Test Brand")
	product, _ := productSvc.Create("Vendor Test Product", brand.ID, nil)
	vendor1, _ := vendorSvc.Create("Vendor 1", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor 2", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quotes for vendor1
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     200.0,
		Currency:  "USD",
	})

	// Create quote for vendor2
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor2.ID,
		ProductID: product.ID,
		Price:     150.0,
		Currency:  "USD",
	})

	// Get vendor spending
	spending, err := dashboardSvc.GetVendorSpending()
	if err != nil {
		t.Errorf("GetVendorSpending() error = %v", err)
		return
	}

	if len(spending) != 2 {
		t.Errorf("GetVendorSpending() count = %v, want 2", len(spending))
	}

	// Should be ordered by total value descending
	if spending[0].TotalValue < spending[1].TotalValue {
		t.Error("GetVendorSpending() should return results ordered by total value descending")
	}

	// Verify vendor1 has higher total
	if spending[0].VendorID == vendor1.ID {
		if spending[0].QuoteCount != 2 {
			t.Errorf("GetVendorSpending() vendor1 QuoteCount = %v, want 2", spending[0].QuoteCount)
		}
		if spending[0].TotalValue != 300.0 {
			t.Errorf("GetVendorSpending() vendor1 TotalValue = %v, want 300.0", spending[0].TotalValue)
		}
		if spending[0].AvgValue != 150.0 {
			t.Errorf("GetVendorSpending() vendor1 AvgValue = %v, want 150.0", spending[0].AvgValue)
		}
	}
}

func TestDashboardService_GetProductPriceComparison(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	dashboardSvc := NewDashboardService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Price Comparison Brand")
	product1, _ := productSvc.Create("Product with Multiple Quotes", brand.ID, nil)
	product2, _ := productSvc.Create("Product with One Quote", brand.ID, nil)
	vendor, _ := vendorSvc.Create("Price Vendor", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create multiple quotes for product1
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     150.0,
		Currency:  "USD",
	})

	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     125.0,
		Currency:  "USD",
	})

	// Create single quote for product2 (should not appear in results)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     200.0,
		Currency:  "USD",
	})

	// Get product price comparison
	comparison, err := dashboardSvc.GetProductPriceComparison()
	if err != nil {
		t.Errorf("GetProductPriceComparison() error = %v", err)
		return
	}

	// Should only return product1 (has > 1 quote)
	if len(comparison) != 1 {
		t.Errorf("GetProductPriceComparison() count = %v, want 1", len(comparison))
	}

	if len(comparison) > 0 {
		result := comparison[0]
		if result.ProductID != product1.ID {
			t.Errorf("GetProductPriceComparison() ProductID = %v, want %v", result.ProductID, product1.ID)
		}
		if result.QuoteCount != 3 {
			t.Errorf("GetProductPriceComparison() QuoteCount = %v, want 3", result.QuoteCount)
		}
		if result.MinPrice != 100.0 {
			t.Errorf("GetProductPriceComparison() MinPrice = %v, want 100.0", result.MinPrice)
		}
		if result.MaxPrice != 150.0 {
			t.Errorf("GetProductPriceComparison() MaxPrice = %v, want 150.0", result.MaxPrice)
		}
		if result.AvgPrice != 125.0 {
			t.Errorf("GetProductPriceComparison() AvgPrice = %v, want 125.0", result.AvgPrice)
		}
	}
}

func TestDashboardService_GetExpiryStats(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	dashboardSvc := NewDashboardService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Expiry Test Brand")
	product, _ := productSvc.Create("Expiry Test Product", brand.ID, nil)
	vendor, _ := vendorSvc.Create("Expiry Vendor", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quote with no expiry
	quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	// Create quote expiring soon (5 days)
	expireSoon := time.Now().AddDate(0, 0, 5)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      100.0,
		Currency:   "USD",
		ValidUntil: &expireSoon,
	})

	// Create quote expiring this month (20 days)
	expireMonth := time.Now().AddDate(0, 0, 20)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      100.0,
		Currency:   "USD",
		ValidUntil: &expireMonth,
	})

	// Create expired quote
	expired := time.Now().AddDate(0, 0, -5)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      100.0,
		Currency:   "USD",
		ValidUntil: &expired,
	})

	// Create valid quote (40 days)
	valid := time.Now().AddDate(0, 0, 40)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      100.0,
		Currency:   "USD",
		ValidUntil: &valid,
	})

	// Get expiry stats
	stats, err := dashboardSvc.GetExpiryStats()
	if err != nil {
		t.Errorf("GetExpiryStats() error = %v", err)
		return
	}

	if stats == nil {
		t.Fatal("GetExpiryStats() returned nil")
	}

	if stats.ExpiringSoon != 1 {
		t.Errorf("GetExpiryStats() ExpiringSoon = %v, want 1", stats.ExpiringSoon)
	}

	if stats.ExpiringMonth != 2 {
		t.Errorf("GetExpiryStats() ExpiringMonth = %v, want 2 (includes ExpiringSoon)", stats.ExpiringMonth)
	}

	if stats.Expired != 1 {
		t.Errorf("GetExpiryStats() Expired = %v, want 1", stats.Expired)
	}

	if stats.Valid != 1 {
		t.Errorf("GetExpiryStats() Valid = %v, want 1", stats.Valid)
	}

	if stats.NoExpiry != 1 {
		t.Errorf("GetExpiryStats() NoExpiry = %v, want 1", stats.NoExpiry)
	}
}

func TestDashboardService_GetRecentQuotes(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	dashboardSvc := NewDashboardService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Recent Test Brand")
	product, _ := productSvc.Create("Recent Test Product", brand.ID, nil)
	vendor, _ := vendorSvc.Create("Recent Vendor", "USD", "")
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create 15 quotes
	for i := 0; i < 15; i++ {
		quoteSvc.Create(CreateQuoteInput{
			VendorID:  vendor.ID,
			ProductID: product.ID,
			Price:     float64(100 + i),
			Currency:  "USD",
		})
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "default limit (10)",
			limit:     0,
			wantCount: 10,
		},
		{
			name:      "custom limit (5)",
			limit:     5,
			wantCount: 5,
		},
		{
			name:      "limit exceeds total",
			limit:     20,
			wantCount: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotes, err := dashboardSvc.GetRecentQuotes(tt.limit)
			if err != nil {
				t.Errorf("GetRecentQuotes() error = %v", err)
				return
			}

			if len(quotes) != tt.wantCount {
				t.Errorf("GetRecentQuotes() count = %v, want %v", len(quotes), tt.wantCount)
			}

			// Verify vendor and product are preloaded
			if len(quotes) > 0 {
				if quotes[0].Vendor == nil {
					t.Error("GetRecentQuotes() should preload Vendor")
				}
				if quotes[0].Product == nil {
					t.Error("GetRecentQuotes() should preload Product")
				}
			}
		})
	}
}
