package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// ProductService handles business logic for products
type ProductService struct {
	db *gorm.DB
}

// NewProductService creates a new product service
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// Create creates a new product
func (s *ProductService) Create(name string, brandID uint, specificationID *uint) (*models.Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "product name cannot be empty"}
	}

	// Verify brand exists
	var brand models.Brand
	if err := s.db.First(&brand, brandID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Brand", ID: brandID}
		}
		return nil, err
	}

	// Verify specification exists if provided
	if specificationID != nil && *specificationID > 0 {
		var spec models.Specification
		if err := s.db.First(&spec, *specificationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &NotFoundError{Entity: "Specification", ID: *specificationID}
			}
			return nil, err
		}
	}

	// Check for duplicate
	var existing models.Product
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Product", Name: name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	product := &models.Product{
		Name:            name,
		BrandID:         brandID,
		SpecificationID: specificationID,
	}

	if err := s.db.Create(product).Error; err != nil {
		return nil, err
	}

	// Reload with brand and specification
	return s.GetByID(product.ID)
}

// GetByID retrieves a product by ID with preloaded relationships
func (s *ProductService) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := s.db.Preload("Brand").Preload("Specification").Preload("Quotes").First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Product", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByName retrieves a product by name
func (s *ProductService) GetByName(name string) (*models.Product, error) {
	var product models.Product
	err := s.db.Preload("Brand").Preload("Specification").Where("name = ?", name).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Product", ID: name}
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// List retrieves all products with optional pagination
func (s *ProductService) List(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	query := s.db.Preload("Brand").Preload("Specification").Preload("Quotes").Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&products).Error
	return products, err
}

// ListByBrand retrieves all products for a specific brand
func (s *ProductService) ListByBrand(brandID uint) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Preload("Brand").Preload("Specification").Where("brand_id = ?", brandID).Order("name ASC").Find(&products).Error
	return products, err
}

// ListBySpecification retrieves all products for a specific specification
func (s *ProductService) ListBySpecification(specificationID uint) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Preload("Brand").Preload("Specification").Where("specification_id = ?", specificationID).Order("name ASC").Find(&products).Error
	return products, err
}

// Update updates a product's name and specification
func (s *ProductService) Update(id uint, newName string, specificationID *uint) (*models.Product, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, &ValidationError{Field: "name", Message: "product name cannot be empty"}
	}

	// Check if product exists
	product, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verify specification exists if provided
	if specificationID != nil && *specificationID > 0 {
		var spec models.Specification
		if err := s.db.First(&spec, *specificationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &NotFoundError{Entity: "Specification", ID: *specificationID}
			}
			return nil, err
		}
	}

	// Check for duplicate name
	var existing models.Product
	err = s.db.Where("name = ? AND id != ?", newName, id).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Product", Name: newName}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	product.Name = newName
	product.SpecificationID = specificationID

	if err := s.db.Save(product).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete deletes a product by ID
func (s *ProductService) Delete(id uint) error {
	result := s.db.Delete(&models.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Product", ID: id}
	}
	return nil
}

// Count returns the total number of products
func (s *ProductService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Product{}).Count(&count).Error
	return count, err
}
