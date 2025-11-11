package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// BrandService handles business logic for brands
type BrandService struct {
	db *gorm.DB
}

// NewBrandService creates a new brand service
func NewBrandService(db *gorm.DB) *BrandService {
	return &BrandService{db: db}
}

// Create creates a new brand
func (s *BrandService) Create(name string) (*models.Brand, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "brand name cannot be empty"}
	}

	// Check for duplicate
	var existing models.Brand
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Brand", Name: name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	brand := &models.Brand{Name: name}
	if err := s.db.Create(brand).Error; err != nil {
		return nil, err
	}

	return brand, nil
}

// GetByID retrieves a brand by ID with preloaded relationships
func (s *BrandService) GetByID(id uint) (*models.Brand, error) {
	var brand models.Brand
	err := s.db.Preload("Vendors").Preload("Products").First(&brand, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Brand", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// GetByName retrieves a brand by name
func (s *BrandService) GetByName(name string) (*models.Brand, error) {
	var brand models.Brand
	err := s.db.Where("name = ?", name).First(&brand).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Brand", ID: name}
	}
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// List retrieves all brands with optional pagination
func (s *BrandService) List(limit, offset int) ([]models.Brand, error) {
	var brands []models.Brand
	query := s.db.Preload("Vendors").Preload("Products").Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&brands).Error
	return brands, err
}

// Update updates a brand's name
func (s *BrandService) Update(id uint, newName string) (*models.Brand, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, &ValidationError{Field: "name", Message: "brand name cannot be empty"}
	}

	// Check if brand exists
	brand, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check for duplicate name
	var existing models.Brand
	err = s.db.Where("name = ? AND id != ?", newName, id).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Brand", Name: newName}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	brand.Name = newName
	if err := s.db.Save(brand).Error; err != nil {
		return nil, err
	}

	return brand, nil
}

// Delete deletes a brand by ID
func (s *BrandService) Delete(id uint) error {
	result := s.db.Delete(&models.Brand{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Brand", ID: id}
	}
	return nil
}

// Count returns the total number of brands
func (s *BrandService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Brand{}).Count(&count).Error
	return count, err
}
