# Project & Bill of Materials (BOM) Design

**Feature:** Project-level organization with Bill of Materials and sequenced Requisitions
**Status:** Design Proposal
**Date:** 2025-11-12

## Overview

This design adds a higher-level organizational structure to buyer by introducing **Projects** with associated **Bills of Materials (BOMs)**. Projects provide budget tracking, deadline management, and the ability to sequence multiple Requisitions according to project requirements.

## Problem Statement

Currently, the buyer application manages:
- Individual Requisitions (purchasing requirements)
- RequisitionItems (line items with specifications and quantities)
- Quote comparisons per Requisition

**Missing capabilities:**
1. No way to group related Requisitions under a common project
2. No project-level budget tracking and deadline management
3. No ability to sequence Requisitions (e.g., Phase 1, Phase 2, etc.)
4. No project-level reporting and analytics
5. No rollup of costs across multiple Requisitions

## Proposed Solution

### Data Model

```
Project (new)
├── BillOfMaterials (new)
│   └── BillOfMaterialsItems (new) - links to Specifications with quantities
└── ProjectRequisitions (new) - sequenced Requisitions
    └── Requisition (existing)
        └── RequisitionItems (existing)
```

The final architecture is:

```text
Project (1) ←→ (1) BillOfMaterials
                    ↓
               (1) ←→ (many) BillOfMaterialsItem
                                  ↓
                             (many) → (1) Specification

Project (1) ←→ (many) ProjectRequisition
                         ↓
                    (1) → (1) Requisition
                               ↓
                          (1) ←→ (many) RequisitionItem
                                           ↓
                                      (many) → (1) Specification
```

### Key relationships

- `Project` has exactly one `BillOfMaterials` (one-to-one)

- `BillOfMaterials` has many `BillOfMaterialsItems` (one-to-many)

- Each `BillOfMaterialsItem` references one `Specification` (many-to-one)

- `Specification` can only appear once per `BillOfMaterials` (unique constraint)


### Model Definitions

#### Project Model
```go
type Project struct {
    ID              uint                  `gorm:"primaryKey" json:"id"`
    Name            string                `gorm:"uniqueIndex;not null" json:"name"`
    Description     string                `gorm:"type:text" json:"description,omitempty"`
    Budget          float64               `json:"budget,omitempty"`          // Overall project budget
    Deadline        *time.Time            `json:"deadline,omitempty"`        // Project deadline
    Status          string                `gorm:"size:20;default:'planning'" json:"status"` // planning, active, completed, cancelled
    BillOfMaterials *BillOfMaterials      `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
    Requisitions    []ProjectRequisition  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"requisitions,omitempty"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}
```

**Status values:**
- `planning` - Initial state, BOM being defined
- `active` - Requisitions being processed
- `completed` - All requisitions fulfilled
- `cancelled` - Project cancelled

#### BillOfMaterials Model
```go
type BillOfMaterials struct {
    ID          uint                   `gorm:"primaryKey" json:"id"`
    ProjectID   uint                   `gorm:"uniqueIndex;not null" json:"project_id"` // One BillOfMaterials per project
    Project     *Project               `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
    Notes       string                 `gorm:"type:text" json:"notes,omitempty"`
    Items       []BillOfMaterialsItem  `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// TableName overrides the default table name
func (BillOfMaterials) TableName() string { return "bills_of_materials" }
```

#### BillOfMaterialsItem Model
```go
type BillOfMaterialsItem struct {
    ID                  uint              `gorm:"primaryKey" json:"id"`
    BillOfMaterialsID   uint              `gorm:"not null;index" json:"bill_of_materials_id"`
    BillOfMaterials     *BillOfMaterials  `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
    SpecificationID     uint              `gorm:"not null;index" json:"specification_id"`
    Specification       *Specification    `gorm:"foreignKey:SpecificationID;constraint:OnDelete:RESTRICT" json:"specification,omitempty"`
    Quantity            int               `gorm:"not null" json:"quantity"`
    Notes               string            `gorm:"type:text" json:"notes,omitempty"`
    CreatedAt           time.Time         `json:"created_at"`
    UpdatedAt           time.Time         `json:"updated_at"`
}

// TableName overrides the default table name
func (BillOfMaterialsItem) TableName() string { return "bill_of_materials_items" }

// Unique constraint: One specification can only appear once in a BillOfMaterials
// Composite index on (bill_of_materials_id, specification_id) with UNIQUE constraint
```

#### ProjectRequisition Model (Join Table)
```go
type ProjectRequisition struct {
    ID             uint         `gorm:"primaryKey" json:"id"`
    ProjectID      uint         `gorm:"not null;index:idx_project_seq" json:"project_id"`
    Project        *Project     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
    RequisitionID  uint         `gorm:"uniqueIndex;not null" json:"requisition_id"` // One requisition belongs to one project
    Requisition    *Requisition `gorm:"foreignKey:RequisitionID;constraint:OnDelete:CASCADE" json:"requisition,omitempty"`
    Sequence       int          `gorm:"not null;index:idx_project_seq" json:"sequence"` // Order within project
    Phase          string       `gorm:"size:100" json:"phase,omitempty"` // e.g., "Phase 1", "Initial Setup", "Expansion"
    CreatedAt      time.Time    `json:"created_at"`
    UpdatedAt      time.Time    `json:"updated_at"`
}

// Composite index on (project_id, sequence) for ordering
```

### Relationships

1. **Project ↔ BillOfMaterials**: One-to-One
   - Each Project has exactly one BillOfMaterials
   - BillOfMaterials CASCADE deletes with Project

2. **BillOfMaterials ↔ BillOfMaterialsItems**: One-to-Many
   - BillOfMaterials can have multiple items
   - BillOfMaterialsItems CASCADE delete with BillOfMaterials

3. **BillOfMaterialsItem → Specification**: Many-to-One
   - Each BillOfMaterialsItem references a Specification
   - RESTRICT delete (can't delete Specification if used in BillOfMaterials)

4. **Project ↔ ProjectRequisition**: One-to-Many
   - Project can have multiple Requisitions
   - ProjectRequisition CASCADE deletes with Project

5. **Requisition ↔ ProjectRequisition**: One-to-One
   - Each Requisition can belong to at most one Project
   - ProjectRequisition CASCADE deletes when Requisition deleted

### Key Design Decisions

#### 1. BillOfMaterials vs Requisition
- **BillOfMaterials**: Master list of ALL specifications needed for the project (the "what")
- **Requisition**: Purchasing action for SOME specifications at a specific time (the "when/how")
- BillOfMaterials items may be split across multiple Requisitions
- Requisitions may contain items not in the BillOfMaterials (ad-hoc purchases)

#### 2. Specification vs Product
- BillOfMaterials uses **Specifications** (generic product types), not Products (brand-specific)
- This allows flexibility in vendor selection during requisition
- Example: BillOfMaterials says "Need 10 laptops with Intel i7", Requisition selects "MacBook Pro 16" or "Dell XPS 15"

#### 3. Sequence and Phases
- Requisitions are ordered within a Project (1, 2, 3...)
- Optional Phase labels for logical grouping
- Enables timeline planning and milestone tracking

## Service Layer

### ProjectService

```go
type ProjectService struct {
    db *gorm.DB
}

// Core CRUD
func (s *ProjectService) Create(name, description string, budget float64, deadline *time.Time) (*models.Project, error)
func (s *ProjectService) GetByID(id uint) (*models.Project, error)
func (s *ProjectService) Update(id uint, name, description string, budget float64, deadline *time.Time, status string) (*models.Project, error)
func (s *ProjectService) Delete(id uint) error
func (s *ProjectService) List(limit, offset int) ([]models.Project, error)

// BillOfMaterials management
func (s *ProjectService) AddBillOfMaterialsItem(projectID, specificationID uint, quantity int, notes string) (*models.BillOfMaterialsItem, error)
func (s *ProjectService) UpdateBillOfMaterialsItem(itemID uint, quantity int, notes string) (*models.BillOfMaterialsItem, error)
func (s *ProjectService) DeleteBillOfMaterialsItem(itemID uint) error

// Requisition management
func (s *ProjectService) AddRequisition(projectID, requisitionID uint, sequence int, phase string) (*models.ProjectRequisition, error)
func (s *ProjectService) UpdateRequisitionSequence(projectReqID uint, sequence int, phase string) error
func (s *ProjectService) RemoveRequisition(projectReqID uint) error
func (s *ProjectService) ReorderRequisitions(projectID uint, sequences map[uint]int) error

// Analytics and reporting
func (s *ProjectService) GetProjectSummary(id uint) (*ProjectSummary, error)
func (s *ProjectService) GetBudgetAnalysis(id uint) (*BudgetAnalysis, error)
func (s *ProjectService) GetBillOfMaterialsCoverage(id uint) (*BillOfMaterialsCoverage, error)
func (s *ProjectService) GetCostProjection(id uint) (*CostProjection, error)
```

### Analytics Structures

#### ProjectSummary
```go
type ProjectSummary struct {
    Project                     *models.Project
    TotalRequisitions           int
    CompletedRequisitions       int
    TotalBillOfMaterialsItems   int
    TotalSpecsInReqs            int
    CoveredSpecs                []uint // Spec IDs in both BillOfMaterials and Requisitions
    MissingSpecs                []uint // Spec IDs in BillOfMaterials but not in Requisitions
    ExtraSpecs                  []uint // Spec IDs in Requisitions but not in BillOfMaterials
    BudgetUsed                  float64
    BudgetRemaining             float64
    DaysUntilDeadline           int
}
```

#### BudgetAnalysis
```go
type BudgetAnalysis struct {
    ProjectBudget       float64
    TotalRequisitionBudget float64  // Sum of all requisition budgets
    EstimatedCostLow    float64     // Based on lowest quotes
    EstimatedCostHigh   float64     // Based on highest quotes
    EstimatedCostAvg    float64     // Based on average quotes
    BudgetUtilization   float64     // Percentage of budget allocated
    CostOverrun         bool        // true if estimated > budget
    CostOverrunAmount   float64
}
```

#### BillOfMaterialsCoverage
```go
type BillOfMaterialsCoverageItem struct {
    SpecificationID   uint
    SpecificationName string
    RequiredQuantity  int    // From BillOfMaterials
    PlannedQuantity   int    // From Requisitions
    CoveragePercent   float64
    Status           string  // "covered", "partial", "missing"
}

type BillOfMaterialsCoverage struct {
    Items            []BillOfMaterialsCoverageItem
    TotalCoverage    float64  // Overall percentage
    FullyCovered     int      // Count of specs 100% covered
    PartiallyCovered int      // Count of specs partially covered
    Missing          int      // Count of specs with 0% coverage
}
```

#### CostProjection
```go
type CostProjectionItem struct {
    SpecificationName string
    Quantity         int
    BestQuotePrice   float64
    BestVendor       string
    TotalCost        float64
}

type CostProjection struct {
    Items              []CostProjectionItem
    TotalProjectedCost float64
    ProjectBudget      float64
    AvailableMargin    float64
    RecommendedVendors map[string]float64 // Vendor -> Total cost if using them
}
```

## CLI Commands

### Project Management
```bash
# Create project
buyer add project "Office Renovation" --budget 50000 --deadline "2025-12-31" --description "Complete office upgrade"

# List projects
buyer list projects [--status active] [--limit N] [--offset N]

# Get project details
buyer show project <id>

# Update project
buyer update project <id> --name "..." --budget X --deadline "..." --status active

# Delete project
buyer delete project <id> [-f|--force]
```

### Bill of Materials Management
```bash
# Add Bill of Materials item to project
buyer add bom-item --project <id> --spec <spec-id> --quantity 10 --notes "High priority"

# List Bill of Materials items
buyer list bom --project <id>

# Update Bill of Materials item
buyer update bom-item <item-id> --quantity 15 --notes "Updated"

# Delete Bill of Materials item
buyer delete bom-item <item-id> [-f|--force]
```

### Requisition Linking
```bash
# Add requisition to project
buyer add project-requisition --project <id> --requisition <req-id> --sequence 1 --phase "Phase 1"

# Reorder requisitions
buyer reorder project-requisitions <project-id> --sequences "1:1,2:3,3:2"  # req_id:new_sequence

# Remove requisition from project
buyer remove project-requisition <project-req-id>
```

### Analytics
```bash
# Project summary
buyer project summary <id>

# Budget analysis
buyer project budget <id>

# Bill of Materials coverage analysis
buyer project coverage <id>

# Cost projection
buyer project cost-projection <id>
```

## Web Interface

### Project Dashboard
- List all projects with status indicators
- Quick stats: Total budget, deadline, progress
- Color coding: Green (on track), Yellow (at risk), Red (overdue/over budget)

### Project Detail View
- **Overview Tab**: Name, description, budget, deadline, status
- **Bill of Materials Tab**: List of specifications with quantities
- **Requisitions Tab**: Sequenced list of requisitions with phase labels
- **Analytics Tab**:
  - Budget utilization chart
  - Bill of Materials coverage visualization
  - Cost projection table
  - Vendor recommendation

### Bill of Materials Editor
- Add/edit/delete Bill of Materials items
- Inline quantity editing
- Search specifications
- Import from existing requisition

### Requisition Sequencer
- Drag-and-drop reordering
- Phase grouping
- Visual timeline

## Database Migrations

### Migration Steps
1. Create `projects` table
2. Create `bills_of_materials` table with FK to projects
3. Create `bill_of_materials_items` table with FK to bills_of_materials and specifications
4. Create `project_requisitions` join table
5. Add composite indexes for performance
6. Add unique constraints

### Foreign Key Constraints
- Project → BillOfMaterials: CASCADE on delete
- BillOfMaterials → BillOfMaterialsItem: CASCADE on delete
- BillOfMaterialsItem → Specification: RESTRICT on delete
- Project → ProjectRequisition: CASCADE on delete
- ProjectRequisition → Requisition: CASCADE on delete

## Implementation Plan

### Phase 1: Core Models and Services (Week 1)
1. Add models to `internal/models/models.go`
2. Create `internal/services/project.go`
3. Write comprehensive tests in `internal/services/project_test.go`
4. Update migrations in `cmd/buyer/main.go`

### Phase 2: CLI Commands (Week 2)
1. Add project commands to `cmd/buyer/add.go`, `list.go`, `update.go`, `delete.go`
2. Add Bill of Materials management commands
3. Add requisition linking commands
4. Add analytics commands

### Phase 3: Web Interface (Week 3)
1. Add project list/detail handlers to `cmd/buyer/web.go`
2. Create HTML templates for project views
3. Add Bill of Materials editor interface
4. Add requisition sequencer

### Phase 4: Analytics and Reporting (Week 4)
1. Implement budget analysis service methods
2. Implement Bill of Materials coverage calculations
3. Implement cost projection
4. Create web dashboard with charts

### Phase 5: Testing and Documentation (Week 5)
1. Integration testing
2. Update README.md with Project/BOM examples
3. Update CLAUDE.md with new architecture
4. Update CODE_REVIEW.md

## Benefits

### For Users
1. **Project-level visibility**: Track multiple related requisitions together
2. **Budget control**: Monitor spending across entire project lifecycle
3. **Bill of Materials validation**: Ensure all required items are covered by requisitions
4. **Timeline management**: Sequence requisitions according to project phases
5. **Cost optimization**: Compare vendor costs at project level

### For System
1. **Better organization**: Hierarchical structure matches real-world workflows
2. **Enhanced reporting**: Project-level analytics provide strategic insights
3. **Scalability**: Can manage large projects with many requisitions
4. **Flexibility**: BillOfMaterials and Requisitions are loosely coupled

## Migration Strategy for Existing Data

### Option 1: Leave Existing Requisitions Standalone
- Existing requisitions continue to work independently
- Users can optionally assign them to projects later
- `ProjectRequisition.RequisitionID` allows NULL or uses soft linking

### Option 2: Create "Uncategorized" Project
- Automatically create a project called "Uncategorized"
- Assign all existing requisitions to it with sequence = creation order
- Users can move them to proper projects later

**Recommendation**: Option 1 (backwards compatible, no forced migration)

## Performance Considerations

### Indexing Strategy
```sql
-- Project lookups
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_deadline ON projects(deadline);

-- Bill of Materials item lookups
CREATE INDEX idx_bom_items_spec ON bill_of_materials_items(specification_id);
CREATE UNIQUE INDEX idx_bom_spec_unique ON bill_of_materials_items(bill_of_materials_id, specification_id);

-- Project requisition ordering
CREATE INDEX idx_project_req_sequence ON project_requisitions(project_id, sequence);
CREATE UNIQUE INDEX idx_req_project ON project_requisitions(requisition_id);
```

### Query Optimization
- Use preloading for nested relationships: `Preload("BillOfMaterials.Items.Specification")`
- Cache project summaries for large projects
- Paginate requisition lists for projects with 100+ requisitions

## Testing Strategy

### Unit Tests
- Test all CRUD operations for Project, BillOfMaterials, BillOfMaterialsItem, ProjectRequisition
- Test validation rules (budget >= 0, sequence uniqueness, etc.)
- Test cascade deletes and foreign key constraints

### Integration Tests
- Test project creation → Bill of Materials population → requisition linking workflow
- Test analytics calculations with realistic data
- Test Bill of Materials coverage calculations

### Edge Cases
- Project with no Bill of Materials
- Project with no requisitions
- Requisition items not in Bill of Materials (ad-hoc purchases)
- Bill of Materials items not covered by any requisition
- Circular references (prevented by schema)

## Future Enhancements

### Phase 2 Features (Future)
1. **Project Templates**: Save Bill of Materials as template for similar projects
2. **Milestone Tracking**: Add milestones with dates to projects
3. **Approval Workflow**: Require approval for requisitions exceeding budget
4. **Multi-project Dashboard**: Compare multiple projects side-by-side
5. **Export/Import**: Export project data to Excel/CSV
6. **Project Cloning**: Duplicate project structure for similar initiatives
7. **Budget Alerts**: Email notifications when approaching budget limits
8. **Gantt Chart**: Visual timeline for requisition sequencing

## Open Questions

1. **Should BillOfMaterialsItems have target prices?**
   - Pro: Enables Bill of Materials-level cost estimation before requisitions
   - Con: Adds complexity, quotes already track prices
   - **Recommendation**: Add optional `target_price` field to BillOfMaterialsItem

2. **Should projects support sub-projects?**
   - Pro: Enables hierarchical project organization
   - Con: Adds significant complexity
   - **Recommendation**: Defer to Phase 2, use tags instead

3. **Should requisitions be required to link to projects?**
   - Pro: Forces organizational structure
   - Con: Breaks backwards compatibility
   - **Recommendation**: Make it optional (allow standalone requisitions)

4. **How to handle Bill of Materials changes after requisitions created?**
   - Pro: Flexibility to adapt to changing requirements
   - Con: Can invalidate coverage analysis
   - **Recommendation**: Allow edits but show warnings if requisitions exist

## Conclusion

This design adds significant value by introducing project-level organization and Bill of Materials management while maintaining the existing requisition workflow. The architecture is:

- **Clean**: Clear separation between Bill of Materials (planning) and Requisitions (execution)
- **Flexible**: Requisitions can be standalone or project-linked
- **Scalable**: Handles projects from small (5 items) to large (500+ items)
- **Backwards Compatible**: Existing requisitions continue to work

The phased implementation approach allows for incremental delivery and validation at each stage.
