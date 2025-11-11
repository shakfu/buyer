package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// SpecificationService handles business logic for specifications
type SpecificationService struct {
	db *gorm.DB
}

// NewSpecificationService creates a new specification service
func NewSpecificationService(db *gorm.DB) *SpecificationService {
	return &SpecificationService{db: db}
}

// Create creates a new specification
func (s *SpecificationService) Create(name, description string) (*models.Specification, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name cannot be empty"}
	}

	// Check for duplicates
	var existing models.Specification
	if err := s.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, &DuplicateError{Entity: "Specification", Name: name}
	}

	spec := &models.Specification{
		Name:        name,
		Description: description,
	}

	if err := s.db.Create(spec).Error; err != nil {
		return nil, err
	}

	return spec, nil
}

// GetByID retrieves a specification by ID
func (s *SpecificationService) GetByID(id uint) (*models.Specification, error) {
	var spec models.Specification
	err := s.db.Preload("Products").First(&spec, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Specification", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

// Update updates a specification's name and description
func (s *SpecificationService) Update(id uint, name, description string) (*models.Specification, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name cannot be empty"}
	}

	// Check if specification exists
	var spec models.Specification
	if err := s.db.First(&spec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Specification", ID: id}
		}
		return nil, err
	}

	// Check for duplicate name (excluding current record)
	var existing models.Specification
	if err := s.db.Where("name = ? AND id != ?", name, id).First(&existing).Error; err == nil {
		return nil, &DuplicateError{Entity: "Specification", Name: name}
	}

	spec.Name = name
	spec.Description = description

	if err := s.db.Save(&spec).Error; err != nil {
		return nil, err
	}

	return &spec, nil
}

// Delete deletes a specification by ID
func (s *SpecificationService) Delete(id uint) error {
	result := s.db.Delete(&models.Specification{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Specification", ID: id}
	}
	return nil
}

// List retrieves all specifications with optional pagination
func (s *SpecificationService) List(limit, offset int) ([]models.Specification, error) {
	var specs []models.Specification
	query := s.db.Order("name")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&specs).Error
	return specs, err
}
