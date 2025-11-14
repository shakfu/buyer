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

// QuoteAttributeComparison represents a quote with attribute compliance information
type QuoteAttributeComparison struct {
	Quote                *models.Quote
	AttributeCompliance  map[uint]bool // attribute_id -> has_value
	MissingRequiredAttrs []string      // names of missing required attributes
	HasAllRequiredAttrs  bool
	ExtraAttributes      []models.ProductAttribute // attributes not in specification
	ComplianceScore      float64                   // 0-100, percentage of required attrs present
}

// AttributeComparisonMatrix represents a full comparison matrix for quotes
type AttributeComparisonMatrix struct {
	Specification       *models.Specification
	SpecificationAttrs  []models.SpecificationAttribute
	QuoteComparisons    []QuoteAttributeComparison
	ShowExtraAttributes bool
}

// GetQuoteComparisonMatrix creates a detailed comparison matrix for quotes of a specification
func (s *QuoteService) GetQuoteComparisonMatrix(specificationID uint, showExtraAttrs bool) (*AttributeComparisonMatrix, error) {
	// Get specification
	var spec models.Specification
	if err := s.db.First(&spec, specificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Specification", ID: specificationID}
		}
		return nil, err
	}

	// Get specification attributes
	var specAttrs []models.SpecificationAttribute
	if err := s.db.Where("specification_id = ?", specificationID).
		Order("is_required DESC, name ASC").Find(&specAttrs).Error; err != nil {
		return nil, err
	}

	// Get all active quotes for products matching this specification
	var quotes []models.Quote
	if err := s.db.Preload("Vendor").
		Preload("Product.Brand").
		Preload("Product.Specification").
		Preload("Product.Attributes.SpecificationAttribute").
		Joins("JOIN products ON products.id = quotes.product_id").
		Where("products.specification_id = ?", specificationID).
		Where("quotes.valid_until IS NULL OR quotes.valid_until > ?", time.Now()).
		Order("quotes.converted_price ASC").
		Find(&quotes).Error; err != nil {
		return nil, err
	}

	// Build comparison data
	comparisons := make([]QuoteAttributeComparison, 0, len(quotes))
	for _, quote := range quotes {
		comparison := s.analyzeQuoteCompliance(&quote, specAttrs, showExtraAttrs)
		comparisons = append(comparisons, comparison)
	}

	return &AttributeComparisonMatrix{
		Specification:       &spec,
		SpecificationAttrs:  specAttrs,
		QuoteComparisons:    comparisons,
		ShowExtraAttributes: showExtraAttrs,
	}, nil
}

// GetProductQuoteComparisonMatrix creates a comparison matrix for quotes of a single product
func (s *QuoteService) GetProductQuoteComparisonMatrix(productID uint, showExtraAttrs bool) (*AttributeComparisonMatrix, error) {
	// Get product with specification
	var product models.Product
	if err := s.db.Preload("Specification").
		Preload("Attributes.SpecificationAttribute").
		First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Product", ID: productID}
		}
		return nil, err
	}

	if product.SpecificationID == nil {
		return nil, &ValidationError{Message: "Product has no specification"}
	}

	// Get specification attributes
	var specAttrs []models.SpecificationAttribute
	if err := s.db.Where("specification_id = ?", *product.SpecificationID).
		Order("is_required DESC, name ASC").Find(&specAttrs).Error; err != nil {
		return nil, err
	}

	// Get all active quotes for this product
	var quotes []models.Quote
	if err := s.db.Preload("Vendor").
		Preload("Product.Brand").
		Preload("Product.Specification").
		Preload("Product.Attributes.SpecificationAttribute").
		Where("product_id = ?", productID).
		Where("valid_until IS NULL OR valid_until > ?", time.Now()).
		Order("converted_price ASC").
		Find(&quotes).Error; err != nil {
		return nil, err
	}

	// Build comparison data
	comparisons := make([]QuoteAttributeComparison, 0, len(quotes))
	for _, quote := range quotes {
		comparison := s.analyzeQuoteCompliance(&quote, specAttrs, showExtraAttrs)
		comparisons = append(comparisons, comparison)
	}

	return &AttributeComparisonMatrix{
		Specification:       product.Specification,
		SpecificationAttrs:  specAttrs,
		QuoteComparisons:    comparisons,
		ShowExtraAttributes: showExtraAttrs,
	}, nil
}

// analyzeQuoteCompliance checks if a product meets specification attribute requirements
func (s *QuoteService) analyzeQuoteCompliance(quote *models.Quote, specAttrs []models.SpecificationAttribute, includeExtra bool) QuoteAttributeComparison {
	compliance := QuoteAttributeComparison{
		Quote:                quote,
		AttributeCompliance:  make(map[uint]bool),
		MissingRequiredAttrs: make([]string, 0),
		ExtraAttributes:      make([]models.ProductAttribute, 0),
		HasAllRequiredAttrs:  true,
	}

	// Build map of product attributes by specification_attribute_id
	productAttrMap := make(map[uint]*models.ProductAttribute)
	for i := range quote.Product.Attributes {
		attr := &quote.Product.Attributes[i]
		productAttrMap[attr.SpecificationAttributeID] = attr
	}

	// Check each specification attribute
	requiredCount := 0
	presentCount := 0

	for _, specAttr := range specAttrs {
		prodAttr, exists := productAttrMap[specAttr.ID]
		hasValue := exists && (prodAttr.ValueNumber != nil || prodAttr.ValueText != nil || prodAttr.ValueBoolean != nil)

		compliance.AttributeCompliance[specAttr.ID] = hasValue

		if specAttr.IsRequired {
			requiredCount++
			if hasValue {
				presentCount++
			} else {
				compliance.MissingRequiredAttrs = append(compliance.MissingRequiredAttrs, specAttr.Name)
				compliance.HasAllRequiredAttrs = false
			}
		}
	}

	// Calculate compliance score
	if requiredCount > 0 {
		compliance.ComplianceScore = (float64(presentCount) / float64(requiredCount)) * 100
	} else {
		compliance.ComplianceScore = 100.0
	}

	// Find extra attributes if requested
	if includeExtra {
		specAttrIDs := make(map[uint]bool)
		for _, specAttr := range specAttrs {
			specAttrIDs[specAttr.ID] = true
		}

		for i := range quote.Product.Attributes {
			attr := &quote.Product.Attributes[i]
			if !specAttrIDs[attr.SpecificationAttributeID] {
				compliance.ExtraAttributes = append(compliance.ExtraAttributes, *attr)
			}
		}
	}

	return compliance
}
