package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
)

func setupProcurementTestData(t *testing.T) (*config.Config, uint) {
	cfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run migrations for all models
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.SpecificationAttribute{},
		&models.Product{},
		&models.ProductAttribute{},
		&models.Quote{},
		&models.Forex{},
		&models.Requisition{},
		&models.RequisitionItem{},
		&models.PurchaseOrder{},
		&models.Document{},
		&models.VendorRating{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
		&models.ProjectProcurementStrategy{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test data
	vendorSvc := services.NewVendorService(cfg.DB)
	brandSvc := services.NewBrandService(cfg.DB)
	specSvc := services.NewSpecificationService(cfg.DB)
	productSvc := services.NewProductService(cfg.DB)
	forexSvc := services.NewForexService(cfg.DB)
	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)

	// Create forex rate for USD
	forexSvc.Create("USD", "USD", 1.0, time.Now())

	vendor1, err := vendorSvc.Create("Test Vendor 1", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor1: %v", err)
	}
	vendor2, err := vendorSvc.Create("Test Vendor 2", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor2: %v", err)
	}
	brand, err := brandSvc.Create("Test Brand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}
	spec1, err := specSvc.Create("Specification 1", "Test spec 1")
	if err != nil {
		t.Fatalf("Failed to create spec1: %v", err)
	}
	spec2, err := specSvc.Create("Specification 2", "Test spec 2")
	if err != nil {
		t.Fatalf("Failed to create spec2: %v", err)
	}

	product1, err := productSvc.Create("Product 1", brand.ID, &spec1.ID)
	if err != nil {
		t.Fatalf("Failed to create product1: %v", err)
	}
	product2, err := productSvc.Create("Product 2", brand.ID, &spec2.ID)
	if err != nil {
		t.Fatalf("Failed to create product2: %v", err)
	}

	// Create quotes
	validUntil := time.Now().AddDate(0, 0, 60)
	quoteSvc.Create(services.CreateQuoteInput{
		VendorID:   vendor1.ID,
		ProductID:  product1.ID,
		Price:      1000.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	quoteSvc.Create(services.CreateQuoteInput{
		VendorID:   vendor2.ID,
		ProductID:  product1.ID,
		Price:      900.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	quoteSvc.Create(services.CreateQuoteInput{
		VendorID:   vendor1.ID,
		ProductID:  product2.ID,
		Price:      2000.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})

	// Create project
	deadline := time.Now().AddDate(0, 0, 90)
	project, _ := projectSvc.Create("Test Procurement Project", "Test project for CLI", 50000.0, &deadline)

	// Add BOM items
	projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 10, "Item 1")
	projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 5, "Item 2")

	return cfg, project.ID
}

func TestProcurementAnalyzeCommand(t *testing.T) {
	testCfg, projectID := setupProcurementTestData(t)
	defer func() { _ = testCfg.Close() }()

	// Set global config for CLI and disable PreRun
	cfg = testCfg
	rootCmd.PersistentPreRun = nil
	defer func() {
		rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			initConfig()
		}
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	rootCmd.SetArgs([]string{"procurement", "analyze", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Command failed: %v", err)
	}

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected content
	if !contains(output, "Bill of Materials Analysis") {
		t.Error("Expected output to contain 'Bill of Materials Analysis'")
	}
}

func TestProcurementDashboardCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	// Execute dashboard command
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "dashboard", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Dashboard command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check for expected sections
	expectedSections := []string{"Progress:", "Financial:", "Procurement Status:"}
	for _, section := range expectedSections {
		if !contains(output, section) {
			t.Errorf("Expected output to contain '%s'", section)
		}
	}
}

func TestProcurementRisksCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "risks", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Risks command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Risk Assessment") {
		t.Error("Expected output to contain 'Risk Assessment'")
	}
}

func TestProcurementSavingsCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "savings", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Savings command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Savings Analysis") {
		t.Error("Expected output to contain 'Savings Analysis'")
	}
}

func TestStrategySetCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "strategy", "set", fmt.Sprintf("%d", projectID), "balanced"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Strategy set command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Strategy") {
		t.Error("Expected output to contain 'Strategy'")
	}

	// Verify strategy was created
	var strategy models.ProjectProcurementStrategy
	err = cfg.DB.Where("project_id = ?", projectID).First(&strategy).Error
	if err != nil {
		t.Errorf("Strategy not created: %v", err)
	}
	if strategy.Strategy != "balanced" {
		t.Errorf("Expected strategy 'balanced', got '%s'", strategy.Strategy)
	}
}

func TestStrategyShowCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	// Create a strategy first
	strategy := models.ProjectProcurementStrategy{
		ProjectID: projectID,
		Strategy:  "lowest_cost",
	}
	cfg.DB.Create(&strategy)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "strategy", "show", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Strategy show command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "lowest_cost") {
		t.Error("Expected output to contain 'lowest_cost'")
	}
}

func TestStrategyCompareCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "strategy", "compare", fmt.Sprintf("%d", projectID)})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Strategy compare command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Scenario Comparison") {
		t.Error("Expected output to contain 'Scenario Comparison'")
	}
}

func TestRecommendGenerateCommand(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"procurement", "recommend", "generate", fmt.Sprintf("%d", projectID), "--strategy", "balanced"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Recommend generate command failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Vendor Recommendations") {
		t.Error("Expected output to contain 'Vendor Recommendations'")
	}
}

func TestInvalidProjectID(t *testing.T) {
	cfg, _ := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	rootCmd.SetArgs([]string{"procurement", "analyze", "invalid"})
	err := rootCmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	if err == nil {
		t.Error("Expected error for invalid project ID")
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Error") {
		t.Error("Expected error message in output")
	}
}

func TestNonExistentProject(t *testing.T) {
	cfg, _ := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	rootCmd.SetArgs([]string{"procurement", "analyze", "999999"})
	err := rootCmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	if err == nil {
		t.Error("Expected error for non-existent project")
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Error") {
		t.Error("Expected error message in output")
	}
}

func TestStrategySetInvalidType(t *testing.T) {
	cfg, projectID := setupProcurementTestData(t)
	defer func() { _ = cfg.Close() }()

	setTestConfig(cfg)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	rootCmd.SetArgs([]string{"procurement", "strategy", "set", fmt.Sprintf("%d", projectID), "invalid_strategy"})
	err := rootCmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	if err == nil {
		t.Error("Expected error for invalid strategy type")
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !contains(output, "Error") || !contains(output, "Invalid") {
		t.Error("Expected error message about invalid strategy type")
	}
}

// Helper function to set the global config for testing
func setTestConfig(testCfg *config.Config) {
	cfg = testCfg
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
