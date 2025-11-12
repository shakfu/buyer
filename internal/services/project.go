package services

import (
	"errors"
	"strings"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService creates a new project service
func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// Create creates a new project with an associated Bill of Materials
func (s *ProjectService) Create(name, description string, budget float64, deadline *time.Time) (*models.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "project name cannot be empty"}
	}

	description = strings.TrimSpace(description)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Check for duplicate
	var existing models.Project
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Project", Name: name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	project := &models.Project{
		Name:        name,
		Description: description,
		Budget:      budget,
		Deadline:    deadline,
		Status:      "planning",
	}

	// Use transaction to create project and BOM together
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create project
		if err := tx.Create(project).Error; err != nil {
			return err
		}

		// Create associated Bill of Materials
		bom := &models.BillOfMaterials{
			ProjectID: project.ID,
		}
		if err := tx.Create(bom).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with BOM
	return s.GetByID(project.ID)
}

// GetByID retrieves a project by ID with preloaded relationships
func (s *ProjectService) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("BillOfMaterials.Items.Specification").
		Preload("Requisitions.Items.BOMItem.Specification").
		First(&project, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Project", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName retrieves a project by name
func (s *ProjectService) GetByName(name string) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("BillOfMaterials.Items.Specification").
		Preload("Requisitions.Items.BOMItem.Specification").
		Where("name = ?", name).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Project", ID: name}
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// List retrieves all projects with optional pagination
func (s *ProjectService) List(limit, offset int) ([]models.Project, error) {
	var projects []models.Project
	query := s.db.Preload("BillOfMaterials").Preload("Requisitions").Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

// Update updates a project's details
func (s *ProjectService) Update(id uint, name, description string, budget float64, deadline *time.Time, status string) (*models.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "project name cannot be empty"}
	}

	description = strings.TrimSpace(description)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Validate status
	validStatuses := map[string]bool{
		"planning":  true,
		"active":    true,
		"completed": true,
		"cancelled": true,
	}
	if status != "" && !validStatuses[status] {
		return nil, &ValidationError{Field: "status", Message: "status must be one of: planning, active, completed, cancelled"}
	}

	// Check if project exists
	project, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check for duplicate name (excluding current project)
	var existing models.Project
	err = s.db.Where("name = ? AND id != ?", name, id).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "Project", Name: name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Update fields
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"budget":      budget,
	}

	if deadline != nil {
		updates["deadline"] = deadline
	}

	if status != "" {
		updates["status"] = status
	}

	if err := s.db.Model(project).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete deletes a project (and cascades to BOM and ProjectRequisitions)
func (s *ProjectService) Delete(id uint) error {
	project, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.db.Delete(project).Error; err != nil {
		return err
	}

	return nil
}

// Count returns the total number of projects
func (s *ProjectService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Project{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AddBillOfMaterialsItem adds an item to the project's Bill of Materials
func (s *ProjectService) AddBillOfMaterialsItem(projectID, specificationID uint, quantity int, notes string) (*models.BillOfMaterialsItem, error) {
	if quantity <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}

	// Verify project exists and get its BOM
	project, err := s.GetByID(projectID)
	if err != nil {
		return nil, err
	}

	if project.BillOfMaterials == nil {
		return nil, &ValidationError{Field: "project", Message: "project has no Bill of Materials"}
	}

	// Verify specification exists
	var spec models.Specification
	if err := s.db.First(&spec, specificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Specification", ID: specificationID}
		}
		return nil, err
	}

	// Check for duplicate specification in this BOM
	var existing models.BillOfMaterialsItem
	err = s.db.Where("bill_of_materials_id = ? AND specification_id = ?",
		project.BillOfMaterials.ID, specificationID).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "BillOfMaterialsItem", Name: spec.Name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create the BOM item
	bomItem := &models.BillOfMaterialsItem{
		BillOfMaterialsID: project.BillOfMaterials.ID,
		SpecificationID:   specificationID,
		Quantity:          quantity,
		Notes:             strings.TrimSpace(notes),
	}

	if err := s.db.Create(bomItem).Error; err != nil {
		return nil, err
	}

	// Reload with specification
	if err := s.db.Preload("Specification").First(bomItem, bomItem.ID).Error; err != nil {
		return nil, err
	}

	return bomItem, nil
}

// UpdateBillOfMaterialsItem updates a BOM item's quantity and notes
func (s *ProjectService) UpdateBillOfMaterialsItem(itemID uint, quantity int, notes string) (*models.BillOfMaterialsItem, error) {
	if quantity <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}

	var bomItem models.BillOfMaterialsItem
	if err := s.db.First(&bomItem, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "BillOfMaterialsItem", ID: itemID}
		}
		return nil, err
	}

	bomItem.Quantity = quantity
	bomItem.Notes = strings.TrimSpace(notes)

	if err := s.db.Save(&bomItem).Error; err != nil {
		return nil, err
	}

	// Reload with specification
	if err := s.db.Preload("Specification").First(&bomItem, itemID).Error; err != nil {
		return nil, err
	}

	return &bomItem, nil
}

// DeleteBillOfMaterialsItem removes an item from a Bill of Materials
func (s *ProjectService) DeleteBillOfMaterialsItem(itemID uint) error {
	var bomItem models.BillOfMaterialsItem
	if err := s.db.First(&bomItem, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &NotFoundError{Entity: "BillOfMaterialsItem", ID: itemID}
		}
		return err
	}

	if err := s.db.Delete(&bomItem).Error; err != nil {
		return err
	}

	return nil
}

