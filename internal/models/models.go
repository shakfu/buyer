package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Vendor represents a selling entity with currency and discount information
type Vendor struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Currency     string `gorm:"size:3;not null" json:"currency"` // ISO 4217 currency code
	DiscountCode string `gorm:"size:50" json:"discount_code,omitempty"`

	// Contact Information
	ContactPerson string `gorm:"size:100" json:"contact_person,omitempty"`
	Email         string `gorm:"size:255" json:"email,omitempty"`
	Phone         string `gorm:"size:50" json:"phone,omitempty"`
	Website       string `gorm:"size:255" json:"website,omitempty"`

	// Address Information
	AddressLine1 string `gorm:"size:255" json:"address_line1,omitempty"`
	AddressLine2 string `gorm:"size:255" json:"address_line2,omitempty"`
	City         string `gorm:"size:100" json:"city,omitempty"`
	State        string `gorm:"size:100" json:"state,omitempty"`
	PostalCode   string `gorm:"size:20" json:"postal_code,omitempty"`
	Country      string `gorm:"size:2" json:"country,omitempty"` // ISO 3166-1 alpha-2

	// Business Information
	TaxID        string `gorm:"size:50" json:"tax_id,omitempty"`         // VAT/EIN/etc
	PaymentTerms string `gorm:"size:100" json:"payment_terms,omitempty"` // e.g., "Net 30"

	// Relationships
	Brands         []*Brand        `gorm:"many2many:vendor_brands;" json:"brands,omitempty"`
	Quotes         []Quote         `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"quotes,omitempty"`
	PurchaseOrders []PurchaseOrder `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"purchase_orders,omitempty"`
	Documents      []Document      `gorm:"-" json:"documents,omitempty"` // Polymorphic - query via EntityType="vendor" and EntityID=ID
	VendorRatings  []VendorRating  `gorm:"foreignKey:VendorID;constraint:OnDelete:CASCADE" json:"vendor_ratings,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
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
	ID          uint                     `gorm:"primaryKey" json:"id"`
	Name        string                   `gorm:"uniqueIndex;not null" json:"name"`
	Description string                   `gorm:"type:text" json:"description,omitempty"`
	Products    []Product                `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"products,omitempty"`
	Attributes  []SpecificationAttribute `gorm:"foreignKey:SpecificationID;constraint:OnDelete:CASCADE" json:"attributes,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// SpecificationAttribute defines what attributes a specification type should have
// Example: "Laptop" specification has attributes like "RAM", "Storage", "Screen Size"
type SpecificationAttribute struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	SpecificationID uint           `gorm:"not null;index:idx_spec_attr,priority:1" json:"specification_id"`
	Specification   *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:CASCADE" json:"specification,omitempty"`
	Name            string         `gorm:"not null;size:100;index:idx_spec_attr,priority:2" json:"name"` // "RAM", "Screen Size", "Storage Type"
	DataType        string         `gorm:"size:20;not null;default:'text'" json:"data_type"`             // "number", "text", "boolean"
	Unit            string         `gorm:"size:50" json:"unit,omitempty"`                                // "GB", "inches", "GHz" (optional)
	IsRequired      bool           `gorm:"default:false" json:"is_required"`                             // Must product have this attribute?
	MinValue        *float64       `json:"min_value,omitempty"`                                          // Validation for numbers
	MaxValue        *float64       `json:"max_value,omitempty"`                                          // Validation for numbers
	Description     string         `gorm:"type:text" json:"description,omitempty"`                       // Help text
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Product represents an item associated with a brand and specification
type Product struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"uniqueIndex;not null" json:"name"`
	SKU             *string        `gorm:"uniqueIndex;size:100" json:"sku,omitempty"` // Stock Keeping Unit (nullable, unique when present)
	Description     string         `gorm:"type:text" json:"description,omitempty"`
	BrandID         uint           `gorm:"not null;index" json:"brand_id"`
	Brand           *Brand         `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"brand,omitempty"`
	SpecificationID *uint          `gorm:"index" json:"specification_id,omitempty"`
	Specification   *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"specification,omitempty"`

	// Product Details
	UnitOfMeasure string `gorm:"size:20;default:'each'" json:"unit_of_measure,omitempty"` // each, box, case, kg, etc.
	MinOrderQty   int    `json:"min_order_qty,omitempty"`                                 // Minimum order quantity
	LeadTimeDays  int    `json:"lead_time_days,omitempty"`                                // Typical delivery time

	// Lifecycle
	IsActive       bool       `gorm:"default:true" json:"is_active"` // Product still available?
	DiscontinuedAt *time.Time `json:"discontinued_at,omitempty"`

	Quotes     []Quote            `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"quotes,omitempty"`
	Attributes []ProductAttribute `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"attributes,omitempty"`

	// Audit fields
	CreatedBy string    `gorm:"size:100" json:"created_by,omitempty"`
	UpdatedBy string    `gorm:"size:100" json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductAttribute stores actual attribute values for a specific product
// Example: iPhone 15 Pro has RAM=8, Storage=256, Screen Size=6.1
type ProductAttribute struct {
	ID                       uint                    `gorm:"primaryKey" json:"id"`
	ProductID                uint                    `gorm:"not null;index:idx_prod_attr,priority:1" json:"product_id"`
	Product                  *Product                `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	SpecificationAttributeID uint                    `gorm:"not null;index:idx_prod_attr,priority:2" json:"specification_attribute_id"`
	SpecificationAttribute   *SpecificationAttribute `gorm:"foreignKey:SpecificationAttributeID;constraint:OnDelete:RESTRICT" json:"specification_attribute,omitempty"`
	ValueText                *string                 `json:"value_text,omitempty"`    // For text/enum values
	ValueNumber              *float64                `json:"value_number,omitempty"`  // For numeric values
	ValueBoolean             *bool                   `json:"value_boolean,omitempty"` // For boolean values
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
}

// Requisition represents a purchasing requirement
type Requisition struct {
	ID             uint              `gorm:"primaryKey" json:"id"`
	Name           string            `gorm:"uniqueIndex;not null" json:"name"`
	Justification  string            `gorm:"type:text" json:"justification,omitempty"`
	Budget         float64           `json:"budget,omitempty"` // Optional overall budget limit
	Items          []RequisitionItem `gorm:"foreignKey:RequisitionID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	PurchaseOrders []PurchaseOrder   `gorm:"foreignKey:RequisitionID;constraint:OnDelete:SET NULL" json:"purchase_orders,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// RequisitionItem represents a line item in a requisition
type RequisitionItem struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	RequisitionID   uint           `gorm:"not null;index" json:"requisition_id"`
	Requisition     *Requisition   `gorm:"foreignKey:RequisitionID;constraint:OnDelete:CASCADE" json:"requisition,omitempty"`
	SpecificationID uint           `gorm:"not null;index" json:"specification_id"`
	Specification   *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:RESTRICT" json:"specification,omitempty"`
	Quantity        int            `gorm:"not null" json:"quantity"`
	BudgetPerUnit   float64        `json:"budget_per_unit,omitempty"`              // Optional budget per unit
	Description     string         `gorm:"type:text" json:"description,omitempty"` // Optional description for details
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Quote represents a price quote from a vendor for a product
type Quote struct {
	ID        uint     `gorm:"primaryKey" json:"id"`
	VendorID  uint     `gorm:"not null;index" json:"vendor_id"`
	Vendor    *Vendor  `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"vendor,omitempty"`
	ProductID uint     `gorm:"not null;index" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT" json:"product,omitempty"`

	// Version Tracking
	Version         int   `gorm:"not null;default:1" json:"version"`        // Quote revision number
	PreviousQuoteID *uint `gorm:"index" json:"previous_quote_id,omitempty"` // Link to previous version
	ReplacedBy      *uint `gorm:"index" json:"replaced_by,omitempty"`       // Link to newer version

	// Pricing
	Price          float64 `gorm:"not null" json:"price"`
	Currency       string  `gorm:"size:3;not null" json:"currency"`
	ConvertedPrice float64 `gorm:"not null" json:"converted_price"` // Price in USD
	ConversionRate float64 `gorm:"not null" json:"conversion_rate"`
	MinQuantity    int     `json:"min_quantity,omitempty"` // Minimum order for this price

	// Quote Details
	QuoteDate  time.Time  `gorm:"not null;index" json:"quote_date"`
	ValidUntil *time.Time `gorm:"index" json:"valid_until,omitempty"` // Optional expiration date

	// Status Tracking
	Status string `gorm:"size:20;default:'active'" json:"status"` // active, superseded, expired, accepted, declined

	Notes          string          `gorm:"type:text" json:"notes,omitempty"`
	PurchaseOrders []PurchaseOrder `gorm:"foreignKey:QuoteID;constraint:OnDelete:RESTRICT" json:"purchase_orders,omitempty"`

	// Audit fields
	CreatedBy string    `gorm:"size:100" json:"created_by,omitempty"`
	UpdatedBy string    `gorm:"size:100" json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PurchaseOrder represents an accepted quote that has been ordered
type PurchaseOrder struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	QuoteID          uint         `gorm:"not null;index" json:"quote_id"`
	Quote            *Quote       `gorm:"foreignKey:QuoteID;constraint:OnDelete:RESTRICT" json:"quote,omitempty"`
	VendorID         uint         `gorm:"not null;index" json:"vendor_id"` // Denormalized for easier queries
	Vendor           *Vendor      `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"vendor,omitempty"`
	ProductID        uint         `gorm:"not null;index" json:"product_id"` // Denormalized for easier queries
	Product          *Product     `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT" json:"product,omitempty"`
	RequisitionID    *uint        `gorm:"index" json:"requisition_id,omitempty"` // Optional link to requisition
	Requisition      *Requisition `gorm:"foreignKey:RequisitionID;constraint:OnDelete:SET NULL" json:"requisition,omitempty"`
	PONumber         string       `gorm:"uniqueIndex;not null;size:50" json:"po_number"`          // Generated or manual PO number
	Status           string       `gorm:"size:20;not null;default:'pending';index" json:"status"` // pending, approved, ordered, shipped, received, cancelled
	OrderDate        time.Time    `gorm:"not null;index" json:"order_date"`
	ExpectedDelivery *time.Time   `json:"expected_delivery,omitempty"`
	ActualDelivery   *time.Time   `json:"actual_delivery,omitempty"`
	Quantity         int          `gorm:"not null" json:"quantity"`        // Can order multiple units
	UnitPrice        float64      `gorm:"not null" json:"unit_price"`      // Price per unit in original currency
	Currency         string       `gorm:"size:3;not null" json:"currency"` // Quote currency
	TotalAmount      float64      `gorm:"not null" json:"total_amount"`    // Total cost (unit_price * quantity)
	ShippingCost     float64      `json:"shipping_cost,omitempty"`
	Tax              float64      `json:"tax,omitempty"`
	GrandTotal       float64      `gorm:"not null" json:"grand_total"` // total_amount + shipping_cost + tax
	InvoiceNumber    string       `gorm:"size:100" json:"invoice_number,omitempty"`
	Notes            string       `gorm:"type:text" json:"notes,omitempty"`

	// Relationships
	Documents     []Document     `gorm:"-" json:"documents,omitempty"` // Polymorphic - query via EntityType="purchase_order" and EntityID=ID
	VendorRatings []VendorRating `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE" json:"vendor_ratings,omitempty"`

	// Audit fields
	CreatedBy string    `gorm:"size:100" json:"created_by,omitempty"`
	UpdatedBy string    `gorm:"size:100" json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VendorRating represents performance ratings for vendors
type VendorRating struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	VendorID        uint           `gorm:"not null;index" json:"vendor_id"`
	Vendor          *Vendor        `gorm:"foreignKey:VendorID;constraint:OnDelete:CASCADE" json:"vendor,omitempty"`
	PurchaseOrderID *uint          `gorm:"index" json:"purchase_order_id,omitempty"` // Optional link to specific order
	PurchaseOrder   *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:SET NULL" json:"purchase_order,omitempty"`

	// Ratings (1-5 scale)
	PriceRating    *int `json:"price_rating,omitempty"`    // 1-5 scale
	QualityRating  *int `json:"quality_rating,omitempty"`  // 1-5 scale
	DeliveryRating *int `json:"delivery_rating,omitempty"` // 1-5 scale
	ServiceRating  *int `json:"service_rating,omitempty"`  // 1-5 scale

	Comments  string    `gorm:"type:text" json:"comments,omitempty"`
	RatedBy   string    `gorm:"size:100" json:"rated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Forex represents currency exchange rates
type Forex struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FromCurrency  string    `gorm:"size:3;not null;index:idx_forex_pair" json:"from_currency"`
	ToCurrency    string    `gorm:"size:3;not null;index:idx_forex_pair" json:"to_currency"`
	Rate          float64   `gorm:"not null" json:"rate"`
	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Project represents a project with budget, deadline, and associated Bill of Materials
type Project struct {
	ID              uint                 `gorm:"primaryKey" json:"id"`
	Name            string               `gorm:"uniqueIndex;not null" json:"name"`
	Description     string               `gorm:"type:text" json:"description,omitempty"`
	Budget          float64              `json:"budget,omitempty"`                         // Overall project budget
	Deadline        *time.Time           `json:"deadline,omitempty"`                       // Project deadline
	Status          string               `gorm:"size:20;default:'planning'" json:"status"` // planning, active, completed, cancelled
	BillOfMaterials *BillOfMaterials     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
	Requisitions    []ProjectRequisition `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"requisitions,omitempty"` // Project-based requisitions
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// BillOfMaterials represents the master list of specifications needed for a project
type BillOfMaterials struct {
	ID        uint                  `gorm:"primaryKey" json:"id"`
	ProjectID uint                  `gorm:"uniqueIndex;not null" json:"project_id"` // One BillOfMaterials per project
	Project   *Project              `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Notes     string                `gorm:"type:text" json:"notes,omitempty"`
	Items     []BillOfMaterialsItem `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// BillOfMaterialsItem represents a line item in a Bill of Materials
type BillOfMaterialsItem struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	BillOfMaterialsID uint             `gorm:"not null;index:idx_bom_spec,priority:1" json:"bill_of_materials_id"`
	BillOfMaterials   *BillOfMaterials `gorm:"foreignKey:BillOfMaterialsID;constraint:OnDelete:CASCADE" json:"bill_of_materials,omitempty"`
	SpecificationID   uint             `gorm:"not null;index:idx_bom_spec,priority:2;uniqueIndex:idx_bom_spec_unique,composite:bom_spec" json:"specification_id"`
	Specification     *Specification   `gorm:"foreignKey:SpecificationID;constraint:OnDelete:RESTRICT" json:"specification,omitempty"`
	Quantity          int              `gorm:"not null" json:"quantity"`
	Notes             string           `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// ProjectRequisition represents a procurement request created from a project's BOM
// Unlike standalone Requisitions, these are always tied to a project and reference BOM items
type ProjectRequisition struct {
	ID            uint                     `gorm:"primaryKey" json:"id"`
	ProjectID     uint                     `gorm:"not null;index" json:"project_id"`
	Project       *Project                 `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Name          string                   `gorm:"not null" json:"name"`
	Justification string                   `gorm:"type:text" json:"justification,omitempty"`
	Budget        float64                  `json:"budget,omitempty"`
	Items         []ProjectRequisitionItem `gorm:"foreignKey:ProjectRequisitionID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
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

	// Procurement tracking fields
	SelectedQuoteID   *uint   `gorm:"index" json:"selected_quote_id,omitempty"`
	SelectedQuote     *Quote  `gorm:"foreignKey:SelectedQuoteID;constraint:OnDelete:SET NULL" json:"selected_quote,omitempty"`
	TargetUnitPrice   float64 `json:"target_unit_price,omitempty"`                          // Budget or target price
	ActualUnitPrice   float64 `json:"actual_unit_price,omitempty"`                          // Final negotiated price
	ProcurementStatus string  `gorm:"size:20;default:'pending'" json:"procurement_status"` // pending, quoted, ordered, received, cancelled

	Notes     string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProjectProcurementStrategy stores user-selected optimization strategies for a project
type ProjectProcurementStrategy struct {
	ID        uint     `gorm:"primaryKey" json:"id"`
	ProjectID uint     `gorm:"uniqueIndex;not null" json:"project_id"`
	Project   *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`

	// Strategy settings
	Strategy            string   `gorm:"size:30;default:'lowest_cost'" json:"strategy"` // lowest_cost, fewest_vendors, balanced, quality_focused
	MaxVendors          *int     `json:"max_vendors,omitempty"`                         // Optional vendor limit
	MinVendorRating     *float64 `json:"min_vendor_rating,omitempty"`                   // Minimum acceptable rating (1-5)
	PreferredVendorIDs  string   `gorm:"type:text" json:"preferred_vendor_ids,omitempty"` // Comma-separated IDs
	ExcludedVendorIDs   string   `gorm:"type:text" json:"excluded_vendor_ids,omitempty"`  // Comma-separated IDs
	AllowPartialFulfill bool     `gorm:"default:true" json:"allow_partial_fulfill"`       // Allow splitting orders

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides for GORM
func (Vendor) TableName() string                 { return "vendors" }
func (Brand) TableName() string                  { return "brands" }
func (Specification) TableName() string          { return "specifications" }
func (SpecificationAttribute) TableName() string { return "specification_attributes" }
func (Product) TableName() string                { return "products" }
func (ProductAttribute) TableName() string       { return "product_attributes" }
func (Requisition) TableName() string            { return "requisitions" }
func (RequisitionItem) TableName() string        { return "requisition_items" }
func (Quote) TableName() string                  { return "quotes" }
func (PurchaseOrder) TableName() string          { return "purchase_orders" }
func (VendorRating) TableName() string           { return "vendor_ratings" }
func (Forex) TableName() string                  { return "forex" }
func (Project) TableName() string                { return "projects" }
func (BillOfMaterials) TableName() string        { return "bills_of_materials" }
func (BillOfMaterialsItem) TableName() string    { return "bill_of_materials_items" }
func (ProjectRequisition) TableName() string          { return "project_requisitions" }
func (ProjectRequisitionItem) TableName() string      { return "project_requisition_items" }
func (ProjectProcurementStrategy) TableName() string  { return "project_procurement_strategies" }
func (Document) TableName() string                    { return "documents" }

// Document represents file attachments for various entities
type Document struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EntityType  string    `gorm:"size:50;not null;index:idx_entity" json:"entity_type"` // vendor, quote, purchase_order, product, etc.
	EntityID    uint      `gorm:"not null;index:idx_entity" json:"entity_id"`
	FileName    string    `gorm:"not null" json:"file_name"`
	FileType    string    `gorm:"size:50" json:"file_type"`  // pdf, xlsx, docx, png, jpg
	FileSize    int64     `json:"file_size"`                 // bytes
	FilePath    string    `gorm:"not null" json:"file_path"` // Storage location or S3 key
	Description string    `gorm:"type:text" json:"description,omitempty"`
	UploadedBy  string    `gorm:"size:100" json:"uploaded_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// BeforeSave hook for RequisitionItem - validates constraints
func (ri *RequisitionItem) BeforeSave(tx *gorm.DB) error {
	// Validate positive quantity
	if ri.Quantity <= 0 {
		return fmt.Errorf("requisition item quantity must be positive, got %d", ri.Quantity)
	}

	// Validate positive budget per unit if set
	if ri.BudgetPerUnit < 0 {
		return fmt.Errorf("requisition item budget per unit cannot be negative, got %.2f", ri.BudgetPerUnit)
	}

	return nil
}

// BeforeSave hook for Product - validates constraints
func (p *Product) BeforeSave(tx *gorm.DB) error {
	// Validate minimum order quantity
	if p.MinOrderQty < 0 {
		return fmt.Errorf("product minimum order quantity cannot be negative, got %d", p.MinOrderQty)
	}

	// Validate lead time days
	if p.LeadTimeDays < 0 {
		return fmt.Errorf("product lead time days cannot be negative, got %d", p.LeadTimeDays)
	}

	return nil
}

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

// BeforeSave hook for Quote - validates constraints
func (q *Quote) BeforeSave(tx *gorm.DB) error {
	// Validate positive price
	if q.Price <= 0 {
		return fmt.Errorf("quote price must be positive, got %.2f", q.Price)
	}
	if q.ConvertedPrice <= 0 {
		return fmt.Errorf("quote converted price must be positive, got %.2f", q.ConvertedPrice)
	}
	if q.ConversionRate <= 0 {
		return fmt.Errorf("quote conversion rate must be positive, got %.2f", q.ConversionRate)
	}

	// Validate status enum
	validStatuses := map[string]bool{
		"active": true, "superseded": true, "expired": true,
		"accepted": true, "declined": true,
	}
	if q.Status != "" && !validStatuses[q.Status] {
		return fmt.Errorf("invalid quote status: %s (must be one of: active, superseded, expired, accepted, declined)", q.Status)
	}

	// Validate minimum quantity
	if q.MinQuantity < 0 {
		return fmt.Errorf("quote minimum quantity cannot be negative, got %d", q.MinQuantity)
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

// BeforeSave hook for Project - validates constraints
func (p *Project) BeforeSave(tx *gorm.DB) error {
	// Validate status enum
	validStatuses := map[string]bool{
		"planning": true, "active": true, "completed": true, "cancelled": true,
	}
	if p.Status != "" && !validStatuses[p.Status] {
		return fmt.Errorf("invalid project status: %s (must be one of: planning, active, completed, cancelled)", p.Status)
	}

	// Validate positive budget
	if p.Budget < 0 {
		return fmt.Errorf("project budget cannot be negative, got %.2f", p.Budget)
	}

	return nil
}

// BeforeCreate hook for PurchaseOrder - sets defaults
func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.OrderDate.IsZero() {
		po.OrderDate = time.Now()
	}
	if po.Status == "" {
		po.Status = "pending"
	}
	// Calculate grand total if not set
	if po.GrandTotal == 0 {
		po.GrandTotal = po.TotalAmount + po.ShippingCost + po.Tax
	}
	return nil
}

// BeforeSave hook for PurchaseOrder - validates constraints
func (po *PurchaseOrder) BeforeSave(tx *gorm.DB) error {
	// Validate status enum
	validStatuses := map[string]bool{
		"pending": true, "approved": true, "ordered": true,
		"shipped": true, "received": true, "cancelled": true,
	}
	if po.Status != "" && !validStatuses[po.Status] {
		return fmt.Errorf("invalid purchase order status: %s (must be one of: pending, approved, ordered, shipped, received, cancelled)", po.Status)
	}

	// Validate positive quantity
	if po.Quantity <= 0 {
		return fmt.Errorf("purchase order quantity must be positive, got %d", po.Quantity)
	}

	// Validate positive amounts
	if po.TotalAmount < 0 {
		return fmt.Errorf("purchase order total amount cannot be negative, got %.2f", po.TotalAmount)
	}
	if po.ShippingCost < 0 {
		return fmt.Errorf("purchase order shipping cost cannot be negative, got %.2f", po.ShippingCost)
	}
	if po.Tax < 0 {
		return fmt.Errorf("purchase order tax cannot be negative, got %.2f", po.Tax)
	}
	if po.GrandTotal < 0 {
		return fmt.Errorf("purchase order grand total cannot be negative, got %.2f", po.GrandTotal)
	}

	// Note: We don't validate that actual delivery must be after expected delivery
	// because items can arrive early, and that's a valid scenario

	return nil
}

// BeforeSave hook for SpecificationAttribute - validates constraints
func (sa *SpecificationAttribute) BeforeSave(tx *gorm.DB) error {
	// Validate data type enum
	validDataTypes := map[string]bool{
		"text": true, "number": true, "boolean": true,
	}
	if sa.DataType != "" && !validDataTypes[sa.DataType] {
		return fmt.Errorf("invalid data type: %s (must be one of: text, number, boolean)", sa.DataType)
	}

	// Validate min/max values
	if sa.MinValue != nil && sa.MaxValue != nil && *sa.MinValue > *sa.MaxValue {
		return fmt.Errorf("min_value (%.2f) cannot be greater than max_value (%.2f)", *sa.MinValue, *sa.MaxValue)
	}

	return nil
}

// BeforeSave hook for ProductAttribute - validates constraints and data types
func (pa *ProductAttribute) BeforeSave(tx *gorm.DB) error {
	// Exactly one value field must be set
	valuesSet := 0
	if pa.ValueText != nil {
		valuesSet++
	}
	if pa.ValueNumber != nil {
		valuesSet++
	}
	if pa.ValueBoolean != nil {
		valuesSet++
	}

	if valuesSet == 0 {
		return fmt.Errorf("product attribute must have at least one value set (value_text, value_number, or value_boolean)")
	}
	if valuesSet > 1 {
		return fmt.Errorf("product attribute can only have one value type set")
	}

	// Validate against specification attribute constraints if available
	if pa.SpecificationAttribute != nil {
		attr := pa.SpecificationAttribute

		// Check data type matches
		switch attr.DataType {
		case "number":
			if pa.ValueNumber == nil {
				return fmt.Errorf("attribute '%s' expects a number value", attr.Name)
			}
			// Validate min/max constraints
			if attr.MinValue != nil && *pa.ValueNumber < *attr.MinValue {
				return fmt.Errorf("value %.2f is below minimum %.2f for attribute '%s'", *pa.ValueNumber, *attr.MinValue, attr.Name)
			}
			if attr.MaxValue != nil && *pa.ValueNumber > *attr.MaxValue {
				return fmt.Errorf("value %.2f exceeds maximum %.2f for attribute '%s'", *pa.ValueNumber, *attr.MaxValue, attr.Name)
			}
		case "text":
			if pa.ValueText == nil {
				return fmt.Errorf("attribute '%s' expects a text value", attr.Name)
			}
		case "boolean":
			if pa.ValueBoolean == nil {
				return fmt.Errorf("attribute '%s' expects a boolean value", attr.Name)
			}
		}
	}

	return nil
}

// BeforeSave hook for ProjectRequisitionItem - validates constraints
func (pri *ProjectRequisitionItem) BeforeSave(tx *gorm.DB) error {
	// Validate positive quantity
	if pri.QuantityRequested <= 0 {
		return fmt.Errorf("project requisition item quantity must be positive, got %d", pri.QuantityRequested)
	}

	// Validate procurement status enum
	validStatuses := map[string]bool{
		"pending": true, "quoted": true, "ordered": true, "received": true, "cancelled": true,
	}
	if pri.ProcurementStatus != "" && !validStatuses[pri.ProcurementStatus] {
		return fmt.Errorf("invalid procurement status: %s (must be one of: pending, quoted, ordered, received, cancelled)", pri.ProcurementStatus)
	}

	// Validate prices are non-negative
	if pri.TargetUnitPrice < 0 {
		return fmt.Errorf("target unit price cannot be negative, got %.2f", pri.TargetUnitPrice)
	}
	if pri.ActualUnitPrice < 0 {
		return fmt.Errorf("actual unit price cannot be negative, got %.2f", pri.ActualUnitPrice)
	}

	return nil
}

// BeforeCreate hook for ProjectRequisitionItem - sets default status
func (pri *ProjectRequisitionItem) BeforeCreate(tx *gorm.DB) error {
	if pri.ProcurementStatus == "" {
		pri.ProcurementStatus = "pending"
	}
	return nil
}

// BeforeSave hook for ProjectProcurementStrategy - validates constraints
func (pps *ProjectProcurementStrategy) BeforeSave(tx *gorm.DB) error {
	// Validate strategy enum
	validStrategies := map[string]bool{
		"lowest_cost": true, "fewest_vendors": true, "balanced": true, "quality_focused": true,
	}
	if pps.Strategy != "" && !validStrategies[pps.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be one of: lowest_cost, fewest_vendors, balanced, quality_focused)", pps.Strategy)
	}

	// Validate max vendors is positive if set
	if pps.MaxVendors != nil && *pps.MaxVendors <= 0 {
		return fmt.Errorf("max vendors must be positive, got %d", *pps.MaxVendors)
	}

	// Validate min vendor rating is in valid range if set
	if pps.MinVendorRating != nil {
		if *pps.MinVendorRating < 1.0 || *pps.MinVendorRating > 5.0 {
			return fmt.Errorf("min vendor rating must be between 1.0 and 5.0, got %.2f", *pps.MinVendorRating)
		}
	}

	return nil
}

// BeforeCreate hook for ProjectProcurementStrategy - sets default strategy
func (pps *ProjectProcurementStrategy) BeforeCreate(tx *gorm.DB) error {
	if pps.Strategy == "" {
		pps.Strategy = "lowest_cost"
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
// Fixed logic: If ValidUntil is set and still valid, the quote is not stale regardless of age
func (q *Quote) IsStale() bool {
	// If expired, it's definitely stale
	if q.IsExpired() {
		return true
	}

	// If ValidUntil is set and still valid, not stale
	if q.ValidUntil != nil {
		return false
	}

	// If no expiration set, consider stale after 90 days
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
