package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// PurchaseOrderService handles business logic for purchase orders
type PurchaseOrderService struct {
	db *gorm.DB
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(db *gorm.DB) *PurchaseOrderService {
	return &PurchaseOrderService{db: db}
}

// CreatePurchaseOrderInput represents input for creating a purchase order
type CreatePurchaseOrderInput struct {
	QuoteID          uint
	RequisitionID    *uint
	PONumber         string
	Quantity         int
	ExpectedDelivery *time.Time
	ShippingCost     float64
	Tax              float64
	Notes            string
}

// Create creates a new purchase order from a quote
func (s *PurchaseOrderService) Create(input CreatePurchaseOrderInput) (*models.PurchaseOrder, error) {
	// Validate PONumber
	poNumber := strings.TrimSpace(input.PONumber)
	if poNumber == "" {
		return nil, &ValidationError{Field: "po_number", Message: "PO number cannot be empty"}
	}

	// Validate quantity
	if input.Quantity <= 0 {
		return nil, &ValidationError{Field: "quantity", Message: "quantity must be greater than zero"}
	}

	// Check if PO number already exists
	var existing models.PurchaseOrder
	err := s.db.Where("po_number = ?", poNumber).First(&existing).Error
	if err == nil {
		return nil, &DuplicateError{Entity: "purchase order", Name: poNumber}
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Get the quote
	var quote models.Quote
	if err := s.db.Preload("Vendor").Preload("Product").First(&quote, input.QuoteID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "quote", ID: input.QuoteID}
		}
		return nil, err
	}

	// Validate requisition if provided
	if input.RequisitionID != nil {
		var req models.Requisition
		if err := s.db.First(&req, *input.RequisitionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, &NotFoundError{Entity: "requisition", ID: *input.RequisitionID}
			}
			return nil, err
		}
	}

	// Create purchase order from quote
	po := &models.PurchaseOrder{
		QuoteID:          quote.ID,
		VendorID:         quote.VendorID,
		ProductID:        quote.ProductID,
		RequisitionID:    input.RequisitionID,
		PONumber:         poNumber,
		Status:           "pending",  // Will be set by BeforeCreate hook if empty
		OrderDate:        time.Now(), // Will be set by BeforeCreate hook if zero
		ExpectedDelivery: input.ExpectedDelivery,
		Quantity:         input.Quantity,
		UnitPrice:        quote.Price,
		Currency:         quote.Currency,
		TotalAmount:      quote.Price * float64(input.Quantity),
		ShippingCost:     input.ShippingCost,
		Tax:              input.Tax,
		Notes:            input.Notes,
	}

	// GrandTotal will be calculated by BeforeCreate hook
	if err := s.db.Create(po).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").Preload("Requisition").First(po, po.ID).Error; err != nil {
		return nil, err
	}

	return po, nil
}

// GetByID retrieves a purchase order by ID
func (s *PurchaseOrderService) GetByID(id uint) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").
		Preload("Requisition").Preload("VendorRatings").First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "purchase order", ID: id}
		}
		return nil, err
	}

	// Load documents separately (polymorphic relationship)
	s.loadDocuments(&po)

	return &po, nil
}

// GetByPONumber retrieves a purchase order by PO number
func (s *PurchaseOrderService) GetByPONumber(poNumber string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").
		Preload("Requisition").Preload("VendorRatings").
		Where("po_number = ?", poNumber).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "purchase order", ID: poNumber}
		}
		return nil, err
	}

	// Load documents separately (polymorphic relationship)
	s.loadDocuments(&po)

	return &po, nil
}

// List retrieves all purchase orders with pagination
func (s *PurchaseOrderService) List(limit, offset int) ([]*models.PurchaseOrder, error) {
	var orders []*models.PurchaseOrder
	query := s.db.Preload("Quote").Preload("Vendor").Preload("Product").
		Preload("Requisition").Preload("VendorRatings").
		Order("order_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	// Load documents for all purchase orders
	for i := range orders {
		s.loadDocuments(orders[i])
	}

	return orders, nil
}

// ListByStatus retrieves purchase orders by status
func (s *PurchaseOrderService) ListByStatus(status string, limit, offset int) ([]*models.PurchaseOrder, error) {
	var orders []*models.PurchaseOrder
	query := s.db.Preload("Quote").Preload("Vendor").Preload("Product").
		Preload("Requisition").Preload("VendorRatings").
		Where("status = ?", status).
		Order("order_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	// Load documents for all purchase orders
	for i := range orders {
		s.loadDocuments(orders[i])
	}

	return orders, nil
}

// ListByVendor retrieves purchase orders for a specific vendor
func (s *PurchaseOrderService) ListByVendor(vendorID uint, limit, offset int) ([]*models.PurchaseOrder, error) {
	var orders []*models.PurchaseOrder
	query := s.db.Preload("Quote").Preload("Vendor").Preload("Product").
		Preload("Requisition").Preload("VendorRatings").
		Where("vendor_id = ?", vendorID).
		Order("order_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	// Load documents for all purchase orders
	for i := range orders {
		s.loadDocuments(orders[i])
	}

	return orders, nil
}

// UpdateStatus updates the status of a purchase order
func (s *PurchaseOrderService) UpdateStatus(id uint, status string) (*models.PurchaseOrder, error) {
	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"approved":  true,
		"ordered":   true,
		"shipped":   true,
		"received":  true,
		"cancelled": true,
	}
	if !validStatuses[status] {
		return nil, &ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", status)}
	}

	var po models.PurchaseOrder
	if err := s.db.First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "purchase order", ID: id}
		}
		return nil, err
	}

	po.Status = status
	if err := s.db.Save(&po).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").Preload("Requisition").First(&po, po.ID).Error; err != nil {
		return nil, err
	}

	return &po, nil
}

// UpdateDeliveryDates updates the delivery dates of a purchase order
func (s *PurchaseOrderService) UpdateDeliveryDates(id uint, expectedDelivery, actualDelivery *time.Time) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := s.db.First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "purchase order", ID: id}
		}
		return nil, err
	}

	if expectedDelivery != nil {
		po.ExpectedDelivery = expectedDelivery
	}
	if actualDelivery != nil {
		po.ActualDelivery = actualDelivery
		// Automatically set status to received if actual delivery is set
		if po.Status != "received" && po.Status != "cancelled" {
			po.Status = "received"
		}
	}

	if err := s.db.Save(&po).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").Preload("Requisition").First(&po, po.ID).Error; err != nil {
		return nil, err
	}

	return &po, nil
}

// UpdateInvoiceNumber updates the invoice number of a purchase order
func (s *PurchaseOrderService) UpdateInvoiceNumber(id uint, invoiceNumber string) (*models.PurchaseOrder, error) {
	invoiceNumber = strings.TrimSpace(invoiceNumber)

	var po models.PurchaseOrder
	if err := s.db.First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Entity: "purchase order", ID: id}
		}
		return nil, err
	}

	po.InvoiceNumber = invoiceNumber
	if err := s.db.Save(&po).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Quote").Preload("Vendor").Preload("Product").Preload("Requisition").First(&po, po.ID).Error; err != nil {
		return nil, err
	}

	return &po, nil
}

// Delete deletes a purchase order
func (s *PurchaseOrderService) Delete(id uint) error {
	var po models.PurchaseOrder
	if err := s.db.First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &NotFoundError{Entity: "purchase order", ID: id}
		}
		return err
	}

	// Only allow deletion of pending or cancelled orders
	if po.Status != "pending" && po.Status != "cancelled" {
		return &ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("cannot delete purchase order with status: %s", po.Status),
		}
	}

	return s.db.Delete(&po).Error
}

// Count returns the total number of purchase orders
func (s *PurchaseOrderService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.PurchaseOrder{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByStatus returns the count of purchase orders by status
func (s *PurchaseOrderService) CountByStatus(status string) (int64, error) {
	var count int64
	if err := s.db.Model(&models.PurchaseOrder{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// loadDocuments loads documents for a purchase order (polymorphic relationship)
func (s *PurchaseOrderService) loadDocuments(po *models.PurchaseOrder) {
	var docs []models.Document
	s.db.Where("entity_type = ? AND entity_id = ?", "purchase_order", po.ID).Find(&docs)
	po.Documents = docs
}
