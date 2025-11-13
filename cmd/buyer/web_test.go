package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestApp creates a Fiber app with test database and services
func setupTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
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
		&models.ProjectRequisitionItem{},
		&models.PurchaseOrder{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	// Create services
	specSvc := services.NewSpecificationService(db)
	brandSvc := services.NewBrandService(db)
	productSvc := services.NewProductService(db)
	vendorSvc := services.NewVendorService(db)
	requisitionSvc := services.NewRequisitionService(db)
	quoteSvc := services.NewQuoteService(db)
	forexSvc := services.NewForexService(db)
	dashboardSvc := services.NewDashboardService(db)
	projectSvc := services.NewProjectService(db)
	projectReqSvc := services.NewProjectRequisitionService(db)
	poSvc := services.NewPurchaseOrderService(db)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})

	// Setup routes
	setupRoutes(app, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc, dashboardSvc, projectSvc, projectReqSvc, poSvc)

	return app, db
}

// seedTestData adds some test data to the database
func seedTestData(t *testing.T, db *gorm.DB) {
	// Create brands
	brand := &models.Brand{Name: "Test Brand"}
	if err := db.Create(brand).Error; err != nil {
		t.Fatalf("failed to create test brand: %v", err)
	}

	// Create specifications
	spec := &models.Specification{Name: "Test Spec", Description: "Test description"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("failed to create test specification: %v", err)
	}

	// Create vendors
	vendor := &models.Vendor{Name: "Test Vendor", Currency: "USD"}
	if err := db.Create(vendor).Error; err != nil {
		t.Fatalf("failed to create test vendor: %v", err)
	}

	// Create products
	product := &models.Product{Name: "Test Product", BrandID: brand.ID, SpecificationID: &spec.ID}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	// Create forex rate
	forex := &models.Forex{FromCurrency: "EUR", ToCurrency: "USD", Rate: 1.1, EffectiveDate: time.Now()}
	if err := db.Create(forex).Error; err != nil {
		t.Fatalf("failed to create test forex: %v", err)
	}

	// Create quote
	quote := &models.Quote{
		VendorID:       vendor.ID,
		ProductID:      product.ID,
		Price:          100.0,
		Currency:       "USD",
		ConvertedPrice: 100.0,
		ConversionRate: 1.0,
		QuoteDate:      time.Now(),
	}
	if err := db.Create(quote).Error; err != nil {
		t.Fatalf("failed to create test quote: %v", err)
	}

	// Create requisition
	req := &models.Requisition{Name: "Test Requisition", Budget: 1000.0}
	if err := db.Create(req).Error; err != nil {
		t.Fatalf("failed to create test requisition: %v", err)
	}

	reqItem := &models.RequisitionItem{
		RequisitionID:   req.ID,
		SpecificationID: spec.ID,
		Quantity:        5,
		BudgetPerUnit:   50.0,
	}
	if err := db.Create(reqItem).Error; err != nil {
		t.Fatalf("failed to create test requisition item: %v", err)
	}

	// Create project
	deadline := time.Now().AddDate(0, 3, 0)
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test project description",
		Budget:      10000.0,
		Deadline:    &deadline,
		Status:      "active",
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	// Create BOM
	bom := &models.BillOfMaterials{ProjectID: project.ID}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("failed to create test BOM: %v", err)
	}

	bomItem := &models.BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          10,
		Notes:             "Test BOM item",
	}
	if err := db.Create(bomItem).Error; err != nil {
		t.Fatalf("failed to create test BOM item: %v", err)
	}

	// Create project requisition
	projectReq := &models.ProjectRequisition{
		ProjectID:     project.ID,
		Name:          "Test Project Requisition",
		Justification: "Test justification",
		Budget:        5000.0,
	}
	if err := db.Create(projectReq).Error; err != nil {
		t.Fatalf("failed to create test project requisition: %v", err)
	}
}

func TestWebHandler_Home(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Dashboard(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Help(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest("GET", "/help", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Specifications(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/specifications", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Brands(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/brands", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Products(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/products", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Vendors(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/vendors", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Quotes(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/quotes", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Requisitions(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/requisitions", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Forex(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/forex", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_Projects(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/projects", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_ProjectDetail(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/projects/1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_ProjectDashboard(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/projects/1/dashboard", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_RequisitionComparison(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	req := httptest.NewRequest("GET", "/requisition-comparison", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateBrand(t *testing.T) {
	app, _ := setupTestApp(t)

	form := url.Values{}
	form.Add("name", "New Test Brand")

	req := httptest.NewRequest("POST", "/brands", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_UpdateBrand(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	form := url.Values{}
	form.Add("name", "Updated Brand")

	req := httptest.NewRequest("PUT", "/brands/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_DeleteBrand(t *testing.T) {
	app, db := setupTestApp(t)

	// Create a brand that won't have foreign key constraints
	brand := &models.Brand{Name: "Deletable Brand"}
	if err := db.Create(brand).Error; err != nil {
		t.Fatalf("failed to create test brand: %v", err)
	}

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/brands/%d", brand.ID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateSpecification(t *testing.T) {
	app, _ := setupTestApp(t)

	form := url.Values{}
	form.Add("name", "New Spec")
	form.Add("description", "New spec description")

	req := httptest.NewRequest("POST", "/specifications", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_UpdateSpecification(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	form := url.Values{}
	form.Add("name", "Updated Spec")
	form.Add("description", "Updated description")

	req := httptest.NewRequest("PUT", "/specifications/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateProduct(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	form := url.Values{}
	form.Add("name", "New Product")
	form.Add("brand_id", "1")
	form.Add("specification_id", "1")

	req := httptest.NewRequest("POST", "/products", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateVendor(t *testing.T) {
	app, _ := setupTestApp(t)

	form := url.Values{}
	form.Add("name", "New Vendor")
	form.Add("currency", "EUR")
	form.Add("discount_code", "DISCOUNT10")

	req := httptest.NewRequest("POST", "/vendors", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateQuote(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	form := url.Values{}
	form.Add("vendor_id", "1")
	form.Add("product_id", "1")
	form.Add("price", "150.50")
	form.Add("currency", "USD")
	form.Add("notes", "Test quote notes")

	req := httptest.NewRequest("POST", "/quotes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateForex(t *testing.T) {
	app, _ := setupTestApp(t)

	form := url.Values{}
	form.Add("from_currency", "GBP")
	form.Add("to_currency", "USD")
	form.Add("rate", "1.25")

	req := httptest.NewRequest("POST", "/forex", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_CreateProject(t *testing.T) {
	app, _ := setupTestApp(t)

	form := url.Values{}
	form.Add("name", "New Project")
	form.Add("description", "New project description")
	form.Add("budget", "25000")
	form.Add("deadline", "2025-12-31")

	req := httptest.NewRequest("POST", "/projects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_UpdateProject(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	form := url.Values{}
	form.Add("name", "Updated Project")
	form.Add("description", "Updated description")
	form.Add("budget", "30000")
	form.Add("status", "completed")

	req := httptest.NewRequest("PUT", "/projects/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_AddBOMItem(t *testing.T) {
	app, db := setupTestApp(t)
	seedTestData(t, db)

	// Create another specification for the BOM item
	spec2 := &models.Specification{Name: "Another Spec", Description: "Another spec"}
	if err := db.Create(spec2).Error; err != nil {
		t.Fatalf("failed to create test specification: %v", err)
	}

	form := url.Values{}
	form.Add("specification_id", fmt.Sprintf("%d", spec2.ID))
	form.Add("quantity", "5")
	form.Add("notes", "Test BOM item notes")

	req := httptest.NewRequest("POST", "/projects/1/bom-items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebHandler_InvalidRoutes(t *testing.T) {
	app, _ := setupTestApp(t)

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"NonExistentProject", "/projects/9999", 404},
		{"InvalidProjectID", "/projects/invalid", 400},
		{"NonExistentProjectDashboard", "/projects/9999/dashboard", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestWebHandler_ValidationErrors(t *testing.T) {
	app, _ := setupTestApp(t)

	tests := []struct {
		name        string
		method      string
		path        string
		formData    map[string]string
		expectError bool
	}{
		{
			name:   "EmptyBrandName",
			method: "POST",
			path:   "/brands",
			formData: map[string]string{
				"name": "",
			},
			expectError: true,
		},
		{
			name:   "EmptySpecName",
			method: "POST",
			path:   "/specifications",
			formData: map[string]string{
				"name":        "",
				"description": "desc",
			},
			expectError: true,
		},
		{
			name:   "InvalidBrandID",
			method: "POST",
			path:   "/products",
			formData: map[string]string{
				"name":     "Product",
				"brand_id": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			for k, v := range tt.formData {
				form.Add(k, v)
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}

			if tt.expectError && resp.StatusCode == 200 {
				t.Errorf("expected error status, got 200")
			}
		})
	}
}

func init() {
	// Initialize config for tests (not used but prevents nil pointer)
	cfg = &config.Config{}
}
