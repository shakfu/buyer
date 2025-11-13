package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
)

func setupPurchaseOrderTestDB(t *testing.T) *config.Config {
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
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.PurchaseOrder{},
		&models.Document{},
		&models.VendorRating{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return cfg
}

func TestPurchaseOrderService_Create(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	reqSvc := NewRequisitionService(cfg.DB)
	req, _ := reqSvc.Create("Test Requisition", "", 0, []RequisitionItemInput{})

	poSvc := NewPurchaseOrderService(cfg.DB)

	tests := []struct {
		name    string
		input   CreatePurchaseOrderInput
		wantErr bool
		errType string
	}{
		{
			name: "valid purchase order",
			input: CreatePurchaseOrderInput{
				QuoteID:      quote.ID,
				PONumber:     "PO-001",
				Quantity:     5,
				ShippingCost: 50.0,
				Tax:          25.0,
			},
			wantErr: false,
		},
		{
			name: "valid with requisition",
			input: CreatePurchaseOrderInput{
				QuoteID:       quote.ID,
				RequisitionID: &req.ID,
				PONumber:      "PO-002",
				Quantity:      10,
			},
			wantErr: false,
		},
		{
			name: "empty po number",
			input: CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "",
				Quantity: 5,
			},
			wantErr: true,
			errType: "validation",
		},
		{
			name: "whitespace po number",
			input: CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "   ",
				Quantity: 5,
			},
			wantErr: true,
			errType: "validation",
		},
		{
			name: "zero quantity",
			input: CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "PO-003",
				Quantity: 0,
			},
			wantErr: true,
			errType: "validation",
		},
		{
			name: "negative quantity",
			input: CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "PO-004",
				Quantity: -5,
			},
			wantErr: true,
			errType: "validation",
		},
		{
			name: "duplicate po number",
			input: CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "PO-001", // Already exists
				Quantity: 5,
			},
			wantErr: true,
			errType: "duplicate",
		},
		{
			name: "non-existent quote",
			input: CreatePurchaseOrderInput{
				QuoteID:  9999,
				PONumber: "PO-005",
				Quantity: 5,
			},
			wantErr: true,
			errType: "not_found",
		},
		{
			name: "non-existent requisition",
			input: CreatePurchaseOrderInput{
				QuoteID:       quote.ID,
				RequisitionID: func() *uint { id := uint(9999); return &id }(),
				PONumber:      "PO-006",
				Quantity:      5,
			},
			wantErr: true,
			errType: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po, err := poSvc.Create(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				switch tt.errType {
				case "validation":
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T: %v", err, err)
					}
				case "duplicate":
					if _, ok := err.(*DuplicateError); !ok {
						t.Errorf("Expected DuplicateError, got %T: %v", err, err)
					}
				case "not_found":
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("Expected NotFoundError, got %T: %v", err, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify purchase order
			if po.PONumber != tt.input.PONumber {
				t.Errorf("PONumber = %v, want %v", po.PONumber, tt.input.PONumber)
			}
			if po.Status != "pending" {
				t.Errorf("Status = %v, want pending", po.Status)
			}
			if po.Quantity != tt.input.Quantity {
				t.Errorf("Quantity = %v, want %v", po.Quantity, tt.input.Quantity)
			}
			if po.UnitPrice != quote.Price {
				t.Errorf("UnitPrice = %v, want %v", po.UnitPrice, quote.Price)
			}
			expectedTotal := quote.Price * float64(tt.input.Quantity)
			if po.TotalAmount != expectedTotal {
				t.Errorf("TotalAmount = %v, want %v", po.TotalAmount, expectedTotal)
			}
			expectedGrand := po.TotalAmount + tt.input.ShippingCost + tt.input.Tax
			if po.GrandTotal != expectedGrand {
				t.Errorf("GrandTotal = %v, want %v", po.GrandTotal, expectedGrand)
			}
			if po.VendorID != vendor.ID {
				t.Errorf("VendorID = %v, want %v", po.VendorID, vendor.ID)
			}
			if po.ProductID != product.ID {
				t.Errorf("ProductID = %v, want %v", po.ProductID, product.ID)
			}
		})
	}
}

func TestPurchaseOrderService_GetByID(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)
	created, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-TEST",
		Quantity: 5,
	})

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing purchase order",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent purchase order",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po, err := poSvc.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("Expected NotFoundError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if po.ID != tt.id {
				t.Errorf("ID = %v, want %v", po.ID, tt.id)
			}
		})
	}
}

func TestPurchaseOrderService_GetByPONumber(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)
	created, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-FIND-ME",
		Quantity: 5,
	})

	tests := []struct {
		name     string
		poNumber string
		wantErr  bool
	}{
		{
			name:     "existing po number",
			poNumber: "PO-FIND-ME",
			wantErr:  false,
		},
		{
			name:     "non-existent po number",
			poNumber: "PO-DOESNT-EXIST",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po, err := poSvc.GetByPONumber(tt.poNumber)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if po.PONumber != tt.poNumber {
				t.Errorf("PONumber = %v, want %v", po.PONumber, tt.poNumber)
			}
			if po.ID != created.ID {
				t.Errorf("ID = %v, want %v", po.ID, created.ID)
			}
		})
	}
}

func TestPurchaseOrderService_UpdateStatus(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)
	po, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-STATUS-TEST",
		Quantity: 5,
	})

	tests := []struct {
		name    string
		id      uint
		status  string
		wantErr bool
		errType string
	}{
		{
			name:    "valid status - approved",
			id:      po.ID,
			status:  "approved",
			wantErr: false,
		},
		{
			name:    "valid status - ordered",
			id:      po.ID,
			status:  "ordered",
			wantErr: false,
		},
		{
			name:    "valid status - shipped",
			id:      po.ID,
			status:  "shipped",
			wantErr: false,
		},
		{
			name:    "valid status - received",
			id:      po.ID,
			status:  "received",
			wantErr: false,
		},
		{
			name:    "valid status - cancelled",
			id:      po.ID,
			status:  "cancelled",
			wantErr: false,
		},
		{
			name:    "invalid status",
			id:      po.ID,
			status:  "invalid_status",
			wantErr: true,
			errType: "validation",
		},
		{
			name:    "non-existent purchase order",
			id:      9999,
			status:  "approved",
			wantErr: true,
			errType: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := poSvc.UpdateStatus(tt.id, tt.status)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				switch tt.errType {
				case "validation":
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T: %v", err, err)
					}
				case "not_found":
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("Expected NotFoundError, got %T: %v", err, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if updated.Status != tt.status {
				t.Errorf("Status = %v, want %v", updated.Status, tt.status)
			}
		})
	}
}

func TestPurchaseOrderService_UpdateDeliveryDates(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)
	po, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-DELIVERY-TEST",
		Quantity: 5,
	})

	expectedDate := time.Now().Add(30 * 24 * time.Hour)
	actualDate := time.Now()

	t.Run("update expected delivery", func(t *testing.T) {
		updated, err := poSvc.UpdateDeliveryDates(po.ID, &expectedDate, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if updated.ExpectedDelivery == nil {
			t.Error("ExpectedDelivery should not be nil")
		} else if !updated.ExpectedDelivery.Equal(expectedDate) {
			t.Errorf("ExpectedDelivery = %v, want %v", updated.ExpectedDelivery, expectedDate)
		}
	})

	t.Run("update actual delivery - auto sets status to received", func(t *testing.T) {
		updated, err := poSvc.UpdateDeliveryDates(po.ID, nil, &actualDate)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if updated.ActualDelivery == nil {
			t.Error("ActualDelivery should not be nil")
		}
		if updated.Status != "received" {
			t.Errorf("Status = %v, want received", updated.Status)
		}
	})
}

func TestPurchaseOrderService_Delete(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)

	tests := []struct {
		name    string
		status  string
		wantErr bool
		errType string
	}{
		{
			name:    "delete pending order",
			status:  "pending",
			wantErr: false,
		},
		{
			name:    "delete cancelled order",
			status:  "cancelled",
			wantErr: false,
		},
		{
			name:    "cannot delete approved order",
			status:  "approved",
			wantErr: true,
			errType: "validation",
		},
		{
			name:    "cannot delete ordered order",
			status:  "ordered",
			wantErr: true,
			errType: "validation",
		},
		{
			name:    "cannot delete shipped order",
			status:  "shipped",
			wantErr: true,
			errType: "validation",
		},
		{
			name:    "cannot delete received order",
			status:  "received",
			wantErr: true,
			errType: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new PO for each test
			po, _ := poSvc.Create(CreatePurchaseOrderInput{
				QuoteID:  quote.ID,
				PONumber: "PO-DELETE-" + tt.status,
				Quantity: 5,
			})

			// Set status
			_, _ = poSvc.UpdateStatus(po.ID, tt.status)

			err := poSvc.Delete(po.ID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				switch tt.errType {
				case "validation":
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T: %v", err, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify deletion
			_, err = poSvc.GetByID(po.ID)
			if err == nil {
				t.Error("Purchase order should have been deleted")
			}
		})
	}
}

func TestPurchaseOrderService_List(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)

	// Create multiple purchase orders
	for i := 1; i <= 5; i++ {
		_, _ = poSvc.Create(CreatePurchaseOrderInput{
			QuoteID:  quote.ID,
			PONumber: fmt.Sprintf("PO-LIST-%d", i),
			Quantity: i,
		})
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all purchase orders",
			limit:     0,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "limited purchase orders",
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "with offset",
			limit:     2,
			offset:    2,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, err := poSvc.List(tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(orders) != tt.wantCount {
				t.Errorf("Got %d orders, want %d", len(orders), tt.wantCount)
			}
		})
	}
}

func TestPurchaseOrderService_ListByStatus(t *testing.T) {
	cfg := setupPurchaseOrderTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Setup test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Test Brand")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Test Product", brand.ID, nil)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	quoteSvc := NewQuoteService(cfg.DB)
	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     100.0,
		Currency:  "USD",
	})

	poSvc := NewPurchaseOrderService(cfg.DB)

	// Create purchase orders with different statuses
	po1, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-STATUS-1",
		Quantity: 1,
	})
	_, _ = poSvc.UpdateStatus(po1.ID, "shipped")

	po2, _ := poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-STATUS-2",
		Quantity: 1,
	})
	_, _ = poSvc.UpdateStatus(po2.ID, "shipped")

	_, _ = poSvc.Create(CreatePurchaseOrderInput{
		QuoteID:  quote.ID,
		PONumber: "PO-STATUS-3",
		Quantity: 1,
	})
	// Leave as pending

	orders, err := poSvc.ListByStatus("shipped", 0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Got %d shipped orders, want 2", len(orders))
	}

	pendingOrders, err := poSvc.ListByStatus("pending", 0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pendingOrders) != 1 {
		t.Errorf("Got %d pending orders, want 1", len(pendingOrders))
	}
}
