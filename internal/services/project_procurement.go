package services

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// ProjectProcurementService handles project-level procurement analysis and optimization
type ProjectProcurementService struct {
	db             *gorm.DB
	quoteService   *QuoteService
	projectService *ProjectService
}

// NewProjectProcurementService creates a new project procurement service
func NewProjectProcurementService(db *gorm.DB, quoteService *QuoteService, projectService *ProjectService) *ProjectProcurementService {
	return &ProjectProcurementService{
		db:             db,
		quoteService:   quoteService,
		projectService: projectService,
	}
}

// BOMItemProcurementAnalysis holds analysis for a single BOM item across all requisitions
type BOMItemProcurementAnalysis struct {
	BOMItem              *models.BillOfMaterialsItem
	Specification        *models.Specification
	TotalQuantityNeeded  int
	TotalQuantityPlanned int
	CoveragePercent      float64
	RequisitionItems     []ProjectRequisitionItemQuotes
	AvailableQuotes      []models.Quote
	BestQuote            *models.Quote
	RecommendedQuote     *models.Quote
	BestTotalCost        float64
	RecommendedTotalCost float64
	TargetTotalCost      float64
	SavingsVsTarget      float64
	HasSufficientQuotes  bool
	HasGaps              bool
	RiskLevel            string
}

// ProjectRequisitionItemQuotes holds quote information for a requisition item
type ProjectRequisitionItemQuotes struct {
	RequisitionItem *models.ProjectRequisitionItem
	RequisitionName string
	Quantity        int
	TargetUnitPrice float64
	BestQuote       *models.Quote
	SelectedQuote   *models.Quote
	Status          string
}

// VendorConsolidationAnalysis analyzes vendor usage and consolidation opportunities
type VendorConsolidationAnalysis struct {
	VendorID            uint
	VendorName          string
	Rating              *VendorRatingsSummary
	BOMItemsAvailable   []uint
	SpecificationsCount int
	TotalQuantity       int
	TotalCostIfUsed     float64
	AveragePriceRank    float64
	EstimatedOrderCount int
	ShippingAdvantage   bool
}

// VendorRatingsSummary holds aggregated vendor ratings
type VendorRatingsSummary struct {
	VendorID     uint
	TotalRatings int
	AvgPrice     *float64
	AvgQuality   *float64
	AvgDelivery  *float64
	AvgService   *float64
	OverallAvg   float64
}

// ProjectProcurementComparison holds complete project-level procurement analysis
type ProjectProcurementComparison struct {
	Project               *models.Project
	Strategy              *models.ProjectProcurementStrategy
	BOMItemAnalyses       []BOMItemProcurementAnalysis
	TotalBOMItems         int
	FullyCoveredItems     int
	PartiallyCoveredItems int
	UncoveredItems        int
	VendorConsolidation   []VendorConsolidationAnalysis
	TotalVendorsNeeded    int
	ProjectBudget         float64
	TotalTargetCost       float64
	BestCaseCost          float64
	RecommendedCost       float64
	WorstCaseCost         float64
	SavingsVsBudget       float64
	SavingsVsTarget       float64
	SavingsPercent        float64
	VendorRecommendations []VendorRecommendation
	RiskAssessment        ProjectRiskAssessment
	AlternativeScenarios  []ProcurementScenario
	AnalysisDate          time.Time
	QuoteFreshness        QuoteFreshnessStats
}

// VendorRecommendation holds vendor assignment recommendation
type VendorRecommendation struct {
	VendorID   uint
	VendorName string
	BOMItems   []uint
	TotalCost  float64
	ItemCount  int
	Rationale  string
	Priority   int
}

// ProjectRiskAssessment holds risk analysis for the project
type ProjectRiskAssessment struct {
	OverallRisk       string
	RiskFactors       []RiskFactor
	MitigationActions []string
}

// RiskFactor represents a specific risk item
type RiskFactor struct {
	Category         string
	Severity         string
	Description      string
	AffectedBOMItems []uint
	Impact           string
}

// ProcurementScenario represents a what-if procurement scenario
type ProcurementScenario struct {
	Name              string
	Description       string
	VendorCount       int
	TotalCost         float64
	SavingsVsBudget   float64
	Tradeoffs         string
	VendorAssignments map[uint][]uint
}

// QuoteFreshnessStats tracks quote age and freshness
type QuoteFreshnessStats struct {
	TotalQuotes    int
	FreshQuotes    int
	StaleQuotes    int
	ExpiredQuotes  int
	AverageAgeDays int
}

// GetProjectProcurementComparison generates comprehensive procurement analysis for a project
func (s *ProjectProcurementService) GetProjectProcurementComparison(projectID uint) (*ProjectProcurementComparison, error) {
	// Load project with all relationships
	var project models.Project
	err := s.db.Preload("BillOfMaterials.Items.Specification").
		Preload("Requisitions.Items.BOMItem.Specification").
		Preload("Requisitions.Items.SelectedQuote.Vendor").
		Preload("Requisitions.Items.SelectedQuote.Product").
		First(&project, projectID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Project", ID: projectID}
		}
		return nil, err
	}

	// Get or create default strategy
	strategy, err := s.GetOrCreateStrategy(projectID)
	if err != nil {
		return nil, err
	}

	comparison := &ProjectProcurementComparison{
		Project:         &project,
		Strategy:        strategy,
		BOMItemAnalyses: make([]BOMItemProcurementAnalysis, 0),
		ProjectBudget:   project.Budget,
		AnalysisDate:    time.Now(),
	}

	// Analyze each BOM item
	if project.BillOfMaterials != nil {
		comparison.TotalBOMItems = len(project.BillOfMaterials.Items)

		for _, bomItem := range project.BillOfMaterials.Items {
			analysis, err := s.analyzeBOMItemInternal(&bomItem, project.Requisitions)
			if err != nil {
				return nil, fmt.Errorf("error analyzing BOM item %d: %w", bomItem.ID, err)
			}

			comparison.BOMItemAnalyses = append(comparison.BOMItemAnalyses, *analysis)

			// Update coverage counters
			if analysis.CoveragePercent >= 100.0 && analysis.HasSufficientQuotes {
				comparison.FullyCoveredItems++
			} else if analysis.HasSufficientQuotes {
				comparison.PartiallyCoveredItems++
			} else {
				comparison.UncoveredItems++
			}

			// Accumulate costs
			comparison.TotalTargetCost += analysis.TargetTotalCost
			comparison.BestCaseCost += analysis.BestTotalCost
			comparison.RecommendedCost += analysis.RecommendedTotalCost
		}
	}

	// Calculate savings
	if comparison.ProjectBudget > 0 {
		comparison.SavingsVsBudget = comparison.ProjectBudget - comparison.RecommendedCost
		comparison.SavingsPercent = (comparison.SavingsVsBudget / comparison.ProjectBudget) * 100
	}
	if comparison.TotalTargetCost > 0 {
		comparison.SavingsVsTarget = comparison.TotalTargetCost - comparison.RecommendedCost
	}

	// Assess risks
	comparison.RiskAssessment = s.assessProjectRisks(&project, comparison.BOMItemAnalyses)

	// Calculate quote freshness
	comparison.QuoteFreshness = s.calculateQuoteFreshness(comparison.BOMItemAnalyses)

	return comparison, nil
}

// AnalyzeBOMItem performs detailed analysis for a specific BOM item
func (s *ProjectProcurementService) AnalyzeBOMItem(bomItemID uint) (*BOMItemProcurementAnalysis, error) {
	// Load BOM item with specification and project requisitions
	var bomItem models.BillOfMaterialsItem
	err := s.db.Preload("Specification").
		Preload("BillOfMaterials.Project.Requisitions.Items.BOMItem").
		Preload("BillOfMaterials.Project.Requisitions.Items.SelectedQuote.Vendor").
		Preload("BillOfMaterials.Project.Requisitions.Items.SelectedQuote.Product").
		First(&bomItem, bomItemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "BillOfMaterialsItem", ID: bomItemID}
		}
		return nil, err
	}

	// Get project requisitions
	var requisitions []models.ProjectRequisition
	if bomItem.BillOfMaterials != nil && bomItem.BillOfMaterials.Project != nil {
		requisitions = bomItem.BillOfMaterials.Project.Requisitions
	}

	return s.analyzeBOMItemInternal(&bomItem, requisitions)
}

// analyzeBOMItemInternal is the internal implementation for BOM item analysis
func (s *ProjectProcurementService) analyzeBOMItemInternal(
	bomItem *models.BillOfMaterialsItem,
	requisitions []models.ProjectRequisition,
) (*BOMItemProcurementAnalysis, error) {
	analysis := &BOMItemProcurementAnalysis{
		BOMItem:             bomItem,
		Specification:       bomItem.Specification,
		TotalQuantityNeeded: bomItem.Quantity,
		RequisitionItems:    make([]ProjectRequisitionItemQuotes, 0),
	}

	// Find all requisition items that reference this BOM item
	for _, req := range requisitions {
		for _, item := range req.Items {
			if item.BillOfMaterialsItemID == bomItem.ID {
				reqItemQuote := ProjectRequisitionItemQuotes{
					RequisitionItem: &item,
					RequisitionName: req.Name,
					Quantity:        item.QuantityRequested,
					TargetUnitPrice: item.TargetUnitPrice,
					SelectedQuote:   item.SelectedQuote,
					Status:          item.ProcurementStatus,
				}

				analysis.RequisitionItems = append(analysis.RequisitionItems, reqItemQuote)
				analysis.TotalQuantityPlanned += item.QuantityRequested

				// Accumulate target costs
				if item.TargetUnitPrice > 0 {
					analysis.TargetTotalCost += item.TargetUnitPrice * float64(item.QuantityRequested)
				}
			}
		}
	}

	// Calculate coverage
	if analysis.TotalQuantityNeeded > 0 {
		analysis.CoveragePercent = (float64(analysis.TotalQuantityPlanned) / float64(analysis.TotalQuantityNeeded)) * 100
	}
	analysis.HasGaps = analysis.TotalQuantityPlanned < analysis.TotalQuantityNeeded

	// Get available quotes for this specification
	if bomItem.Specification != nil {
		quotes, err := s.quoteService.CompareQuotesForSpecification(bomItem.SpecificationID)
		if err != nil {
			return nil, err
		}
		analysis.AvailableQuotes = quotes
		analysis.HasSufficientQuotes = len(quotes) > 0

		// Find best quote (lowest price)
		if len(quotes) > 0 {
			analysis.BestQuote = &quotes[0]
			analysis.RecommendedQuote = &quotes[0] // Default to best quote
			analysis.BestTotalCost = quotes[0].ConvertedPrice * float64(analysis.TotalQuantityNeeded)
			analysis.RecommendedTotalCost = analysis.BestTotalCost

			// Update requisition items with best quotes
			for i := range analysis.RequisitionItems {
				analysis.RequisitionItems[i].BestQuote = &quotes[0]
			}
		}
	}

	// Calculate savings vs target
	if analysis.TargetTotalCost > 0 && analysis.RecommendedTotalCost > 0 {
		analysis.SavingsVsTarget = analysis.TargetTotalCost - analysis.RecommendedTotalCost
	}

	// Assess risk level
	analysis.RiskLevel = s.assessBOMItemRisk(analysis)

	return analysis, nil
}

// assessBOMItemRisk determines risk level for a BOM item
func (s *ProjectProcurementService) assessBOMItemRisk(analysis *BOMItemProcurementAnalysis) string {
	riskScore := 0

	// No quotes available
	if !analysis.HasSufficientQuotes {
		riskScore += 3
	} else if len(analysis.AvailableQuotes) == 1 {
		// Single source risk
		riskScore += 2
	}

	// Coverage gaps
	if analysis.HasGaps {
		riskScore += 2
	} else if analysis.CoveragePercent < 50 {
		riskScore += 1
	}

	// Quote freshness
	if len(analysis.AvailableQuotes) > 0 {
		staleCount := 0
		for _, quote := range analysis.AvailableQuotes {
			if quote.IsStale() {
				staleCount++
			}
		}
		if staleCount == len(analysis.AvailableQuotes) {
			riskScore += 2
		} else if staleCount > len(analysis.AvailableQuotes)/2 {
			riskScore += 1
		}
	}

	// Map score to risk level
	if riskScore >= 5 {
		return "critical"
	} else if riskScore >= 3 {
		return "high"
	} else if riskScore >= 1 {
		return "medium"
	}
	return "low"
}

// assessProjectRisks generates project-level risk assessment
func (s *ProjectProcurementService) assessProjectRisks(
	project *models.Project,
	bomAnalyses []BOMItemProcurementAnalysis,
) ProjectRiskAssessment {
	assessment := ProjectRiskAssessment{
		RiskFactors:       make([]RiskFactor, 0),
		MitigationActions: make([]string, 0),
	}

	// Analyze quote coverage risks
	uncoveredItems := make([]uint, 0)
	singleSourceItems := make([]uint, 0)

	for _, analysis := range bomAnalyses {
		if !analysis.HasSufficientQuotes {
			uncoveredItems = append(uncoveredItems, analysis.BOMItem.ID)
		} else if len(analysis.AvailableQuotes) == 1 {
			singleSourceItems = append(singleSourceItems, analysis.BOMItem.ID)
		}
	}

	if len(uncoveredItems) > 0 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Category:         "quote_coverage",
			Severity:         "critical",
			Description:      fmt.Sprintf("%d BOM items have no available quotes", len(uncoveredItems)),
			AffectedBOMItems: uncoveredItems,
			Impact:           "Cannot proceed with procurement for these items",
		})
		assessment.MitigationActions = append(assessment.MitigationActions, "Request quotes from vendors for uncovered items")
	}

	if len(singleSourceItems) > 0 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Category:         "vendor_capacity",
			Severity:         "medium",
			Description:      fmt.Sprintf("%d BOM items have only one vendor option", len(singleSourceItems)),
			AffectedBOMItems: singleSourceItems,
			Impact:           "Limited negotiation leverage and supply chain risk",
		})
		assessment.MitigationActions = append(assessment.MitigationActions, "Identify additional vendor sources for single-source items")
	}

	// Determine overall risk
	criticalCount := 0
	highCount := 0
	for _, factor := range assessment.RiskFactors {
		if factor.Severity == "critical" {
			criticalCount++
		} else if factor.Severity == "high" {
			highCount++
		}
	}

	if criticalCount > 0 {
		assessment.OverallRisk = "critical"
	} else if highCount > 0 {
		assessment.OverallRisk = "high"
	} else if len(assessment.RiskFactors) > 0 {
		assessment.OverallRisk = "medium"
	} else {
		assessment.OverallRisk = "low"
	}

	return assessment
}

// calculateQuoteFreshness analyzes quote age and freshness
func (s *ProjectProcurementService) calculateQuoteFreshness(bomAnalyses []BOMItemProcurementAnalysis) QuoteFreshnessStats {
	stats := QuoteFreshnessStats{}
	totalAgeDays := 0

	seenQuotes := make(map[uint]bool)

	for _, analysis := range bomAnalyses {
		for _, quote := range analysis.AvailableQuotes {
			// Avoid counting the same quote multiple times
			if seenQuotes[quote.ID] {
				continue
			}
			seenQuotes[quote.ID] = true

			stats.TotalQuotes++
			ageDays := int(time.Since(quote.QuoteDate).Hours() / 24)
			totalAgeDays += ageDays

			if quote.IsExpired() {
				stats.ExpiredQuotes++
			} else if quote.IsStale() {
				stats.StaleQuotes++
			} else {
				stats.FreshQuotes++
			}
		}
	}

	if stats.TotalQuotes > 0 {
		stats.AverageAgeDays = totalAgeDays / stats.TotalQuotes
	}

	return stats
}

// getOrCreateStrategy gets existing strategy or creates default one
// GetOrCreateStrategy retrieves or creates a procurement strategy for a project
func (s *ProjectProcurementService) GetOrCreateStrategy(projectID uint) (*models.ProjectProcurementStrategy, error) {
	var strategy models.ProjectProcurementStrategy
	err := s.db.Where("project_id = ?", projectID).First(&strategy).Error

	if err == nil {
		return &strategy, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create default strategy
	strategy = models.ProjectProcurementStrategy{
		ProjectID:           projectID,
		Strategy:            "lowest_cost",
		AllowPartialFulfill: true,
	}

	if err := s.db.Create(&strategy).Error; err != nil {
		return nil, err
	}

	return &strategy, nil
}

// GetVendorRatingSummary calculates aggregated ratings for a vendor
func (s *ProjectProcurementService) GetVendorRatingSummary(vendorID uint) (*VendorRatingsSummary, error) {
	var ratings []models.VendorRating
	err := s.db.Where("vendor_id = ?", vendorID).Find(&ratings).Error
	if err != nil {
		return nil, err
	}

	summary := &VendorRatingsSummary{
		VendorID:     vendorID,
		TotalRatings: len(ratings),
	}

	if len(ratings) == 0 {
		return summary, nil
	}

	var priceSum, qualitySum, deliverySum, serviceSum float64
	var priceCount, qualityCount, deliveryCount, serviceCount int

	for _, rating := range ratings {
		if rating.PriceRating != nil {
			priceSum += float64(*rating.PriceRating)
			priceCount++
		}
		if rating.QualityRating != nil {
			qualitySum += float64(*rating.QualityRating)
			qualityCount++
		}
		if rating.DeliveryRating != nil {
			deliverySum += float64(*rating.DeliveryRating)
			deliveryCount++
		}
		if rating.ServiceRating != nil {
			serviceSum += float64(*rating.ServiceRating)
			serviceCount++
		}
	}

	if priceCount > 0 {
		avg := priceSum / float64(priceCount)
		summary.AvgPrice = &avg
	}
	if qualityCount > 0 {
		avg := qualitySum / float64(qualityCount)
		summary.AvgQuality = &avg
	}
	if deliveryCount > 0 {
		avg := deliverySum / float64(deliveryCount)
		summary.AvgDelivery = &avg
	}
	if serviceCount > 0 {
		avg := serviceSum / float64(serviceCount)
		summary.AvgService = &avg
	}

	// Calculate overall average
	totalSum := 0.0
	totalCount := 0
	if summary.AvgPrice != nil {
		totalSum += *summary.AvgPrice
		totalCount++
	}
	if summary.AvgQuality != nil {
		totalSum += *summary.AvgQuality
		totalCount++
	}
	if summary.AvgDelivery != nil {
		totalSum += *summary.AvgDelivery
		totalCount++
	}
	if summary.AvgService != nil {
		totalSum += *summary.AvgService
		totalCount++
	}

	if totalCount > 0 {
		summary.OverallAvg = totalSum / float64(totalCount)
	}

	return summary, nil
}

// GetVendorConsolidationAnalysis analyzes opportunities to consolidate vendors
func (s *ProjectProcurementService) GetVendorConsolidationAnalysis(projectID uint) ([]VendorConsolidationAnalysis, error) {
	// Get project with BOM
	var project models.Project
	err := s.db.Preload("BillOfMaterials.Items.Specification").First(&project, projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Project", ID: projectID}
		}
		return nil, err
	}

	if project.BillOfMaterials == nil || len(project.BillOfMaterials.Items) == 0 {
		return []VendorConsolidationAnalysis{}, nil
	}

	// Collect all specification IDs from BOM
	specIDs := make([]uint, 0)
	specQuantities := make(map[uint]int)
	for _, bomItem := range project.BillOfMaterials.Items {
		specIDs = append(specIDs, bomItem.SpecificationID)
		specQuantities[bomItem.SpecificationID] = bomItem.Quantity
	}

	// Get all quotes for these specifications
	var quotes []models.Quote
	err = s.db.Preload("Vendor").Preload("Product.Specification").
		Joins("JOIN products ON products.id = quotes.product_id").
		Where("products.specification_id IN ?", specIDs).
		Where("quotes.valid_until IS NULL OR quotes.valid_until > ?", time.Now()).
		Find(&quotes).Error
	if err != nil {
		return nil, err
	}

	// Build vendor capability map
	vendorCapabilities := make(map[uint]map[uint]float64) // vendorID -> specID -> best price
	vendorInfo := make(map[uint]*models.Vendor)

	for _, quote := range quotes {
		if quote.Product == nil || quote.Product.Specification == nil {
			continue
		}

		vendorID := quote.VendorID
		specID := quote.Product.Specification.ID

		if vendorCapabilities[vendorID] == nil {
			vendorCapabilities[vendorID] = make(map[uint]float64)
			vendorInfo[vendorID] = quote.Vendor
		}

		// Track best (lowest) price per vendor per specification
		if existingPrice, exists := vendorCapabilities[vendorID][specID]; !exists || quote.ConvertedPrice < existingPrice {
			vendorCapabilities[vendorID][specID] = quote.ConvertedPrice
		}
	}

	// Build consolidation analyses
	analyses := make([]VendorConsolidationAnalysis, 0)

	for vendorID, capabilities := range vendorCapabilities {
		analysis := VendorConsolidationAnalysis{
			VendorID:            vendorID,
			VendorName:          vendorInfo[vendorID].Name,
			BOMItemsAvailable:   make([]uint, 0),
			SpecificationsCount: len(capabilities),
			TotalQuantity:       0,
			TotalCostIfUsed:     0,
		}

		// Calculate total cost if using this vendor for all items they can supply
		priceRanks := make([]float64, 0)
		for specID, price := range capabilities {
			quantity := specQuantities[specID]
			analysis.TotalQuantity += quantity
			analysis.TotalCostIfUsed += price * float64(quantity)

			// Calculate price rank for this spec
			rank := s.calculateVendorPriceRank(specID, vendorID, quotes)
			priceRanks = append(priceRanks, rank)
		}

		// Calculate average price rank
		if len(priceRanks) > 0 {
			sum := 0.0
			for _, rank := range priceRanks {
				sum += rank
			}
			analysis.AveragePriceRank = sum / float64(len(priceRanks))
		}

		// Get vendor ratings
		rating, err := s.GetVendorRatingSummary(vendorID)
		if err == nil {
			analysis.Rating = rating
		}

		// Estimate shipping advantage (vendors supplying >50% of items)
		analysis.ShippingAdvantage = len(capabilities) > len(specIDs)/2

		analyses = append(analyses, analysis)
	}

	// Sort by coverage (specs count) descending, then by total cost ascending
	for i := 0; i < len(analyses); i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].SpecificationsCount < analyses[j].SpecificationsCount ||
				(analyses[i].SpecificationsCount == analyses[j].SpecificationsCount &&
					analyses[i].TotalCostIfUsed > analyses[j].TotalCostIfUsed) {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}

	return analyses, nil
}

// calculateVendorPriceRank determines where a vendor ranks on price for a specification
func (s *ProjectProcurementService) calculateVendorPriceRank(specID, vendorID uint, allQuotes []models.Quote) float64 {
	// Get all quotes for this spec
	specQuotes := make([]float64, 0)
	vendorPrice := 0.0

	for _, quote := range allQuotes {
		if quote.Product != nil && quote.Product.SpecificationID != nil && *quote.Product.SpecificationID == specID {
			specQuotes = append(specQuotes, quote.ConvertedPrice)
			if quote.VendorID == vendorID {
				vendorPrice = quote.ConvertedPrice
			}
		}
	}

	if len(specQuotes) == 0 {
		return 1.0
	}

	// Sort prices
	for i := 0; i < len(specQuotes); i++ {
		for j := i + 1; j < len(specQuotes); j++ {
			if specQuotes[i] > specQuotes[j] {
				specQuotes[i], specQuotes[j] = specQuotes[j], specQuotes[i]
			}
		}
	}

	// Find rank
	rank := 1.0
	for i, price := range specQuotes {
		if price == vendorPrice {
			rank = float64(i + 1)
			break
		}
	}

	return rank
}

// GenerateVendorRecommendations creates optimized vendor assignments based on strategy
func (s *ProjectProcurementService) GenerateVendorRecommendations(
	projectID uint,
	strategy string,
) ([]VendorRecommendation, error) {
	// Get consolidation analysis
	consolidation, err := s.GetVendorConsolidationAnalysis(projectID)
	if err != nil {
		return nil, err
	}

	if len(consolidation) == 0 {
		return []VendorRecommendation{}, nil
	}

	// Get project with BOM
	var project models.Project
	err = s.db.Preload("BillOfMaterials.Items").First(&project, projectID).Error
	if err != nil {
		return nil, err
	}

	// Apply strategy-specific algorithm
	switch strategy {
	case "lowest_cost":
		return s.generateLowestCostRecommendations(project, consolidation)
	case "fewest_vendors":
		return s.generateFewestVendorsRecommendations(project, consolidation)
	case "balanced":
		return s.generateBalancedRecommendations(project, consolidation)
	case "quality_focused":
		return s.generateQualityFocusedRecommendations(project, consolidation)
	default:
		return s.generateLowestCostRecommendations(project, consolidation)
	}
}

// generateLowestCostRecommendations always picks cheapest vendor per item
func (s *ProjectProcurementService) generateLowestCostRecommendations(
	project models.Project,
	consolidation []VendorConsolidationAnalysis,
) ([]VendorRecommendation, error) {
	recommendations := make([]VendorRecommendation, 0)
	vendorAssignments := make(map[uint][]uint) // vendorID -> []bomItemIDs
	vendorCosts := make(map[uint]float64)

	// For each BOM item, find cheapest vendor
	for _, bomItem := range project.BillOfMaterials.Items {
		bestVendorID := uint(0)
		bestCost := 0.0

		// Find vendor with best price for this spec
		for _, vendor := range consolidation {
			quotes, _ := s.quoteService.CompareQuotesForSpecification(bomItem.SpecificationID)
			for _, quote := range quotes {
				if quote.VendorID == vendor.VendorID {
					cost := quote.ConvertedPrice * float64(bomItem.Quantity)
					if bestVendorID == 0 || cost < bestCost {
						bestVendorID = vendor.VendorID
						bestCost = cost
					}
					break
				}
			}
		}

		if bestVendorID != 0 {
			vendorAssignments[bestVendorID] = append(vendorAssignments[bestVendorID], bomItem.ID)
			vendorCosts[bestVendorID] += bestCost
		}
	}

	// Build recommendations
	priority := 1
	for vendorID, bomItems := range vendorAssignments {
		vendorName := ""
		for _, v := range consolidation {
			if v.VendorID == vendorID {
				vendorName = v.VendorName
				break
			}
		}

		recommendations = append(recommendations, VendorRecommendation{
			VendorID:   vendorID,
			VendorName: vendorName,
			BOMItems:   bomItems,
			TotalCost:  vendorCosts[vendorID],
			ItemCount:  len(bomItems),
			Rationale:  "Selected for lowest cost on assigned items",
			Priority:   priority,
		})
		priority++
	}

	return recommendations, nil
}

// generateFewestVendorsRecommendations minimizes vendor count
func (s *ProjectProcurementService) generateFewestVendorsRecommendations(
	project models.Project,
	consolidation []VendorConsolidationAnalysis,
) ([]VendorRecommendation, error) {
	// Sort vendors by coverage (most specs covered first)
	// This is already done in GetVendorConsolidationAnalysis

	recommendations := make([]VendorRecommendation, 0)
	coveredSpecs := make(map[uint]bool)
	vendorAssignments := make(map[uint][]uint)
	vendorCosts := make(map[uint]float64)

	// Greedy assignment: pick vendor with most uncovered specs
	for len(coveredSpecs) < len(project.BillOfMaterials.Items) {
		bestVendorIdx := -1
		bestCoverage := 0

		// Find vendor covering most uncovered specs
		for i, vendor := range consolidation {
			coverage := 0
			quotes, _ := s.quoteService.CompareQuotesForSpecification(0) // We'll check per spec

			for _, bomItem := range project.BillOfMaterials.Items {
				if coveredSpecs[bomItem.SpecificationID] {
					continue
				}

				// Check if this vendor can supply this spec
				quotes, _ = s.quoteService.CompareQuotesForSpecification(bomItem.SpecificationID)
				canSupply := false
				for _, quote := range quotes {
					if quote.VendorID == vendor.VendorID {
						canSupply = true
						break
					}
				}
				if canSupply {
					coverage++
				}
			}

			if coverage > bestCoverage {
				bestCoverage = coverage
				bestVendorIdx = i
			}
		}

		if bestVendorIdx == -1 {
			break // No more vendors can help
		}

		// Assign this vendor
		vendor := consolidation[bestVendorIdx]
		for _, bomItem := range project.BillOfMaterials.Items {
			if coveredSpecs[bomItem.SpecificationID] {
				continue
			}

			quotes, _ := s.quoteService.CompareQuotesForSpecification(bomItem.SpecificationID)
			for _, quote := range quotes {
				if quote.VendorID == vendor.VendorID {
					coveredSpecs[bomItem.SpecificationID] = true
					vendorAssignments[vendor.VendorID] = append(vendorAssignments[vendor.VendorID], bomItem.ID)
					vendorCosts[vendor.VendorID] += quote.ConvertedPrice * float64(bomItem.Quantity)
					break
				}
			}
		}
	}

	// Build recommendations
	priority := 1
	for vendorID, bomItems := range vendorAssignments {
		vendorName := ""
		for _, v := range consolidation {
			if v.VendorID == vendorID {
				vendorName = v.VendorName
				break
			}
		}

		recommendations = append(recommendations, VendorRecommendation{
			VendorID:   vendorID,
			VendorName: vendorName,
			BOMItems:   bomItems,
			TotalCost:  vendorCosts[vendorID],
			ItemCount:  len(bomItems),
			Rationale:  fmt.Sprintf("Selected to minimize vendor count (covers %d items)", len(bomItems)),
			Priority:   priority,
		})
		priority++
	}

	return recommendations, nil
}

// generateBalancedRecommendations optimizes both cost and vendor count
func (s *ProjectProcurementService) generateBalancedRecommendations(
	project models.Project,
	consolidation []VendorConsolidationAnalysis,
) ([]VendorRecommendation, error) {
	// Get both strategies
	lowestCost, _ := s.generateLowestCostRecommendations(project, consolidation)
	fewestVendors, _ := s.generateFewestVendorsRecommendations(project, consolidation)

	// Calculate scores
	lowestCostTotal := 0.0
	for _, rec := range lowestCost {
		lowestCostTotal += rec.TotalCost
	}

	fewestVendorsTotal := 0.0
	for _, rec := range fewestVendors {
		fewestVendorsTotal += rec.TotalCost
	}

	// If fewest vendors is within 10% of lowest cost, prefer it
	if fewestVendorsTotal <= lowestCostTotal*1.10 {
		// Update rationale
		for i := range fewestVendors {
			fewestVendors[i].Rationale = "Balanced approach: minimizes vendors with acceptable cost"
		}
		return fewestVendors, nil
	}

	// Otherwise, use lowest cost
	for i := range lowestCost {
		lowestCost[i].Rationale = "Balanced approach: prioritizes cost optimization"
	}
	return lowestCost, nil
}

// generateQualityFocusedRecommendations prioritizes highly-rated vendors
func (s *ProjectProcurementService) generateQualityFocusedRecommendations(
	project models.Project,
	consolidation []VendorConsolidationAnalysis,
) ([]VendorRecommendation, error) {
	// Filter vendors by minimum rating (4.0+)
	qualityVendors := make([]VendorConsolidationAnalysis, 0)
	for _, vendor := range consolidation {
		if vendor.Rating != nil && vendor.Rating.OverallAvg >= 4.0 {
			qualityVendors = append(qualityVendors, vendor)
		}
	}

	// If no quality vendors, fall back to lowest cost
	if len(qualityVendors) == 0 {
		return s.generateLowestCostRecommendations(project, consolidation)
	}

	// Use fewest vendors strategy among quality vendors
	recommendations, err := s.generateFewestVendorsRecommendations(project, qualityVendors)
	if err != nil {
		return nil, err
	}

	// Update rationale
	for i := range recommendations {
		rating := "N/A"
		for _, v := range qualityVendors {
			if v.VendorID == recommendations[i].VendorID && v.Rating != nil {
				rating = fmt.Sprintf("%.1f/5.0", v.Rating.OverallAvg)
				break
			}
		}
		recommendations[i].Rationale = fmt.Sprintf("Quality-focused: highly-rated vendor (rating: %s)", rating)
	}

	return recommendations, nil
}

// CompareScenarios generates multiple procurement scenarios for comparison
func (s *ProjectProcurementService) CompareScenarios(projectID uint) ([]ProcurementScenario, error) {
	// Get project
	var project models.Project
	err := s.db.Preload("BillOfMaterials.Items").First(&project, projectID).Error
	if err != nil {
		return nil, err
	}

	scenarios := make([]ProcurementScenario, 0)

	// Scenario 1: Lowest Cost
	lowestCostRecs, err := s.GenerateVendorRecommendations(projectID, "lowest_cost")
	if err == nil {
		totalCost := 0.0
		vendorAssignments := make(map[uint][]uint)
		for _, rec := range lowestCostRecs {
			totalCost += rec.TotalCost
			vendorAssignments[rec.VendorID] = rec.BOMItems
		}

		scenarios = append(scenarios, ProcurementScenario{
			Name:              "Lowest Cost",
			Description:       "Minimizes total cost by selecting cheapest vendor for each item independently",
			VendorCount:       len(lowestCostRecs),
			TotalCost:         totalCost,
			SavingsVsBudget:   project.Budget - totalCost,
			Tradeoffs:         "Highest savings, but may involve many vendors (increased admin overhead)",
			VendorAssignments: vendorAssignments,
		})
	}

	// Scenario 2: Fewest Vendors
	fewestVendorsRecs, err := s.GenerateVendorRecommendations(projectID, "fewest_vendors")
	if err == nil {
		totalCost := 0.0
		vendorAssignments := make(map[uint][]uint)
		for _, rec := range fewestVendorsRecs {
			totalCost += rec.TotalCost
			vendorAssignments[rec.VendorID] = rec.BOMItems
		}

		scenarios = append(scenarios, ProcurementScenario{
			Name:              "Fewest Vendors",
			Description:       "Minimizes number of vendors to reduce administrative complexity",
			VendorCount:       len(fewestVendorsRecs),
			TotalCost:         totalCost,
			SavingsVsBudget:   project.Budget - totalCost,
			Tradeoffs:         "Simplifies ordering/management, but may cost slightly more than lowest cost",
			VendorAssignments: vendorAssignments,
		})
	}

	// Scenario 3: Balanced
	balancedRecs, err := s.GenerateVendorRecommendations(projectID, "balanced")
	if err == nil {
		totalCost := 0.0
		vendorAssignments := make(map[uint][]uint)
		for _, rec := range balancedRecs {
			totalCost += rec.TotalCost
			vendorAssignments[rec.VendorID] = rec.BOMItems
		}

		scenarios = append(scenarios, ProcurementScenario{
			Name:              "Balanced",
			Description:       "Optimizes both cost and vendor count for best overall value",
			VendorCount:       len(balancedRecs),
			TotalCost:         totalCost,
			SavingsVsBudget:   project.Budget - totalCost,
			Tradeoffs:         "Good balance between savings and simplicity",
			VendorAssignments: vendorAssignments,
		})
	}

	// Scenario 4: Quality Focused
	qualityRecs, err := s.GenerateVendorRecommendations(projectID, "quality_focused")
	if err == nil {
		totalCost := 0.0
		vendorAssignments := make(map[uint][]uint)
		for _, rec := range qualityRecs {
			totalCost += rec.TotalCost
			vendorAssignments[rec.VendorID] = rec.BOMItems
		}

		scenarios = append(scenarios, ProcurementScenario{
			Name:              "Quality Focused",
			Description:       "Prioritizes vendors with highest quality ratings (4.0+)",
			VendorCount:       len(qualityRecs),
			TotalCost:         totalCost,
			SavingsVsBudget:   project.Budget - totalCost,
			Tradeoffs:         "Higher quality/reliability, may have higher costs",
			VendorAssignments: vendorAssignments,
		})
	}

	return scenarios, nil
}

// ProjectSavingsSummary holds detailed savings analysis
type ProjectSavingsSummary struct {
	TotalSavingsUSD      float64
	SavingsPercent       float64
	SavingsByCategory    map[string]float64
	SavingsByVendor      map[string]float64
	ConsolidationSavings float64
	DetailedBreakdown    []SavingsLineItem
}

// SavingsLineItem represents savings for a single BOM item
type SavingsLineItem struct {
	BOMItemID            uint
	SpecificationName    string
	Quantity             int
	TargetPrice          float64
	RecommendedPrice     float64
	BestPrice            float64
	SavingsPerUnit       float64
	TotalSavings         float64
	SavingsPercent       float64
}

// CalculateProjectSavings performs comprehensive savings analysis
func (s *ProjectProcurementService) CalculateProjectSavings(projectID uint) (*ProjectSavingsSummary, error) {
	// Get project comparison data
	comparison, err := s.GetProjectProcurementComparison(projectID)
	if err != nil {
		return nil, err
	}

	summary := &ProjectSavingsSummary{
		SavingsByCategory:    make(map[string]float64),
		SavingsByVendor:      make(map[string]float64),
		DetailedBreakdown:    make([]SavingsLineItem, 0),
		ConsolidationSavings: 0,
	}

	// Calculate line item savings
	totalTargetCost := 0.0
	totalRecommendedCost := 0.0

	for _, bomAnalysis := range comparison.BOMItemAnalyses {
		if bomAnalysis.Specification == nil || bomAnalysis.BOMItem == nil {
			continue
		}

		lineItem := SavingsLineItem{
			BOMItemID:         bomAnalysis.BOMItem.ID,
			SpecificationName: bomAnalysis.Specification.Name,
			Quantity:          bomAnalysis.TotalQuantityNeeded,
		}

		// Calculate target price (from requisition items or default)
		// Use TotalQuantityPlanned when we have target costs from requisitions
		if bomAnalysis.TargetTotalCost > 0 && bomAnalysis.TotalQuantityPlanned > 0 {
			lineItem.TargetPrice = bomAnalysis.TargetTotalCost / float64(bomAnalysis.TotalQuantityPlanned)
			// Use planned quantity for savings calculation when we have target data
			lineItem.Quantity = bomAnalysis.TotalQuantityPlanned
		}

		// Get recommended and best prices
		if bomAnalysis.RecommendedQuote != nil {
			lineItem.RecommendedPrice = bomAnalysis.RecommendedQuote.ConvertedPrice
		}
		if bomAnalysis.BestQuote != nil {
			lineItem.BestPrice = bomAnalysis.BestQuote.ConvertedPrice
		}

		// Calculate savings
		if lineItem.TargetPrice > 0 && lineItem.RecommendedPrice > 0 {
			lineItem.SavingsPerUnit = lineItem.TargetPrice - lineItem.RecommendedPrice
			lineItem.TotalSavings = lineItem.SavingsPerUnit * float64(lineItem.Quantity)
			if lineItem.TargetPrice > 0 {
				lineItem.SavingsPercent = (lineItem.SavingsPerUnit / lineItem.TargetPrice) * 100
			}

			// Aggregate by category (specification)
			summary.SavingsByCategory[bomAnalysis.Specification.Name] += lineItem.TotalSavings

			// Track vendor contributions
			if bomAnalysis.RecommendedQuote != nil && bomAnalysis.RecommendedQuote.Vendor != nil {
				vendorName := bomAnalysis.RecommendedQuote.Vendor.Name
				summary.SavingsByVendor[vendorName] += lineItem.TotalSavings
			}
		}

		totalTargetCost += lineItem.TargetPrice * float64(lineItem.Quantity)
		totalRecommendedCost += lineItem.RecommendedPrice * float64(lineItem.Quantity)

		summary.DetailedBreakdown = append(summary.DetailedBreakdown, lineItem)
	}

	// Calculate overall savings
	summary.TotalSavingsUSD = totalTargetCost - totalRecommendedCost
	if totalTargetCost > 0 {
		summary.SavingsPercent = (summary.TotalSavingsUSD / totalTargetCost) * 100
	}

	// Estimate consolidation savings (administrative overhead reduction)
	if len(comparison.VendorRecommendations) > 0 {
		// Estimate $100-500 per vendor in administrative costs avoided
		// Base on how many vendors we're NOT using
		consolidation, _ := s.GetVendorConsolidationAnalysis(projectID)
		potentialVendors := len(consolidation)
		actualVendors := len(comparison.VendorRecommendations)
		if potentialVendors > actualVendors {
			vendorsAvoided := potentialVendors - actualVendors
			summary.ConsolidationSavings = float64(vendorsAvoided) * 250 // $250 per vendor avoided
		}
	}

	return summary, nil
}

// EnhancedRiskAssessment provides more detailed risk analysis
type EnhancedRiskAssessment struct {
	OverallRisk           string
	RiskScore             int // 0-100
	CategoryRisks         map[string]CategoryRisk
	TimelineRisk          TimelineRisk
	BudgetRisk            BudgetRisk
	SupplyChainRisk       SupplyChainRisk
	QualityRisk           QualityRisk
	MitigationActions     []MitigationAction
	HighPriorityActions   []string
}

// CategoryRisk represents risk in a specific category
type CategoryRisk struct {
	Level               string
	Score               int // 0-100
	Issues              []string
	AffectedItems       int
	EstimatedImpact     string
}

// TimelineRisk assesses schedule-related risks
type TimelineRisk struct {
	Level                string
	QuotesExpiringSoon   int
	QuotesExpired        int
	AverageQuoteAge      int
	LeadTimeRisks        []string
}

// BudgetRisk assesses financial risks
type BudgetRisk struct {
	Level                string
	ProjectedOverrun     float64
	OverrunPercent       float64
	ItemsOverBudget      int
	ContingencyNeeded    float64
}

// SupplyChainRisk assesses vendor and availability risks
type SupplyChainRisk struct {
	Level                string
	SingleSourceItems    int
	NoQuoteItems         int
	LowVendorDiversity   bool
	VendorCapacityIssues []string
}

// QualityRisk assesses quality-related risks
type QualityRisk struct {
	Level                string
	LowRatedVendors      int
	UnratedVendors       int
	QualityIssues        []string
}

// MitigationAction represents a recommended action
type MitigationAction struct {
	Priority    string // critical, high, medium, low
	Category    string
	Action      string
	Impact      string
	Effort      string // low, medium, high
	Timeline    string // immediate, short-term, long-term
}

// AssessEnhancedProjectRisks performs comprehensive risk analysis
func (s *ProjectProcurementService) AssessEnhancedProjectRisks(projectID uint) (*EnhancedRiskAssessment, error) {
	// Get project comparison
	comparison, err := s.GetProjectProcurementComparison(projectID)
	if err != nil {
		return nil, err
	}

	assessment := &EnhancedRiskAssessment{
		CategoryRisks:       make(map[string]CategoryRisk),
		MitigationActions:   make([]MitigationAction, 0),
		HighPriorityActions: make([]string, 0),
	}

	// Assess quote coverage risk
	coverageRisk := s.assessQuoteCoverageRisk(comparison)
	assessment.CategoryRisks["quote_coverage"] = coverageRisk

	// Assess timeline risk
	assessment.TimelineRisk = s.assessTimelineRisk(comparison)

	// Assess budget risk
	assessment.BudgetRisk = s.assessBudgetRisk(comparison)

	// Assess supply chain risk
	assessment.SupplyChainRisk = s.assessSupplyChainRisk(comparison)

	// Assess quality risk
	assessment.QualityRisk = s.assessQualityRisk(comparison)

	// Calculate overall risk score
	assessment.RiskScore = s.calculateOverallRiskScore(assessment)

	// Determine overall risk level
	if assessment.RiskScore >= 75 {
		assessment.OverallRisk = "critical"
	} else if assessment.RiskScore >= 50 {
		assessment.OverallRisk = "high"
	} else if assessment.RiskScore >= 25 {
		assessment.OverallRisk = "medium"
	} else {
		assessment.OverallRisk = "low"
	}

	// Generate mitigation actions
	assessment.MitigationActions = s.generateMitigationActions(assessment)

	// Extract high-priority actions
	for _, action := range assessment.MitigationActions {
		if action.Priority == "critical" || action.Priority == "high" {
			assessment.HighPriorityActions = append(assessment.HighPriorityActions, action.Action)
		}
	}

	return assessment, nil
}

// assessQuoteCoverageRisk evaluates quote availability
func (s *ProjectProcurementService) assessQuoteCoverageRisk(comparison *ProjectProcurementComparison) CategoryRisk {
	risk := CategoryRisk{
		Issues:        make([]string, 0),
		AffectedItems: 0,
	}

	uncovered := 0
	partial := 0
	singleSource := 0

	for _, analysis := range comparison.BOMItemAnalyses {
		if !analysis.HasSufficientQuotes {
			uncovered++
			risk.AffectedItems++
			risk.Issues = append(risk.Issues, fmt.Sprintf("%s has no quotes", analysis.Specification.Name))
		} else if len(analysis.AvailableQuotes) == 1 {
			singleSource++
			risk.AffectedItems++
		}
	}

	// Calculate score
	totalItems := len(comparison.BOMItemAnalyses)
	if totalItems > 0 {
		riskPct := float64(uncovered+partial) / float64(totalItems) * 100
		risk.Score = int(riskPct)
	}

	// Determine level
	if uncovered > 0 {
		risk.Level = "critical"
		risk.EstimatedImpact = "Project cannot proceed without quotes for all items"
	} else if singleSource > totalItems/2 {
		risk.Level = "high"
		risk.EstimatedImpact = "Limited negotiation leverage and supply chain vulnerability"
	} else if singleSource > 0 {
		risk.Level = "medium"
		risk.EstimatedImpact = "Some items have limited vendor options"
	} else {
		risk.Level = "low"
		risk.EstimatedImpact = "Good vendor diversity and quote coverage"
	}

	return risk
}

// assessTimelineRisk evaluates schedule-related risks
func (s *ProjectProcurementService) assessTimelineRisk(comparison *ProjectProcurementComparison) TimelineRisk {
	risk := TimelineRisk{
		LeadTimeRisks: make([]string, 0),
	}

	risk.QuotesExpiringSoon = comparison.QuoteFreshness.StaleQuotes
	risk.QuotesExpired = comparison.QuoteFreshness.ExpiredQuotes
	risk.AverageQuoteAge = comparison.QuoteFreshness.AverageAgeDays

	// Determine level
	if risk.QuotesExpired > 0 {
		risk.Level = "high"
		risk.LeadTimeRisks = append(risk.LeadTimeRisks, fmt.Sprintf("%d quotes have expired and need renewal", risk.QuotesExpired))
	} else if risk.QuotesExpiringSoon > 0 {
		risk.Level = "medium"
		risk.LeadTimeRisks = append(risk.LeadTimeRisks, fmt.Sprintf("%d quotes are becoming stale", risk.QuotesExpiringSoon))
	} else {
		risk.Level = "low"
	}

	return risk
}

// assessBudgetRisk evaluates financial risks
func (s *ProjectProcurementService) assessBudgetRisk(comparison *ProjectProcurementComparison) BudgetRisk {
	risk := BudgetRisk{}

	if comparison.ProjectBudget > 0 {
		risk.ProjectedOverrun = comparison.RecommendedCost - comparison.ProjectBudget
		if comparison.ProjectBudget > 0 {
			risk.OverrunPercent = (risk.ProjectedOverrun / comparison.ProjectBudget) * 100
		}

		// Count items over budget
		for _, analysis := range comparison.BOMItemAnalyses {
			if analysis.TargetTotalCost > 0 && analysis.RecommendedTotalCost > analysis.TargetTotalCost {
				risk.ItemsOverBudget++
			}
		}

		// Determine level and contingency
		if risk.OverrunPercent > 20 {
			risk.Level = "critical"
			risk.ContingencyNeeded = risk.ProjectedOverrun * 1.2 // 20% buffer
		} else if risk.OverrunPercent > 10 {
			risk.Level = "high"
			risk.ContingencyNeeded = risk.ProjectedOverrun * 1.15
		} else if risk.OverrunPercent > 0 {
			risk.Level = "medium"
			risk.ContingencyNeeded = risk.ProjectedOverrun * 1.10
		} else {
			risk.Level = "low"
			risk.ContingencyNeeded = 0
		}
	}

	return risk
}

// assessSupplyChainRisk evaluates vendor availability risks
func (s *ProjectProcurementService) assessSupplyChainRisk(comparison *ProjectProcurementComparison) SupplyChainRisk {
	risk := SupplyChainRisk{
		VendorCapacityIssues: make([]string, 0),
	}

	for _, analysis := range comparison.BOMItemAnalyses {
		if !analysis.HasSufficientQuotes {
			risk.NoQuoteItems++
		} else if len(analysis.AvailableQuotes) == 1 {
			risk.SingleSourceItems++
		}
	}

	// Check vendor diversity
	if len(comparison.VendorRecommendations) < 2 && len(comparison.BOMItemAnalyses) > 5 {
		risk.LowVendorDiversity = true
	}

	// Determine level
	if risk.NoQuoteItems > 0 {
		risk.Level = "critical"
	} else if risk.SingleSourceItems > len(comparison.BOMItemAnalyses)/2 {
		risk.Level = "high"
	} else if risk.LowVendorDiversity {
		risk.Level = "medium"
	} else {
		risk.Level = "low"
	}

	return risk
}

// assessQualityRisk evaluates quality-related risks
func (s *ProjectProcurementService) assessQualityRisk(comparison *ProjectProcurementComparison) QualityRisk {
	risk := QualityRisk{
		QualityIssues: make([]string, 0),
	}

	vendorRatings := make(map[uint]*VendorRatingsSummary)

	// Get ratings for recommended vendors
	for _, rec := range comparison.VendorRecommendations {
		rating, err := s.GetVendorRatingSummary(rec.VendorID)
		if err == nil && rating != nil {
			vendorRatings[rec.VendorID] = rating

			if rating.TotalRatings > 0 && rating.OverallAvg < 3.0 {
				risk.LowRatedVendors++
				risk.QualityIssues = append(risk.QualityIssues,
					fmt.Sprintf("%s has low average rating (%.1f/5.0)", rec.VendorName, rating.OverallAvg))
			}
		} else {
			risk.UnratedVendors++
		}
	}

	// Determine level
	if risk.LowRatedVendors > 0 {
		risk.Level = "high"
	} else if risk.UnratedVendors > len(comparison.VendorRecommendations)/2 {
		risk.Level = "medium"
	} else {
		risk.Level = "low"
	}

	return risk
}

// calculateOverallRiskScore computes aggregate risk score
func (s *ProjectProcurementService) calculateOverallRiskScore(assessment *EnhancedRiskAssessment) int {
	// Weighted risk scoring
	weights := map[string]float64{
		"quote_coverage": 0.30,
		"timeline":       0.15,
		"budget":         0.25,
		"supply_chain":   0.20,
		"quality":        0.10,
	}

	score := 0.0

	// Coverage risk
	if coverageRisk, ok := assessment.CategoryRisks["quote_coverage"]; ok {
		score += float64(coverageRisk.Score) * weights["quote_coverage"]
	}

	// Timeline risk
	timelineScore := 0
	switch assessment.TimelineRisk.Level {
	case "critical":
		timelineScore = 100
	case "high":
		timelineScore = 75
	case "medium":
		timelineScore = 50
	case "low":
		timelineScore = 25
	}
	score += float64(timelineScore) * weights["timeline"]

	// Budget risk
	budgetScore := 0
	switch assessment.BudgetRisk.Level {
	case "critical":
		budgetScore = 100
	case "high":
		budgetScore = 75
	case "medium":
		budgetScore = 50
	case "low":
		budgetScore = 25
	}
	score += float64(budgetScore) * weights["budget"]

	// Supply chain risk
	supplyScore := 0
	switch assessment.SupplyChainRisk.Level {
	case "critical":
		supplyScore = 100
	case "high":
		supplyScore = 75
	case "medium":
		supplyScore = 50
	case "low":
		supplyScore = 25
	}
	score += float64(supplyScore) * weights["supply_chain"]

	// Quality risk
	qualityScore := 0
	switch assessment.QualityRisk.Level {
	case "critical":
		qualityScore = 100
	case "high":
		qualityScore = 75
	case "medium":
		qualityScore = 50
	case "low":
		qualityScore = 25
	}
	score += float64(qualityScore) * weights["quality"]

	return int(score)
}

// generateMitigationActions creates actionable recommendations
func (s *ProjectProcurementService) generateMitigationActions(assessment *EnhancedRiskAssessment) []MitigationAction {
	actions := make([]MitigationAction, 0)

	// Quote coverage actions
	if coverageRisk, ok := assessment.CategoryRisks["quote_coverage"]; ok {
		if coverageRisk.Level == "critical" || coverageRisk.Level == "high" {
			actions = append(actions, MitigationAction{
				Priority: "critical",
				Category: "quote_coverage",
				Action:   "Request quotes from additional vendors for items with no or limited quotes",
				Impact:   "Enables procurement and improves negotiation leverage",
				Effort:   "medium",
				Timeline: "immediate",
			})
		}
	}

	// Timeline actions
	if assessment.TimelineRisk.QuotesExpired > 0 {
		actions = append(actions, MitigationAction{
			Priority: "high",
			Category: "timeline",
			Action:   fmt.Sprintf("Renew %d expired quotes before proceeding with procurement", assessment.TimelineRisk.QuotesExpired),
			Impact:   "Ensures current pricing and availability",
			Effort:   "low",
			Timeline: "immediate",
		})
	}

	// Budget actions
	if assessment.BudgetRisk.Level == "critical" || assessment.BudgetRisk.Level == "high" {
		actions = append(actions, MitigationAction{
			Priority: "high",
			Category: "budget",
			Action:   fmt.Sprintf("Secure additional budget of $%.2f or negotiate better pricing", assessment.BudgetRisk.ContingencyNeeded),
			Impact:   "Prevents project delays due to funding shortfall",
			Effort:   "high",
			Timeline: "short-term",
		})
	}

	// Supply chain actions
	if assessment.SupplyChainRisk.SingleSourceItems > 0 {
		actions = append(actions, MitigationAction{
			Priority: "medium",
			Category: "supply_chain",
			Action:   fmt.Sprintf("Identify backup vendors for %d single-source items", assessment.SupplyChainRisk.SingleSourceItems),
			Impact:   "Reduces supply chain risk and improves resilience",
			Effort:   "medium",
			Timeline: "short-term",
		})
	}

	// Quality actions
	if assessment.QualityRisk.LowRatedVendors > 0 {
		actions = append(actions, MitigationAction{
			Priority: "medium",
			Category: "quality",
			Action:   "Review low-rated vendors and consider alternatives with better track records",
			Impact:   "Reduces risk of quality issues and delays",
			Effort:   "low",
			Timeline: "immediate",
		})
	}

	return actions
}

// ============================================================================
// PHASE 4: DASHBOARD AND REPORTING
// ============================================================================

// ProjectDashboard provides comprehensive project overview
type ProjectDashboard struct {
	Project           *models.Project
	Progress          ProjectProgress
	Financial         ProjectFinancialOverview
	Procurement       ProjectProcurementStatus
	VendorPerformance []VendorPerformanceSummary
	Risks             []RiskIndicator
	RecentActivity    []ActivityItem
	ChartsData        ProjectChartsData
}

// ProjectProgress tracks project completion metrics
type ProjectProgress struct {
	BOMCoverage          float64 // % of BOM items with quotes
	RequisitionsComplete int
	RequisitionsTotal    int
	OrdersPlaced         int
	OrdersReceived       int
	TimelineStatus       string // on_track, at_risk, delayed
	DaysToDeadline       int
}

// ProjectFinancialOverview summarizes financial status
type ProjectFinancialOverview struct {
	Budget         float64
	Committed      float64 // Orders placed
	Estimated      float64 // Best quotes for remaining items
	Remaining      float64 // Budget - (Committed + Estimated)
	Savings        float64
	SavingsPercent float64
	BudgetHealth   string // healthy, warning, critical
}

// ProjectProcurementStatus tracks procurement progress
type ProjectProcurementStatus struct {
	TotalItems          int
	ItemsWithQuotes     int
	ItemsOrdered        int
	ItemsReceived       int
	AverageLeadTimeDays int
	VendorsEngaged      int
	QuoteFreshness      string // fresh, aging, stale
}

// VendorPerformanceSummary summarizes vendor activity
type VendorPerformanceSummary struct {
	VendorID       uint
	VendorName     string
	ItemsSupplied  int
	TotalValue     float64
	AverageRating  float64
	OnTimeDelivery float64 // Percentage
	Status         string  // active, pending, completed
}

// RiskIndicator summarizes risk category
type RiskIndicator struct {
	Category string
	Level    string // low, medium, high, critical
	Count    int    // Number of issues
	TopIssue string // Most critical issue description
}

// ActivityItem tracks project events
type ActivityItem struct {
	Timestamp   time.Time
	Type        string // quote_added, order_placed, item_received, etc.
	Description string
	Impact      string // Positive, Negative, Neutral
}

// ProjectChartsData contains data for visualizations
type ProjectChartsData struct {
	BudgetUtilization  []ChartDataPoint // Time series of budget allocation
	SavingsByCategory  []ChartDataPoint // Pie chart: savings per specification category
	VendorDistribution []ChartDataPoint // Pie chart: spending per vendor
	TimelineGantt      []GanttItem      // Gantt chart of requisitions
	CostComparison     []ChartDataPoint // Bar chart: Budget vs. Estimated vs. Committed
}

// ChartDataPoint represents a single data point in a chart
type ChartDataPoint struct {
	Label    string
	Value    float64
	Color    string
	Metadata map[string]interface{}
}

// GanttItem represents an item in a Gantt chart
type GanttItem struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    string
	DependsOn []uint
}

// GetProjectDashboard generates comprehensive project dashboard
func (s *ProjectProcurementService) GetProjectDashboard(projectID uint) (*ProjectDashboard, error) {
	// Load project
	var project models.Project
	if err := s.db.Preload("BillOfMaterials.Items.Specification").
		Preload("Requisitions.Items.BOMItem.Specification").
		Preload("Requisitions.Items.SelectedQuote.Vendor").
		First(&project, projectID).Error; err != nil {
		return nil, &NotFoundError{Entity: "Project", ID: projectID}
	}

	dashboard := &ProjectDashboard{
		Project: &project,
	}

	// Calculate progress metrics
	progress, err := s.calculateProjectProgress(&project)
	if err != nil {
		return nil, err
	}
	dashboard.Progress = progress

	// Calculate financial overview
	financial, err := s.calculateFinancialOverview(&project)
	if err != nil {
		return nil, err
	}
	dashboard.Financial = financial

	// Calculate procurement status
	procurementStatus, err := s.calculateProcurementStatus(&project)
	if err != nil {
		return nil, err
	}
	dashboard.Procurement = procurementStatus

	// Aggregate vendor performance
	vendorPerf, err := s.aggregateVendorPerformance(&project)
	if err != nil {
		return nil, err
	}
	dashboard.VendorPerformance = vendorPerf

	// Identify risk indicators
	risks, err := s.identifyRiskIndicators(projectID)
	if err != nil {
		return nil, err
	}
	dashboard.Risks = risks

	// Fetch recent activity
	activity, err := s.getRecentActivity(&project)
	if err != nil {
		return nil, err
	}
	dashboard.RecentActivity = activity

	// Generate chart data
	chartsData, err := s.generateChartData(&project, financial, vendorPerf)
	if err != nil {
		return nil, err
	}
	dashboard.ChartsData = chartsData

	return dashboard, nil
}

// calculateProjectProgress computes progress metrics
func (s *ProjectProcurementService) calculateProjectProgress(project *models.Project) (ProjectProgress, error) {
	progress := ProjectProgress{}

	if project.BillOfMaterials == nil {
		return progress, nil
	}

	// BOM coverage
	totalBOMItems := len(project.BillOfMaterials.Items)
	itemsWithQuotes := 0

	for _, bomItem := range project.BillOfMaterials.Items {
		if bomItem.SpecificationID > 0 {
			// Check if this specification has quotes
			var quoteCount int64
			s.db.Model(&models.Quote{}).
				Joins("JOIN products ON products.id = quotes.product_id").
				Where("products.specification_id = ?", bomItem.SpecificationID).
				Count(&quoteCount)

			if quoteCount > 0 {
				itemsWithQuotes++
			}
		}
	}

	if totalBOMItems > 0 {
		progress.BOMCoverage = float64(itemsWithQuotes) / float64(totalBOMItems) * 100
	}

	// Requisition counts
	progress.RequisitionsTotal = len(project.Requisitions)
	// Note: ProjectRequisitions don't have status, count all as complete if they exist
	progress.RequisitionsComplete = len(project.Requisitions)

	// Order counts
	var orders []models.PurchaseOrder
	s.db.Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Find(&orders)

	progress.OrdersPlaced = len(orders)
	for _, order := range orders {
		if order.Status == "received" {
			progress.OrdersReceived++
		}
	}

	// Timeline status
	if project.Deadline != nil {
		now := time.Now()
		progress.DaysToDeadline = int(project.Deadline.Sub(now).Hours() / 24)

		if progress.DaysToDeadline < 0 {
			progress.TimelineStatus = "delayed"
		} else if progress.DaysToDeadline < 30 && progress.BOMCoverage < 80 {
			progress.TimelineStatus = "at_risk"
		} else {
			progress.TimelineStatus = "on_track"
		}
	} else {
		progress.TimelineStatus = "on_track"
		progress.DaysToDeadline = 0
	}

	return progress, nil
}

// calculateFinancialOverview computes financial metrics
func (s *ProjectProcurementService) calculateFinancialOverview(project *models.Project) (ProjectFinancialOverview, error) {
	financial := ProjectFinancialOverview{
		Budget: project.Budget,
	}

	// Calculate committed (orders placed)
	var committedSum float64
	if project.BillOfMaterials != nil {
		s.db.Model(&models.PurchaseOrder{}).
			Select("COALESCE(SUM(purchase_orders.quantity * quotes.converted_price), 0)").
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
			Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
			Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
			Scan(&committedSum)
	}
	financial.Committed = committedSum

	// Calculate estimated (best quotes for remaining items)
	if project.BillOfMaterials != nil {
		for _, bomItem := range project.BillOfMaterials.Items {
			if bomItem.SpecificationID == 0 {
				continue
			}

			// Check if already ordered
			var orderedQty int
			s.db.Model(&models.PurchaseOrder{}).
				Select("COALESCE(SUM(purchase_orders.quantity), 0)").
				Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
				Joins("JOIN products ON products.id = quotes.product_id").
				Where("products.specification_id = ?", bomItem.SpecificationID).
				Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
				Scan(&orderedQty)

			remainingQty := bomItem.Quantity - orderedQty
			if remainingQty > 0 {
				// Get best quote
				var bestPrice float64
				err := s.db.Model(&models.Quote{}).
					Select("MIN(quotes.converted_price)").
					Joins("JOIN products ON products.id = quotes.product_id").
					Where("products.specification_id = ?", bomItem.SpecificationID).
					Where("quotes.valid_until > ?", time.Now()).
					Scan(&bestPrice).Error

				if err == nil && bestPrice > 0 {
					financial.Estimated += bestPrice * float64(remainingQty)
				}
			}
		}
	}

	financial.Remaining = financial.Budget - (financial.Committed + financial.Estimated)

	// Calculate savings
	savingsSummary, err := s.CalculateProjectSavings(project.ID)
	if err == nil {
		financial.Savings = savingsSummary.TotalSavingsUSD
		financial.SavingsPercent = savingsSummary.SavingsPercent
	}

	// Budget health
	utilizationPercent := ((financial.Committed + financial.Estimated) / financial.Budget) * 100
	if utilizationPercent > 100 {
		financial.BudgetHealth = "critical"
	} else if utilizationPercent > 90 {
		financial.BudgetHealth = "warning"
	} else {
		financial.BudgetHealth = "healthy"
	}

	return financial, nil
}

// calculateProcurementStatus computes procurement metrics
func (s *ProjectProcurementService) calculateProcurementStatus(project *models.Project) (ProjectProcurementStatus, error) {
	status := ProjectProcurementStatus{}

	if project.BillOfMaterials == nil {
		return status, nil
	}

	status.TotalItems = len(project.BillOfMaterials.Items)

	// Items with quotes
	itemsWithQuotes := 0
	for _, bomItem := range project.BillOfMaterials.Items {
		if bomItem.SpecificationID > 0 {
			var quoteCount int64
			s.db.Model(&models.Quote{}).
				Joins("JOIN products ON products.id = quotes.product_id").
				Where("products.specification_id = ?", bomItem.SpecificationID).
				Count(&quoteCount)

			if quoteCount > 0 {
				itemsWithQuotes++
			}
		}
	}
	status.ItemsWithQuotes = itemsWithQuotes

	// Items ordered and received
	for _, bomItem := range project.BillOfMaterials.Items {
		if bomItem.SpecificationID == 0 {
			continue
		}

		var orderCount int64
		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Where("products.specification_id = ?", bomItem.SpecificationID).
			Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
			Count(&orderCount)

		if orderCount > 0 {
			status.ItemsOrdered++
		}

		var receivedCount int64
		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Where("products.specification_id = ?", bomItem.SpecificationID).
			Where("purchase_orders.status = ?", "received").
			Count(&receivedCount)

		if receivedCount > 0 {
			status.ItemsReceived++
		}
	}

	// Average lead time
	var avgLeadTime float64
	s.db.Model(&models.PurchaseOrder{}).
		Select("AVG(JULIANDAY(actual_delivery_date) - JULIANDAY(order_date))").
		Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Where("purchase_orders.actual_delivery_date IS NOT NULL").
		Scan(&avgLeadTime)
	status.AverageLeadTimeDays = int(avgLeadTime)

	// Vendors engaged
	var vendorIDs []uint
	s.db.Model(&models.Quote{}).
		Select("DISTINCT quotes.vendor_id").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Pluck("vendor_id", &vendorIDs)
	status.VendorsEngaged = len(vendorIDs)

	// Quote freshness
	var freshCount, staleCount, expiredCount int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)

	s.db.Model(&models.Quote{}).
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Where("quotes.quote_date > ?", thirtyDaysAgo).
		Count(&freshCount)

	s.db.Model(&models.Quote{}).
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Where("quotes.quote_date BETWEEN ? AND ?", ninetyDaysAgo, thirtyDaysAgo).
		Count(&staleCount)

	s.db.Model(&models.Quote{}).
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Where("quotes.valid_until < ?", time.Now()).
		Count(&expiredCount)

	totalQuotes := freshCount + staleCount + expiredCount
	if totalQuotes > 0 {
		if float64(freshCount)/float64(totalQuotes) > 0.7 {
			status.QuoteFreshness = "fresh"
		} else if float64(staleCount)/float64(totalQuotes) > 0.5 {
			status.QuoteFreshness = "stale"
		} else {
			status.QuoteFreshness = "aging"
		}
	} else {
		status.QuoteFreshness = "none"
	}

	return status, nil
}

// aggregateVendorPerformance summarizes vendor activity
func (s *ProjectProcurementService) aggregateVendorPerformance(project *models.Project) ([]VendorPerformanceSummary, error) {
	summaries := make([]VendorPerformanceSummary, 0)

	if project.BillOfMaterials == nil {
		return summaries, nil
	}

	// Get all vendors with orders for this project
	var vendorIDs []uint
	s.db.Model(&models.PurchaseOrder{}).
		Select("DISTINCT quotes.vendor_id").
		Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
		Pluck("vendor_id", &vendorIDs)

	for _, vendorID := range vendorIDs {
		var vendor models.Vendor
		if err := s.db.First(&vendor, vendorID).Error; err != nil {
			continue
		}

		summary := VendorPerformanceSummary{
			VendorID:   vendorID,
			VendorName: vendor.Name,
		}

		// Items supplied
		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
			Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
			Count(&[]int64{int64(summary.ItemsSupplied)}[0])

		// Total value
		var totalValue float64
		s.db.Model(&models.PurchaseOrder{}).
			Select("COALESCE(SUM(purchase_orders.quantity * quotes.converted_price), 0)").
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
			Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status NOT IN (?)", []string{"cancelled"}).
			Scan(&totalValue)
		summary.TotalValue = totalValue

		// Average rating
		ratingsSummary, err := s.GetVendorRatingSummary(vendorID)
		if err == nil && ratingsSummary != nil {
			summary.AverageRating = ratingsSummary.OverallAvg
		}

		// On-time delivery
		var totalOrders, onTimeOrders int64
		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
			Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status = ?", "received").
			Where("purchase_orders.actual_delivery_date IS NOT NULL").
			Count(&totalOrders)

		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Joins("JOIN products ON products.id = quotes.product_id").
			Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
			Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status = ?", "received").
			Where("purchase_orders.actual_delivery_date <= purchase_orders.expected_delivery_date").
			Count(&onTimeOrders)

		if totalOrders > 0 {
			summary.OnTimeDelivery = float64(onTimeOrders) / float64(totalOrders) * 100
		}

		// Status
		var pendingCount, completedCount int64
		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status IN (?)", []string{"pending", "approved", "ordered", "shipped"}).
			Count(&pendingCount)

		s.db.Model(&models.PurchaseOrder{}).
			Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
			Where("quotes.vendor_id = ?", vendorID).
			Where("purchase_orders.status = ?", "received").
			Count(&completedCount)

		if pendingCount > 0 {
			summary.Status = "active"
		} else if completedCount > 0 {
			summary.Status = "completed"
		} else {
			summary.Status = "pending"
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// identifyRiskIndicators summarizes risks by category
func (s *ProjectProcurementService) identifyRiskIndicators(projectID uint) ([]RiskIndicator, error) {
	indicators := make([]RiskIndicator, 0)

	assessment, err := s.AssessEnhancedProjectRisks(projectID)
	if err != nil {
		return indicators, err
	}

	// Convert category risks to indicators
	for category, risk := range assessment.CategoryRisks {
		topIssue := risk.EstimatedImpact
		if len(risk.Issues) > 0 {
			topIssue = risk.Issues[0]
		}
		indicator := RiskIndicator{
			Category: category,
			Level:    risk.Level,
			Count:    risk.AffectedItems,
			TopIssue: topIssue,
		}
		indicators = append(indicators, indicator)
	}

	return indicators, nil
}

// getRecentActivity fetches recent project events
func (s *ProjectProcurementService) getRecentActivity(project *models.Project) ([]ActivityItem, error) {
	activities := make([]ActivityItem, 0)

	// Get recent quotes added
	var recentQuotes []models.Quote
	s.db.Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Order("quotes.created_at DESC").
		Limit(10).
		Preload("Vendor").
		Preload("Product").
		Find(&recentQuotes)

	for _, quote := range recentQuotes {
		activities = append(activities, ActivityItem{
			Timestamp:   quote.CreatedAt,
			Type:        "quote_added",
			Description: fmt.Sprintf("New quote from %s for %s at $%.2f", quote.Vendor.Name, quote.Product.Name, quote.ConvertedPrice),
			Impact:      "Positive",
		})
	}

	// Get recent orders placed
	var recentOrders []models.PurchaseOrder
	s.db.Joins("JOIN quotes ON quotes.id = purchase_orders.quote_id").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("JOIN bill_of_materials_items ON bill_of_materials_items.specification_id = products.specification_id").
		Where("bill_of_materials_items.bill_of_materials_id = ?", project.BillOfMaterials.ID).
		Order("purchase_orders.created_at DESC").
		Limit(10).
		Preload("Quote.Vendor").
		Preload("Quote.Product").
		Find(&recentOrders)

	for _, order := range recentOrders {
		activities = append(activities, ActivityItem{
			Timestamp:   order.CreatedAt,
			Type:        "order_placed",
			Description: fmt.Sprintf("Order %s placed with %s for %d units", order.PONumber, order.Quote.Vendor.Name, order.Quantity),
			Impact:      "Positive",
		})
	}

	// Sort all activities by timestamp
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	// Limit to 20 most recent
	if len(activities) > 20 {
		activities = activities[:20]
	}

	return activities, nil
}

// generateChartData creates visualization data
func (s *ProjectProcurementService) generateChartData(
	project *models.Project,
	financial ProjectFinancialOverview,
	vendorPerf []VendorPerformanceSummary,
) (ProjectChartsData, error) {
	chartsData := ProjectChartsData{}

	// Budget utilization
	chartsData.BudgetUtilization = []ChartDataPoint{
		{
			Label: "Committed",
			Value: financial.Committed,
			Color: "#4CAF50",
		},
		{
			Label: "Estimated",
			Value: financial.Estimated,
			Color: "#FFC107",
		},
		{
			Label: "Remaining",
			Value: financial.Remaining,
			Color: "#2196F3",
		},
	}

	// Cost comparison
	chartsData.CostComparison = []ChartDataPoint{
		{
			Label: "Budget",
			Value: financial.Budget,
			Color: "#9E9E9E",
		},
		{
			Label: "Estimated Total",
			Value: financial.Committed + financial.Estimated,
			Color: "#FF9800",
		},
		{
			Label: "Committed",
			Value: financial.Committed,
			Color: "#4CAF50",
		},
	}

	// Vendor distribution
	for _, vp := range vendorPerf {
		chartsData.VendorDistribution = append(chartsData.VendorDistribution, ChartDataPoint{
			Label: vp.VendorName,
			Value: vp.TotalValue,
			Color: "", // Will be assigned by frontend
		})
	}

	// Savings by category
	savingsSummary, err := s.CalculateProjectSavings(project.ID)
	if err == nil {
		for category, savings := range savingsSummary.SavingsByCategory {
			if savings > 0 {
				chartsData.SavingsByCategory = append(chartsData.SavingsByCategory, ChartDataPoint{
					Label: category,
					Value: savings,
					Color: "", // Will be assigned by frontend
				})
			}
		}
	}

	// Timeline Gantt
	for _, req := range project.Requisitions {
		ganttItem := GanttItem{
			Name:      req.Name,
			StartDate: req.CreatedAt,
			Status:    "active", // ProjectRequisitions don't have status
		}

		// Estimate end date based on lead time (default 30 days)
		ganttItem.EndDate = req.CreatedAt.AddDate(0, 0, 30)

		chartsData.TimelineGantt = append(chartsData.TimelineGantt, ganttItem)
	}

	return chartsData, nil
}
