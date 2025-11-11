package services

import (
	"errors"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// QuoteService handles business logic for quotes
type QuoteService struct {
	db           *gorm.DB
	forexService *ForexService
}

// NewQuoteService creates a new quote service
func NewQuoteService(db *gorm.DB) *QuoteService {
	return &QuoteService{
		db:           db,
		forexService: NewForexService(db),
	}
}

// CreateQuoteInput holds the input for creating a quote
type CreateQuoteInput struct {
	VendorID   uint
	ProductID  uint
	Price      float64
	Currency   string
	QuoteDate  time.Time
	ValidUntil *time.Time
	Notes      string
}

// Create creates a new quote with automatic currency conversion
func (s *QuoteService) Create(input CreateQuoteInput) (*models.Quote, error) {
	// Validate vendor exists
	var vendor models.Vendor
	if err := s.db.First(&vendor, input.VendorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Vendor", ID: input.VendorID}
		}
		return nil, err
	}

	// Validate product exists
	var product models.Product
	if err := s.db.First(&product, input.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Product", ID: input.ProductID}
		}
		return nil, err
	}

	if input.Price <= 0 {
		return nil, &ValidationError{Field: "price", Message: "price must be positive"}
	}

	// Use vendor's currency if not specified
	currency := input.Currency
	if currency == "" {
		currency = vendor.Currency
	}

	// Convert to USD for standardized comparison
	convertedPrice, conversionRate, err := s.forexService.Convert(input.Price, currency, "USD")
	if err != nil {
		return nil, err
	}

	quoteDate := input.QuoteDate
	if quoteDate.IsZero() {
		quoteDate = time.Now()
	}

	quote := &models.Quote{
		VendorID:       input.VendorID,
		ProductID:      input.ProductID,
		Price:          input.Price,
		Currency:       currency,
		ConvertedPrice: convertedPrice,
		ConversionRate: conversionRate,
		QuoteDate:      quoteDate,
		ValidUntil:     input.ValidUntil,
		Notes:          input.Notes,
	}

	if err := s.db.Create(quote).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	return s.GetByID(quote.ID)
}

// GetByID retrieves a quote by ID with preloaded relationships
func (s *QuoteService) GetByID(id uint) (*models.Quote, error) {
	var quote models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").First(&quote, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Quote", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

// List retrieves all quotes with optional pagination
func (s *QuoteService) List(limit, offset int) ([]models.Quote, error) {
	var quotes []models.Quote
	query := s.db.Preload("Vendor").Preload("Product.Brand").Order("quote_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&quotes).Error
	return quotes, err
}

// ListByProduct retrieves all quotes for a specific product
func (s *QuoteService) ListByProduct(productID uint) ([]models.Quote, error) {
	var quotes []models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").
		Where("product_id = ?", productID).
		Order("quote_date DESC").
		Find(&quotes).Error
	return quotes, err
}

// ListByVendor retrieves all quotes from a specific vendor
func (s *QuoteService) ListByVendor(vendorID uint) ([]models.Quote, error) {
	var quotes []models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").
		Where("vendor_id = ?", vendorID).
		Order("quote_date DESC").
		Find(&quotes).Error
	return quotes, err
}

// GetBestQuote finds the lowest price quote for a product (in USD)
func (s *QuoteService) GetBestQuote(productID uint) (*models.Quote, error) {
	var quote models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").
		Where("product_id = ?", productID).
		Order("converted_price ASC").
		First(&quote).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Quote for product", ID: productID}
	}
	if err != nil {
		return nil, err
	}

	return &quote, nil
}

// Delete deletes a quote by ID
func (s *QuoteService) Delete(id uint) error {
	result := s.db.Delete(&models.Quote{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Quote", ID: id}
	}
	return nil
}

// Count returns the total number of quotes
func (s *QuoteService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Quote{}).Count(&count).Error
	return count, err
}

// ListActiveQuotes retrieves all non-expired quotes
func (s *QuoteService) ListActiveQuotes(limit, offset int) ([]models.Quote, error) {
	var quotes []models.Quote
	query := s.db.Preload("Vendor").Preload("Product.Brand").
		Where("valid_until IS NULL OR valid_until > ?", time.Now()).
		Order("quote_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&quotes).Error
	return quotes, err
}

// CompareQuotesForProduct retrieves all active quotes for a product with comparison data
func (s *QuoteService) CompareQuotesForProduct(productID uint) ([]models.Quote, error) {
	var quotes []models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").
		Where("product_id = ?", productID).
		Where("valid_until IS NULL OR valid_until > ?", time.Now()).
		Order("converted_price ASC").
		Find(&quotes).Error
	return quotes, err
}

// CompareQuotesForSpecification retrieves all active quotes for products matching a specification
func (s *QuoteService) CompareQuotesForSpecification(specificationID uint) ([]models.Quote, error) {
	var quotes []models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").Preload("Product.Specification").
		Joins("JOIN products ON products.id = quotes.product_id").
		Where("products.specification_id = ?", specificationID).
		Where("quotes.valid_until IS NULL OR quotes.valid_until > ?", time.Now()).
		Order("quotes.converted_price ASC").
		Find(&quotes).Error
	return quotes, err
}

// GetBestQuoteForSpecification finds the lowest price quote for products matching a specification
func (s *QuoteService) GetBestQuoteForSpecification(specificationID uint) (*models.Quote, error) {
	var quote models.Quote
	err := s.db.Preload("Vendor").Preload("Product.Brand").Preload("Product.Specification").
		Joins("JOIN products ON products.id = quotes.product_id").
		Where("products.specification_id = ?", specificationID).
		Where("quotes.valid_until IS NULL OR quotes.valid_until > ?", time.Now()).
		Order("quotes.converted_price ASC").
		First(&quote).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Quote for specification", ID: specificationID}
	}
	if err != nil {
		return nil, err
	}

	return &quote, nil
}
