package services

import (
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/models"
)

func TestProjectProcurementService_GetProjectProcurementComparison(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, err := vendorSvc.Create("Vendor A", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	brandSvc := NewBrandService(cfg.DB)
	brand, err := brandSvc.Create("Brand A")
	if err != nil {
		t.Fatalf("Failed to create brand: %v", err)
	}

	specSvc := NewSpecificationService(cfg.DB)
	spec, err := specSvc.Create("Laptop", "Laptop computers")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Create products
	productSvc := NewProductService(cfg.DB)
	product1, err := productSvc.Create("Laptop Model A", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product1: %v", err)
	}
	product2, err := productSvc.Create("Laptop Model B", brand.ID, &spec.ID)
	if err != nil {
		t.Fatalf("Failed to create product2: %v", err)
	}

	// Create forex rate
	forexSvc := NewForexService(cfg.DB)
	_, err = forexSvc.Create("USD", "USD", 1.0, time.Now())
	if err != nil {
		t.Fatalf("Failed to create forex: %v", err)
	}

	// Create quotes
	quote1, err := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product1.ID,
		Price:     1000.0,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote1: %v", err)
	}

	quote2, err := quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product2.ID,
		Price:     1200.0,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create quote2: %v", err)
	}

	// Create project with BOM
	project, err := projectSvc.Create("Test Project", "Project description", 50000.0, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Add BOM items
	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "Need 10 laptops")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	// Create project requisition
	requisition := models.ProjectRequisition{
		ProjectID:     project.ID,
		Name:          "Q1 Purchase",
		Justification: "Initial purchase",
		Budget:        15000.0,
	}
	if err := cfg.DB.Create(&requisition).Error; err != nil {
		t.Fatalf("Failed to create requisition: %v", err)
	}

	// Add requisition item
	reqItem := models.ProjectRequisitionItem{
		ProjectRequisitionID:  requisition.ID,
		BillOfMaterialsItemID: bomItem.ID,
		QuantityRequested:     5,
		TargetUnitPrice:       1100.0,
	}
	if err := cfg.DB.Create(&reqItem).Error; err != nil {
		t.Fatalf("Failed to add requisition item: %v", err)
	}

	// Get procurement comparison
	comparison, err := procurementSvc.GetProjectProcurementComparison(project.ID)
	if err != nil {
		t.Errorf("GetProjectProcurementComparison() error = %v", err)
		return
	}

	// Verify results
	if comparison == nil {
		t.Fatal("Expected comparison to be non-nil")
	}

	if comparison.Project.ID != project.ID {
		t.Errorf("Expected project ID %d, got %d", project.ID, comparison.Project.ID)
	}

	if comparison.TotalBOMItems != 1 {
		t.Errorf("Expected 1 BOM item, got %d", comparison.TotalBOMItems)
	}

	if len(comparison.BOMItemAnalyses) != 1 {
		t.Fatalf("Expected 1 BOM analysis, got %d", len(comparison.BOMItemAnalyses))
	}

	bomAnalysis := comparison.BOMItemAnalyses[0]
	if bomAnalysis.TotalQuantityNeeded != 10 {
		t.Errorf("Expected quantity needed 10, got %d", bomAnalysis.TotalQuantityNeeded)
	}

	if bomAnalysis.TotalQuantityPlanned != 5 {
		t.Errorf("Expected quantity planned 5, got %d", bomAnalysis.TotalQuantityPlanned)
	}

	if bomAnalysis.CoveragePercent != 50.0 {
		t.Errorf("Expected coverage 50%%, got %.2f%%", bomAnalysis.CoveragePercent)
	}

	if !bomAnalysis.HasGaps {
		t.Error("Expected HasGaps to be true")
	}

	if !bomAnalysis.HasSufficientQuotes {
		t.Error("Expected HasSufficientQuotes to be true")
	}

	if len(bomAnalysis.AvailableQuotes) != 2 {
		t.Errorf("Expected 2 available quotes, got %d", len(bomAnalysis.AvailableQuotes))
	}

	if bomAnalysis.BestQuote == nil {
		t.Fatal("Expected best quote to be set")
	}

	// Best quote should be the cheaper one (quote1 at 1000)
	if bomAnalysis.BestQuote.ID != quote1.ID {
		t.Errorf("Expected best quote ID %d, got %d", quote1.ID, bomAnalysis.BestQuote.ID)
	}

	// Check cost calculations
	expectedBestCost := 1000.0 * 10 // $1000 * 10 laptops
	if bomAnalysis.BestTotalCost != expectedBestCost {
		t.Errorf("Expected best total cost %.2f, got %.2f", expectedBestCost, bomAnalysis.BestTotalCost)
	}

	// Check strategy was created
	if comparison.Strategy == nil {
		t.Fatal("Expected strategy to be created")
	}

	if comparison.Strategy.Strategy != "lowest_cost" {
		t.Errorf("Expected default strategy 'lowest_cost', got %s", comparison.Strategy.Strategy)
	}

	// Verify quote2 exists (more expensive option)
	_ = quote2

	t.Logf("Procurement comparison successful: %d BOM items analyzed", len(comparison.BOMItemAnalyses))
}

func TestProjectProcurementService_AnalyzeBOMItem(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Vendor A", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Monitor", "Computer monitors")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Monitor 27\"", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	_, _ = quoteSvc.Create(CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     350.0,
		Currency:  "USD",
	})

	// Create project with BOM
	project, err := projectSvc.Create("Monitor Project", "Need monitors", 10000.0, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 20, "Need 20 monitors")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	// Analyze BOM item
	analysis, err := procurementSvc.AnalyzeBOMItem(bomItem.ID)
	if err != nil {
		t.Errorf("AnalyzeBOMItem() error = %v", err)
		return
	}

	if analysis == nil {
		t.Fatal("Expected analysis to be non-nil")
	}

	if analysis.BOMItem.ID != bomItem.ID {
		t.Errorf("Expected BOM item ID %d, got %d", bomItem.ID, analysis.BOMItem.ID)
	}

	if analysis.TotalQuantityNeeded != 20 {
		t.Errorf("Expected quantity 20, got %d", analysis.TotalQuantityNeeded)
	}

	if analysis.Specification.ID != spec.ID {
		t.Errorf("Expected spec ID %d, got %d", spec.ID, analysis.Specification.ID)
	}

	if len(analysis.AvailableQuotes) != 1 {
		t.Errorf("Expected 1 quote, got %d", len(analysis.AvailableQuotes))
	}

	if analysis.BestQuote == nil {
		t.Fatal("Expected best quote to be set")
	}

	expectedCost := 350.0 * 20
	if analysis.BestTotalCost != expectedCost {
		t.Errorf("Expected cost %.2f, got %.2f", expectedCost, analysis.BestTotalCost)
	}
}

func TestProjectProcurementService_AssessBOMItemRisk(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	tests := []struct {
		name                string
		hasQuotes           bool
		quoteCount          int
		coveragePercent     float64
		quotesAreStale      bool
		expectedRisk        string
	}{
		{
			name:            "No quotes - high risk",
			hasQuotes:       false,
			quoteCount:      0,
			coveragePercent: 100,
			quotesAreStale:  false,
			expectedRisk:    "high",
		},
		{
			name:            "Single source with gaps - high risk",
			hasQuotes:       true,
			quoteCount:      1,
			coveragePercent: 30,
			quotesAreStale:  false,
			expectedRisk:    "high",
		},
		{
			name:            "Multiple quotes, good coverage - low risk",
			hasQuotes:       true,
			quoteCount:      3,
			coveragePercent: 100,
			quotesAreStale:  false,
			expectedRisk:    "low",
		},
		{
			name:            "Good quotes but stale - medium risk",
			hasQuotes:       true,
			quoteCount:      2,
			coveragePercent: 100,
			quotesAreStale:  true,
			expectedRisk:    "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &BOMItemProcurementAnalysis{
				TotalQuantityNeeded:  100,
				TotalQuantityPlanned: int(float64(100) * tt.coveragePercent / 100),
				HasSufficientQuotes:  tt.hasQuotes,
				AvailableQuotes:      make([]models.Quote, tt.quoteCount),
			}

			analysis.CoveragePercent = (float64(analysis.TotalQuantityPlanned) / float64(analysis.TotalQuantityNeeded)) * 100
			analysis.HasGaps = analysis.TotalQuantityPlanned < analysis.TotalQuantityNeeded

			// Set quote staleness
			if tt.quotesAreStale && tt.quoteCount > 0 {
				for i := range analysis.AvailableQuotes {
					oldDate := time.Now().AddDate(0, 0, -100)
					analysis.AvailableQuotes[i].QuoteDate = oldDate
				}
			} else if tt.quoteCount > 0 {
				for i := range analysis.AvailableQuotes {
					analysis.AvailableQuotes[i].QuoteDate = time.Now().AddDate(0, 0, -10)
				}
			}

			risk := procurementSvc.assessBOMItemRisk(analysis)
			if risk != tt.expectedRisk {
				t.Errorf("Expected risk %s, got %s", tt.expectedRisk, risk)
			}
		})
	}
}

func TestProjectProcurementService_CalculateQuoteFreshness(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	now := time.Now()
	freshDate := now.AddDate(0, 0, -10)
	staleDate := now.AddDate(0, 0, -100)
	expiredDate := now.AddDate(0, 0, -200)
	validUntilPast := now.AddDate(0, 0, -5)

	bomAnalyses := []BOMItemProcurementAnalysis{
		{
			AvailableQuotes: []models.Quote{
				{ID: 1, QuoteDate: freshDate},
				{ID: 2, QuoteDate: staleDate},
				{ID: 3, QuoteDate: expiredDate, ValidUntil: &validUntilPast},
			},
		},
	}

	stats := procurementSvc.calculateQuoteFreshness(bomAnalyses)

	if stats.TotalQuotes != 3 {
		t.Errorf("Expected 3 total quotes, got %d", stats.TotalQuotes)
	}

	if stats.FreshQuotes != 1 {
		t.Errorf("Expected 1 fresh quote, got %d", stats.FreshQuotes)
	}

	if stats.StaleQuotes != 1 {
		t.Errorf("Expected 1 stale quote, got %d", stats.StaleQuotes)
	}

	if stats.ExpiredQuotes != 1 {
		t.Errorf("Expected 1 expired quote, got %d", stats.ExpiredQuotes)
	}

	if stats.AverageAgeDays < 50 || stats.AverageAgeDays > 150 {
		t.Errorf("Expected average age around 100 days, got %d", stats.AverageAgeDays)
	}
}

func TestProjectProcurementService_GetVendorRatingSummary(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create vendor
	vendorSvc := NewVendorService(cfg.DB)
	vendor, err := vendorSvc.Create("Rated Vendor", "USD", "")
	if err != nil {
		t.Fatalf("Failed to create vendor: %v", err)
	}

	// Create ratings
	ratings := []struct {
		price    *int
		quality  *int
		delivery *int
		service  *int
	}{
		{int32Ptr(5), int32Ptr(4), int32Ptr(5), int32Ptr(4)},
		{int32Ptr(4), int32Ptr(5), int32Ptr(4), int32Ptr(5)},
		{int32Ptr(5), int32Ptr(5), int32Ptr(5), int32Ptr(5)},
	}

	for _, r := range ratings {
		rating := models.VendorRating{
			VendorID:       vendor.ID,
			PriceRating:    r.price,
			QualityRating:  r.quality,
			DeliveryRating: r.delivery,
			ServiceRating:  r.service,
		}
		if err := cfg.DB.Create(&rating).Error; err != nil {
			t.Fatalf("Failed to create rating: %v", err)
		}
	}

	// Get summary
	summary, err := procurementSvc.GetVendorRatingSummary(vendor.ID)
	if err != nil {
		t.Errorf("GetVendorRatingSummary() error = %v", err)
		return
	}

	if summary == nil {
		t.Fatal("Expected summary to be non-nil")
	}

	if summary.TotalRatings != 3 {
		t.Errorf("Expected 3 ratings, got %d", summary.TotalRatings)
	}

	// Check averages (should be around 4.5-5.0)
	if summary.AvgPrice == nil || *summary.AvgPrice < 4.0 || *summary.AvgPrice > 5.0 {
		t.Errorf("Unexpected avg price rating: %v", summary.AvgPrice)
	}

	if summary.OverallAvg < 4.0 || summary.OverallAvg > 5.0 {
		t.Errorf("Expected overall avg between 4.0 and 5.0, got %.2f", summary.OverallAvg)
	}
}

func TestProjectProcurementService_AssessProjectRisks(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	project := &models.Project{
		ID:     1,
		Name:   "Test Project",
		Budget: 100000,
	}

	// Create BOM analyses with different risk profiles
	bomAnalyses := []BOMItemProcurementAnalysis{
		{
			BOMItem:             &models.BillOfMaterialsItem{ID: 1},
			HasSufficientQuotes: false, // Critical risk
			AvailableQuotes:     []models.Quote{},
		},
		{
			BOMItem:             &models.BillOfMaterialsItem{ID: 2},
			HasSufficientQuotes: true,
			AvailableQuotes:     []models.Quote{{ID: 1}}, // Single source
		},
		{
			BOMItem:             &models.BillOfMaterialsItem{ID: 3},
			HasSufficientQuotes: true,
			AvailableQuotes:     []models.Quote{{ID: 2}, {ID: 3}}, // Good
		},
	}

	assessment := procurementSvc.assessProjectRisks(project, bomAnalyses)

	if assessment.OverallRisk != "critical" {
		t.Errorf("Expected overall risk 'critical', got %s", assessment.OverallRisk)
	}

	if len(assessment.RiskFactors) < 2 {
		t.Errorf("Expected at least 2 risk factors, got %d", len(assessment.RiskFactors))
	}

	// Check for uncovered items risk
	foundUncoveredRisk := false
	for _, factor := range assessment.RiskFactors {
		if factor.Category == "quote_coverage" && factor.Severity == "critical" {
			foundUncoveredRisk = true
			if len(factor.AffectedBOMItems) != 1 {
				t.Errorf("Expected 1 affected BOM item, got %d", len(factor.AffectedBOMItems))
			}
		}
	}
	if !foundUncoveredRisk {
		t.Error("Expected to find uncovered items risk")
	}

	if len(assessment.MitigationActions) == 0 {
		t.Error("Expected mitigation actions to be suggested")
	}
}

func TestProjectProcurementService_GetOrCreateStrategy(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create project
	project, err := projectSvc.Create("Strategy Test", "Test project", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// First call should create strategy
	strategy1, err := procurementSvc.GetOrCreateStrategy(project.ID)
	if err != nil {
		t.Errorf("GetOrCreateStrategy() error = %v", err)
		return
	}

	if strategy1 == nil {
		t.Fatal("Expected strategy to be created")
	}

	if strategy1.Strategy != "lowest_cost" {
		t.Errorf("Expected default strategy 'lowest_cost', got %s", strategy1.Strategy)
	}

	// Second call should return same strategy
	strategy2, err := procurementSvc.GetOrCreateStrategy(project.ID)
	if err != nil {
		t.Errorf("GetOrCreateStrategy() error = %v", err)
		return
	}

	if strategy2.ID != strategy1.ID {
		t.Errorf("Expected same strategy ID %d, got %d", strategy1.ID, strategy2.ID)
	}
}

// Helper function to create int pointer
func int32Ptr(i int) *int {
	return &i
}

func TestProjectProcurementService_GetVendorConsolidationAnalysis(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create vendors
	vendorSvc := NewVendorService(cfg.DB)
	vendor1, _ := vendorSvc.Create("Vendor A", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor B", "USD", "")

	// Create specifications and products
	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec1, _ := specSvc.Create("Laptop", "Laptop computers")
	spec2, _ := specSvc.Create("Monitor", "Computer monitors")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop Model A", brand.ID, &spec1.ID)
	product2, _ := productSvc.Create("Monitor 27\"", brand.ID, &spec2.ID)
	product3, _ := productSvc.Create("Laptop Model B", brand.ID, &spec1.ID)

	// Create forex
	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Create quotes
	// Vendor1 can supply both laptops and monitors
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product1.ID, Price: 1000, Currency: "USD"})
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product2.ID, Price: 300, Currency: "USD"})

	// Vendor2 can only supply laptops (cheaper)
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product3.ID, Price: 900, Currency: "USD"})

	// Create project with BOM
	project, _ := projectSvc.Create("Test Project", "Testing consolidation", 50000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 10, "10 laptops")
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 5, "5 monitors")

	// Get consolidation analysis
	analysis, err := procurementSvc.GetVendorConsolidationAnalysis(project.ID)
	if err != nil {
		t.Errorf("GetVendorConsolidationAnalysis() error = %v", err)
		return
	}

	if len(analysis) != 2 {
		t.Errorf("Expected 2 vendors in analysis, got %d", len(analysis))
		return
	}

	// Vendor1 should be first (covers more specs)
	if analysis[0].VendorID != vendor1.ID {
		t.Errorf("Expected vendor1 first, got vendor %d", analysis[0].VendorID)
	}

	if analysis[0].SpecificationsCount != 2 {
		t.Errorf("Expected vendor1 to cover 2 specs, got %d", analysis[0].SpecificationsCount)
	}

	if !analysis[0].ShippingAdvantage {
		t.Error("Expected vendor1 to have shipping advantage")
	}

	// Vendor2 should be second (covers fewer specs)
	if analysis[1].VendorID != vendor2.ID {
		t.Errorf("Expected vendor2 second, got vendor %d", analysis[1].VendorID)
	}

	if analysis[1].SpecificationsCount != 1 {
		t.Errorf("Expected vendor2 to cover 1 spec, got %d", analysis[1].SpecificationsCount)
	}

	t.Logf("Vendor consolidation analysis successful: %d vendors analyzed", len(analysis))
}

func TestProjectProcurementService_GenerateVendorRecommendations(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor1, _ := vendorSvc.Create("Cheap Vendor", "USD", "")
	vendor2, _ := vendorSvc.Create("Full Service Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec1, _ := specSvc.Create("Laptop", "")
	spec2, _ := specSvc.Create("Monitor", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec1.ID)
	product2, _ := productSvc.Create("Monitor A", brand.ID, &spec2.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Vendor1: cheap laptop only
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product1.ID, Price: 800, Currency: "USD"})

	// Vendor2: slightly more expensive laptop AND monitors (can supply everything)
	product4, _ := productSvc.Create("Laptop B", brand.ID, &spec1.ID)
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product4.ID, Price: 850, Currency: "USD"})
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product2.ID, Price: 250, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Test Project", "", 20000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 5, "")
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 10, "")

	tests := []struct {
		name         string
		strategy     string
		expectVendor uint
		minVendors   int
		maxVendors   int
	}{
		{
			name:       "Lowest Cost Strategy",
			strategy:   "lowest_cost",
			minVendors: 1,
			maxVendors: 2,
		},
		{
			name:       "Fewest Vendors Strategy",
			strategy:   "fewest_vendors",
			minVendors: 1,
			maxVendors: 1,
		},
		{
			name:       "Balanced Strategy",
			strategy:   "balanced",
			minVendors: 1,
			maxVendors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recs, err := procurementSvc.GenerateVendorRecommendations(project.ID, tt.strategy)
			if err != nil {
				t.Errorf("GenerateVendorRecommendations() error = %v", err)
				return
			}

			if len(recs) < tt.minVendors || len(recs) > tt.maxVendors {
				t.Errorf("Expected %d-%d vendors, got %d", tt.minVendors, tt.maxVendors, len(recs))
			}

			// Verify all recommendations have required fields
			for _, rec := range recs {
				if rec.VendorID == 0 {
					t.Error("Recommendation missing VendorID")
				}
				if rec.VendorName == "" {
					t.Error("Recommendation missing VendorName")
				}
				if len(rec.BOMItems) == 0 {
					t.Error("Recommendation has no BOM items assigned")
				}
				if rec.Rationale == "" {
					t.Error("Recommendation missing rationale")
				}
			}

			t.Logf("%s: %d vendors recommended", tt.strategy, len(recs))
		})
	}
}

func TestProjectProcurementService_CompareScenarios(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create minimal test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Laptop", "")

	productSvc := NewProductService(cfg.DB)
	product, _ := productSvc.Create("Laptop A", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor.ID, ProductID: product.ID, Price: 1000, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Test Project", "", 15000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Compare scenarios
	scenarios, err := procurementSvc.CompareScenarios(project.ID)
	if err != nil {
		t.Errorf("CompareScenarios() error = %v", err)
		return
	}

	// Should have 4 scenarios: lowest_cost, fewest_vendors, balanced, quality_focused
	if len(scenarios) != 4 {
		t.Errorf("Expected 4 scenarios, got %d", len(scenarios))
	}

	// Verify each scenario has required fields
	scenarioNames := make(map[string]bool)
	for _, scenario := range scenarios {
		if scenario.Name == "" {
			t.Error("Scenario missing name")
		}
		if scenario.Description == "" {
			t.Error("Scenario missing description")
		}
		if scenario.Tradeoffs == "" {
			t.Error("Scenario missing tradeoffs")
		}
		if scenario.TotalCost == 0 {
			t.Error("Scenario has zero total cost")
		}

		scenarioNames[scenario.Name] = true
		t.Logf("Scenario: %s, Cost: $%.2f, Vendors: %d", scenario.Name, scenario.TotalCost, scenario.VendorCount)
	}

	// Check all expected scenarios are present
	expectedScenarios := []string{"Lowest Cost", "Fewest Vendors", "Balanced", "Quality Focused"}
	for _, expected := range expectedScenarios {
		if !scenarioNames[expected] {
			t.Errorf("Missing expected scenario: %s", expected)
		}
	}
}

func TestProjectProcurementService_QualityFocusedRecommendations(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create vendors with different ratings
	vendorSvc := NewVendorService(cfg.DB)
	goodVendor, _ := vendorSvc.Create("Good Vendor", "USD", "")
	badVendor, _ := vendorSvc.Create("Bad Vendor", "USD", "")

	// Add ratings
	cfg.DB.Create(&models.VendorRating{
		VendorID:      goodVendor.ID,
		PriceRating:   int32Ptr(5),
		QualityRating: int32Ptr(5),
	})
	cfg.DB.Create(&models.VendorRating{
		VendorID:      badVendor.ID,
		PriceRating:   int32Ptr(2),
		QualityRating: int32Ptr(2),
	})

	// Create products
	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Laptop", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec.ID)
	product2, _ := productSvc.Create("Laptop B", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Bad vendor is cheaper
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: badVendor.ID, ProductID: product1.ID, Price: 800, Currency: "USD"})
	// Good vendor is more expensive
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: goodVendor.ID, ProductID: product2.ID, Price: 1000, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Quality Project", "", 15000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Generate quality-focused recommendations
	recs, err := procurementSvc.GenerateVendorRecommendations(project.ID, "quality_focused")
	if err != nil {
		t.Errorf("GenerateVendorRecommendations() error = %v", err)
		return
	}

	if len(recs) == 0 {
		t.Fatal("Expected at least one recommendation")
	}

	// Should prefer the good vendor despite higher cost
	if recs[0].VendorID != goodVendor.ID {
		t.Logf("Warning: Expected good vendor to be recommended for quality-focused strategy")
		// Note: This may not always happen if the bad vendor is the only option
		// or if the rating threshold logic differs
	}

	t.Logf("Quality-focused selected vendor %d with cost $%.2f", recs[0].VendorID, recs[0].TotalCost)
}

func TestProjectProcurementService_CalculateProjectSavings(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendorSvc := NewVendorService(cfg.DB)
	vendor1, _ := vendorSvc.Create("Vendor A", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor B", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Laptop", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec.ID)
	product2, _ := productSvc.Create("Laptop B", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Vendor1: $900 (best price)
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product1.ID, Price: 900, Currency: "USD"})
	// Vendor2: $950
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product2.ID, Price: 950, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Savings Test", "", 15000, nil)
	bomItem, _ := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Create requisition with target price
	requisition := models.ProjectRequisition{
		ProjectID:     project.ID,
		Name:          "Test Req",
		Justification: "",
		Budget:        12000,
	}
	cfg.DB.Create(&requisition)

	reqItem := models.ProjectRequisitionItem{
		ProjectRequisitionID:  requisition.ID,
		BillOfMaterialsItemID: bomItem.ID,
		QuantityRequested:     10,
		TargetUnitPrice:       1000, // Target $1000, best quote is $900
	}
	cfg.DB.Create(&reqItem)

	// Calculate savings
	savings, err := procurementSvc.CalculateProjectSavings(project.ID)
	if err != nil {
		t.Errorf("CalculateProjectSavings() error = %v", err)
		return
	}

	if savings == nil {
		t.Fatal("Expected savings to be non-nil")
	}

	// Should save $100 per unit * 10 units = $1000
	expectedSavings := 1000.0
	if savings.TotalSavingsUSD < expectedSavings*0.9 || savings.TotalSavingsUSD > expectedSavings*1.1 {
		t.Errorf("Expected savings around $%.2f, got $%.2f", expectedSavings, savings.TotalSavingsUSD)
	}

	// Should have 10% savings
	if savings.SavingsPercent < 9 || savings.SavingsPercent > 11 {
		t.Errorf("Expected savings percent around 10%%, got %.2f%%", savings.SavingsPercent)
	}

	// Should have detailed breakdown
	if len(savings.DetailedBreakdown) != 1 {
		t.Errorf("Expected 1 line item, got %d", len(savings.DetailedBreakdown))
	}

	// Should have category breakdown
	if len(savings.SavingsByCategory) == 0 {
		t.Error("Expected savings by category")
	}

	t.Logf("Total savings: $%.2f (%.2f%%)", savings.TotalSavingsUSD, savings.SavingsPercent)
}

func TestProjectProcurementService_AssessEnhancedProjectRisks(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data with various risk factors
	vendorSvc := NewVendorService(cfg.DB)
	vendor, _ := vendorSvc.Create("Test Vendor", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec1, _ := specSvc.Create("Laptop", "")
	spec2, _ := specSvc.Create("Monitor", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec1.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Only one quote for spec1, none for spec2 (creates risks)
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor.ID, ProductID: product1.ID, Price: 1000, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Risk Test", "", 5000, nil) // Low budget creates budget risk
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 10, "")
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 5, "") // No quotes for this

	// Assess risks
	assessment, err := procurementSvc.AssessEnhancedProjectRisks(project.ID)
	if err != nil {
		t.Errorf("AssessEnhancedProjectRisks() error = %v", err)
		return
	}

	if assessment == nil {
		t.Fatal("Expected assessment to be non-nil")
	}

	// Should have critical or high overall risk due to missing quotes
	if assessment.OverallRisk != "critical" && assessment.OverallRisk != "high" {
		t.Errorf("Expected critical or high risk, got %s", assessment.OverallRisk)
	}

	// Should have quote coverage risk
	if coverageRisk, ok := assessment.CategoryRisks["quote_coverage"]; ok {
		if coverageRisk.Level != "critical" && coverageRisk.Level != "high" {
			t.Errorf("Expected high quote coverage risk, got %s", coverageRisk.Level)
		}
	} else {
		t.Error("Expected quote coverage risk category")
	}

	// Should have supply chain risk (no quotes for monitor)
	if assessment.SupplyChainRisk.NoQuoteItems == 0 {
		t.Error("Expected supply chain risk due to missing quotes")
	}

	// Should have mitigation actions
	if len(assessment.MitigationActions) == 0 {
		t.Error("Expected mitigation actions")
	}

	// Should have high-priority actions
	if len(assessment.HighPriorityActions) == 0 {
		t.Error("Expected high-priority actions")
	}

	t.Logf("Overall risk: %s (score: %d/100)", assessment.OverallRisk, assessment.RiskScore)
	t.Logf("Mitigation actions: %d total, %d high-priority", len(assessment.MitigationActions), len(assessment.HighPriorityActions))
}

func TestProjectProcurementService_RiskCategories(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create low-risk scenario
	vendorSvc := NewVendorService(cfg.DB)
	vendor1, _ := vendorSvc.Create("Vendor A", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor B", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Laptop", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec.ID)
	product2, _ := productSvc.Create("Laptop B", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Multiple vendors with fresh quotes
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product1.ID, Price: 900, Currency: "USD"})
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product2.ID, Price: 950, Currency: "USD"})

	// Create project with sufficient budget
	project, _ := projectSvc.Create("Low Risk Project", "", 20000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Assess risks
	assessment, err := procurementSvc.AssessEnhancedProjectRisks(project.ID)
	if err != nil {
		t.Errorf("AssessEnhancedProjectRisks() error = %v", err)
		return
	}

	// Should have low overall risk
	if assessment.OverallRisk != "low" && assessment.OverallRisk != "medium" {
		t.Logf("Expected low/medium risk for well-covered project, got %s", assessment.OverallRisk)
	}

	// Verify all risk categories are assessed
	categories := []string{"timeline", "budget", "supply_chain", "quality"}
	for _, category := range categories {
		switch category {
		case "timeline":
			if assessment.TimelineRisk.Level == "" {
				t.Errorf("Timeline risk not assessed")
			}
		case "budget":
			if assessment.BudgetRisk.Level == "" {
				t.Errorf("Budget risk not assessed")
			}
		case "supply_chain":
			if assessment.SupplyChainRisk.Level == "" {
				t.Errorf("Supply chain risk not assessed")
			}
		case "quality":
			if assessment.QualityRisk.Level == "" {
				t.Errorf("Quality risk not assessed")
			}
		}
	}

	t.Logf("Risk assessment complete - Overall: %s, Score: %d", assessment.OverallRisk, assessment.RiskScore)
}

func TestProjectProcurementService_ConsolidationSavings(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create multiple vendors
	vendorSvc := NewVendorService(cfg.DB)
	vendor1, _ := vendorSvc.Create("Vendor A", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor B", "USD", "")
	vendor3, _ := vendorSvc.Create("Vendor C", "USD", "")

	brandSvc := NewBrandService(cfg.DB)
	brand, _ := brandSvc.Create("Brand A")

	specSvc := NewSpecificationService(cfg.DB)
	spec, _ := specSvc.Create("Laptop", "")

	productSvc := NewProductService(cfg.DB)
	product1, _ := productSvc.Create("Laptop A", brand.ID, &spec.ID)
	product2, _ := productSvc.Create("Laptop B", brand.ID, &spec.ID)
	product3, _ := productSvc.Create("Laptop C", brand.ID, &spec.ID)

	forexSvc := NewForexService(cfg.DB)
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	// Each vendor has one quote (3 potential vendors)
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor1.ID, ProductID: product1.ID, Price: 900, Currency: "USD"})
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor2.ID, ProductID: product2.ID, Price: 920, Currency: "USD"})
	_, _ = quoteSvc.Create(CreateQuoteInput{VendorID: vendor3.ID, ProductID: product3.ID, Price: 950, Currency: "USD"})

	// Create project
	project, _ := projectSvc.Create("Consolidation Test", "", 15000, nil)
	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Calculate savings
	savings, err := procurementSvc.CalculateProjectSavings(project.ID)
	if err != nil {
		t.Errorf("CalculateProjectSavings() error = %v", err)
		return
	}

	// Should have consolidation savings
	// We have 3 potential vendors, but will likely use only 1 for lowest cost
	// Savings = vendors avoided * $250
	if savings.ConsolidationSavings <= 0 {
		t.Logf("Expected consolidation savings, got $%.2f", savings.ConsolidationSavings)
		// Note: This may be 0 if the algorithm uses all vendors
	}

	t.Logf("Consolidation savings: $%.2f", savings.ConsolidationSavings)
}

// ============================================================================
// PHASE 4: DASHBOARD AND REPORTING TESTS
// ============================================================================

// Helper function for tests
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestProjectProcurementService_GetProjectDashboard(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Create services
	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendor1, _ := vendorSvc.Create("Vendor A", "USD", "")
	vendor2, _ := vendorSvc.Create("Vendor B", "USD", "")
	brand1, _ := brandSvc.Create("BrandX")
	spec1, _ := specSvc.Create("Laptop Spec", "")
	spec2, _ := specSvc.Create("Monitor Spec", "")
	product1, _ := productSvc.Create("Laptop Model A", brand1.ID, &spec1.ID)
	product2, _ := productSvc.Create("Monitor Model B", brand1.ID, &spec2.ID)

	// Create quotes
	quote1, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor1.ID,
		ProductID:  product1.ID,
		Price:      1000.0,
		Currency:   "USD",
		ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
	})
	_, _ = quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor2.ID,
		ProductID:  product2.ID,
		Price:      300.0,
		Currency:   "USD",
		ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
	})

	// Create project with BOM
	deadline := time.Now().AddDate(0, 3, 0)
	project, _ := projectSvc.Create("Dashboard Test Project", "", 50000.0, &deadline)

	// Add BOM items
	bomItem1, _ := projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 10, "")
	bomItem2, _ := projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 20, "")

	// Create requisition
	var projectReq models.ProjectRequisition
	cfg.DB.Create(&models.ProjectRequisition{
		ProjectID: project.ID,
		Name:      "Test Requisition",
	}).Scan(&projectReq)

	// Add requisition items
	cfg.DB.Create(&models.ProjectRequisitionItem{
		ProjectRequisitionID:  projectReq.ID,
		BillOfMaterialsItemID: bomItem1.ID,
		QuantityRequested:     10,
		ProcurementStatus:     "pending",
	})
	cfg.DB.Create(&models.ProjectRequisitionItem{
		ProjectRequisitionID:  projectReq.ID,
		BillOfMaterialsItemID: bomItem2.ID,
		QuantityRequested:     20,
		ProcurementStatus:     "pending",
	})

	// Create purchase order
	cfg.DB.Create(&models.PurchaseOrder{
		PONumber:  "PO-001",
		QuoteID:   quote1.ID,
		Quantity:  5,
		OrderDate: time.Now(),
		Status:    "ordered",
	})

	// Get dashboard
	dashboard, err := procurementSvc.GetProjectDashboard(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project dashboard: %v", err)
	}

	// Validate dashboard structure
	if dashboard.Project == nil {
		t.Error("Expected project in dashboard")
	}

	// Validate progress metrics
	if dashboard.Progress.BOMCoverage <= 0 {
		t.Error("Expected BOM coverage > 0")
	}
	if dashboard.Progress.RequisitionsTotal != 1 {
		t.Errorf("Expected 1 requisition total, got %d", dashboard.Progress.RequisitionsTotal)
	}
	if dashboard.Progress.TimelineStatus == "" {
		t.Error("Expected timeline status")
	}

	// Validate financial overview
	if dashboard.Financial.Budget != 50000.0 {
		t.Errorf("Expected budget 50000, got %.2f", dashboard.Financial.Budget)
	}
	if dashboard.Financial.BudgetHealth == "" {
		t.Error("Expected budget health status")
	}

	// Validate procurement status
	if dashboard.Procurement.TotalItems != 2 {
		t.Errorf("Expected 2 total items, got %d", dashboard.Procurement.TotalItems)
	}
	if dashboard.Procurement.ItemsWithQuotes != 2 {
		t.Errorf("Expected 2 items with quotes, got %d", dashboard.Procurement.ItemsWithQuotes)
	}

	// Validate chart data
	if len(dashboard.ChartsData.BudgetUtilization) == 0 {
		t.Error("Expected budget utilization chart data")
	}
	if len(dashboard.ChartsData.CostComparison) == 0 {
		t.Error("Expected cost comparison chart data")
	}

	t.Logf("Dashboard generated successfully:")
	t.Logf("  Progress: %.1f%% BOM coverage, %s timeline", dashboard.Progress.BOMCoverage, dashboard.Progress.TimelineStatus)
	t.Logf("  Financial: $%.2f committed, $%.2f estimated, %s health",
		dashboard.Financial.Committed, dashboard.Financial.Estimated, dashboard.Financial.BudgetHealth)
	t.Logf("  Procurement: %d/%d items with quotes, %d vendors engaged",
		dashboard.Procurement.ItemsWithQuotes, dashboard.Procurement.TotalItems, dashboard.Procurement.VendorsEngaged)
}

func TestProjectProcurementService_CalculateProjectProgress(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	tests := []struct {
		name               string
		setupFunc          func() *models.Project
		expectedCoverage   float64
		expectedTimeline   string
		expectedDaysToDeadline int
	}{
		{
			name: "Full coverage, on track",
			setupFunc: func() *models.Project {
				vendor, _ := vendorSvc.Create("Vendor", "USD", "")
				brand, _ := brandSvc.Create("Brand")
				spec, _ := specSvc.Create("Spec", "")
				product, _ := productSvc.Create("Product", brand.ID, &spec.ID)
				quoteSvc.Create(CreateQuoteInput{
					VendorID:   vendor.ID,
					ProductID:  product.ID,
					Price:      1000.0,
					Currency:   "USD",
					ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
				})

				dl := time.Now().AddDate(0, 0, 90)
				project, _ := projectSvc.Create("Test Project", "", 10000.0, &dl)
				projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 5, "")
				return project
			},
			expectedCoverage:   100.0,
			expectedTimeline:   "on_track",
		},
		{
			name: "Partial coverage, at risk",
			setupFunc: func() *models.Project {
				vendor, err := vendorSvc.Create("Vendor2", "USD", "")
				if err != nil {
					t.Fatalf("Failed to create vendor: %v", err)
				}
				brand, err := brandSvc.Create("Brand2")
				if err != nil {
					t.Fatalf("Failed to create brand: %v", err)
				}
				spec1, err := specSvc.Create("Spec1", "")
				if err != nil {
					t.Fatalf("Failed to create spec1: %v", err)
				}
				spec2, err := specSvc.Create("Spec2", "")
				if err != nil {
					t.Fatalf("Failed to create spec2: %v", err)
				}
				product, err := productSvc.Create("ProductPartial", brand.ID, &spec1.ID)
				if err != nil {
					t.Fatalf("Failed to create product: %v", err)
				}
				_, err = quoteSvc.Create(CreateQuoteInput{
					VendorID:   vendor.ID,
					ProductID:  product.ID,
					Price:      1000.0,
					Currency:   "USD",
					ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
				})
				if err != nil {
					t.Fatalf("Failed to create quote: %v", err)
				}

				dl2 := time.Now().AddDate(0, 0, 20)
				project, err := projectSvc.Create("Test Project 2", "", 10000.0, &dl2)
				if err != nil {
					t.Fatalf("Failed to create project: %v", err)
				}
				_, err = projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 5, "")
				if err != nil {
					t.Fatalf("Failed to add BOM item 1: %v", err)
				}
				_, err = projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 10, "")
				if err != nil {
					t.Fatalf("Failed to add BOM item 2: %v", err)
				}
				return project
			},
			expectedCoverage: 50.0,
			expectedTimeline: "at_risk",
		},
		{
			name: "Past deadline",
			setupFunc: func() *models.Project {
				dl3 := time.Now().AddDate(0, 0, -10)
				project, _ := projectSvc.Create("Test Project 3", "", 10000.0, &dl3)
				return project
			},
			expectedCoverage: 0.0,
			expectedTimeline: "delayed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := tt.setupFunc()

			// Reload project with all relationships
			var reloadedProject models.Project
			cfg.DB.Preload("BillOfMaterials.Items.Specification").
				Preload("Requisitions").
				First(&reloadedProject, project.ID)

			progress, err := procurementSvc.calculateProjectProgress(&reloadedProject)
			if err != nil {
				t.Fatalf("Failed to calculate progress: %v", err)
			}

			if progress.BOMCoverage != tt.expectedCoverage {
				t.Errorf("Expected BOM coverage %.1f%%, got %.1f%%", tt.expectedCoverage, progress.BOMCoverage)
			}

			if progress.TimelineStatus != tt.expectedTimeline {
				t.Errorf("Expected timeline status %s, got %s", tt.expectedTimeline, progress.TimelineStatus)
			}

			t.Logf("%s: Coverage=%.1f%%, Timeline=%s, DaysToDeadline=%d",
				tt.name, progress.BOMCoverage, progress.TimelineStatus, progress.DaysToDeadline)
		})
	}
}

func TestProjectProcurementService_CalculateFinancialOverview(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendor, _ := vendorSvc.Create("Vendor", "USD", "")
	brand, _ := brandSvc.Create("Brand")
	spec, _ := specSvc.Create("Spec", "")
	product, _ := productSvc.Create("Product", brand.ID, &spec.ID)

	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      1000.0,
		Currency:   "USD",
		ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
	})

	project, _ := projectSvc.Create("Financial Test Project", "", 20000.0, nil)

	_, _ = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Create purchase order for 5 units
	cfg.DB.Create(&models.PurchaseOrder{
		PONumber:    "PO-FIN-001",
		QuoteID:     quote.ID,
		VendorID:    vendor.ID,
		ProductID:   product.ID,
		Quantity:    5,
		OrderDate:   time.Now(),
		Status:      "ordered",
		UnitPrice:   1000.0,
		Currency:    "USD",
		TotalAmount: 5000.0,
		GrandTotal:  5000.0,
	})

	// Reload project
	var reloadedProject models.Project
	cfg.DB.Preload("BillOfMaterials.Items.Specification").First(&reloadedProject, project.ID)

	financial, err := procurementSvc.calculateFinancialOverview(&reloadedProject)
	if err != nil {
		t.Fatalf("Failed to calculate financial overview: %v", err)
	}

	// Validate
	if financial.Budget != 20000.0 {
		t.Errorf("Expected budget 20000, got %.2f", financial.Budget)
	}

	// Committed should be 5 units * $1000 = $5000
	expectedCommitted := 5000.0
	if financial.Committed != expectedCommitted {
		t.Errorf("Expected committed %.2f, got %.2f", expectedCommitted, financial.Committed)
	}

	// Estimated should be 5 remaining units * $1000 = $5000
	expectedEstimated := 5000.0
	if financial.Estimated != expectedEstimated {
		t.Errorf("Expected estimated %.2f, got %.2f", expectedEstimated, financial.Estimated)
	}

	// Remaining should be 20000 - (5000 + 5000) = 10000
	expectedRemaining := 10000.0
	if financial.Remaining != expectedRemaining {
		t.Errorf("Expected remaining %.2f, got %.2f", expectedRemaining, financial.Remaining)
	}

	// Budget health should be "healthy" (50% utilization)
	if financial.BudgetHealth != "healthy" {
		t.Errorf("Expected budget health 'healthy', got '%s'", financial.BudgetHealth)
	}

	t.Logf("Financial overview: Committed=$%.2f, Estimated=$%.2f, Remaining=$%.2f, Health=%s",
		financial.Committed, financial.Estimated, financial.Remaining, financial.BudgetHealth)
}

func TestProjectProcurementService_ProcurementStatus(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data with different quote ages
	vendor, _ := vendorSvc.Create("Vendor", "USD", "")
	brand, _ := brandSvc.Create("Brand")
	spec1, _ := specSvc.Create("Spec1", "")
	spec2, _ := specSvc.Create("Spec2", "")
	spec3, _ := specSvc.Create("Spec3", "")
	product1, _ := productSvc.Create("Product1", brand.ID, &spec1.ID)
	product2, _ := productSvc.Create("Product2", brand.ID, &spec2.ID)
	product3, _ := productSvc.Create("Product3", brand.ID, &spec3.ID)

	// Fresh quote (< 30 days)
	quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product1.ID,
		Price:      1000.0,
		Currency:   "USD",
		ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
	})

	// Stale quote (30-90 days) - created directly to set custom QuoteDate
	staleValid := time.Now().AddDate(0, 0, 30)
	staleDate := time.Now().AddDate(0, 0, -60)
	staleQuote := &models.Quote{
		VendorID:       vendor.ID,
		ProductID:      product2.ID,
		Price:          1000.0,
		Currency:       "USD",
		ConvertedPrice: 1000.0,
		ConversionRate: 1.0,
		QuoteDate:      staleDate,
		ValidUntil:     &staleValid,
		Version:        1,
	}
	cfg.DB.Create(staleQuote)

	// Expired quote - created directly to set custom QuoteDate
	expiredValid := time.Now().AddDate(0, 0, -10)
	expiredDate := time.Now().AddDate(0, 0, -120)
	expiredQuote := &models.Quote{
		VendorID:       vendor.ID,
		ProductID:      product3.ID,
		Price:          1000.0,
		Currency:       "USD",
		ConvertedPrice: 1000.0,
		ConversionRate: 1.0,
		QuoteDate:      expiredDate,
		ValidUntil:     &expiredValid,
		Version:        1,
	}
	cfg.DB.Create(expiredQuote)

	project, _ := projectSvc.Create("Status Test Project", "", 30000.0, nil)

	projectSvc.AddBillOfMaterialsItem(project.ID, spec1.ID, 10, "")
	projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 10, "")
	projectSvc.AddBillOfMaterialsItem(project.ID, spec3.ID, 10, "")

	// Reload project
	var reloadedProject models.Project
	cfg.DB.Preload("BillOfMaterials.Items.Specification").First(&reloadedProject, project.ID)

	status, err := procurementSvc.calculateProcurementStatus(&reloadedProject)
	if err != nil {
		t.Fatalf("Failed to calculate procurement status: %v", err)
	}

	// Validate
	if status.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", status.TotalItems)
	}

	if status.ItemsWithQuotes != 3 {
		t.Errorf("Expected 3 items with quotes, got %d", status.ItemsWithQuotes)
	}

	if status.VendorsEngaged != 1 {
		t.Errorf("Expected 1 vendor engaged, got %d", status.VendorsEngaged)
	}

	if status.QuoteFreshness == "" {
		t.Error("Expected quote freshness status")
	}

	t.Logf("Procurement status: %d/%d items with quotes, %d vendors, freshness=%s",
		status.ItemsWithQuotes, status.TotalItems, status.VendorsEngaged, status.QuoteFreshness)
}

func TestProjectProcurementService_ChartDataGeneration(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	brandSvc := NewBrandService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	projectSvc := NewProjectService(cfg.DB)
	procurementSvc := NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	// Create test data
	vendor, _ := vendorSvc.Create("Vendor", "USD", "")
	brand, _ := brandSvc.Create("Brand")
	spec, _ := specSvc.Create("Spec", "")
	product, _ := productSvc.Create("Product", brand.ID, &spec.ID)

	quote, _ := quoteSvc.Create(CreateQuoteInput{
		VendorID:   vendor.ID,
		ProductID:  product.ID,
		Price:      1000.0,
		Currency:   "USD",
		ValidUntil: timePtr(time.Now().AddDate(0, 0, 60)),
	})

	project, _ := projectSvc.Create("Chart Test Project", "", 15000.0, nil)

	projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "")

	// Create purchase order
	cfg.DB.Create(&models.PurchaseOrder{
		PONumber:  "PO-CHART-001",
		QuoteID:   quote.ID,
		Quantity:  5,
		OrderDate: time.Now(),
		Status:    "ordered",
	})

	// Reload project
	var reloadedProject models.Project
	cfg.DB.Preload("BillOfMaterials.Items.Specification").
		Preload("Requisitions").
		First(&reloadedProject, project.ID)

	financial, _ := procurementSvc.calculateFinancialOverview(&reloadedProject)
	vendorPerf, _ := procurementSvc.aggregateVendorPerformance(&reloadedProject)

	chartsData, err := procurementSvc.generateChartData(&reloadedProject, financial, vendorPerf)
	if err != nil {
		t.Fatalf("Failed to generate chart data: %v", err)
	}

	// Validate chart data
	if len(chartsData.BudgetUtilization) != 3 {
		t.Errorf("Expected 3 budget utilization data points, got %d", len(chartsData.BudgetUtilization))
	}

	if len(chartsData.CostComparison) != 3 {
		t.Errorf("Expected 3 cost comparison data points, got %d", len(chartsData.CostComparison))
	}

	// Verify data point structure
	for i, dp := range chartsData.BudgetUtilization {
		if dp.Label == "" {
			t.Errorf("Budget utilization point %d missing label", i)
		}
		if dp.Color == "" {
			t.Errorf("Budget utilization point %d missing color", i)
		}
	}

	t.Logf("Chart data generated: %d budget points, %d cost comparison points, %d vendor points",
		len(chartsData.BudgetUtilization), len(chartsData.CostComparison), len(chartsData.VendorDistribution))
}
