package main

import (
	"fmt"
	"io"
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
		&models.SpecificationAttribute{},
		&models.Product{},
		&models.ProductAttribute{},
		&models.Requisition{},
		&models.RequisitionItem{},
		&models.Quote{},
		&models.Forex{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
		&models.ProjectProcurementStrategy{},
		&models.PurchaseOrder{},
		&models.Document{},
		&models.VendorRating{},
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
	docSvc := services.NewDocumentService(db)
	ratingsSvc := services.NewVendorRatingService(db)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})

	// Setup routes
	setupRoutes(app, db, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc, dashboardSvc, projectSvc, projectReqSvc, poSvc, docSvc, ratingsSvc)

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

func TestWebHandler_ProductDetailWithAttributes(t *testing.T) {
	app, db := setupTestApp(t)

	// Create spec and brand
	spec := &models.Specification{Name: "Laptop"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	brand := &models.Brand{Name: "Dell"}
	if err := db.Create(brand).Error; err != nil {
		t.Fatal(err)
	}

	// Create specification attributes
	ramAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "RAM",
		DataType:        "number",
		Unit:            "GB",
		IsRequired:      true,
	}
	storageAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "Storage",
		DataType:        "number",
		Unit:            "GB",
		IsRequired:      true,
	}
	if err := db.Create(&ramAttr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&storageAttr).Error; err != nil {
		t.Fatal(err)
	}

	// Create product
	product := &models.Product{
		Name:            "XPS 15",
		BrandID:         brand.ID,
		SpecificationID: &spec.ID,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatal(err)
	}

	// Create product attributes
	ram32 := 32.0
	storage512 := 512.0
	attrs := []models.ProductAttribute{
		{
			ProductID:                product.ID,
			SpecificationAttributeID: ramAttr.ID,
			ValueNumber:              &ram32,
		},
		{
			ProductID:                product.ID,
			SpecificationAttributeID: storageAttr.ID,
			ValueNumber:              &storage512,
		},
	}
	for _, attr := range attrs {
		if err := db.Create(&attr).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Test product detail page
	req := httptest.NewRequest("GET", fmt.Sprintf("/products/%d", product.ID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Read body to check if attributes are present
	body := make([]byte, resp.ContentLength)
	resp.Body.Read(body)
	bodyStr := string(body)

	// Check if attribute section is present
	if !strings.Contains(bodyStr, "Product Specifications") {
		t.Error("expected 'Product Specifications' section in response")
	}

	// Check if RAM attribute is present
	if !strings.Contains(bodyStr, "RAM") {
		t.Error("expected 'RAM' attribute in response")
	}

	// Check if Storage attribute is present
	if !strings.Contains(bodyStr, "Storage") {
		t.Error("expected 'Storage' attribute in response")
	}
}

func TestWebHandler_ProductComparison(t *testing.T) {
	app, db := setupTestApp(t)

	// Create spec and brand
	spec := &models.Specification{Name: "Smartphone"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	brand1 := &models.Brand{Name: "Apple"}
	brand2 := &models.Brand{Name: "Samsung"}
	if err := db.Create(&brand1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&brand2).Error; err != nil {
		t.Fatal(err)
	}

	// Create specification attributes
	ramAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "RAM",
		DataType:        "number",
		Unit:            "GB",
		IsRequired:      true,
	}
	if err := db.Create(&ramAttr).Error; err != nil {
		t.Fatal(err)
	}

	// Create products
	product1 := &models.Product{
		Name:            "iPhone 15",
		BrandID:         brand1.ID,
		SpecificationID: &spec.ID,
	}
	product2 := &models.Product{
		Name:            "Galaxy S24",
		BrandID:         brand2.ID,
		SpecificationID: &spec.ID,
	}
	if err := db.Create(&product1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&product2).Error; err != nil {
		t.Fatal(err)
	}

	// Create product attributes
	ram8 := 8.0
	ram12 := 12.0
	attrs := []models.ProductAttribute{
		{
			ProductID:                product1.ID,
			SpecificationAttributeID: ramAttr.ID,
			ValueNumber:              &ram8,
		},
		{
			ProductID:                product2.ID,
			SpecificationAttributeID: ramAttr.ID,
			ValueNumber:              &ram12,
		},
	}
	for _, attr := range attrs {
		if err := db.Create(&attr).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Test product comparison page
	req := httptest.NewRequest("GET", fmt.Sprintf("/products/compare/%d", spec.ID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Read body to check if comparison is present
	body := make([]byte, resp.ContentLength)
	resp.Body.Read(body)
	bodyStr := string(body)

	// Check if both products are present
	if !strings.Contains(bodyStr, "iPhone 15") {
		t.Error("expected 'iPhone 15' in comparison")
	}
	if !strings.Contains(bodyStr, "Galaxy S24") {
		t.Error("expected 'Galaxy S24' in comparison")
	}

	// Check if attribute comparison is present
	if !strings.Contains(bodyStr, "RAM") {
		t.Error("expected 'RAM' attribute in comparison")
	}
}

func TestWebHandler_SpecificationAttributesPage(t *testing.T) {
	app, db := setupTestApp(t)

	// Create specification
	spec := &models.Specification{Name: "Laptop", Description: "Laptop computers"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	// Create some attributes
	ramAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "RAM",
		DataType:        "number",
		Unit:            "GB",
		IsRequired:      true,
		MinValue:        ptrFloat64(4),
		MaxValue:        ptrFloat64(128),
	}
	cpuAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "CPU",
		DataType:        "text",
		IsRequired:      false,
	}
	if err := db.Create(&ramAttr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&cpuAttr).Error; err != nil {
		t.Fatal(err)
	}

	// Test GET /specifications/:id/attributes
	req := httptest.NewRequest("GET", fmt.Sprintf("/specifications/%d/attributes", spec.ID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyStr := string(body)

	// Check for spec info and attributes
	if !strings.Contains(bodyStr, "Laptop - Attributes") {
		t.Error("expected page title in response")
	}
	if !strings.Contains(bodyStr, "RAM") {
		t.Error("expected 'RAM' attribute in response")
	}
	if !strings.Contains(bodyStr, "CPU") {
		t.Error("expected 'CPU' attribute in response")
	}
	if !strings.Contains(bodyStr, "4.00") && !strings.Contains(bodyStr, "Min: 4") {
		t.Error("expected min value for RAM")
	}
}

func TestWebHandler_CreateSpecificationAttribute(t *testing.T) {
	app, db := setupTestApp(t)

	// Create specification
	spec := &models.Specification{Name: "Monitor"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	// Test POST /specifications/:id/attributes
	form := url.Values{}
	form.Set("name", "Screen Size")
	form.Set("data_type", "number")
	form.Set("unit", "inches")
	form.Set("is_required", "on")
	form.Set("min_value", "21")
	form.Set("max_value", "49")
	form.Set("description", "Diagonal screen size")

	req := httptest.NewRequest("POST", fmt.Sprintf("/specifications/%d/attributes", spec.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify attribute was created
	var attr models.SpecificationAttribute
	if err := db.Where("specification_id = ? AND name = ?", spec.ID, "Screen Size").First(&attr).Error; err != nil {
		t.Fatal("attribute not created:", err)
	}
	if attr.DataType != "number" {
		t.Errorf("expected data_type 'number', got %s", attr.DataType)
	}
	if attr.Unit != "inches" {
		t.Errorf("expected unit 'inches', got %s", attr.Unit)
	}
	if !attr.IsRequired {
		t.Error("expected is_required to be true")
	}
	if attr.MinValue == nil || *attr.MinValue != 21 {
		t.Error("expected min_value 21")
	}
	if attr.MaxValue == nil || *attr.MaxValue != 49 {
		t.Error("expected max_value 49")
	}
}

func TestWebHandler_DeleteSpecificationAttribute(t *testing.T) {
	app, db := setupTestApp(t)

	// Create specification and attribute
	spec := &models.Specification{Name: "Keyboard"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	attr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "Switch Type",
		DataType:        "text",
	}
	if err := db.Create(&attr).Error; err != nil {
		t.Fatal(err)
	}

	// Test DELETE /specifications/:specId/attributes/:attrId
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/specifications/%d/attributes/%d", spec.ID, attr.ID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify attribute was deleted
	var count int64
	db.Model(&models.SpecificationAttribute{}).Where("id = ?", attr.ID).Count(&count)
	if count != 0 {
		t.Error("attribute should be deleted")
	}
}

func TestWebHandler_UpdateProductAttributes(t *testing.T) {
	app, db := setupTestApp(t)

	// Create spec, brand, and attributes
	spec := &models.Specification{Name: "Laptop"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	brand := &models.Brand{Name: "Dell"}
	if err := db.Create(brand).Error; err != nil {
		t.Fatal(err)
	}

	ramAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "RAM",
		DataType:        "number",
		Unit:            "GB",
		IsRequired:      true,
		MinValue:        ptrFloat64(4),
		MaxValue:        ptrFloat64(128),
	}
	storageAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "Storage",
		DataType:        "number",
		Unit:            "GB",
	}
	hasSSDAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "Has SSD",
		DataType:        "boolean",
	}
	cpuAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "CPU Model",
		DataType:        "text",
	}
	if err := db.Create(&ramAttr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&storageAttr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&hasSSDAttr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&cpuAttr).Error; err != nil {
		t.Fatal(err)
	}

	// Create product
	product := &models.Product{
		Name:            "XPS 15",
		BrandID:         brand.ID,
		SpecificationID: &spec.ID,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	// Test POST /products/:id/attributes
	form := url.Values{}
	form.Set(fmt.Sprintf("attr_%d", ramAttr.ID), "16")
	form.Set(fmt.Sprintf("attr_%d", storageAttr.ID), "512")
	form.Set(fmt.Sprintf("attr_%d", hasSSDAttr.ID), "true")
	form.Set(fmt.Sprintf("attr_%d", cpuAttr.ID), "Intel i7-12700H")

	req := httptest.NewRequest("POST", fmt.Sprintf("/products/%d/attributes", product.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify attributes were created
	var attrs []models.ProductAttribute
	if err := db.Where("product_id = ?", product.ID).Preload("SpecificationAttribute").Find(&attrs).Error; err != nil {
		t.Fatal(err)
	}
	if len(attrs) != 4 {
		t.Fatalf("expected 4 attributes, got %d", len(attrs))
	}

	// Check values
	for _, attr := range attrs {
		switch attr.SpecificationAttribute.Name {
		case "RAM":
			if attr.ValueNumber == nil || *attr.ValueNumber != 16 {
				t.Error("expected RAM value 16")
			}
		case "Storage":
			if attr.ValueNumber == nil || *attr.ValueNumber != 512 {
				t.Error("expected Storage value 512")
			}
		case "Has SSD":
			if attr.ValueBoolean == nil || !*attr.ValueBoolean {
				t.Error("expected Has SSD to be true")
			}
		case "CPU Model":
			if attr.ValueText == nil || *attr.ValueText != "Intel i7-12700H" {
				t.Error("expected CPU Model 'Intel i7-12700H'")
			}
		}
	}

	// Test response contains the updated display
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "16.00") && !strings.Contains(bodyStr, "16") {
		t.Error("expected RAM value in response")
	}
	if !strings.Contains(bodyStr, "Intel i7-12700H") {
		t.Error("expected CPU model in response")
	}
}

func TestWebHandler_ProductAttributeValidation(t *testing.T) {
	app, db := setupTestApp(t)

	// Create spec, brand, and attributes
	spec := &models.Specification{Name: "Mouse"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatal(err)
	}

	brand := &models.Brand{Name: "Logitech"}
	if err := db.Create(brand).Error; err != nil {
		t.Fatal(err)
	}

	// Required attribute
	dpiAttr := &models.SpecificationAttribute{
		SpecificationID: spec.ID,
		Name:            "DPI",
		DataType:        "number",
		IsRequired:      true,
		MinValue:        ptrFloat64(800),
		MaxValue:        ptrFloat64(25600),
	}
	if err := db.Create(&dpiAttr).Error; err != nil {
		t.Fatal(err)
	}

	product := &models.Product{
		Name:            "MX Master 3",
		BrandID:         brand.ID,
		SpecificationID: &spec.ID,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	// Test 1: Missing required attribute
	form := url.Values{}
	req := httptest.NewRequest("POST", fmt.Sprintf("/products/%d/attributes", product.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400 for missing required attribute, got %d", resp.StatusCode)
	}

	// Test 2: Value below minimum
	form = url.Values{}
	form.Set(fmt.Sprintf("attr_%d", dpiAttr.ID), "500")
	req = httptest.NewRequest("POST", fmt.Sprintf("/products/%d/attributes", product.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400 for value below minimum, got %d", resp.StatusCode)
	}

	// Test 3: Value above maximum
	form = url.Values{}
	form.Set(fmt.Sprintf("attr_%d", dpiAttr.ID), "30000")
	req = httptest.NewRequest("POST", fmt.Sprintf("/products/%d/attributes", product.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400 for value above maximum, got %d", resp.StatusCode)
	}

	// Test 4: Valid value
	form = url.Values{}
	form.Set(fmt.Sprintf("attr_%d", dpiAttr.ID), "4000")
	req = httptest.NewRequest("POST", fmt.Sprintf("/products/%d/attributes", product.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200 for valid value, got %d: %s", resp.StatusCode, string(body))
	}
}

// Helper function for float64 pointers
func ptrFloat64(f float64) *float64 {
	return &f
}
