package main

import (
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
)

// setupTestDB creates a test database configuration
func setupTestDB(t *testing.T) *config.Config {
	t.Helper()
	testCfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run auto-migration
	if err := testCfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.Product{},
		&models.Requisition{},
		&models.RequisitionItem{},
		&models.Quote{},
		&models.Forex{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.PurchaseOrder{},
		&models.Document{},
		&models.VendorRating{},
	); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return testCfg
}

// TestCLI_AddBrand tests the brand creation workflow through CLI
func TestCLI_AddBrand(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM brands")

	// Set global cfg for CLI commands
	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	svc := services.NewBrandService(testCfg.DB)

	// Create brand through service (simulating CLI add command logic)
	brand, err := svc.Create("TestBrand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	if brand.Name != "TestBrand" {
		t.Errorf("Expected brand name 'TestBrand', got %q", brand.Name)
	}

	// Verify we can retrieve it (simulating CLI list command logic)
	brands, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list brands: %v", err)
	}

	if len(brands) != 1 {
		t.Fatalf("Expected 1 brand, got %d", len(brands))
	}

	if brands[0].Name != "TestBrand" {
		t.Errorf("Expected brand name 'TestBrand', got %q", brands[0].Name)
	}

	// Test update (simulating CLI update command logic)
	updatedBrand, err := svc.Update(brand.ID, "UpdatedBrand")
	if err != nil {
		t.Fatalf("Failed to update brand: %v", err)
	}

	if updatedBrand.Name != "UpdatedBrand" {
		t.Errorf("Expected updated brand name 'UpdatedBrand', got %q", updatedBrand.Name)
	}

	// Test delete (simulating CLI delete command logic)
	if err := svc.Delete(brand.ID); err != nil {
		t.Fatalf("Failed to delete brand: %v", err)
	}

	// Verify deletion
	brands, err = svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list brands after deletion: %v", err)
	}

	if len(brands) != 0 {
		t.Errorf("Expected 0 brands after deletion, got %d", len(brands))
	}
}

// TestCLI_AddSpecification tests specification creation workflow
func TestCLI_AddSpecification(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM specifications")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	svc := services.NewSpecificationService(testCfg.DB)

	// Create specification with description
	spec, err := svc.Create("Laptop", "High-performance laptop")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	if spec.Name != "Laptop" {
		t.Errorf("Expected spec name 'Laptop', got %q", spec.Name)
	}

	if spec.Description != "High-performance laptop" {
		t.Errorf("Expected spec description 'High-performance laptop', got %q", spec.Description)
	}

	// List specifications
	specs, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list specifications: %v", err)
	}

	if len(specs) != 1 {
		t.Fatalf("Expected 1 specification, got %d", len(specs))
	}
}

// TestCLI_AddProduct tests product creation workflow
func TestCLI_AddProduct(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM products")
	defer testCfg.DB.Exec("DELETE FROM brands")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	// Create brand first
	brandSvc := services.NewBrandService(testCfg.DB)
	brand, err := brandSvc.Create("Lenovo")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	// Create product
	productSvc := services.NewProductService(testCfg.DB)
	product, err := productSvc.Create("ThinkPad X1", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	if product.Name != "ThinkPad X1" {
		t.Errorf("Expected product name 'ThinkPad X1', got %q", product.Name)
	}

	if product.Brand == nil || product.Brand.Name != "Lenovo" {
		t.Errorf("Expected brand 'Lenovo', got %v", product.Brand)
	}

	// List products
	products, err := productSvc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(products) != 1 {
		t.Fatalf("Expected 1 product, got %d", len(products))
	}
}

// TestCLI_AddVendor tests vendor creation workflow
func TestCLI_AddVendor(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM vendors")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	svc := services.NewVendorService(testCfg.DB)

	// Create vendor with EUR currency
	vendor, err := svc.Create("TechSupplier", "EUR", "SAVE20")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	if vendor.Name != "TechSupplier" {
		t.Errorf("Expected vendor name 'TechSupplier', got %q", vendor.Name)
	}

	if vendor.Currency != "EUR" {
		t.Errorf("Expected currency 'EUR', got %q", vendor.Currency)
	}

	if vendor.DiscountCode != "SAVE20" {
		t.Errorf("Expected discount code 'SAVE20', got %q", vendor.DiscountCode)
	}

	// List vendors
	vendors, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list vendors: %v", err)
	}

	if len(vendors) != 1 {
		t.Fatalf("Expected 1 vendor, got %d", len(vendors))
	}
}

// TestCLI_AddForex tests forex rate creation workflow
func TestCLI_AddForex(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM forex_rates")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	svc := services.NewForexService(testCfg.DB)

	// Create forex rate
	forex, err := svc.Create("EUR", "USD", 1.10, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex rate: %v", err)
	}

	if forex.FromCurrency != "EUR" {
		t.Errorf("Expected from currency 'EUR', got %q", forex.FromCurrency)
	}

	if forex.ToCurrency != "USD" {
		t.Errorf("Expected to currency 'USD', got %q", forex.ToCurrency)
	}

	if forex.Rate != 1.10 {
		t.Errorf("Expected rate 1.10, got %.2f", forex.Rate)
	}

	// Test conversion
	converted, rate, err := svc.Convert(100, "EUR", "USD")
	if err != nil {
		t.Fatalf("Failed to convert currency: %v", err)
	}

	// Use approximate comparison for floating point
	if converted < 109.99 || converted > 110.01 {
		t.Errorf("Expected converted value ~110.0, got %.2f", converted)
	}

	if rate < 1.09 || rate > 1.11 {
		t.Errorf("Expected rate ~1.10, got %.2f", rate)
	}
}

// TestCLI_AddRequisition tests requisition creation workflow
func TestCLI_AddRequisition(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM requisition_items")
	defer testCfg.DB.Exec("DELETE FROM requisitions")
	defer testCfg.DB.Exec("DELETE FROM specifications")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	reqSvc := services.NewRequisitionService(testCfg.DB)
	specSvc := services.NewSpecificationService(testCfg.DB)

	// Create requisition
	req, err := reqSvc.Create("Office Equipment", "Upgrade office", 5000.0, []services.RequisitionItemInput{})
	if err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	if req.Name != "Office Equipment" {
		t.Errorf("Expected requisition name 'Office Equipment', got %q", req.Name)
	}

	if req.Justification != "Upgrade office" {
		t.Errorf("Expected justification 'Upgrade office', got %q", req.Justification)
	}

	if req.Budget != 5000.0 {
		t.Errorf("Expected budget 5000.0, got %.2f", req.Budget)
	}

	// Create specification for requisition item
	spec, err := specSvc.Create("Desktop PC", "High-end workstation")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Add requisition item
	item, err := reqSvc.AddItem(req.ID, spec.ID, 5, 800.0, "For development team")
	if err != nil {
		t.Fatalf("Failed to add requisition item: %v", err)
	}

	if item.Quantity != 5 {
		t.Errorf("Expected quantity 5, got %d", item.Quantity)
	}

	if item.BudgetPerUnit != 800.0 {
		t.Errorf("Expected budget per unit 800.0, got %.2f", item.BudgetPerUnit)
	}

	// Get requisition with items
	reqWithItems, err := reqSvc.GetByID(req.ID)
	if err != nil {
		t.Fatalf("Failed to get requisition: %v", err)
	}

	if len(reqWithItems.Items) != 1 {
		t.Fatalf("Expected 1 requisition item, got %d", len(reqWithItems.Items))
	}
}

// TestCLI_CompleteWorkflow tests a complete workflow from brand to quote
func TestCLI_CompleteWorkflow(t *testing.T) {
	testCfg := setupTestDB(t)

	// Cleanup in reverse order of dependencies
	defer testCfg.DB.Exec("DELETE FROM quotes")
	defer testCfg.DB.Exec("DELETE FROM products")
	defer testCfg.DB.Exec("DELETE FROM brands")
	defer testCfg.DB.Exec("DELETE FROM vendors")
	defer testCfg.DB.Exec("DELETE FROM forex_rates")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	// 1. Create brand
	brandSvc := services.NewBrandService(testCfg.DB)
	brand, err := brandSvc.Create("Apple")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	// 2. Create product
	productSvc := services.NewProductService(testCfg.DB)
	product, err := productSvc.Create("MacBook Pro 16\"", brand.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// 3. Create vendor
	vendorSvc := services.NewVendorService(testCfg.DB)
	vendor, err := vendorSvc.Create("Apple Store", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	// 4. Create quote
	quoteSvc := services.NewQuoteService(testCfg.DB)
	quote, err := quoteSvc.Create(services.CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     2499.00,
		Currency:  "USD",
		Notes:     "Base model",
	})
	if err != nil {
		t.Fatalf("Failed to create quote: %v", err)
	}

	if quote.Price != 2499.00 {
		t.Errorf("Expected price 2499.00, got %.2f", quote.Price)
	}

	// 5. List quotes for product
	quotes, err := quoteSvc.ListByProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to list quotes: %v", err)
	}

	if len(quotes) != 1 {
		t.Fatalf("Expected 1 quote, got %d", len(quotes))
	}

	// 6. Get best quote
	bestQuote, err := quoteSvc.GetBestQuote(product.ID)
	if err != nil {
		t.Fatalf("Failed to get best quote: %v", err)
	}

	if bestQuote.ID != quote.ID {
		t.Errorf("Expected best quote ID %d, got %d", quote.ID, bestQuote.ID)
	}
}

// TestCLI_ErrorHandling tests error handling in CLI workflows
func TestCLI_ErrorHandling(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM brands")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	brandSvc := services.NewBrandService(testCfg.DB)

	// Test duplicate brand
	_, err := brandSvc.Create("Dell")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	_, err = brandSvc.Create("Dell")
	if err == nil {
		t.Error("Expected error for duplicate brand, got nil")
	}

	// Test empty name
	_, err = brandSvc.Create("")
	if err == nil {
		t.Error("Expected error for empty brand name, got nil")
	}

	// Test update non-existent brand
	_, err = brandSvc.Update(999, "NewName")
	if err == nil {
		t.Error("Expected error for updating non-existent brand, got nil")
	}

	// Test delete non-existent brand
	err = brandSvc.Delete(999)
	if err == nil {
		t.Error("Expected error for deleting non-existent brand, got nil")
	}
}

// TestCLI_AddProject tests project creation workflow
func TestCLI_AddProject(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM bill_of_materials_items")
	defer testCfg.DB.Exec("DELETE FROM bills_of_materials")
	defer testCfg.DB.Exec("DELETE FROM project_requisitions")
	defer testCfg.DB.Exec("DELETE FROM projects")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	svc := services.NewProjectService(testCfg.DB)

	// Create project with deadline
	deadline := time.Now().Add(60 * 24 * time.Hour)
	project, err := svc.Create("Office Renovation", "Complete office renovation", 100000.0, &deadline)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.Name != "Office Renovation" {
		t.Errorf("Expected project name 'Office Renovation', got %q", project.Name)
	}

	if project.Description != "Complete office renovation" {
		t.Errorf("Expected description 'Complete office renovation', got %q", project.Description)
	}

	if project.Budget != 100000.0 {
		t.Errorf("Expected budget 100000.0, got %.2f", project.Budget)
	}

	if project.Status != "planning" {
		t.Errorf("Expected status 'planning', got %q", project.Status)
	}

	if project.BillOfMaterials == nil {
		t.Error("Expected BillOfMaterials to be automatically created")
	}

	// List projects
	projects, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}

	// Test update
	updatedProject, err := svc.Update(project.ID, "Office Renovation Phase 1", "First phase", 80000.0, nil, "active")
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	if updatedProject.Name != "Office Renovation Phase 1" {
		t.Errorf("Expected updated name 'Office Renovation Phase 1', got %q", updatedProject.Name)
	}

	if updatedProject.Status != "active" {
		t.Errorf("Expected status 'active', got %q", updatedProject.Status)
	}

	// Test delete
	if err := svc.Delete(project.ID); err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify deletion
	projects, err = svc.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list projects after deletion: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after deletion, got %d", len(projects))
	}
}

// TestCLI_AddBOMItem tests Bill of Materials item workflow
func TestCLI_AddBOMItem(t *testing.T) {
	testCfg := setupTestDB(t)
	defer testCfg.DB.Exec("DELETE FROM bill_of_materials_items")
	defer testCfg.DB.Exec("DELETE FROM bills_of_materials")
	defer testCfg.DB.Exec("DELETE FROM projects")
	defer testCfg.DB.Exec("DELETE FROM specifications")

	oldCfg := cfg
	cfg = testCfg
	defer func() { cfg = oldCfg }()

	projectSvc := services.NewProjectService(testCfg.DB)
	specSvc := services.NewSpecificationService(testCfg.DB)

	// Create project
	project, err := projectSvc.Create("Test Project", "Test", 50000.0, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create specification
	spec, err := specSvc.Create("Laptop - Intel i7", "High-performance laptop")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Add BOM item
	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "For development team")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	if bomItem.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", bomItem.Quantity)
	}

	if bomItem.Notes != "For development team" {
		t.Errorf("Expected notes 'For development team', got %q", bomItem.Notes)
	}

	if bomItem.Specification.Name != "Laptop - Intel i7" {
		t.Errorf("Expected specification 'Laptop - Intel i7', got %q", bomItem.Specification.Name)
	}

	// Update BOM item
	updatedBOMItem, err := projectSvc.UpdateBillOfMaterialsItem(bomItem.ID, 15, "Updated quantity")
	if err != nil {
		t.Fatalf("Failed to update BOM item: %v", err)
	}

	if updatedBOMItem.Quantity != 15 {
		t.Errorf("Expected quantity 15, got %d", updatedBOMItem.Quantity)
	}

	if updatedBOMItem.Notes != "Updated quantity" {
		t.Errorf("Expected notes 'Updated quantity', got %q", updatedBOMItem.Notes)
	}

	// Delete BOM item
	if err := projectSvc.DeleteBillOfMaterialsItem(bomItem.ID); err != nil {
		t.Fatalf("Failed to delete BOM item: %v", err)
	}

	// Verify deletion by getting project
	projectWithBOM, err := projectSvc.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(projectWithBOM.BillOfMaterials.Items) != 0 {
		t.Errorf("Expected 0 BOM items after deletion, got %d", len(projectWithBOM.BillOfMaterials.Items))
	}
}

// TestCLI_AddProjectRequisition tests project requisition linking workflow

// TestCLI_ProjectCompleteWorkflow tests complete project workflow
