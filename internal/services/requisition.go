package services

import (
	"errors"
	"strings"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// RequisitionService handles business logic for requisitions
type RequisitionService struct {
	db *gorm.DB
}

// NewRequisitionService creates a new requisition service
func NewRequisitionService(db *gorm.DB) *RequisitionService {
	return &RequisitionService{db: db}
}

// RequisitionItemInput represents a requisition item input
type RequisitionItemInput struct {
	SpecificationID uint
	Quantity        int
	BudgetPerUnit   float64 // Optional budget per unit
	Description     string  // Optional description for details
}

// Create creates a new requisition with items
func (s *RequisitionService) Create(name, justification string, budget float64, items []RequisitionItemInput) (*models.Requisition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name cannot be empty"}
	}

	justification = strings.TrimSpace(justification)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Check for duplicates
	var existing models.Requisition
	if err := s.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, &DuplicateError{Entity: "Requisition", Name: name}
	}

	// Validate each item (if any items are provided)
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, &ValidationError{Field: "quantity", Message: "quantity must be positive"}
		}

		if item.BudgetPerUnit < 0 {
			return nil, &ValidationError{Field: "budget_per_unit", Message: "budget per unit cannot be negative"}
		}

		// Verify specification exists
		var spec models.Specification
		if err := s.db.First(&spec, item.SpecificationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &NotFoundError{Entity: "Specification", ID: item.SpecificationID}
			}
			return nil, err
		}
	}

	// Create requisition with items in a transaction
	requisition := &models.Requisition{
		Name:          name,
		Justification: justification,
		Budget:        budget,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create requisition
		if err := tx.Create(requisition).Error; err != nil {
			return err
		}

		// Create requisition items
		for _, item := range items {
			reqItem := &models.RequisitionItem{
				RequisitionID:   requisition.ID,
				SpecificationID: item.SpecificationID,
				Quantity:        item.Quantity,
				BudgetPerUnit:   item.BudgetPerUnit,
				Description:     strings.TrimSpace(item.Description),
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

	// Reload with relationships
	return s.GetByID(requisition.ID)
}

// GetByID retrieves a requisition by ID with preloaded items
func (s *RequisitionService) GetByID(id uint) (*models.Requisition, error) {
	var requisition models.Requisition
	err := s.db.Preload("Items.Specification").Preload("PurchaseOrders").
		First(&requisition, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Requisition", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &requisition, nil
}

// Update updates a requisition's name, justification, and budget
func (s *RequisitionService) Update(id uint, name, justification string, budget float64) (*models.Requisition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name cannot be empty"}
	}

	justification = strings.TrimSpace(justification)

	if budget < 0 {
		return nil, &ValidationError{Field: "budget", Message: "budget cannot be negative"}
	}

	// Check if requisition exists
	var requisition models.Requisition
	if err := s.db.First(&requisition, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Requisition", ID: id}
		}
		return nil, err
	}

	// Check for duplicate name (excluding current record)
	var existing models.Requisition
	if err := s.db.Where("name = ? AND id != ?", name, id).First(&existing).Error; err == nil {
		return nil, &DuplicateError{Entity: "Requisition", Name: name}
	}

	requisition.Name = name
	requisition.Justification = justification
	requisition.Budget = budget

	if err := s.db.Save(&requisition).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// AddItem adds an item to an existing requisition
func (s *RequisitionService) AddItem(requisitionID, specificationID uint, quantity int, budgetPerUnit float64, description string) (*models.RequisitionItem, error) {
	if quantity <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be positive"}
	}

	if budgetPerUnit < 0 {
		return nil, &ValidationError{Field: "budget_per_unit", Message: "budget per unit cannot be negative"}
	}

	// Verify requisition exists
	var requisition models.Requisition
	if err := s.db.First(&requisition, requisitionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Requisition", ID: requisitionID}
		}
		return nil, err
	}

	// Verify specification exists
	var spec models.Specification
	if err := s.db.First(&spec, specificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Specification", ID: specificationID}
		}
		return nil, err
	}

	item := &models.RequisitionItem{
		RequisitionID:   requisitionID,
		SpecificationID: specificationID,
		Quantity:        quantity,
		BudgetPerUnit:   budgetPerUnit,
		Description:     strings.TrimSpace(description),
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	var reloadedItem models.RequisitionItem
	if err := s.db.Preload("Specification").First(&reloadedItem, item.ID).Error; err != nil {
		return nil, err
	}

	return &reloadedItem, nil
}

// UpdateItem updates a requisition item
func (s *RequisitionService) UpdateItem(itemID, specificationID uint, quantity int, budgetPerUnit float64, description string) (*models.RequisitionItem, error) {
	if quantity <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be positive"}
	}

	if budgetPerUnit < 0 {
		return nil, &ValidationError{Field: "budget_per_unit", Message: "budget per unit cannot be negative"}
	}

	// Check if item exists
	var item models.RequisitionItem
	if err := s.db.First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "RequisitionItem", ID: itemID}
		}
		return nil, err
	}

	// Verify specification exists
	var spec models.Specification
	if err := s.db.First(&spec, specificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Specification", ID: specificationID}
		}
		return nil, err
	}

	item.SpecificationID = specificationID
	item.Quantity = quantity
	item.BudgetPerUnit = budgetPerUnit
	item.Description = strings.TrimSpace(description)

	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	var reloadedItem models.RequisitionItem
	if err := s.db.Preload("Specification").First(&reloadedItem, item.ID).Error; err != nil {
		return nil, err
	}

	return &reloadedItem, nil
}

// DeleteItem removes an item from a requisition
func (s *RequisitionService) DeleteItem(itemID uint) error {
	result := s.db.Delete(&models.RequisitionItem{}, itemID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "RequisitionItem", ID: itemID}
	}
	return nil
}

// Delete deletes a requisition by ID (cascades to items)
func (s *RequisitionService) Delete(id uint) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all items first
		if err := tx.Where("requisition_id = ?", id).Delete(&models.RequisitionItem{}).Error; err != nil {
			return err
		}

		// Delete requisition
		result := tx.Delete(&models.Requisition{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return &NotFoundError{Entity: "Requisition", ID: id}
		}

		return nil
	})

	return err
}

// List retrieves all requisitions with optional pagination
func (s *RequisitionService) List(limit, offset int) ([]models.Requisition, error) {
	var requisitions []models.Requisition
	query := s.db.Preload("Items.Specification").Preload("PurchaseOrders").
		Order("name")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&requisitions).Error
	return requisitions, err
}

// QuoteComparison holds quote comparison data for a requisition item
type QuoteComparison struct {
	Item              *models.RequisitionItem
	Specification     *models.Specification
	Quotes            []models.Quote
	BestQuote         *models.Quote
	TotalCostBest     float64 // Best quote price * quantity
	TotalCostBudget   float64 // Budget per unit * quantity (if set)
	SavingsVsBudget   float64 // Difference between budget and best quote
	HasQuotes         bool
	MissingQuotes     bool
}

// RequisitionQuoteComparison holds full comparison report for a requisition
type RequisitionQuoteComparison struct {
	Requisition      *models.Requisition
	ItemComparisons  []QuoteComparison
	TotalEstimate    float64 // Sum of all best quote totals
	TotalBudget      float64 // Sum of all item budgets (if set) or requisition budget
	TotalSavings     float64 // Budget - Estimate
	AllItemsHaveQuotes bool
}

// GetQuoteComparison generates a comprehensive quote comparison for a requisition
func (s *RequisitionService) GetQuoteComparison(requisitionID uint, quoteService *QuoteService) (*RequisitionQuoteComparison, error) {
	// Get requisition with items
	requisition, err := s.GetByID(requisitionID)
	if err != nil {
		return nil, err
	}

	comparison := &RequisitionQuoteComparison{
		Requisition:     requisition,
		ItemComparisons: make([]QuoteComparison, 0),
		TotalBudget:     requisition.Budget,
	}

	allHaveQuotes := true

	for _, item := range requisition.Items {
		itemComp := QuoteComparison{
			Item:          &item,
			Specification: item.Specification,
		}

		// Get quotes for this specification
		quotes, err := quoteService.CompareQuotesForSpecification(item.SpecificationID)
		if err != nil {
			return nil, err
		}

		itemComp.Quotes = quotes
		itemComp.HasQuotes = len(quotes) > 0

		if itemComp.HasQuotes {
			// Best quote is first (ordered by price)
			itemComp.BestQuote = &quotes[0]
			itemComp.TotalCostBest = quotes[0].ConvertedPrice * float64(item.Quantity)
			comparison.TotalEstimate += itemComp.TotalCostBest

			// Calculate savings vs budget if set
			if item.BudgetPerUnit > 0 {
				itemComp.TotalCostBudget = item.BudgetPerUnit * float64(item.Quantity)
				itemComp.SavingsVsBudget = itemComp.TotalCostBudget - itemComp.TotalCostBest
			}
		} else {
			itemComp.MissingQuotes = true
			allHaveQuotes = false
		}

		comparison.ItemComparisons = append(comparison.ItemComparisons, itemComp)
	}

	comparison.AllItemsHaveQuotes = allHaveQuotes

	if comparison.TotalBudget > 0 {
		comparison.TotalSavings = comparison.TotalBudget - comparison.TotalEstimate
	}

	return comparison, nil
}
