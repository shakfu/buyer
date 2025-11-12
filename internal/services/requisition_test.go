package services

import (
	"testing"
	"time"
)

func TestRequisitionService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)

	// Create test specifications
	spec1, err := specService.Create("Laptop", "Portable computer")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	spec2, err := specService.Create("Monitor", "Display device")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	tests := []struct {
		name          string
		reqName       string
		justification string
		budget        float64
		items         []RequisitionItemInput
		wantErr       bool
		errType       interface{}
	}{
		{
			name:          "valid requisition with items",
			reqName:       "Office Equipment",
			justification: "Need new equipment for office",
			budget:        5000.0,
			items: []RequisitionItemInput{
				{SpecificationID: spec1.ID, Quantity: 2, BudgetPerUnit: 1500.0, Description: "For developers"},
				{SpecificationID: spec2.ID, Quantity: 3, BudgetPerUnit: 500.0, Description: "For workstations"},
			},
			wantErr: false,
		},
		{
			name:          "valid requisition without items",
			reqName:       "Empty Requisition",
			justification: "Will add items later",
			budget:        1000.0,
			items:         []RequisitionItemInput{},
			wantErr:       false,
		},
		{
			name:          "empty name",
			reqName:       "",
			justification: "Test",
			budget:        100.0,
			items:         []RequisitionItemInput{},
			wantErr:       true,
			errType:       &ValidationError{},
		},
		{
			name:          "whitespace name",
			reqName:       "   ",
			justification: "Test",
			budget:        100.0,
			items:         []RequisitionItemInput{},
			wantErr:       true,
			errType:       &ValidationError{},
		},
		{
			name:          "negative budget",
			reqName:       "Negative Budget",
			justification: "Test",
			budget:        -100.0,
			items:         []RequisitionItemInput{},
			wantErr:       true,
			errType:       &ValidationError{},
		},
		{
			name:          "duplicate requisition",
			reqName:       "Office Equipment", // Already created in first test
			justification: "Test",
			budget:        100.0,
			items:         []RequisitionItemInput{},
			wantErr:       true,
			errType:       &DuplicateError{},
		},
		{
			name:          "item with zero quantity",
			reqName:       "Bad Quantity",
			justification: "Test",
			budget:        100.0,
			items: []RequisitionItemInput{
				{SpecificationID: spec1.ID, Quantity: 0, BudgetPerUnit: 100.0},
			},
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:          "item with negative quantity",
			reqName:       "Negative Quantity",
			justification: "Test",
			budget:        100.0,
			items: []RequisitionItemInput{
				{SpecificationID: spec1.ID, Quantity: -1, BudgetPerUnit: 100.0},
			},
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:          "item with negative budget per unit",
			reqName:       "Negative Unit Budget",
			justification: "Test",
			budget:        100.0,
			items: []RequisitionItemInput{
				{SpecificationID: spec1.ID, Quantity: 1, BudgetPerUnit: -100.0},
			},
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:          "item with non-existent specification",
			reqName:       "Bad Spec",
			justification: "Test",
			budget:        100.0,
			items: []RequisitionItemInput{
				{SpecificationID: 9999, Quantity: 1, BudgetPerUnit: 100.0},
			},
			wantErr: true,
			errType: &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := reqService.Create(tt.reqName, tt.justification, tt.budget, tt.items)

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
				case *NotFoundError:
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("Create() error type = %T, want NotFoundError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if req.Name != tt.reqName {
				t.Errorf("Create() name = %v, want %v", req.Name, tt.reqName)
			}
			if req.Budget != tt.budget {
				t.Errorf("Create() budget = %v, want %v", req.Budget, tt.budget)
			}
			if len(req.Items) != len(tt.items) {
				t.Errorf("Create() items count = %v, want %v", len(req.Items), len(tt.items))
			}
		})
	}
}

func TestRequisitionService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)

	// Create a test requisition
	req, err := reqService.Create("Test Req", "Test justification", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create test requisition: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing requisition",
			id:      req.ID,
			wantErr: false,
		},
		{
			name:    "non-existent requisition",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reqService.GetByID(tt.id)

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
		})
	}
}

func TestRequisitionService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)

	// Create test requisitions
	req1, err := reqService.Create("Original Req", "Original", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	_, err = reqService.Create("Other Req", "Other", 2000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	tests := []struct {
		name          string
		id            uint
		newName       string
		justification string
		budget        float64
		wantErr       bool
		errType       interface{}
	}{
		{
			name:          "valid update",
			id:            req1.ID,
			newName:       "Updated Req",
			justification: "Updated justification",
			budget:        1500.0,
			wantErr:       false,
		},
		{
			name:          "empty name",
			id:            req1.ID,
			newName:       "",
			justification: "Test",
			budget:        100.0,
			wantErr:       true,
			errType:       &ValidationError{},
		},
		{
			name:          "negative budget",
			id:            req1.ID,
			newName:       "Test",
			justification: "Test",
			budget:        -100.0,
			wantErr:       true,
			errType:       &ValidationError{},
		},
		{
			name:          "duplicate name",
			id:            req1.ID,
			newName:       "Other Req", // Already exists as req2
			justification: "Test",
			budget:        100.0,
			wantErr:       true,
			errType:       &DuplicateError{},
		},
		{
			name:          "non-existent requisition",
			id:            9999,
			newName:       "Test",
			justification: "Test",
			budget:        100.0,
			wantErr:       true,
			errType:       &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reqService.Update(tt.id, tt.newName, tt.justification, tt.budget)

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
			if result.Budget != tt.budget {
				t.Errorf("Update() budget = %v, want %v", result.Budget, tt.budget)
			}
		})
	}
}

func TestRequisitionService_AddItem(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)

	// Create test data
	req, err := reqService.Create("Test Req", "Test", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	spec, err := specService.Create("Test Spec", "Test")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	tests := []struct {
		name            string
		requisitionID   uint
		specificationID uint
		quantity        int
		budgetPerUnit   float64
		description     string
		wantErr         bool
		errType         interface{}
	}{
		{
			name:            "valid item",
			requisitionID:   req.ID,
			specificationID: spec.ID,
			quantity:        5,
			budgetPerUnit:   100.0,
			description:     "Test item",
			wantErr:         false,
		},
		{
			name:            "zero quantity",
			requisitionID:   req.ID,
			specificationID: spec.ID,
			quantity:        0,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &ValidationError{},
		},
		{
			name:            "negative budget per unit",
			requisitionID:   req.ID,
			specificationID: spec.ID,
			quantity:        1,
			budgetPerUnit:   -100.0,
			description:     "",
			wantErr:         true,
			errType:         &ValidationError{},
		},
		{
			name:            "non-existent requisition",
			requisitionID:   9999,
			specificationID: spec.ID,
			quantity:        1,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &NotFoundError{},
		},
		{
			name:            "non-existent specification",
			requisitionID:   req.ID,
			specificationID: 9999,
			quantity:        1,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := reqService.AddItem(tt.requisitionID, tt.specificationID, tt.quantity, tt.budgetPerUnit, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Error("AddItem() error = nil, wantErr true")
					return
				}
				// Check error type
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("AddItem() error type = %T, want ValidationError", err)
					}
				case *NotFoundError:
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("AddItem() error type = %T, want NotFoundError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("AddItem() unexpected error = %v", err)
				return
			}

			if item.Quantity != tt.quantity {
				t.Errorf("AddItem() quantity = %v, want %v", item.Quantity, tt.quantity)
			}
			if item.BudgetPerUnit != tt.budgetPerUnit {
				t.Errorf("AddItem() budgetPerUnit = %v, want %v", item.BudgetPerUnit, tt.budgetPerUnit)
			}
		})
	}
}

func TestRequisitionService_UpdateItem(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)

	// Create test data
	req, err := reqService.Create("Test Req", "Test", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	spec1, err := specService.Create("Spec 1", "Test")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	spec2, err := specService.Create("Spec 2", "Test")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	item, err := reqService.AddItem(req.ID, spec1.ID, 2, 100.0, "Original")
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	tests := []struct {
		name            string
		itemID          uint
		specificationID uint
		quantity        int
		budgetPerUnit   float64
		description     string
		wantErr         bool
		errType         interface{}
	}{
		{
			name:            "valid update",
			itemID:          item.ID,
			specificationID: spec2.ID,
			quantity:        5,
			budgetPerUnit:   200.0,
			description:     "Updated",
			wantErr:         false,
		},
		{
			name:            "zero quantity",
			itemID:          item.ID,
			specificationID: spec1.ID,
			quantity:        0,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &ValidationError{},
		},
		{
			name:            "non-existent item",
			itemID:          9999,
			specificationID: spec1.ID,
			quantity:        1,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &NotFoundError{},
		},
		{
			name:            "non-existent specification",
			itemID:          item.ID,
			specificationID: 9999,
			quantity:        1,
			budgetPerUnit:   100.0,
			description:     "",
			wantErr:         true,
			errType:         &NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reqService.UpdateItem(tt.itemID, tt.specificationID, tt.quantity, tt.budgetPerUnit, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateItem() error = nil, wantErr true")
					return
				}
				// Check error type
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("UpdateItem() error type = %T, want ValidationError", err)
					}
				case *NotFoundError:
					if _, ok := err.(*NotFoundError); !ok {
						t.Errorf("UpdateItem() error type = %T, want NotFoundError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateItem() unexpected error = %v", err)
				return
			}

			if result.Quantity != tt.quantity {
				t.Errorf("UpdateItem() quantity = %v, want %v", result.Quantity, tt.quantity)
			}
		})
	}
}

func TestRequisitionService_DeleteItem(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)

	// Create test data
	req, err := reqService.Create("Test Req", "Test", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	spec, err := specService.Create("Test Spec", "Test")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	item, err := reqService.AddItem(req.ID, spec.ID, 1, 100.0, "")
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	tests := []struct {
		name    string
		itemID  uint
		wantErr bool
	}{
		{
			name:    "delete existing item",
			itemID:  item.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent item",
			itemID:  9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reqService.DeleteItem(tt.itemID)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteItem() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteItem() unexpected error = %v", err)
			}
		})
	}
}

func TestRequisitionService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)

	// Create test data
	spec, err := specService.Create("Test Spec", "Test")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	req, err := reqService.Create("To Delete", "Test", 1000.0, []RequisitionItemInput{
		{SpecificationID: spec.ID, Quantity: 1, BudgetPerUnit: 100.0},
	})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing requisition",
			id:      req.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent requisition",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reqService.Delete(tt.id)

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
			_, err = reqService.GetByID(tt.id)
			if err == nil {
				t.Error("Requisition should be deleted but still exists")
			}
		})
	}
}

func TestRequisitionService_List(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)

	// Create test requisitions
	_, err := reqService.Create("Req A", "Test A", 1000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	_, err = reqService.Create("Req B", "Test B", 2000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}
	_, err = reqService.Create("Req C", "Test C", 3000.0, []RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all requisitions",
			limit:     0,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited requisitions",
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
			reqs, err := reqService.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(reqs) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(reqs), tt.wantCount)
			}
		})
	}
}

func TestRequisitionService_GetQuoteComparison(t *testing.T) {
	cfg := setupTestDB(t)
	reqService := NewRequisitionService(cfg.DB)
	specService := NewSpecificationService(cfg.DB)
	brandService := NewBrandService(cfg.DB)
	productService := NewProductService(cfg.DB)
	vendorService := NewVendorService(cfg.DB)
	forexService := NewForexService(cfg.DB)
	quoteService := NewQuoteService(cfg.DB)

	// Create test data
	spec, err := specService.Create("Laptop", "Portable computer")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}
	brand, err := brandService.Create("Dell")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	product, err := productService.Create("Dell Latitude", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	vendor, err := vendorService.Create("Vendor A", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	// Create a forex rate
	_, err = forexService.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create a quote
	_, err = quoteService.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     1200.0,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	// Create requisition with item
	req, err := reqService.Create("Office Equipment", "Need laptops", 3000.0, []RequisitionItemInput{
		{SpecificationID: spec.ID, Quantity: 2, BudgetPerUnit: 1500.0},
	})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	comparison, err := reqService.GetQuoteComparison(req.ID, quoteService)
	if err != nil {
		t.Fatalf("GetQuoteComparison() error = %v", err)
	}

	if comparison == nil {
		t.Fatal("GetQuoteComparison() returned nil")
	}

	if comparison.Requisition.ID != req.ID {
		t.Errorf("GetQuoteComparison() requisition ID = %v, want %v", comparison.Requisition.ID, req.ID)
	}

	if len(comparison.ItemComparisons) != 1 {
		t.Errorf("GetQuoteComparison() item comparisons count = %v, want 1", len(comparison.ItemComparisons))
	}

	if !comparison.AllItemsHaveQuotes {
		t.Error("GetQuoteComparison() AllItemsHaveQuotes = false, want true")
	}

	if comparison.TotalEstimate != 2400.0 { // 1200 * 2
		t.Errorf("GetQuoteComparison() TotalEstimate = %v, want 2400.0", comparison.TotalEstimate)
	}
}
