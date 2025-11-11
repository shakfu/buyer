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
	Quotes       []Quote   `gorm:"foreignKey:VendorID" json:"quotes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Brand represents a manufacturing entity
type Brand struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Vendors   []*Vendor `gorm:"many2many:vendor_brands;" json:"vendors,omitempty"`
	Products  []Product `gorm:"foreignKey:BrandID" json:"products,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product represents an item associated with a brand
type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	BrandID   uint      `gorm:"not null;index" json:"brand_id"`
	Brand     *Brand    `gorm:"foreignKey:BrandID" json:"brand,omitempty"`
	Quotes    []Quote   `gorm:"foreignKey:ProductID" json:"quotes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Quote represents a price quote from a vendor for a product
type Quote struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	VendorID         uint      `gorm:"not null;index" json:"vendor_id"`
	Vendor           *Vendor   `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
	ProductID        uint      `gorm:"not null;index" json:"product_id"`
	Product          *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Price            float64   `gorm:"not null" json:"price"`
	Currency         string    `gorm:"size:3;not null" json:"currency"`
	ConvertedPrice   float64   `gorm:"not null" json:"converted_price"` // Price in USD
	ConversionRate   float64   `gorm:"not null" json:"conversion_rate"`
	QuoteDate        time.Time `gorm:"not null;index" json:"quote_date"`
	Notes            string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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

// TableName overrides for GORM
func (Vendor) TableName() string  { return "vendors" }
func (Brand) TableName() string   { return "brands" }
func (Product) TableName() string { return "products" }
func (Quote) TableName() string   { return "quotes" }
func (Forex) TableName() string   { return "forex" }

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
