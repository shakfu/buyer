package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// ProjectRequisitionService handles business logic for project-based requisitions
type ProjectRequisitionService struct {
	db *gorm.DB
}

// NewProjectRequisitionService creates a new project requisition service
func NewProjectRequisitionService(db *gorm.DB) *ProjectRequisitionService {
	return &ProjectRequisitionService{db: db}
}

// ProjectRequisitionItemInput represents input for creating a project requisition item
type ProjectRequisitionItemInput struct {
	BOMItemID         uint
	QuantityRequested int
	Notes             string
}

// Create creates a new project requisition from BOM items
func (s *ProjectRequisitionService) Create(projectID uint, name, justification string, budget float64, items []ProjectRequisitionItemInput) (*models.ProjectRequisition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "requisition name cannot be empty"}
	}

	justification = strings.TrimSpace(justification)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Verify project exists
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Project", ID: projectID}
		}
		return nil, err
	}

	// Validate items
	if len(items) == 0 {
		return nil, &ValidationError{Field: "items", Message: "at least one BOM item is required"}
	}

	for i, item := range items {
		if item.QuantityRequested <= 0 {
			return nil, &ValidationError{Field: "items", Message: "quantity must be greater than 0 for all items"}
		}

		// Verify BOM item exists and belongs to this project
		var bomItem models.BillOfMaterialsItem
		err := s.db.Preload("BillOfMaterials").First(&bomItem, item.BOMItemID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &NotFoundError{Entity: "BillOfMaterialsItem", ID: item.BOMItemID}
			}
			return nil, err
		}

		if bomItem.BillOfMaterials.ProjectID != projectID {
			return nil, &ValidationError{Field: "items", Message: "BOM item does not belong to this project"}
		}

		// Validate quantity doesn't exceed BOM item quantity
		if item.QuantityRequested > bomItem.Quantity {
			return nil, &ValidationError{
				Field:   "items",
				Message: "quantity requested exceeds BOM item quantity",
			}
		}

		items[i].Notes = strings.TrimSpace(item.Notes)
	}

	// Create requisition with items in a transaction
	var requisition *models.ProjectRequisition
	err := s.db.Transaction(func(tx *gorm.DB) error {
		requisition = &models.ProjectRequisition{
			ProjectID:     projectID,
			Name:          name,
			Justification: justification,
			Budget:        budget,
		}

		if err := tx.Create(requisition).Error; err != nil {
			return err
		}

		// Create items
		for _, item := range items {
			reqItem := &models.ProjectRequisitionItem{
				ProjectRequisitionID:  requisition.ID,
				BillOfMaterialsItemID: item.BOMItemID,
				QuantityRequested:     item.QuantityRequested,
				Notes:                 item.Notes,
			}
			if err := tx.Create(reqItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with items
	return s.GetByID(requisition.ID)
}

// GetByID retrieves a project requisition by ID with preloaded relationships
func (s *ProjectRequisitionService) GetByID(id uint) (*models.ProjectRequisition, error) {
	var requisition models.ProjectRequisition
	err := s.db.Preload("Items.BOMItem.Specification").
		Preload("Project").
		First(&requisition, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "ProjectRequisition", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &requisition, nil
}

// List retrieves all project requisitions with optional pagination
func (s *ProjectRequisitionService) List(limit, offset int) ([]models.ProjectRequisition, error) {
	var requisitions []models.ProjectRequisition
	query := s.db.Preload("Items.BOMItem.Specification").
		Preload("Project").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&requisitions).Error; err != nil {
		return nil, err
	}

	return requisitions, nil
}

// ListByProject retrieves all project requisitions for a specific project
func (s *ProjectRequisitionService) ListByProject(projectID uint) ([]models.ProjectRequisition, error) {
	var requisitions []models.ProjectRequisition
	err := s.db.Where("project_id = ?", projectID).
		Preload("Items.BOMItem.Specification").
		Order("created_at DESC").
		Find(&requisitions).Error
	if err != nil {
		return nil, err
	}
	return requisitions, nil
}

// Update updates a project requisition's basic details (not items)
func (s *ProjectRequisitionService) Update(id uint, name, justification string, budget float64) (*models.ProjectRequisition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "requisition name cannot be empty"}
	}

	justification = strings.TrimSpace(justification)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Check if requisition exists
	requisition, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	updates := map[string]interface{}{
		"name":          name,
		"justification": justification,
		"budget":        budget,
	}

	if err := s.db.Model(requisition).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete deletes a project requisition (cascades to items)
func (s *ProjectRequisitionService) Delete(id uint) error {
	requisition, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.db.Delete(requisition).Error; err != nil {
		return err
	}

	return nil
}

// AddItem adds an item to an existing project requisition
func (s *ProjectRequisitionService) AddItem(requisitionID, bomItemID uint, quantityRequested int, notes string) (*models.ProjectRequisitionItem, error) {
	if quantityRequested <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}

	// Verify requisition exists
	requisition, err := s.GetByID(requisitionID)
	if err != nil {
		return nil, err
	}

	// Verify BOM item exists and belongs to the same project
	var bomItem models.BillOfMaterialsItem
	err = s.db.Preload("BillOfMaterials").First(&bomItem, bomItemID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "BillOfMaterialsItem", ID: bomItemID}
		}
		return nil, err
	}

	if bomItem.BillOfMaterials.ProjectID != requisition.ProjectID {
		return nil, &ValidationError{Field: "bom_item", Message: "BOM item does not belong to this project"}
	}

	if quantityRequested > bomItem.Quantity {
		return nil, &ValidationError{Field: "quantity", Message: "quantity requested exceeds BOM item quantity"}
	}

	// Check if item already exists in this requisition
	var existing models.ProjectRequisitionItem
	err = s.db.Where("project_requisition_id = ? AND bill_of_materials_item_id = ?", requisitionID, bomItemID).
		First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "ProjectRequisitionItem", Name: "BOM item already in requisition"}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create item
	item := &models.ProjectRequisitionItem{
		ProjectRequisitionID:  requisitionID,
		BillOfMaterialsItemID: bomItemID,
		QuantityRequested:     quantityRequested,
		Notes:                 strings.TrimSpace(notes),
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// Reload with BOM item
	if err := s.db.Preload("BOMItem.Specification").First(item, item.ID).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateItem updates a project requisition item
func (s *ProjectRequisitionService) UpdateItem(itemID uint, quantityRequested int, notes string) (*models.ProjectRequisitionItem, error) {
	if quantityRequested <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}

	var item models.ProjectRequisitionItem
	if err := s.db.Preload("BOMItem").First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "ProjectRequisitionItem", ID: itemID}
		}
		return nil, err
	}

	// Validate quantity
	if quantityRequested > item.BOMItem.Quantity {
		return nil, &ValidationError{Field: "quantity", Message: "quantity requested exceeds BOM item quantity"}
	}

	item.QuantityRequested = quantityRequested
	item.Notes = strings.TrimSpace(notes)

	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}

	// Reload with specification
	if err := s.db.Preload("BOMItem.Specification").First(&item, itemID).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// DeleteItem removes an item from a project requisition
func (s *ProjectRequisitionService) DeleteItem(itemID uint) error {
	var item models.ProjectRequisitionItem
	if err := s.db.First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &NotFoundError{Entity: "ProjectRequisitionItem", ID: itemID}
		}
		return err
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return err
	}

	return nil
}

// Count returns the total number of project requisitions
func (s *ProjectRequisitionService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.ProjectRequisition{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
