# Data Model Analysis - Buyer Vendor Quote Management System

**Analysis Date:** 2025-11-13
**Codebase Version:** 068d6d3 (with recent refactorings)

## Executive Summary

The current data model is well-structured with proper relationships and constraints. However, several critical business entities and fields are missing that would be needed for a production procurement system. This analysis identifies gaps and proposes enhancements organized by priority.

**Overall Assessment:**
- ✅ **Strong Foundation:** Clean relationships, proper constraints, good separation of concerns
- ⚠️ **Missing Critical Features:** Purchase orders, vendor contacts, shipping/receiving
- ⚠️ **Limited Audit Trail:** No user tracking, soft deletes not implemented
- ✅ **Good Domain Modeling:** Clear distinction between specifications and products

---

## 1. Current Models Overview

### Core Entities (Well-Implemented)
- ✅ **Vendor** - Selling entities with currency
- ✅ **Brand** - Manufacturing entities
- ✅ **Specification** - Generic product types (e.g., "smartphone", "17-inch 4K Monitor")
- ✅ **Product** - Specific products from brands (e.g., "iPhone 15", "Dell XPS 17")
- ✅ **Quote** - Price quotes with currency conversion
- ✅ **Forex** - Exchange rate tracking

### Procurement Workflow (Good)
- ✅ **Requisition** - Purchasing requirements
- ✅ **RequisitionItem** - Line items with specifications
- ✅ **Project** - Project tracking with budget/deadline
- ✅ **BillOfMaterials** - Project material requirements
- ✅ **ProjectRequisition** - Project-based procurement requests

### Strengths
1. **Proper normalization** - No obvious redundancy
2. **Good constraint modeling** - Cascade/restrict deletes where appropriate
3. **Flexible relationships** - M:N between Vendor and Brand
4. **Currency handling** - Automatic conversion to USD
5. **Temporal tracking** - Quote expiration, creation timestamps

---

## 2. CRITICAL Missing Entities

### D1: Purchase Orders (SEVERITY: CRITICAL)
**Status:** MISSING
**Impact:** Cannot track actual purchases or fulfillment

**Current Gap:**
Quotes exist, but there's no way to:
- Accept a quote and create a purchase order
- Track order status (pending, approved, shipped, received)
- Record delivery dates
- Match invoices to orders
- Track fulfillment

**Proposed Model:**
```go
// PurchaseOrder represents an accepted quote that has been ordered
type PurchaseOrder struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    QuoteID          uint           `gorm:"not null;index" json:"quote_id"`
    Quote            *Quote         `gorm:"foreignKey:QuoteID;constraint:OnDelete:RESTRICT" json:"quote,omitempty"`
    RequisitionID    *uint          `gorm:"index" json:"requisition_id,omitempty"` // Optional link to requisition
    Requisition      *Requisition   `gorm:"foreignKey:RequisitionID;constraint:OnDelete:SET NULL" json:"requisition,omitempty"`
    PONumber         string         `gorm:"uniqueIndex;not null;size:50" json:"po_number"` // Generated or manual PO number
    Status           string         `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, approved, ordered, shipped, received, cancelled
    OrderDate        time.Time      `gorm:"not null;index" json:"order_date"`
    ExpectedDelivery *time.Time     `json:"expected_delivery,omitempty"`
    ActualDelivery   *time.Time     `json:"actual_delivery,omitempty"`
    Quantity         int            `gorm:"not null" json:"quantity"` // Can order multiple units
    TotalAmount      float64        `gorm:"not null" json:"total_amount"` // Total cost (price * quantity)
    ShippingCost     float64        `json:"shipping_cost,omitempty"`
    Tax              float64        `json:"tax,omitempty"`
    InvoiceNumber    string         `gorm:"size:100" json:"invoice_number,omitempty"`
    Notes            string         `gorm:"type:text" json:"notes,omitempty"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`
}
```

**Business Value:**
- Track order fulfillment end-to-end
- Match invoices to orders
- Calculate shipping costs
- Monitor delivery performance
- Essential for any real procurement system

---

## 3. HIGH PRIORITY Missing Fields

### D2: Vendor Contact Information (SEVERITY: HIGH)
**Status:** MISSING
**Impact:** Cannot communicate with vendors

**Current Gap:**
```go
type Vendor struct {
    ID           uint
    Name         string
    Currency     string
    DiscountCode string
    // Include URL
    // Missing: All contact information
}
```

**Proposed Enhancement:**
```go
type Vendor struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Name         string    `gorm:"uniqueIndex;not null" json:"name"`
    Currency     string    `gorm:"size:3;not null" json:"currency"`
    DiscountCode string    `gorm:"size:50" json:"discount_code,omitempty"`

    // Contact Information (NEW)
    ContactPerson string    `gorm:"size:100" json:"contact_person,omitempty"`
    Email         string    `gorm:"size:255" json:"email,omitempty"`
    Phone         string    `gorm:"size:50" json:"phone,omitempty"`
    Website       string    `gorm:"size:255" json:"website,omitempty"`

    // Address Information (NEW)
    AddressLine1  string    `gorm:"size:255" json:"address_line1,omitempty"`
    AddressLine2  string    `gorm:"size:255" json:"address_line2,omitempty"`
    City          string    `gorm:"size:100" json:"city,omitempty"`
    State         string    `gorm:"size:100" json:"state,omitempty"`
    PostalCode    string    `gorm:"size:20" json:"postal_code,omitempty"`
    Country       string    `gorm:"size:2" json:"country,omitempty"` // ISO 3166-1 alpha-2

    // Business Information (NEW)
    TaxID         string    `gorm:"size:50" json:"tax_id,omitempty"` // VAT/EIN/etc
    PaymentTerms  string    `gorm:"size:100" json:"payment_terms,omitempty"` // e.g., "Net 30"

    Brands        []*Brand  `gorm:"many2many:vendor_brands;" json:"brands,omitempty"`
    Quotes        []Quote   `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"quotes,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

**Business Value:**
- Enable communication with vendors
- Track payment terms
- Store shipping addresses
- Required for purchase order generation
- Tax compliance

### D3: Product Extended Information (SEVERITY: HIGH)
**Status:** MINIMAL
**Impact:** Limited product comparison and specification tracking

**Current Gap:**
```go
type Product struct {
    ID              uint
    Name            string
    BrandID         uint
    SpecificationID *uint
    // Missing: SKU, description, technical specs, units
}
```

**Proposed Enhancement:**
```go
type Product struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    Name            string         `gorm:"uniqueIndex;not null" json:"name"`
    SKU             string         `gorm:"uniqueIndex;size:100" json:"sku,omitempty"` // NEW
    Description     string         `gorm:"type:text" json:"description,omitempty"` // NEW
    BrandID         uint           `gorm:"not null;index" json:"brand_id"`
    Brand           *Brand         `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"brand,omitempty"`
    SpecificationID *uint          `gorm:"index" json:"specification_id,omitempty"`
    Specification   *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"specification,omitempty"`

    // Product Details (NEW)
    UnitOfMeasure   string         `gorm:"size:20;default:'each'" json:"unit_of_measure,omitempty"` // each, box, case, kg, etc.
    MinOrderQty     int            `json:"min_order_qty,omitempty"` // Minimum order quantity
    LeadTimeDays    int            `json:"lead_time_days,omitempty"` // Typical delivery time

    // Lifecycle (NEW)
    IsActive        bool           `gorm:"default:true" json:"is_active"` // Product still available?
    DiscontinuedAt  *time.Time     `json:"discontinued_at,omitempty"`

    Quotes          []Quote        `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"quotes,omitempty"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
}
```

**Business Value:**
- Track product availability
- Enforce minimum order quantities
- Estimate delivery times
- Manage product lifecycle
- Better procurement planning

### D4: Quote History and Versioning (SEVERITY: MEDIUM)
**Status:** MISSING
**Impact:** Cannot track quote changes or negotiations

**Current Gap:**
- No quote version tracking
- Cannot see quote change history
- Cannot track vendor negotiations

**Proposed Enhancement:**
```go
type Quote struct {
    ID               uint       `gorm:"primaryKey" json:"id"`
    VendorID         uint       `gorm:"not null;index" json:"vendor_id"`
    Vendor           *Vendor    `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"vendor,omitempty"`
    ProductID        uint       `gorm:"not null;index" json:"product_id"`
    Product          *Product   `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT" json:"product,omitempty"`

    // Version Tracking (NEW)
    Version          int        `gorm:"not null;default:1" json:"version"` // Quote revision number
    PreviousQuoteID  *uint      `gorm:"index" json:"previous_quote_id,omitempty"` // Link to previous version
    ReplacedBy       *uint      `gorm:"index" json:"replaced_by,omitempty"` // Link to newer version

    // Pricing
    Price            float64    `gorm:"not null" json:"price"`
    Currency         string     `gorm:"size:3;not null" json:"currency"`
    ConvertedPrice   float64    `gorm:"not null" json:"converted_price"`
    ConversionRate   float64    `gorm:"not null" json:"conversion_rate"`

    // Quote Details
    MinQuantity      int        `json:"min_quantity,omitempty"` // NEW - Minimum order for this price
    QuoteDate        time.Time  `gorm:"not null;index" json:"quote_date"`
    ValidUntil       *time.Time `gorm:"index" json:"valid_until,omitempty"`

    // Status Tracking (NEW)
    Status           string     `gorm:"size:20;default:'active'" json:"status"` // active, superseded, expired, accepted, declined

    Notes            string     `gorm:"type:text" json:"notes,omitempty"`
    CreatedAt        time.Time  `json:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at"`
}
```

**Business Value:**
- Track price negotiations
- Maintain quote audit trail
- See historical pricing trends
- Track which quotes were accepted

---

## 4. MEDIUM PRIORITY Enhancements

### D5: Audit Trail (SEVERITY: MEDIUM)
**Status:** PARTIAL (only timestamps)
**Impact:** Cannot track who made changes

**Current State:**
All models have `CreatedAt` and `UpdatedAt`, but no user tracking.

**Proposed Enhancement:**
```go
// Add to all models that need audit tracking
type AuditFields struct {
    CreatedBy   string     `gorm:"size:100" json:"created_by,omitempty"` // Username or user ID
    UpdatedBy   string     `gorm:"size:100" json:"updated_by,omitempty"`
    DeletedBy   string     `gorm:"size:100" json:"deleted_by,omitempty"`
    DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete support
}
```

**Implementation Note:**
This requires implementing user authentication and session management first.

**Business Value:**
- Accountability for changes
- Compliance requirements
- Troubleshooting who changed what
- Soft delete recovery

### D6: Specification Versioning (SEVERITY: LOW)
**Status:** MISSING
**Impact:** Historical requisitions lose context when specs change

**Current Gap:**
```go
type Specification struct {
    ID          uint
    Name        string
    Description string
    // No versioning - if description changes, old requisitions lose context
}
```

**Proposed Enhancement:**
Consider making specifications immutable with versioning:
```go
type Specification struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"not null;index" json:"name"` // Not unique anymore
    Version     int       `gorm:"not null;default:1" json:"version"`
    Description string    `gorm:"type:text" json:"description,omitempty"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    ReplacesID  *uint     `gorm:"index" json:"replaces_id,omitempty"` // Points to previous version
    Products    []Product `gorm:"foreignKey:SpecificationID;constraint:OnDelete:SET NULL" json:"products,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**Business Value:**
- Maintain historical accuracy
- Track specification evolution
- Compliance and auditing

### D7: Quote IsStale() Logic Issue (SEVERITY: LOW)
**Status:** INCORRECT LOGIC
**Location:** `internal/models/models.go` lines 221-228

**Current Implementation:**
```go
func (q *Quote) IsStale() bool {
    if q.IsExpired() {
        return true
    }
    // Consider quotes older than 90 days as stale
    return time.Since(q.QuoteDate) > 90*24*time.Hour
}
```

**Problem:**
A quote dated yesterday but valid for 1 year would be marked as stale after 90 days, even though it's still valid.

**Proposed Fix:**
```go
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
```

**Business Value:**
- Accurate quote status
- Don't ignore valid long-term quotes
- Better decision making

---

## 5. Additional Missing Entities (NICE TO HAVE)

### D8: Attachments/Documents (SEVERITY: LOW)
**Status:** MISSING
**Impact:** Cannot store quote PDFs, invoices, contracts

**Proposed Model:**
```go
type Document struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    EntityType   string    `gorm:"size:50;not null;index" json:"entity_type"` // vendor, quote, purchase_order, etc.
    EntityID     uint      `gorm:"not null;index" json:"entity_id"`
    FileName     string    `gorm:"not null" json:"file_name"`
    FileType     string    `gorm:"size:50" json:"file_type"` // pdf, xlsx, docx
    FileSize     int64     `json:"file_size"` // bytes
    FilePath     string    `gorm:"not null" json:"file_path"` // Storage location or S3 key
    Description  string    `gorm:"type:text" json:"description,omitempty"`
    UploadedBy   string    `gorm:"size:100" json:"uploaded_by,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
}
```

### D9: Vendor Performance Tracking (SEVERITY: LOW)
**Status:** MISSING
**Impact:** Cannot evaluate vendor reliability

**Proposed Model:**
```go
type VendorRating struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    VendorID        uint      `gorm:"not null;index" json:"vendor_id"`
    Vendor          *Vendor   `gorm:"foreignKey:VendorID;constraint:OnDelete:CASCADE" json:"vendor,omitempty"`
    PurchaseOrderID *uint     `gorm:"index" json:"purchase_order_id,omitempty"` // Optional link to specific order

    // Ratings (1-5 scale)
    PriceRating     int       `json:"price_rating,omitempty"`
    QualityRating   int       `json:"quality_rating,omitempty"`
    DeliveryRating  int       `json:"delivery_rating,omitempty"`
    ServiceRating   int       `json:"service_rating,omitempty"`

    Comments        string    `gorm:"type:text" json:"comments,omitempty"`
    RatedBy         string    `gorm:"size:100" json:"rated_by,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
}
```

### D10: Budget Tracking and Approval Workflow (SEVERITY: LOW)
**Status:** MINIMAL
**Impact:** No formal approval process

**Current State:**
- Projects have budgets
- Requisitions have optional budgets
- No approval workflow

**Proposed Enhancement:**
```go
type Approval struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    EntityType    string    `gorm:"size:50;not null;index" json:"entity_type"` // requisition, purchase_order
    EntityID      uint      `gorm:"not null;index" json:"entity_id"`
    ApproverName  string    `gorm:"not null" json:"approver_name"`
    Status        string    `gorm:"size:20;not null" json:"status"` // pending, approved, rejected
    Comments      string    `gorm:"type:text" json:"comments,omitempty"`
    ApprovedAt    *time.Time `json:"approved_at,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}
```

---

## 6. Relationship Analysis

### Current Relationships (Good)

#### Many-to-Many
✅ **Vendor ↔ Brand** - Vendors can sell multiple brands, brands available from multiple vendors

#### One-to-Many (with CASCADE)
✅ **Product → Quote** - Delete product removes quotes
✅ **Requisition → RequisitionItem** - Delete requisition removes items
✅ **Project → BillOfMaterials** - Delete project removes BOM
✅ **BillOfMaterials → BillOfMaterialsItem** - Delete BOM removes items
✅ **Project → ProjectRequisition** - Delete project removes requisitions

#### One-to-Many (with RESTRICT)
✅ **Brand → Product** - Cannot delete brand with products
✅ **Vendor → Quote** - Cannot delete vendor with quotes
✅ **Specification → Product** - Cannot delete spec with products
✅ **Specification → RequisitionItem** - Cannot delete spec with items

#### One-to-Many (with SET NULL)
✅ **Specification → Product** - Delete spec sets product.specification_id to NULL

### Missing Relationships

❌ **Quote → PurchaseOrder** - Need to track which quotes were accepted
❌ **Requisition → PurchaseOrder** - Link requisitions to fulfillment
❌ **PurchaseOrder → Document** - Attach invoices, receipts
❌ **Vendor → Document** - Attach contracts, certifications
❌ **PurchaseOrder → VendorRating** - Rate vendors based on order performance

---

## 7. Index Analysis

### Current Indexes (Good)
✅ Unique indexes on all name fields
✅ Foreign key indexes
✅ Composite index on `BillOfMaterialsItem` (bill_of_materials_id, specification_id)
✅ Date indexes on `Quote.QuoteDate` and `Quote.ValidUntil`
✅ Forex pair composite index

### Recommended Additional Indexes
```sql
-- For purchase order queries
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_order_date ON purchase_orders(order_date);
CREATE INDEX idx_purchase_orders_vendor ON purchase_orders(vendor_id, status);

-- For vendor search
CREATE INDEX idx_vendors_email ON vendors(email);
CREATE INDEX idx_vendors_country ON vendors(country);

-- For product search
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_is_active ON products(is_active);

-- For audit trail
CREATE INDEX idx_deleted_at ON <all_tables>(deleted_at);
```

---

## 8. Implementation Priority

### Phase 1: Critical Business Features (MUST HAVE)
1. **PurchaseOrder model** - Core procurement workflow
2. **Vendor contact information** - Enable vendor communication
3. **Product extended fields** - SKU, description, units

### Phase 2: Enhanced Operations (SHOULD HAVE)
4. **Quote versioning** - Track negotiations
5. **Audit trail** - User tracking and soft deletes
6. **Product lifecycle** - Active/discontinued status

### Phase 3: Advanced Features (NICE TO HAVE)
7. **Document attachments** - Store PDFs, invoices
8. **Vendor ratings** - Performance tracking
9. **Approval workflow** - Budget compliance
10. **Specification versioning** - Historical accuracy

---

## 9. Database Constraints to Add

### Validation Constraints
```sql
-- Ensure positive values
ALTER TABLE quotes ADD CONSTRAINT chk_quote_price_positive CHECK (price > 0);
ALTER TABLE purchase_orders ADD CONSTRAINT chk_po_quantity_positive CHECK (quantity > 0);
ALTER TABLE requisition_items ADD CONSTRAINT chk_req_qty_positive CHECK (quantity > 0);

-- Ensure valid status values
ALTER TABLE purchase_orders ADD CONSTRAINT chk_po_status
    CHECK (status IN ('pending', 'approved', 'ordered', 'shipped', 'received', 'cancelled'));

ALTER TABLE projects ADD CONSTRAINT chk_project_status
    CHECK (status IN ('planning', 'active', 'completed', 'cancelled'));

-- Ensure logical date ordering
ALTER TABLE purchase_orders ADD CONSTRAINT chk_po_delivery_dates
    CHECK (expected_delivery IS NULL OR actual_delivery IS NULL OR actual_delivery >= expected_delivery);
```

---

## 10. Recommendations Summary

### Immediate Actions (Critical)
1. ✅ Add `PurchaseOrder` model to complete procurement workflow
2. ✅ Enhance `Vendor` with contact information
3. ✅ Add product SKU and lifecycle fields
4. ✅ Fix `Quote.IsStale()` logic

### Short-term (High Priority)
5. Add quote versioning for negotiation tracking
6. Implement audit fields (CreatedBy, UpdatedBy)
7. Add document attachment support
8. Add product minimum order quantities and lead times

### Long-term (Nice to Have)
9. Vendor performance ratings
10. Approval workflow system
11. Specification versioning
12. Budget tracking enhancements

### Code Quality
- Add database constraint checks for positive values
- Add enum validation for status fields
- Implement soft delete support
- Add comprehensive indexes for common queries

---

## Conclusion

The current model provides a solid foundation for quote comparison, but lacks critical features for a complete procurement system. The most urgent gap is **purchase order tracking** - without it, the system can help compare quotes but cannot track actual purchases or fulfillment.

Adding the Phase 1 enhancements would transform this from a quote comparison tool into a functional procurement system suitable for production use.
