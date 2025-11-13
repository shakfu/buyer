package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// VendorService handles business logic for vendors
type VendorService struct {
	db *gorm.DB
}

// NewVendorService creates a new vendor service
func NewVendorService(db *gorm.DB) *VendorService {
	return &VendorService{db: db}
}

// Create creates a new vendor
func (s *VendorService) Create(name, currency, discountCode string) (*models.Vendor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "vendor name cannot be empty"}
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}
	if len(currency) != 3 {
		return nil, &ValidationError{Field: "currency", Message: "currency must be a 3-letter ISO 4217 code"}
	}

	// Check for duplicate
	var existing models.Vendor
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Vendor", Name: name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	vendor := &models.Vendor{
		Name:         name,
		Currency:     currency,
		DiscountCode: strings.TrimSpace(discountCode),
	}
	if err := s.db.Create(vendor).Error; err != nil {
		return nil, err
	}

	return vendor, nil
}

// GetByID retrieves a vendor by ID with preloaded relationships
func (s *VendorService) GetByID(id uint) (*models.Vendor, error) {
	var vendor models.Vendor
	err := s.db.Preload("Brands").Preload("Quotes").Preload("VendorRatings").
		Preload("PurchaseOrders").First(&vendor, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Vendor", ID: id}
	}
	if err != nil {
		return nil, err
	}

	// Load documents separately (polymorphic relationship)
	s.loadDocuments(&vendor)

	return &vendor, nil
}

// GetByName retrieves a vendor by name
func (s *VendorService) GetByName(name string) (*models.Vendor, error) {
	var vendor models.Vendor
	err := s.db.Where("name = ?", name).First(&vendor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Vendor", ID: name}
	}
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// List retrieves all vendors with optional pagination
func (s *VendorService) List(limit, offset int) ([]models.Vendor, error) {
	var vendors []models.Vendor
	query := s.db.Preload("Brands").Preload("Quotes").Preload("VendorRatings").
		Preload("PurchaseOrders").Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&vendors).Error
	if err != nil {
		return vendors, err
	}

	// Load documents for all vendors
	for i := range vendors {
		s.loadDocuments(&vendors[i])
	}

	return vendors, nil
}

// Update updates a vendor's information
func (s *VendorService) Update(id uint, newName string) (*models.Vendor, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, &ValidationError{Field: "name", Message: "vendor name cannot be empty"}
	}

	// Check if vendor exists
	vendor, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check for duplicate name
	var existing models.Vendor
	err = s.db.Where("name = ? AND id != ?", newName, id).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Vendor", Name: newName}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	vendor.Name = newName
	if err := s.db.Save(vendor).Error; err != nil {
		return nil, err
	}

	return vendor, nil
}

// Delete deletes a vendor by ID
func (s *VendorService) Delete(id uint) error {
	result := s.db.Delete(&models.Vendor{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Vendor", ID: id}
	}
	return nil
}

// AddBrand adds a brand association to a vendor
func (s *VendorService) AddBrand(vendorID, brandID uint) error {
	vendor, err := s.GetByID(vendorID)
	if err != nil {
		return err
	}

	var brand models.Brand
	if err := s.db.First(&brand, brandID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &NotFoundError{Entity: "Brand", ID: brandID}
		}
		return err
	}

	return s.db.Model(vendor).Association("Brands").Append(&brand)
}

// RemoveBrand removes a brand association from a vendor
func (s *VendorService) RemoveBrand(vendorID, brandID uint) error {
	vendor, err := s.GetByID(vendorID)
	if err != nil {
		return err
	}

	var brand models.Brand
	if err := s.db.First(&brand, brandID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &NotFoundError{Entity: "Brand", ID: brandID}
		}
		return err
	}

	return s.db.Model(vendor).Association("Brands").Delete(&brand)
}

// Count returns the total number of vendors
func (s *VendorService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Vendor{}).Count(&count).Error
	return count, err
}

// loadDocuments loads documents for a vendor (polymorphic relationship)
func (s *VendorService) loadDocuments(vendor *models.Vendor) {
	var docs []models.Document
	s.db.Where("entity_type = ? AND entity_id = ?", "vendor", vendor.ID).Find(&docs)
	vendor.Documents = docs
}
