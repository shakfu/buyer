package models

import (
	"time"

	"gorm.io/gorm"
)

// Vendor represents a selling entity with currency and discount information
type Vendor struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;not null" json:"name"`
	Currency     string    `gorm:"size:3;not null" json:"currency"` // ISO 4217 currency code
	DiscountCode string    `gorm:"size:50" json:"discount_code,omitempty"`
	Brands       []*Brand  `gorm:"many2many:vendor_brands;" json:"brands,omitempty"`
	Quotes       []Quote   `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"quotes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Brand represents a manufacturing entity
type Brand struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Vendors   []*Vendor `gorm:"many2many:vendor_brands;" json:"vendors,omitempty"`
	Products  []Product `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"products,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Specification represents a general description of a type of product
type Specification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Products    []Product `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"products,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Product represents an item associated with a brand and specification
type Product struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	Name            string          `gorm:"uniqueIndex;not null" json:"name"`
	BrandID         uint            `gorm:"not null;index" json:"brand_id"`
	Brand           *Brand          `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"brand,omitempty"`
	SpecificationID *uint           `gorm:"index" json:"specification_id,omitempty"`
	Specification   *Specification  `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"specification,omitempty"`
	Quotes          []Quote         `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"quotes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Requisition represents a purchasing requirement
type Requisition struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	Name          string            `gorm:"uniqueIndex;not null" json:"name"`
	Justification string            `gorm:"type:text" json:"justification,omitempty"`
	Budget        float64           `json:"budget,omitempty"` // Optional overall budget limit
	Items         []RequisitionItem `gorm:"foreignKey:RequisitionID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// RequisitionItem represents a line item in a requisition
type RequisitionItem struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	RequisitionID   uint           `gorm:"not null;index" json:"requisition_id"`
	Requisition     *Requisition   `gorm:"foreignKey:RequisitionID;constraint:OnDelete:CASCADE" json:"requisition,omitempty"`
	SpecificationID uint           `gorm:"not null;index" json:"specification_id"`
	Specification   *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:RESTRICT" json:"specification,omitempty"`
	Quantity        int            `gorm:"not null" json:"quantity"`
	BudgetPerUnit   float64        `json:"budget_per_unit,omitempty"` // Optional budget per unit
	Description     string         `gorm:"type:text" json:"description,omitempty"` // Optional description for details
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Quote represents a price quote from a vendor for a product
type Quote struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	VendorID         uint       `gorm:"not null;index" json:"vendor_id"`
	Vendor           *Vendor    `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"vendor,omitempty"`
	ProductID        uint       `gorm:"not null;index" json:"product_id"`
	Product          *Product   `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT" json:"product,omitempty"`
	Price            float64    `gorm:"not null" json:"price"`
	Currency         string     `gorm:"size:3;not null" json:"currency"`
	ConvertedPrice   float64    `gorm:"not null" json:"converted_price"` // Price in USD
	ConversionRate   float64    `gorm:"not null" json:"conversion_rate"`
	QuoteDate        time.Time  `gorm:"not null;index" json:"quote_date"`
	ValidUntil       *time.Time `gorm:"index" json:"valid_until,omitempty"` // Optional expiration date
	Notes            string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Forex represents currency exchange rates
type Forex struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FromCurrency string    `gorm:"size:3;not null;index:idx_forex_pair" json:"from_currency"`
	ToCurrency   string    `gorm:"size:3;not null;index:idx_forex_pair" json:"to_currency"`
	Rate         float64   `gorm:"not null" json:"rate"`
	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Project represents a project with budget, deadline, and associated Bill of Materials
type Project struct {
	ID              uint                  `gorm:"primaryKey" json:"id"`
	Name            string                `gorm:"uniqueIndex;not null" json:"name"`
	Description     string                `gorm:"type:text" json:"description,omitempty"`
	Budget          float64               `json:"budget,omitempty"`          // Overall project budget
	Deadline        *time.Time            `json:"deadline,omitempty"`        // Project deadline
	Status          string                `gorm:"size:20;default:'planning'" json:"status"` // planning, active, completed, cancelled
	BillOfMaterials *BillOfMaterials      `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
	Requisitions    []ProjectRequisition  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"requisitions,omitempty"` // Project-based requisitions
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// BillOfMaterials represents the master list of specifications needed for a project
type BillOfMaterials struct {
	ID        uint                   `gorm:"primaryKey" json:"id"`
	ProjectID uint                   `gorm:"uniqueIndex;not null" json:"project_id"` // One BillOfMaterials per project
	Project   *Project               `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Notes     string                 `gorm:"type:text" json:"notes,omitempty"`
	Items     []BillOfMaterialsItem  `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// BillOfMaterialsItem represents a line item in a Bill of Materials
type BillOfMaterialsItem struct {
	ID                uint              `gorm:"primaryKey" json:"id"`
	BillOfMaterialsID uint              `gorm:"not null;index:idx_bom_spec,priority:1" json:"bill_of_materials_id"`
	BillOfMaterials   *BillOfMaterials  `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
	SpecificationID   uint              `gorm:"not null;index:idx_bom_spec,priority:2;uniqueIndex:idx_bom_spec_unique,composite:bom_spec" json:"specification_id"`
	Specification     *Specification    `gorm:"foreignKey:SpecificationID;constraint:OnDelete:RESTRICT" json:"specification,omitempty"`
	Quantity          int               `gorm:"not null" json:"quantity"`
	Notes             string            `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ProjectRequisition represents a procurement request created from a project's BOM
// Unlike standalone Requisitions, these are always tied to a project and reference BOM items
type ProjectRequisition struct {
	ID            uint                       `gorm:"primaryKey" json:"id"`
	ProjectID     uint                       `gorm:"not null;index" json:"project_id"`
	Project       *Project                   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Name          string                     `gorm:"not null" json:"name"`
	Justification string                     `gorm:"type:text" json:"justification,omitempty"`
	Budget        float64                    `json:"budget,omitempty"`
	Items         []ProjectRequisitionItem   `gorm:"foreignKey:ProjectRequisitionID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

// ProjectRequisitionItem represents a line item in a project requisition
// Links to specific BOM items rather than general specifications
type ProjectRequisitionItem struct {
	ID                    uint                 `gorm:"primaryKey" json:"id"`
	ProjectRequisitionID  uint                 `gorm:"not null;index" json:"project_requisition_id"`
	ProjectRequisition    *ProjectRequisition  `gorm:"foreignKey:ProjectRequisitionID;constraint:OnDelete:CASCADE" json:"project_requisition,omitempty"`
	BillOfMaterialsItemID uint                 `gorm:"not null;index" json:"bill_of_materials_item_id"`
	BOMItem               *BillOfMaterialsItem `gorm:"foreignKey:BillOfMaterialsItemID;constraint:OnDelete:RESTRICT" json:"bom_item,omitempty"`
	QuantityRequested     int                  `gorm:"not null" json:"quantity_requested"` // How much of this BOM item to procure
	Notes                 string               `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
}

// TableName overrides for GORM
func (Vendor) TableName() string                  { return "vendors" }
func (Brand) TableName() string                   { return "brands" }
func (Specification) TableName() string           { return "specifications" }
func (Product) TableName() string                 { return "products" }
func (Requisition) TableName() string             { return "requisitions" }
func (RequisitionItem) TableName() string         { return "requisition_items" }
func (Quote) TableName() string                   { return "quotes" }
func (Forex) TableName() string                   { return "forex" }
func (Project) TableName() string                 { return "projects" }
func (BillOfMaterials) TableName() string         { return "bills_of_materials" }
func (BillOfMaterialsItem) TableName() string     { return "bill_of_materials_items" }
func (ProjectRequisition) TableName() string      { return "project_requisitions" }
func (ProjectRequisitionItem) TableName() string  { return "project_requisition_items" }

// BeforeCreate hook for Vendor
func (v *Vendor) BeforeCreate(tx *gorm.DB) error {
	if v.Currency == "" {
		v.Currency = "USD"
	}
	return nil
}

// BeforeCreate hook for Quote - sets quote_date to now if not set
func (q *Quote) BeforeCreate(tx *gorm.DB) error {
	if q.QuoteDate.IsZero() {
		q.QuoteDate = time.Now()
	}
	return nil
}

// BeforeCreate hook for Project - sets default status if not provided
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = "planning"
	}
	return nil
}

// IsExpired checks if the quote has passed its expiration date
func (q *Quote) IsExpired() bool {
	if q.ValidUntil == nil {
		return false
	}
	return time.Now().After(*q.ValidUntil)
}

// IsStale checks if the quote is older than 90 days (or expired if ValidUntil is set)
func (q *Quote) IsStale() bool {
	if q.IsExpired() {
		return true
	}
	// Consider quotes older than 90 days as stale
	return time.Since(q.QuoteDate) > 90*24*time.Hour
}

// DaysUntilExpiration returns the number of days until expiration, or -1 if no expiration
func (q *Quote) DaysUntilExpiration() int {
	if q.ValidUntil == nil {
		return -1
	}
	duration := time.Until(*q.ValidUntil)
	return int(duration.Hours() / 24)
}
