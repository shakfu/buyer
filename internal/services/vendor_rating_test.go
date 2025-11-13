package services

import (
	"testing"
	"time"
)

func TestVendorRatingService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)
	brandService := NewBrandService(cfg.DB)
	productService := NewProductService(cfg.DB)
	quoteService := NewQuoteService(cfg.DB)
	poService := NewPurchaseOrderService(cfg.DB)

	// Setup test data
	vendor, _ := vendorService.Create("Tech Supplier", "USD", "")
	brand, _ := brandService.Create("TechBrand")
	product, _ := productService.Create("Laptop", brand.ID, nil)
	quote, _ := quoteService.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     999.99,
		Currency:  "USD",
	})
	po, _ := poService.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-TEST-001",
		Quantity: 1,
	})

	tests := []struct {
		name    string
		input   CreateVendorRatingInput
		wantErr bool
		errType string
	}{
		{
			name: "valid rating with all fields",
			input: CreateVendorRatingInput{
				VendorID:        vendor.ID,
				PurchaseOrderID: &po.ID,
				PriceRating:     intPtr(5),
				QualityRating:   intPtr(4),
				DeliveryRating:  intPtr(5),
				ServiceRating:   intPtr(4),
				Comments:        "Excellent service",
				RatedBy:         "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid rating without purchase order",
			input: CreateVendorRatingInput{
				VendorID:      vendor.ID,
				PriceRating:   intPtr(3),
				QualityRating: intPtr(4),
				Comments:      "Good overall",
				RatedBy:       "admin",
			},
			wantErr: false,
		},
		{
			name: "valid rating with only one rating field",
			input: CreateVendorRatingInput{
				VendorID:    vendor.ID,
				PriceRating: intPtr(5),
				RatedBy:     "buyer",
			},
			wantErr: false,
		},
		{
			name: "invalid vendor",
			input: CreateVendorRatingInput{
				VendorID:    99999,
				PriceRating: intPtr(5),
			},
			wantErr: true,
			errType: "NotFoundError",
		},
		{
			name: "invalid purchase order",
			input: CreateVendorRatingInput{
				VendorID:        vendor.ID,
				PurchaseOrderID: uintPtr(99999),
				PriceRating:     intPtr(5),
			},
			wantErr: true,
			errType: "NotFoundError",
		},
		{
			name: "rating out of range - too low",
			input: CreateVendorRatingInput{
				VendorID:    vendor.ID,
				PriceRating: intPtr(0),
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "rating out of range - too high",
			input: CreateVendorRatingInput{
				VendorID:      vendor.ID,
				QualityRating: intPtr(6),
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "no ratings provided",
			input: CreateVendorRatingInput{
				VendorID: vendor.ID,
				Comments: "Just a comment",
			},
			wantErr: true,
			errType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := service.Create(tt.input)

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

			if rating == nil {
				t.Error("Create() returned nil rating")
				return
			}

			if rating.VendorID != tt.input.VendorID {
				t.Errorf("VendorID = %v, want %v", rating.VendorID, tt.input.VendorID)
			}

			if rating.Vendor == nil {
				t.Error("Vendor association not loaded")
			}
		})
	}
}

func TestVendorRatingService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")

	// Create a rating
	created, _ := service.Create(CreateVendorRatingInput{
		VendorID:    vendor.ID,
		PriceRating: intPtr(5),
		RatedBy:     "tester",
	})

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing rating",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent rating",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := service.GetByID(tt.id)

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

			if rating.ID != tt.id {
				t.Errorf("ID = %v, want %v", rating.ID, tt.id)
			}
		})
	}
}

func TestVendorRatingService_ListByVendor(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor1, _ := vendorService.Create("Vendor 1", "USD", "")
	vendor2, _ := vendorService.Create("Vendor 2", "USD", "")

	// Create ratings for vendor 1
	service.Create(CreateVendorRatingInput{
		VendorID:    vendor1.ID,
		PriceRating: intPtr(5),
	})
	service.Create(CreateVendorRatingInput{
		VendorID:      vendor1.ID,
		QualityRating: intPtr(4),
	})

	// Create rating for vendor 2
	service.Create(CreateVendorRatingInput{
		VendorID:    vendor2.ID,
		PriceRating: intPtr(3),
	})

	tests := []struct {
		name      string
		vendorID  uint
		wantCount int
	}{
		{
			name:      "vendor with 2 ratings",
			vendorID:  vendor1.ID,
			wantCount: 2,
		},
		{
			name:      "vendor with 1 rating",
			vendorID:  vendor2.ID,
			wantCount: 1,
		},
		{
			name:      "vendor with no ratings",
			vendorID:  99999,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratings, err := service.ListByVendor(tt.vendorID, 0, 0)
			if err != nil {
				t.Errorf("ListByVendor() error: %v", err)
				return
			}

			if len(ratings) != tt.wantCount {
				t.Errorf("ListByVendor() count = %v, want %v", len(ratings), tt.wantCount)
			}
		})
	}
}

func TestVendorRatingService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	created, _ := service.Create(CreateVendorRatingInput{
		VendorID:    vendor.ID,
		PriceRating: intPtr(3),
		RatedBy:     "original",
	})

	tests := []struct {
		name    string
		id      uint
		input   CreateVendorRatingInput
		wantErr bool
	}{
		{
			name: "valid update",
			id:   created.ID,
			input: CreateVendorRatingInput{
				VendorID:      vendor.ID,
				PriceRating:   intPtr(5),
				QualityRating: intPtr(4),
				Comments:      "Updated rating",
				RatedBy:       "updater",
			},
			wantErr: false,
		},
		{
			name: "update non-existent rating",
			id:   99999,
			input: CreateVendorRatingInput{
				VendorID:    vendor.ID,
				PriceRating: intPtr(5),
			},
			wantErr: true,
		},
		{
			name: "update with invalid rating",
			id:   created.ID,
			input: CreateVendorRatingInput{
				VendorID:    vendor.ID,
				PriceRating: intPtr(10),
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

			if updated.PriceRating != nil && tt.input.PriceRating != nil {
				if *updated.PriceRating != *tt.input.PriceRating {
					t.Errorf("PriceRating = %v, want %v", *updated.PriceRating, *tt.input.PriceRating)
				}
			}
		})
	}
}

func TestVendorRatingService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")
	rating, _ := service.Create(CreateVendorRatingInput{
		VendorID:    vendor.ID,
		PriceRating: intPtr(5),
	})

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing rating",
			id:      rating.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent rating",
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
					t.Error("Rating still exists after deletion")
				}
			}
		})
	}
}

func TestVendorRatingService_GetAverageRatings(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)

	vendor, _ := vendorService.Create("Test Vendor", "USD", "")

	// Create multiple ratings
	service.Create(CreateVendorRatingInput{
		VendorID:       vendor.ID,
		PriceRating:    intPtr(5),
		QualityRating:  intPtr(4),
		DeliveryRating: intPtr(5),
		ServiceRating:  intPtr(5),
	})
	service.Create(CreateVendorRatingInput{
		VendorID:       vendor.ID,
		PriceRating:    intPtr(3),
		QualityRating:  intPtr(4),
		DeliveryRating: intPtr(3),
		ServiceRating:  intPtr(3),
	})

	averages, err := service.GetAverageRatings(vendor.ID)
	if err != nil {
		t.Fatalf("GetAverageRatings() error: %v", err)
	}

	// Expected averages
	expectedPrice := 4.0    // (5 + 3) / 2
	expectedQuality := 4.0  // (4 + 4) / 2
	expectedDelivery := 4.0 // (5 + 3) / 2
	expectedService := 4.0  // (5 + 3) / 2
	expectedOverall := 4.0  // (5+4+5+5 + 3+4+3+3) / 8

	if averages["price"] != expectedPrice {
		t.Errorf("Average price = %v, want %v", averages["price"], expectedPrice)
	}
	if averages["quality"] != expectedQuality {
		t.Errorf("Average quality = %v, want %v", averages["quality"], expectedQuality)
	}
	if averages["delivery"] != expectedDelivery {
		t.Errorf("Average delivery = %v, want %v", averages["delivery"], expectedDelivery)
	}
	if averages["service"] != expectedService {
		t.Errorf("Average service = %v, want %v", averages["service"], expectedService)
	}
	if averages["overall"] != expectedOverall {
		t.Errorf("Average overall = %v, want %v", averages["overall"], expectedOverall)
	}
	if averages["count"] != 2.0 {
		t.Errorf("Count = %v, want 2", averages["count"])
	}
}

func TestVendorRatingService_VendorPurchaseOrderMismatch(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()
	service := NewVendorRatingService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)
	brandService := NewBrandService(cfg.DB)
	productService := NewProductService(cfg.DB)
	quoteService := NewQuoteService(cfg.DB)
	poService := NewPurchaseOrderService(cfg.DB)

	// Create two vendors
	vendor1, _ := vendorService.Create("Vendor 1", "USD", "")
	vendor2, _ := vendorService.Create("Vendor 2", "USD", "")

	// Create PO for vendor1
	brand, _ := brandService.Create("Brand")
	product, _ := productService.Create("Product", brand.ID, nil)
	quote, _ := quoteService.Create(CreateQuoteInput{
		VendorID:  vendor1.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})
	po, _ := poService.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-MISMATCH",
		Quantity: 1,
	})

	// Try to create rating for vendor2 with vendor1's PO
	_, err := service.Create(CreateVendorRatingInput{
		VendorID:        vendor2.ID,
		PurchaseOrderID: &po.ID,
		PriceRating:     intPtr(5),
	})

	if err == nil {
		t.Error("Expected error for vendor/PO mismatch, got none")
	}

	_, isValidationError := err.(*ValidationError)
	if !isValidationError {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func uintPtr(u uint) *uint {
	return &u
}

func timePtr(t time.Time) *time.Time {
	return &t
}
