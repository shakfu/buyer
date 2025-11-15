package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
)

// TestProcurementIntegration_EndToEndWorkflow tests the complete procurement workflow
func TestProcurementIntegration_EndToEndWorkflow(t *testing.T) {
	cfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer func() { _ = cfg.Close() }()

	// Migrate all models
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.SpecificationAttribute{},
		&models.Product{},
		&models.ProductAttribute{},
		&models.Quote{},
		&models.Forex{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
		&models.ProjectProcurementStrategy{},
		&models.VendorRating{},
		&models.PurchaseOrder{},
		&models.Document{},
	); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Initialize services
	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Setup forex rate
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex rate: %v", err)
	}

	// Step 1: Create vendors
	vendor1, err := vendorSvc.Create("Acme Corp", "USD", "CORP2024")
	if err != nil {
		t.Fatalf("Failed to create vendor1: %v", err)
	}

	vendor2, err := vendorSvc.Create("Global Supply", "USD", "GLOBAL123")
	if err != nil {
		t.Fatalf("Failed to create vendor2: %v", err)
	}

	vendor3, err := vendorSvc.Create("Quality Parts Inc", "USD", "QPI2024")
	if err != nil {
		t.Fatalf("Failed to create vendor3: %v", err)
	}

	// Step 2: Create brand and specifications
	brand, err := brandSvc.Create("TechBrand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	spec1, err := specSvc.Create("Widget A", "Standard widget")
	if err != nil {
		t.Fatalf("Failed to create spec1: %v", err)
	}

	spec2, err := specSvc.Create("Widget B", "Premium widget")
	if err != nil {
		t.Fatalf("Failed to create spec2: %v", err)
	}

	spec3, err := specSvc.Create("Widget C", "Economy widget")
	if err != nil {
		t.Fatalf("Failed to create spec3: %v", err)
	}

	// Step 3: Create products
	product1, err := productSvc.Create("Widget A Pro", brand.ID, &spec1.ID)
	if err != nil {
		t.Fatalf("Failed to create product1: %v", err)
	}

	product2, err := productSvc.Create("Widget B Elite", brand.ID, &spec2.ID)
	if err != nil {
		t.Fatalf("Failed to create product2: %v", err)
	}

	product3, err := productSvc.Create("Widget C Basic", brand.ID, &spec3.ID)
	if err != nil {
		t.Fatalf("Failed to create product3: %v", err)
	}

	// Step 4: Create quotes from multiple vendors
	validUntil := time.Now().AddDate(0, 0, 60)

	// Vendor1 quotes (competitive on product1, expensive on product2)
	quote1_1, err := quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor1.ID,
		ProductID:  product1.ID,
		Price:      100.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote1_1: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor1.ID,
		ProductID:  product2.ID,
		Price:      250.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote1_2: %v", err)
	}

	// Vendor2 quotes (best on product2, competitive on product3)
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor2.ID,
		ProductID:  product2.ID,
		Price:      200.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote2_2: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor2.ID,
		ProductID:  product3.ID,
		Price:      75.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote2_3: %v", err)
	}

	// Vendor3 quotes (all products available, medium prices)
	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor3.ID,
		ProductID:  product1.ID,
		Price:      110.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote3_1: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor3.ID,
		ProductID:  product2.ID,
		Price:      220.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote3_2: %v", err)
	}

	_, err = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor3.ID,
		ProductID:  product3.ID,
		Price:      80.0,
		Currency:   "USD",
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("Failed to create quote3_3: %v", err)
	}

	// Step 5: Create project with BOM
	deadline := time.Now().AddDate(0, 0, 90)
	project, err := projectSvc.Create("Integration Test Project", "End-to-end test", 50000.0, &deadline)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Add BOM items
	bomItem1, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 100, "Widget A batch")
	if err != nil {
		t.Fatalf("Failed to add BOM item 1: %v", err)
	}

	bomItem2, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 50, "Widget B batch")
	if err != nil {
		t.Fatalf("Failed to add BOM item 2: %v", err)
	}

	bomItem3, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec3.ID, 200, "Widget C batch")
	if err != nil {
		t.Fatalf("Failed to add BOM item 3: %v", err)
	}

	// Step 6: Create project requisitions
	req1 := models.ProjectRequisition{
		ProjectID:     project.ID,
		Name:          "Phase 1 Requisition",
		Justification: "Initial procurement phase",
		Budget:        25000.0,
	}
	if err := cfg.DB.Create(&req1).Error; err != nil {
		t.Fatalf("Failed to create requisition 1: %v", err)
	}

	// Add requisition items with target prices for savings calculation
	reqItem1 := models.ProjectRequisitionItem{
		ProjectRequisitionID:  req1.ID,
		BillOfMaterialsItemID: bomItem1.ID,
		QuantityRequested:     50,
		TargetUnitPrice:       120.0, // Higher than best quote (100) to show savings
		ProcurementStatus:     "pending",
		Notes:                 "Half of Widget A",
	}
	if err := cfg.DB.Create(&reqItem1).Error; err != nil {
		t.Fatalf("Failed to add req item 1: %v", err)
	}

	reqItem2 := models.ProjectRequisitionItem{
		ProjectRequisitionID:  req1.ID,
		BillOfMaterialsItemID: bomItem2.ID,
		QuantityRequested:     25,
		TargetUnitPrice:       250.0, // Higher than best quote (200) to show savings
		ProcurementStatus:     "pending",
		Notes:                 "Half of Widget B",
	}
	if err := cfg.DB.Create(&reqItem2).Error; err != nil {
		t.Fatalf("Failed to add req item 2: %v", err)
	}

	reqItem3 := models.ProjectRequisitionItem{
		ProjectRequisitionID:  req1.ID,
		BillOfMaterialsItemID: bomItem3.ID,
		QuantityRequested:     100,
		TargetUnitPrice:       90.0, // Higher than best quote (75) to show savings
		ProcurementStatus:     "pending",
		Notes:                 "Half of Widget C",
	}
	if err := cfg.DB.Create(&reqItem3).Error; err != nil {
		t.Fatalf("Failed to add req item 3: %v", err)
	}

	// Step 7: Test procurement analysis
	t.Run("GetProjectProcurementComparison", func(t *testing.T) {
		comparison, err := procurementSvc.GetProjectProcurementComparison(project.ID)
		if err != nil {
			t.Fatalf("GetProjectProcurementComparison failed: %v", err)
		}

		if comparison.Project.ID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, comparison.Project.ID)
		}

		if comparison.TotalBOMItems != 3 {
			t.Errorf("Expected 3 BOM items, got %d", comparison.TotalBOMItems)
		}

		if len(comparison.BOMItemAnalyses) != 3 {
			t.Errorf("Expected 3 BOM item analyses, got %d", len(comparison.BOMItemAnalyses))
		}

		// Verify each BOM item has quotes
		for _, analysis := range comparison.BOMItemAnalyses {
			if len(analysis.AvailableQuotes) == 0 {
				t.Errorf("BOM item %d has no available quotes", analysis.BOMItem.ID)
			}

			if analysis.BestQuote == nil {
				t.Errorf("BOM item %d has no best quote", analysis.BOMItem.ID)
			}
		}
	})

	// Step 8: Test savings calculation
	t.Run("CalculateProjectSavings", func(t *testing.T) {
		savings, err := procurementSvc.CalculateProjectSavings(project.ID)
		if err != nil {
			t.Fatalf("CalculateProjectSavings failed: %v", err)
		}

		if len(savings.DetailedBreakdown) == 0 {
			t.Error("Expected detailed breakdown")
		} else {
			// Debug output
			t.Logf("Detailed Breakdown:")
			for i, item := range savings.DetailedBreakdown {
				t.Logf("  Item %d: %s", i+1, item.SpecificationName)
				t.Logf("    Quantity: %d", item.Quantity)
				t.Logf("    Target Price: %.2f", item.TargetPrice)
				t.Logf("    Recommended Price: %.2f", item.RecommendedPrice)
				t.Logf("    Best Price: %.2f", item.BestPrice)
				t.Logf("    Savings Per Unit: %.2f", item.SavingsPerUnit)
				t.Logf("    Total Savings: %.2f", item.TotalSavings)
			}
		}

		if savings.TotalSavingsUSD < 0 {
			t.Errorf("Expected non-negative savings, got %.2f", savings.TotalSavingsUSD)
		}
	})

	// Step 9: Test risk assessment
	t.Run("AssessEnhancedProjectRisks", func(t *testing.T) {
		risks, err := procurementSvc.AssessEnhancedProjectRisks(project.ID)
		if err != nil {
			t.Fatalf("AssessEnhancedProjectRisks failed: %v", err)
		}

		if risks.OverallRisk == "" {
			t.Error("Expected overall risk level")
		}

		if len(risks.CategoryRisks) == 0 {
			t.Error("Expected category risks")
		}

		// Verify risk scoring
		if risks.RiskScore < 0 || risks.RiskScore > 100 {
			t.Errorf("Risk score %d out of valid range [0-100]", risks.RiskScore)
		}
	})

	// Step 10: Test vendor recommendations
	t.Run("GenerateVendorRecommendations_LowestCost", func(t *testing.T) {
		recs, err := procurementSvc.GenerateVendorRecommendations(project.ID, "lowest_cost")
		if err != nil {
			t.Fatalf("GenerateVendorRecommendations failed: %v", err)
		}

		if len(recs) == 0 {
			t.Error("Expected at least one vendor recommendation")
		}

		// Verify that lowest_cost strategy picks cheapest quotes
		for _, vendorRec := range recs {
			if vendorRec.VendorID != vendor1.ID && vendorRec.VendorID != vendor2.ID && vendorRec.VendorID != vendor3.ID {
				t.Errorf("Unexpected vendor ID: %d", vendorRec.VendorID)
			}
		}

		// Calculate total cost from recommendations
		var totalCost float64
		for _, rec := range recs {
			totalCost += rec.TotalCost
		}

		// Expected: Vendor1 for Widget A ($100*100=$10k), Vendor2 for Widget B ($200*50=$10k), Vendor2 for Widget C ($75*200=$15k)
		// Total BOM-based cost = $35,000
		expectedCost := 35000.0
		if totalCost < expectedCost*0.9 || totalCost > expectedCost*1.1 {
			t.Errorf("Expected total cost around %.2f, got %.2f", expectedCost, totalCost)
		}
	})

	t.Run("GenerateVendorRecommendations_FewestVendors", func(t *testing.T) {
		recs, err := procurementSvc.GenerateVendorRecommendations(project.ID, "fewest_vendors")
		if err != nil {
			t.Fatalf("GenerateVendorRecommendations failed: %v", err)
		}

		// Should recommend vendor3 since they can supply all products
		vendorIDs := make(map[uint]bool)
		for _, vendorRec := range recs {
			vendorIDs[vendorRec.VendorID] = true
		}

		// Fewest vendors should ideally use 1-2 vendors
		if len(vendorIDs) > 2 {
			t.Errorf("Expected <= 2 vendors for fewest_vendors strategy, got %d", len(vendorIDs))
		}
	})

	// Step 11: Test scenario comparison
	t.Run("CompareScenarios", func(t *testing.T) {
		scenarios, err := procurementSvc.CompareScenarios(project.ID)
		if err != nil {
			t.Fatalf("CompareScenarios failed: %v", err)
		}

		if len(scenarios) < 3 {
			t.Errorf("Expected at least 3 scenarios, got %d", len(scenarios))
		}

		// Verify scenarios have different vendor counts and costs
		seenNames := make(map[string]bool)
		for _, scenario := range scenarios {
			if seenNames[scenario.Name] {
				t.Errorf("Duplicate scenario name: %s", scenario.Name)
			}
			seenNames[scenario.Name] = true

			if scenario.TotalCost <= 0 {
				t.Errorf("Scenario %s has invalid cost: %.2f", scenario.Name, scenario.TotalCost)
			}

			if scenario.VendorCount <= 0 {
				t.Errorf("Scenario %s has invalid vendor count: %d", scenario.Name, scenario.VendorCount)
			}
		}
	})

	// Step 12: Test dashboard
	t.Run("GetProjectDashboard", func(t *testing.T) {
		dashboard, err := procurementSvc.GetProjectDashboard(project.ID)
		if err != nil {
			t.Fatalf("GetProjectDashboard failed: %v", err)
		}

		if dashboard.Project.ID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, dashboard.Project.ID)
		}

		// Verify progress metrics
		if dashboard.Progress.BOMCoverage < 0 || dashboard.Progress.BOMCoverage > 100 {
			t.Errorf("Invalid BOM coverage: %.2f", dashboard.Progress.BOMCoverage)
		}

		// Verify financial metrics
		if dashboard.Financial.Budget != project.Budget {
			t.Errorf("Expected budget %.2f, got %.2f", project.Budget, dashboard.Financial.Budget)
		}

		// Verify procurement status
		if dashboard.Procurement.TotalItems != 3 {
			t.Errorf("Expected 3 total items, got %d", dashboard.Procurement.TotalItems)
		}
	})

	// Step 13: Test vendor consolidation
	t.Run("GetVendorConsolidationAnalysis", func(t *testing.T) {
		consolidation, err := procurementSvc.GetVendorConsolidationAnalysis(project.ID)
		if err != nil {
			t.Fatalf("GetVendorConsolidationAnalysis failed: %v", err)
		}

		if len(consolidation) == 0 {
			t.Error("Expected vendor analyses")
		}

		// Verify vendor3 can supply all items
		var foundFullCoverage bool
		for _, va := range consolidation {
			if va.SpecificationsCount == 3 {
				foundFullCoverage = true
				if va.VendorID != vendor3.ID {
					t.Errorf("Expected vendor3 (%d) to have full coverage, got %d", vendor3.ID, va.VendorID)
				}
			}
		}

		if !foundFullCoverage {
			t.Error("Expected at least one vendor with full coverage")
		}
	})

	// Step 14: Test strategy persistence
	t.Run("GetOrCreateStrategy", func(t *testing.T) {
		strategy, err := procurementSvc.GetOrCreateStrategy(project.ID)
		if err != nil {
			t.Fatalf("GetOrCreateStrategy failed: %v", err)
		}

		if strategy.ProjectID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, strategy.ProjectID)
		}

		if strategy.Strategy == "" {
			t.Error("Expected default strategy to be set")
		}

		// Update strategy
		strategy.Strategy = "fewest_vendors"
		if err := cfg.DB.Save(strategy).Error; err != nil {
			t.Fatalf("Failed to update strategy: %v", err)
		}

		// Verify it persists
		strategy2, err := procurementSvc.GetOrCreateStrategy(project.ID)
		if err != nil {
			t.Fatalf("GetOrCreateStrategy failed on second call: %v", err)
		}

		if strategy2.Strategy != "fewest_vendors" {
			t.Errorf("Expected strategy 'fewest_vendors', got '%s'", strategy2.Strategy)
		}
	})

	// Step 15: Test with quote selections
	t.Run("SelectQuotesAndRecalculate", func(t *testing.T) {
		// Select specific quotes for requisition items
		var reqItem models.ProjectRequisitionItem
		err := cfg.DB.Preload("BOMItem.Specification").
			Where("project_requisition_id = ?", req1.ID).
			Where("EXISTS (SELECT 1 FROM bill_of_materials_items WHERE id = bill_of_materials_item_id AND specification_id = ?)", spec1.ID).
			First(&reqItem).Error
		if err != nil {
			t.Fatalf("Failed to find requisition item: %v", err)
		}

		// Select quote1_1 (vendor1, best price for product1)
		reqItem.SelectedQuoteID = &quote1_1.ID
		if err := cfg.DB.Save(&reqItem).Error; err != nil {
			t.Fatalf("Failed to save selected quote: %v", err)
		}

		// Recalculate comparison
		comparison, err := procurementSvc.GetProjectProcurementComparison(project.ID)
		if err != nil {
			t.Fatalf("GetProjectProcurementComparison failed after selection: %v", err)
		}

		// Verify selection is reflected
		var foundSelectedQuote bool
		for _, analysis := range comparison.BOMItemAnalyses {
			for _, reqItemQuote := range analysis.RequisitionItems {
				if reqItemQuote.SelectedQuote != nil && reqItemQuote.SelectedQuote.ID == quote1_1.ID {
					foundSelectedQuote = true
				}
			}
		}

		if !foundSelectedQuote {
			t.Error("Selected quote not reflected in analysis")
		}
	})

	t.Logf("Integration test completed successfully!")
	t.Logf("Project ID: %d", project.ID)
	t.Logf("Total Vendors: %d", 3)
	t.Logf("Total BOM Items: %d", 3)
	t.Logf("Total Quotes: %d", 6)
	t.Logf("Total Requisitions: %d", 1)
}

// TestProcurementIntegration_LargeProject tests performance with larger datasets
func TestProcurementIntegration_LargeProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large project test in short mode")
	}

	cfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer func() { _ = cfg.Close() }()

	// Migrate
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.SpecificationAttribute{},
		&models.Product{},
		&models.ProductAttribute{},
		&models.Quote{},
		&models.Forex{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
		&models.ProjectProcurementStrategy{},
		&models.VendorRating{},
		&models.PurchaseOrder{},
		&models.Document{},
	); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Setup services
	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Setup forex
	if _, err := forexSvc.Create("USD", "USD", 1.0, time.Now()); err != nil {
		t.Fatalf("Failed to create forex rate: %v", err)
	}

	// Create multiple vendors (10)
	vendors := make([]*models.Vendor, 10)
	for i := 0; i < 10; i++ {
		vendor, err := vendorSvc.Create(fmt.Sprintf("Vendor %d", i+1), "USD", "")
		if err != nil {
			t.Fatalf("Failed to create vendor %d: %v", i+1, err)
		}
		vendors[i] = vendor
	}

	// Create brand
	brand, err := brandSvc.Create("TestBrand")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	// Create multiple specifications (50)
	specs := make([]*models.Specification, 50)
	for i := 0; i < 50; i++ {
		spec, err := specSvc.Create(fmt.Sprintf("Spec %d", i+1), "Test spec")
		if err != nil {
			t.Fatalf("Failed to create spec %d: %v", i+1, err)
		}
		specs[i] = spec
	}

	// Create products and quotes
	validUntil := time.Now().AddDate(0, 0, 60)
	for _, spec := range specs {
		product, err := productSvc.Create(fmt.Sprintf("Product for %s", spec.Name), brand.ID, &spec.ID)
		if err != nil {
			t.Fatalf("Failed to create product for spec %s: %v", spec.Name, err)
		}

		// Each product gets 3 quotes from different vendors
		for j := 0; j < 3; j++ {
			vendorIdx := (int(spec.ID) + j) % len(vendors)
			price := 100.0 + float64(j)*10.0 + float64(spec.ID)*2.0

			_, err := quoteSvc.Create(CreateQuoteInput{
				VendorID:   vendors[vendorIdx].ID,
				ProductID:  product.ID,
				Price:      price,
				Currency:   "USD",
				ValidUntil: &validUntil,
			})
			if err != nil {
				t.Fatalf("Failed to create quote for product %d: %v", product.ID, err)
			}
		}
	}

	// Create large project with many BOM items
	deadline := time.Now().AddDate(0, 0, 90)
	project, err := projectSvc.Create("Large Project", "Performance test", 500000.0, &deadline)
	if err != nil {
		t.Fatalf("Failed to create large project: %v", err)
	}

	// Add all specifications to BOM
	for _, spec := range specs {
		qty := 10 + int(spec.ID)*5
		_, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, qty, "Batch")
		if err != nil {
			t.Fatalf("Failed to add BOM item for spec %d: %v", spec.ID, err)
		}
	}

	// Performance test: Analysis should complete in reasonable time
	start := time.Now()
	comparison, err := procurementSvc.GetProjectProcurementComparison(project.ID)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Large project analysis failed: %v", err)
	}

	t.Logf("Large project analysis completed in %v", duration)
	t.Logf("BOM Items: %d", comparison.TotalBOMItems)
	t.Logf("Vendor count: %d", len(vendors))

	// Verify performance target: < 5 seconds for 50 items
	if duration > 5*time.Second {
		t.Errorf("Analysis took too long: %v (target: < 5s)", duration)
	}

	// Test dashboard performance
	start = time.Now()
	dashboard, err := procurementSvc.GetProjectDashboard(project.ID)
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("Large project dashboard failed: %v", err)
	}

	t.Logf("Dashboard generation completed in %v", duration)

	// Dashboard should be fast: < 2 seconds
	if duration > 2*time.Second {
		t.Errorf("Dashboard generation took too long: %v (target: < 2s)", duration)
	}

	// Verify dashboard completeness
	if dashboard.Procurement.TotalItems != 50 {
		t.Errorf("Expected 50 total items, got %d", dashboard.Procurement.TotalItems)
	}
}
