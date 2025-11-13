package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// DocumentService handles business logic for documents
type DocumentService struct {
	db *gorm.DB
}

// NewDocumentService creates a new document service
func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{db: db}
}

// CreateDocumentInput represents input for creating a document
type CreateDocumentInput struct {
	EntityType  string
	EntityID    uint
	FileName    string
	FileType    string
	FileSize    int64
	FilePath    string
	Description string
	UploadedBy  string
}

// Create creates a new document
func (s *DocumentService) Create(input CreateDocumentInput) (*models.Document, error) {
	// Validate required fields
	entityType := strings.TrimSpace(input.EntityType)
	if entityType == "" {
		return nil, &ValidationError{Field: "entity_type", Message: "entity type cannot be empty"}
	}

	fileName := strings.TrimSpace(input.FileName)
	if fileName == "" {
		return nil, &ValidationError{Field: "file_name", Message: "file name cannot be empty"}
	}

	filePath := strings.TrimSpace(input.FilePath)
	if filePath == "" {
		return nil, &ValidationError{Field: "file_path", Message: "file path cannot be empty"}
	}

	// Validate entity type
	validEntityTypes := map[string]bool{
		"vendor":         true,
		"brand":          true,
		"product":        true,
		"quote":          true,
		"purchase_order": true,
		"requisition":    true,
		"project":        true,
	}

	if !validEntityTypes[entityType] {
		return nil, &ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type: must be one of vendor, brand, product, quote, purchase_order, requisition, project",
		}
	}

	// Validate that the entity exists (basic check - could be enhanced per entity type)
	if input.EntityID == 0 {
		return nil, &ValidationError{Field: "entity_id", Message: "entity ID must be greater than 0"}
	}

	doc := &models.Document{
		EntityType:  entityType,
		EntityID:    input.EntityID,
		FileName:    fileName,
		FileType:    strings.TrimSpace(input.FileType),
		FileSize:    input.FileSize,
		FilePath:    filePath,
		Description: strings.TrimSpace(input.Description),
		UploadedBy:  strings.TrimSpace(input.UploadedBy),
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// GetByID retrieves a document by ID
func (s *DocumentService) GetByID(id uint) (*models.Document, error) {
	var doc models.Document
	if err := s.db.First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Document", ID: id}
		}
		return nil, err
	}
	return &doc, nil
}

// List retrieves all documents with pagination
func (s *DocumentService) List(limit, offset int) ([]*models.Document, error) {
	var docs []*models.Document
	query := s.db.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// ListByEntity retrieves all documents for a specific entity
func (s *DocumentService) ListByEntity(entityType string, entityID uint) ([]*models.Document, error) {
	var docs []*models.Document
	if err := s.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// ListByEntityType retrieves all documents for a specific entity type
func (s *DocumentService) ListByEntityType(entityType string, limit, offset int) ([]*models.Document, error) {
	var docs []*models.Document
	query := s.db.Where("entity_type = ?", entityType).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// Update updates a document's information
func (s *DocumentService) Update(id uint, input CreateDocumentInput) (*models.Document, error) {
	var doc models.Document
	if err := s.db.First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Document", ID: id}
		}
		return nil, err
	}

	// Update fields if provided
	if input.FileName != "" {
		doc.FileName = strings.TrimSpace(input.FileName)
	}
	if input.FileType != "" {
		doc.FileType = strings.TrimSpace(input.FileType)
	}
	if input.FileSize > 0 {
		doc.FileSize = input.FileSize
	}
	if input.FilePath != "" {
		doc.FilePath = strings.TrimSpace(input.FilePath)
	}
	doc.Description = strings.TrimSpace(input.Description)
	if input.UploadedBy != "" {
		doc.UploadedBy = strings.TrimSpace(input.UploadedBy)
	}

	if err := s.db.Save(&doc).Error; err != nil {
		return nil, err
	}

	return &doc, nil
}

// Delete deletes a document
func (s *DocumentService) Delete(id uint) error {
	result := s.db.Delete(&models.Document{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Document", ID: id}
	}
	return nil
}

// DeleteByEntity deletes all documents for a specific entity
func (s *DocumentService) DeleteByEntity(entityType string, entityID uint) error {
	return s.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Delete(&models.Document{}).Error
}

// Count returns the total number of documents
func (s *DocumentService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Document{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByEntity returns the count of documents for a specific entity
func (s *DocumentService) CountByEntity(entityType string, entityID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Document{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
