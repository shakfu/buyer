# Project-Level Procurement Comparisons and Savings Analysis

**Feature:** Multi-Level Procurement Intelligence System
**Status:** Design Proposal
**Date:** 2025-11-14
**Version:** 1.0

## Executive Summary

This document proposes a comprehensive project-level procurement comparison and savings analysis system that extends the existing specification-level quote comparisons to the project context. The system will analyze Bill of Materials (BOM) items across project requisitions, recommend optimal vendor selections, calculate total project savings, and provide a project-level dashboard for procurement decision-making.

## Problem Statement

### Current Capabilities

The buyer application currently supports:
- Specification-level quote comparisons via `QuoteService.CompareQuotesForSpecification()`
- Requisition-level quote comparisons via `RequisitionService.GetQuoteComparison()`
- Individual requisition savings analysis (budget vs. best quotes)
- Basic project structure with BOM and requisitions

### Gaps and Limitations

1. **No Project-Level Intelligence**: Comparisons happen at specification or requisition level, not holistically across the project
2. **No Cross-Requisition Optimization**: Cannot identify bulk purchasing opportunities across multiple requisitions
3. **No Vendor Consolidation Analysis**: Cannot assess cost/benefit of using fewer vendors
4. **Limited Savings Visibility**: Project-level savings and cost optimization opportunities are invisible
5. **No BOM-Driven Procurement**: BOM items are not directly linked to procurement recommendations
6. **Missing Strategic Insights**: No vendor performance analysis at project level
7. **No What-If Scenarios**: Cannot model different vendor selection strategies

## Proposed Solution Architecture

### Multi-Level Comparison Hierarchy

```
Project Level (NEW)
  |
  +-- BillOfMaterials (Master Requirements)
  |     |
  |     +-- BillOfMaterialsItem (Specifications + Quantities)
  |
  +-- ProjectRequisitions (Procurement Actions)
        |
        +-- ProjectRequisitionItem (Links to BOM Items)
              |
              +-- Specification Matching
                    |
                    +-- Product Quotes (Price Comparisons)
```

### Key Design Principles

1. **BOM-Centric Analysis**: All comparisons trace back to BOM requirements
2. **Hierarchical Aggregation**: Roll up savings from quote -> item -> requisition -> project
3. **Vendor Intelligence**: Analyze vendor performance, consolidation opportunities, and ratings
4. **Flexible Recommendations**: Support multiple optimization strategies (lowest cost, fewest vendors, highest quality)
5. **Real-Time Calculations**: Dynamic recalculation as quotes and requisitions change
6. **Actionable Insights**: Clear recommendations with measurable impact

## Data Model Extensions

### ProjectRequisitionItem Enhancement

The existing `ProjectRequisitionItem` model needs to track selected quotes and actual costs:

```go
type ProjectRequisitionItem struct {
    ID                    uint                 `gorm:"primaryKey"`
    ProjectRequisitionID  uint                 `gorm:"not null;index"`
    ProjectRequisition    *ProjectRequisition
    BillOfMaterialsItemID uint                 `gorm:"not null;index"`
    BOMItem               *BillOfMaterialsItem
    QuantityRequested     int                  `gorm:"not null"`

    // NEW: Procurement tracking fields
    SelectedQuoteID       *uint                `gorm:"index" json:"selected_quote_id,omitempty"`
    SelectedQuote         *Quote               `gorm:"foreignKey:SelectedQuoteID;constraint:OnDelete:SET NULL"`
    TargetUnitPrice       float64              `json:"target_unit_price,omitempty"` // Budget or target price
    ActualUnitPrice       float64              `json:"actual_unit_price,omitempty"` // Final negotiated price
    ProcurementStatus     string               `gorm:"size:20;default:'pending'" json:"procurement_status"` // pending, quoted, ordered, received

    Notes                 string               `gorm:"type:text"`
    CreatedAt             time.Time
    UpdatedAt             time.Time
}

// Valid procurement statuses: pending, quoted, ordered, received, cancelled
```

### New: ProjectProcurementStrategy

Stores user-selected optimization strategies for a project:

```go
type ProjectProcurementStrategy struct {
    ID                    uint      `gorm:"primaryKey"`
    ProjectID             uint      `gorm:"uniqueIndex;not null"`
    Project               *Project

    // Strategy settings
    Strategy              string    `gorm:"size:30;default:'lowest_cost'"` // lowest_cost, fewest_vendors, balanced, quality_focused
    MaxVendors            *int      `json:"max_vendors,omitempty"`         // Optional vendor limit
    MinVendorRating       *float64  `json:"min_vendor_rating,omitempty"`   // Minimum acceptable rating (1-5)
    PreferredVendorIDs    string    `gorm:"type:text" json:"preferred_vendor_ids,omitempty"` // Comma-separated IDs
    ExcludedVendorIDs     string    `gorm:"type:text" json:"excluded_vendor_ids,omitempty"`  // Comma-separated IDs
    AllowPartialFulfill   bool      `gorm:"default:true" json:"allow_partial_fulfill"`       // Allow splitting orders

    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

## Service Layer: ProjectProcurementService

### Core Service Structure

```go
type ProjectProcurementService struct {
    db              *gorm.DB
    quoteService    *QuoteService
    projectService  *ProjectService
    vendorService   *VendorService
}

func NewProjectProcurementService(
    db *gorm.DB,
    quoteService *QuoteService,
    projectService *ProjectService,
    vendorService *VendorService,
) *ProjectProcurementService {
    return &ProjectProcurementService{
        db:             db,
        quoteService:   quoteService,
        projectService: projectService,
        vendorService:  vendorService,
    }
}
```

### Analysis Data Structures

#### BOMItemProcurementAnalysis

Analysis for a single BOM item across all requisitions:

```go
type BOMItemProcurementAnalysis struct {
    BOMItem              *models.BillOfMaterialsItem
    Specification        *models.Specification
    TotalQuantityNeeded  int                          // From BOM

    // Requisition breakdown
    RequisitionItems     []ProjectRequisitionItemQuotes
    TotalQuantityPlanned int                          // Sum across requisitions
    CoveragePercent      float64                      // Planned / Needed * 100

    // Quote analysis
    AvailableQuotes      []models.Quote               // All active quotes for this spec
    BestQuote            *models.Quote                // Lowest price quote
    RecommendedQuote     *models.Quote                // Based on strategy

    // Cost calculations (all in USD)
    BestTotalCost        float64                      // Best quote * total quantity
    RecommendedTotalCost float64                      // Recommended quote * total quantity
    TargetTotalCost      float64                      // Target price * total quantity
    SavingsVsTarget      float64                      // Target - Recommended

    // Status
    HasSufficientQuotes  bool
    HasGaps              bool                         // Quantity planned < needed
    RiskLevel            string                       // low, medium, high
}

type ProjectRequisitionItemQuotes struct {
    RequisitionItem      *models.ProjectRequisitionItem
    RequisitionName      string
    Quantity             int
    TargetUnitPrice      float64
    BestQuote            *models.Quote
    SelectedQuote        *models.Quote
    Status               string
}
```

#### VendorConsolidationAnalysis

Analyzes vendor usage and consolidation opportunities:

```go
type VendorConsolidationAnalysis struct {
    VendorID             uint
    VendorName           string
    Rating               *VendorRatingsSummary        // Average ratings

    // Items this vendor can supply
    BOMItemsAvailable    []uint                       // BOM item IDs
    SpecificationsCount  int                          // Number of different specs
    TotalQuantity        int                          // Total units across all items

    // Cost analysis
    TotalCostIfUsed      float64                      // Total cost for all items from this vendor
    AveragePriceRank     float64                      // 1.0 = always cheapest, 2.0 = always 2nd, etc.

    // Logistics
    EstimatedOrderCount  int                          // Number of separate orders needed
    ShippingAdvantage    bool                         // Can consolidate shipping
}

type VendorRatingsSummary struct {
    VendorID        uint
    TotalRatings    int
    AvgPrice        *float64
    AvgQuality      *float64
    AvgDelivery     *float64
    AvgService      *float64
    OverallAvg      float64
}
```

#### ProjectProcurementComparison

Complete project-level analysis:

```go
type ProjectProcurementComparison struct {
    Project              *models.Project
    Strategy             *models.ProjectProcurementStrategy

    // BOM analysis
    BOMItemAnalyses      []BOMItemProcurementAnalysis
    TotalBOMItems        int
    FullyCoveredItems    int                          // 100% quote coverage
    PartiallyCoveredItems int                         // Some quotes available
    UncoveredItems       int                          // No quotes available

    // Vendor analysis
    VendorConsolidation  []VendorConsolidationAnalysis
    TotalVendorsNeeded   int                          // For recommended solution

    // Financial summary (all in USD)
    ProjectBudget        float64
    TotalTargetCost      float64                      // Sum of target costs
    BestCaseCost         float64                      // Always picking cheapest per item
    RecommendedCost      float64                      // Based on strategy
    WorstCaseCost        float64                      // Highest cost scenario

    // Savings analysis
    SavingsVsBudget      float64                      // Budget - Recommended
    SavingsVsTarget      float64                      // Target - Recommended
    SavingsPercent       float64                      // Savings / Budget * 100

    // Recommendations
    VendorRecommendations []VendorRecommendation
    RiskAssessment       ProjectRiskAssessment
    AlternativeScenarios []ProcurementScenario        // What-if analyses

    // Metadata
    AnalysisDate         time.Time
    QuoteFreshness       QuoteFreshnessStats
}

type VendorRecommendation struct {
    VendorID             uint
    VendorName           string
    BOMItems             []uint                       // Which BOM items to procure from this vendor
    TotalCost            float64
    ItemCount            int
    Rationale            string                       // Why this vendor for these items
    Priority             int                          // 1 = highest priority
}

type ProjectRiskAssessment struct {
    OverallRisk          string                       // low, medium, high, critical
    RiskFactors          []RiskFactor
    MitigationActions    []string
}

type RiskFactor struct {
    Category             string                       // quote_coverage, vendor_capacity, budget, timeline
    Severity             string                       // low, medium, high, critical
    Description          string
    AffectedBOMItems     []uint
    Impact               string                       // Potential consequences
}

type ProcurementScenario struct {
    Name                 string                       // "Lowest Cost", "Fewest Vendors", "Quality Focus"
    Description          string
    VendorCount          int
    TotalCost            float64
    SavingsVsBudget      float64
    Tradeoffs            string                       // What you gain/lose with this approach
    VendorAssignments    map[uint][]uint              // VendorID -> BOM Item IDs
}

type QuoteFreshnessStats struct {
    TotalQuotes          int
    FreshQuotes          int                          // < 30 days old
    StaleQuotes          int                          // 30-90 days old
    ExpiredQuotes        int                          // Past valid_until date
    AverageAgeDays       int
}
```

### Key Service Methods

#### 1. GetProjectProcurementComparison

Main entry point for comprehensive project analysis:

```go
func (s *ProjectProcurementService) GetProjectProcurementComparison(
    projectID uint,
) (*ProjectProcurementComparison, error) {
    // 1. Load project with BOM and requisitions
    // 2. For each BOM item:
    //    - Find all matching requisition items
    //    - Get all active quotes for the specification
    //    - Perform BOM item analysis
    // 3. Analyze vendor consolidation opportunities
    // 4. Calculate financial summaries
    // 5. Generate recommendations based on strategy
    // 6. Assess risks
    // 7. Generate alternative scenarios
    // 8. Return comprehensive comparison
}
```

#### 2. AnalyzeBOMItem

Detailed analysis for a specific BOM item:

```go
func (s *ProjectProcurementService) AnalyzeBOMItem(
    bomItemID uint,
) (*BOMItemProcurementAnalysis, error) {
    // 1. Get BOM item with specification
    // 2. Find all project requisition items referencing this BOM item
    // 3. Get all active quotes for this specification
    // 4. Calculate coverage, costs, and savings
    // 5. Assess risk based on quote availability and freshness
    // 6. Return analysis
}
```

#### 3. GetVendorConsolidationAnalysis

Analyze opportunities to consolidate vendors:

```go
func (s *ProjectProcurementService) GetVendorConsolidationAnalysis(
    projectID uint,
) ([]VendorConsolidationAnalysis, error) {
    // 1. Get all BOM items for project
    // 2. For each vendor that has quotes for project specifications:
    //    - Identify which BOM items they can supply
    //    - Calculate total cost if using this vendor for those items
    //    - Get vendor ratings
    //    - Assess shipping/logistics advantages
    // 3. Rank vendors by coverage and cost-effectiveness
    // 4. Return analysis sorted by best consolidation opportunity
}
```

#### 4. GenerateVendorRecommendations

Generate optimized vendor assignment recommendations:

```go
func (s *ProjectProcurementService) GenerateVendorRecommendations(
    projectID uint,
    strategy string, // "lowest_cost", "fewest_vendors", "balanced", "quality_focused"
) ([]VendorRecommendation, error) {
    // 1. Get BOM items and available quotes
    // 2. Apply strategy-specific algorithm:
    //    - lowest_cost: Pick cheapest vendor for each item independently
    //    - fewest_vendors: Minimize vendor count (bin packing problem)
    //    - balanced: Optimize cost vs vendor count (weighted scoring)
    //    - quality_focused: Prioritize highly-rated vendors within budget
    // 3. Generate vendor assignments with rationale
    // 4. Calculate total costs and savings
    // 5. Return prioritized recommendations
}
```

#### 5. CompareScenarios

Generate and compare multiple procurement scenarios:

```go
func (s *ProjectProcurementService) CompareScenarios(
    projectID uint,
) ([]ProcurementScenario, error) {
    // Generate scenarios for:
    // 1. Lowest Cost (always pick cheapest)
    // 2. Fewest Vendors (minimize vendor count)
    // 3. Balanced (optimize cost vs. vendor count)
    // 4. Quality Focus (prefer high-rated vendors)
    // 5. Preferred Vendors Only (if configured)
    // Return all scenarios with cost, vendor count, and tradeoff analysis
}
```

#### 6. AssessProjectRisks

Identify and categorize procurement risks:

```go
func (s *ProjectProcurementService) AssessProjectRisks(
    projectID uint,
) (*ProjectRiskAssessment, error) {
    // Assess risks in categories:
    // 1. Quote Coverage: Missing or insufficient quotes
    // 2. Budget: Estimated costs exceeding budget
    // 3. Timeline: Expired/expiring quotes, long lead times
    // 4. Vendor Capacity: Single-source dependencies
    // 5. Quality: Low-rated vendors in critical path
    // Calculate overall risk level and suggest mitigations
}
```

#### 7. CalculateProjectSavings

Comprehensive savings calculation:

```go
func (s *ProjectProcurementService) CalculateProjectSavings(
    projectID uint,
) (*ProjectSavingsSummary, error) {
    // Calculate savings across multiple dimensions:
    // 1. Budget vs. Recommended Cost
    // 2. Target Prices vs. Actual Quotes
    // 3. Worst Case vs. Best Case
    // 4. Vendor Consolidation Savings (shipping, admin overhead)
    // Break down by BOM item category/specification
}

type ProjectSavingsSummary struct {
    TotalSavingsUSD      float64
    SavingsPercent       float64
    SavingsByCategory    map[string]float64           // Specification -> Savings
    SavingsByVendor      map[string]float64           // Vendor -> Contribution to savings
    ConsolidationSavings float64                      // Estimated from vendor reduction
    DetailedBreakdown    []SavingsLineItem
}

type SavingsLineItem struct {
    BOMItemID            uint
    SpecificationName    string
    Quantity             int
    TargetPrice          float64
    RecommendedPrice     float64
    SavingsPerUnit       float64
    TotalSavings         float64
    SavingsPercent       float64
}
```

## Project Dashboard Enhancements

### Dashboard Structure

```go
type ProjectDashboard struct {
    // Basic info
    Project              *models.Project

    // Progress metrics
    Progress             ProjectProgress

    // Financial overview
    Financial            ProjectFinancialOverview

    // Procurement status
    Procurement          ProjectProcurementStatus

    // Vendor performance
    VendorPerformance    []VendorPerformanceSummary

    // Risk indicators
    Risks                []RiskIndicator

    // Recent activity
    RecentActivity       []ActivityItem

    // Charts data
    ChartsData           ProjectChartsData
}

type ProjectProgress struct {
    BOMCoverage          float64                      // % of BOM items with quotes
    RequisitionsComplete int
    RequisitionsTotal    int
    OrdersPlaced         int
    OrdersReceived       int
    TimelineStatus       string                       // on_track, at_risk, delayed
    DaysToDeadline       int
}

type ProjectFinancialOverview struct {
    Budget               float64
    Committed            float64                      // Orders placed
    Estimated            float64                      // Best quotes for remaining items
    Remaining            float64                      // Budget - (Committed + Estimated)
    Savings              float64
    SavingsPercent       float64
    BudgetHealth         string                       // healthy, warning, critical
}

type ProjectProcurementStatus struct {
    TotalItems           int
    ItemsWithQuotes      int
    ItemsOrdered         int
    ItemsReceived        int
    AverageLeadTimeDays  int
    VendorsEngaged       int
    QuoteFreshness       string                       // fresh, aging, stale
}

type VendorPerformanceSummary struct {
    VendorID             uint
    VendorName           string
    ItemsSupplied        int
    TotalValue           float64
    AverageRating        float64
    OnTimeDelivery       float64                      // Percentage
    Status               string                       // active, pending, completed
}

type RiskIndicator struct {
    Category             string
    Level                string                       // low, medium, high, critical
    Count                int                          // Number of issues
    TopIssue             string                       // Most critical issue description
}

type ActivityItem struct {
    Timestamp            time.Time
    Type                 string                       // quote_added, order_placed, item_received, etc.
    Description          string
    Impact               string                       // Positive, Negative, Neutral
}

type ProjectChartsData struct {
    BudgetUtilization    []ChartDataPoint             // Time series of budget allocation
    SavingsByCategory    []ChartDataPoint             // Pie chart: savings per specification category
    VendorDistribution   []ChartDataPoint             // Pie chart: spending per vendor
    TimelineGantt        []GanttItem                  // Gantt chart of requisitions
    CostComparison       []ChartDataPoint             // Bar chart: Budget vs. Estimated vs. Committed
}

type ChartDataPoint struct {
    Label                string
    Value                float64
    Color                string
    Metadata             map[string]interface{}
}

type GanttItem struct {
    Name                 string
    StartDate            time.Time
    EndDate              time.Time
    Status               string
    DependsOn            []uint
}
```

### Dashboard Service Method

```go
func (s *ProjectProcurementService) GetProjectDashboard(
    projectID uint,
) (*ProjectDashboard, error) {
    // 1. Load project data
    // 2. Calculate progress metrics
    // 3. Compute financial overview
    // 4. Assess procurement status
    // 5. Aggregate vendor performance
    // 6. Identify risk indicators
    // 7. Fetch recent activity
    // 8. Generate chart data
    // 9. Return dashboard
}
```

## CLI Commands

### Procurement Analysis Commands

```bash
# Complete project procurement analysis
buyer project procurement-analysis <project-id> [--strategy lowest_cost|fewest_vendors|balanced|quality_focused]

# Analyze specific BOM item
buyer project bom-item-analysis <bom-item-id>

# Vendor consolidation opportunities
buyer project vendor-consolidation <project-id> [--max-vendors N]

# Compare procurement scenarios
buyer project compare-scenarios <project-id>

# Project savings summary
buyer project savings <project-id> [--detailed]

# Risk assessment
buyer project risks <project-id>

# Project dashboard (text version)
buyer project dashboard <project-id>
```

### Strategy Management

```bash
# Set project procurement strategy
buyer project set-strategy <project-id> --strategy lowest_cost [--max-vendors N] [--min-rating 4.0]

# Add preferred vendors
buyer project add-preferred-vendor <project-id> <vendor-id>

# Exclude vendors
buyer project exclude-vendor <project-id> <vendor-id>
```

### Recommendation Application

```bash
# Apply recommendations to requisition items
buyer project apply-recommendations <project-id> [--dry-run]

# Select quote for requisition item
buyer project select-quote <requisition-item-id> <quote-id>

# Set target prices from BOM
buyer project sync-targets <project-id>
```

## Web Interface

### Project Procurement Page

New page: `/projects/{id}/procurement`

**Sections:**

1. **Overview Panel**
   - Budget health indicator (green/yellow/red)
   - Total savings amount and percentage
   - Vendor count
   - Quote coverage percentage
   - Risk level indicator

2. **BOM Analysis Table**
   - Columns: Specification | Required Qty | Planned Qty | Best Quote | Recommended Quote | Savings | Coverage | Risk
   - Sortable and filterable
   - Drill-down to item details
   - Color coding for risk levels

3. **Vendor Consolidation Panel**
   - List of vendors with consolidation score
   - What items each vendor can supply
   - Total cost if using this vendor
   - Rating indicators
   - "Apply Recommendation" button

4. **Scenario Comparison**
   - Side-by-side comparison table
   - Columns: Scenario | Vendors | Total Cost | Savings | Trade-offs
   - Visual indicators (bar charts)
   - Select and apply scenario

5. **Savings Breakdown**
   - Pie chart: Savings by specification category
   - Bar chart: Budget vs. Estimated vs. Committed
   - Table: Top savings opportunities
   - Export to CSV/Excel

6. **Risk Dashboard**
   - Risk category cards with counts
   - Critical issues list
   - Mitigation recommendations
   - Timeline impact warnings

### Project Dashboard Page

Enhanced page: `/projects/{id}/dashboard`

**New Visualizations:**

1. **Budget Utilization Chart**
   - Stacked area chart showing committed vs. estimated vs. remaining
   - Timeline on X-axis
   - Threshold lines for budget limits

2. **Vendor Distribution**
   - Pie chart: Spending distribution across vendors
   - Hover shows vendor ratings

3. **Procurement Timeline**
   - Gantt chart of requisitions
   - Color-coded by status
   - Deadline indicators

4. **Savings Trend**
   - Line chart showing projected savings over time
   - As more quotes are added

5. **Risk Heat Map**
   - Matrix: BOM Items (Y) × Risk Categories (X)
   - Color intensity = severity

## Implementation Plan

### Phase 1: Core Data Models and Service (Week 1-2)

1. Add `ProjectProcurementStrategy` model
2. Extend `ProjectRequisitionItem` with procurement fields
3. Create `ProjectProcurementService` with basic structure
4. Implement `GetProjectProcurementComparison()`
5. Implement `AnalyzeBOMItem()`
6. Write comprehensive tests

### Phase 2: Vendor Analysis and Recommendations (Week 3)

1. Implement `GetVendorConsolidationAnalysis()`
2. Implement `GenerateVendorRecommendations()` with all strategies
3. Implement `CompareScenarios()`
4. Add vendor rating aggregation methods
5. Write tests for recommendation logic

### Phase 3: Savings and Risk Analysis (Week 4)

1. Implement `CalculateProjectSavings()`
2. Implement `AssessProjectRisks()`
3. Add quote freshness tracking
4. Implement risk mitigation suggestions
5. Write tests for savings calculations

### Phase 4: Dashboard and Reporting (Week 5)

1. Implement `GetProjectDashboard()`
2. Add chart data generation methods
3. Create dashboard templates and handlers
4. Add procurement analysis page
5. Implement data export (CSV/Excel)

### Phase 5: CLI Commands (Week 6)

1. Add procurement analysis commands
2. Add strategy management commands
3. Add recommendation application commands
4. Format output tables
5. Write CLI tests

### Phase 6: Web Interface (Week 7-8)

1. Create procurement analysis page HTML/CSS/JS
2. Add vendor consolidation UI
3. Add scenario comparison interface
4. Enhance project dashboard with new charts
5. Add interactive recommendation application
6. Implement real-time updates

### Phase 7: Testing and Optimization (Week 9)

1. Integration testing across all components
2. Performance optimization for large projects
3. Load testing with realistic data
4. Edge case handling
5. Documentation updates

## Algorithm Details

### Vendor Consolidation Optimizer (Fewest Vendors Strategy)

This is a variant of the bin packing problem with cost constraints:

```
GOAL: Minimize number of vendors while staying within budget

INPUT:
  - BOM items with quantities and specifications
  - Quotes matrix: BOMItem × Vendor → Price
  - Budget constraint
  - Optional: Max vendors, vendor ratings threshold

ALGORITHM (Greedy with Backtracking):
  1. For each vendor, calculate:
     - Coverage score: How many BOM items they can supply
     - Cost score: Average price rank across their items
     - Rating score: Average vendor rating

  2. Sort vendors by composite score:
     Score = w1×Coverage + w2×(1/Cost) + w3×Rating

  3. Greedy assignment:
     - Start with highest-scoring vendor
     - Assign all BOM items they can supply at competitive price
     - Mark items as assigned
     - Move to next vendor
     - Repeat until all items assigned or budget exceeded

  4. Optimization pass:
     - For each item, check if reassigning to existing vendor saves cost
     - If yes and within budget, reassign

  5. Validate budget constraint
     - If over budget, apply cost reduction strategies:
       a) Replace expensive items with cheaper quotes
       b) Remove lowest-priority vendor if alternatives exist

OUTPUT:
  - Vendor assignments: Map[VendorID][]BOMItemID
  - Total cost
  - Savings vs. best-case
```

### Balanced Strategy Algorithm

Optimizes both cost and vendor count using weighted scoring:

```
GOAL: Minimize weighted sum of cost and vendor count

ALGORITHM:
  1. Calculate best-case cost (always cheapest) = C_min
  2. Calculate fewest-vendors cost = C_fewest
  3. Calculate vendor range: V_min to V_max

  4. For each vendor count V from V_min to V_max:
     - Find lowest-cost assignment using exactly V vendors
     - Calculate score = α×(Cost/C_min) + β×(V/V_min)
       where α + β = 1 (typically α=0.6, β=0.4)

  5. Select vendor count with best score
  6. Return corresponding assignment

EFFICIENCY: O(V² × B × log(B)) where V=vendors, B=BOM items
```

## Performance Considerations

### Caching Strategy

```go
type ProcurementAnalysisCache struct {
    ProjectID       uint
    AnalysisData    *ProjectProcurementComparison
    GeneratedAt     time.Time
    ExpiresAt       time.Time
    QuoteDigest     string                          // Hash of quote IDs and prices
}

// Invalidate cache when:
// - New quote added for project specifications
// - Quote price updated
// - BOM modified
// - Requisition items changed
// - Strategy updated
```

### Query Optimization

1. **Eager Loading**: Preload all relationships in single query
2. **Materialized Views**: For vendor rating aggregations
3. **Indexed Queries**: Add composite indexes on frequently-joined columns
4. **Batch Processing**: Process BOM items in batches for large projects
5. **Parallel Analysis**: Analyze BOM items concurrently using goroutines

```go
// Example: Parallel BOM analysis
func (s *ProjectProcurementService) analyzeBOMItemsParallel(
    bomItems []models.BillOfMaterialsItem,
) ([]BOMItemProcurementAnalysis, error) {
    resultChan := make(chan BOMItemProcurementAnalysis, len(bomItems))
    errorChan := make(chan error, len(bomItems))

    for _, item := range bomItems {
        go func(bomItem models.BillOfMaterialsItem) {
            analysis, err := s.AnalyzeBOMItem(bomItem.ID)
            if err != nil {
                errorChan <- err
                return
            }
            resultChan <- *analysis
        }(item)
    }

    // Collect results...
}
```

### Database Indexes

```sql
-- Project procurement queries
CREATE INDEX idx_proj_req_items_status ON project_requisition_items(procurement_status);
CREATE INDEX idx_proj_req_items_quote ON project_requisition_items(selected_quote_id);

-- Vendor ratings aggregation
CREATE INDEX idx_vendor_ratings_vendor ON vendor_ratings(vendor_id, created_at);

-- Quote freshness queries
CREATE INDEX idx_quotes_spec_date ON quotes(product_id, quote_date DESC);
```

## Testing Strategy

### Unit Tests

1. Test each analysis method with mock data
2. Test recommendation algorithms with known inputs
3. Test savings calculations with edge cases
4. Test risk assessment logic
5. Test scenario generation

### Integration Tests

1. End-to-end procurement analysis workflow
2. Dashboard generation with realistic project data
3. Recommendation application and validation
4. Multi-strategy comparison
5. Vendor consolidation with constraints

### Performance Tests

1. Large project (500+ BOM items) analysis time < 5s
2. Dashboard generation for 100-item project < 2s
3. Scenario comparison < 3s
4. Concurrent analysis requests handling

### Edge Cases

1. Project with no quotes
2. Project over budget
3. Single vendor for all items
4. All vendors excluded/none preferred
5. Expired quotes only
6. Requisitions exceeding BOM quantities
7. BOM items with no matching requisitions

## Benefits and Impact

### For Procurement Teams

1. **Strategic Visibility**: See entire project procurement landscape at a glance
2. **Data-Driven Decisions**: Recommendations based on comprehensive analysis
3. **Time Savings**: Automated vendor consolidation reduces manual analysis
4. **Risk Mitigation**: Early identification of procurement risks
5. **Cost Optimization**: Identify and realize savings opportunities
6. **Vendor Management**: Better vendor selection and relationship management

### For Finance Teams

1. **Budget Control**: Real-time tracking of budget utilization
2. **Savings Tracking**: Measurable savings documentation
3. **Forecasting**: Accurate cost projections
4. **Audit Trail**: Complete procurement decision documentation

### For Project Managers

1. **Timeline Confidence**: Procurement risks visible early
2. **Status Clarity**: Clear progress indicators
3. **Stakeholder Reporting**: Executive-ready dashboards
4. **Scenario Planning**: What-if analysis for decision making

## Future Enhancements

### Phase 2 Features

1. **Machine Learning Recommendations**
   - Learn from past procurement decisions
   - Predict optimal vendor selection
   - Forecast price trends

2. **Automated Negotiations**
   - Generate RFQ documents
   - Track vendor responses
   - Compare against historical prices

3. **Supplier Diversity Tracking**
   - Monitor diversity goals
   - Recommend minority/women-owned vendors
   - Report on diversity metrics

4. **Carbon Footprint Analysis**
   - Track shipping distances
   - Calculate environmental impact
   - Recommend greener alternatives

5. **Integration with Procurement Systems**
   - Export to ERP systems
   - Import POs from external systems
   - Sync with accounting software

6. **Advanced Analytics**
   - Predictive lead time analysis
   - Seasonal pricing trends
   - Vendor capacity forecasting

## Conclusion

This project-level procurement system transforms the buyer application from a quote management tool into a comprehensive procurement intelligence platform. By analyzing Bill of Materials in the context of project requisitions and vendor capabilities, it enables:

- **Holistic optimization** across the entire project lifecycle
- **Strategic vendor management** with consolidation and performance tracking
- **Measurable cost savings** through data-driven recommendations
- **Risk-aware procurement** with early issue identification
- **Flexible decision-making** through scenario comparison

The phased implementation approach ensures incremental value delivery while maintaining system stability and test coverage.
