package services

import (
	"errors"
	"fmt"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// VendorRatingService handles business logic for vendor ratings
type VendorRatingService struct {
	db *gorm.DB
}

// NewVendorRatingService creates a new vendor rating service
func NewVendorRatingService(db *gorm.DB) *VendorRatingService {
	return &VendorRatingService{db: db}
}

// CreateVendorRatingInput represents input for creating a vendor rating
type CreateVendorRatingInput struct {
	VendorID        uint
	PurchaseOrderID *uint
	PriceRating     *int
	QualityRating   *int
	DeliveryRating  *int
	ServiceRating   *int
	Comments        string
	RatedBy         string
}

// Create creates a new vendor rating
func (s *VendorRatingService) Create(input CreateVendorRatingInput) (*models.VendorRating, error) {
	// Validate vendor exists
	var vendor models.Vendor
	if err := s.db.First(&vendor, input.VendorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "Vendor", ID: input.VendorID}
		}
		return nil, err
	}

	// Validate purchase order if provided
	if input.PurchaseOrderID != nil {
		var po models.PurchaseOrder
		if err := s.db.First(&po, *input.PurchaseOrderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &NotFoundError{Entity: "PurchaseOrder", ID: *input.PurchaseOrderID}
			}
			return nil, err
		}
		// Ensure the purchase order is for this vendor
		if po.VendorID != input.VendorID {
			return nil, &ValidationError{
				Field:   "purchase_order_id",
				Message: "purchase order does not belong to this vendor",
			}
		}
	}

	// Validate ratings are in 1-5 range
	if err := validateRating(input.PriceRating, "price_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.QualityRating, "quality_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.DeliveryRating, "delivery_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.ServiceRating, "service_rating"); err != nil {
		return nil, err
	}

	// At least one rating must be provided
	if input.PriceRating == nil && input.QualityRating == nil &&
		input.DeliveryRating == nil && input.ServiceRating == nil {
		return nil, &ValidationError{
			Field:   "ratings",
			Message: "at least one rating must be provided",
		}
	}

	rating := &models.VendorRating{
		VendorID:        input.VendorID,
		PurchaseOrderID: input.PurchaseOrderID,
		PriceRating:     input.PriceRating,
		QualityRating:   input.QualityRating,
		DeliveryRating:  input.DeliveryRating,
		ServiceRating:   input.ServiceRating,
		Comments:        input.Comments,
		RatedBy:         input.RatedBy,
	}

	if err := s.db.Create(rating).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Vendor").Preload("PurchaseOrder").First(rating, rating.ID).Error; err != nil {
		return nil, err
	}

	return rating, nil
}

// validateRating checks if a rating value is in the valid range (1-5)
func validateRating(rating *int, field string) error {
	if rating != nil {
		if *rating < 1 || *rating > 5 {
			return &ValidationError{
				Field:   field,
				Message: fmt.Sprintf("%s must be between 1 and 5", field),
			}
		}
	}
	return nil
}

// GetByID retrieves a vendor rating by ID
func (s *VendorRatingService) GetByID(id uint) (*models.VendorRating, error) {
	var rating models.VendorRating
	if err := s.db.Preload("Vendor").Preload("PurchaseOrder").First(&rating, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "VendorRating", ID: id}
		}
		return nil, err
	}
	return &rating, nil
}

// List retrieves all vendor ratings with pagination
func (s *VendorRatingService) List(limit, offset int) ([]*models.VendorRating, error) {
	var ratings []*models.VendorRating
	query := s.db.Preload("Vendor").Preload("PurchaseOrder").Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}

// ListByVendor retrieves all ratings for a specific vendor
func (s *VendorRatingService) ListByVendor(vendorID uint, limit, offset int) ([]*models.VendorRating, error) {
	var ratings []*models.VendorRating
	query := s.db.Preload("Vendor").Preload("PurchaseOrder").
		Where("vendor_id = ?", vendorID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}

// ListByPurchaseOrder retrieves all ratings for a specific purchase order
func (s *VendorRatingService) ListByPurchaseOrder(poID uint, limit, offset int) ([]*models.VendorRating, error) {
	var ratings []*models.VendorRating
	query := s.db.Preload("Vendor").Preload("PurchaseOrder").
		Where("purchase_order_id = ?", poID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}

// Update updates a vendor rating
func (s *VendorRatingService) Update(id uint, input CreateVendorRatingInput) (*models.VendorRating, error) {
	var rating models.VendorRating
	if err := s.db.First(&rating, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{Entity: "VendorRating", ID: id}
		}
		return nil, err
	}

	// Validate ratings are in 1-5 range
	if err := validateRating(input.PriceRating, "price_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.QualityRating, "quality_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.DeliveryRating, "delivery_rating"); err != nil {
		return nil, err
	}
	if err := validateRating(input.ServiceRating, "service_rating"); err != nil {
		return nil, err
	}

	// At least one rating must be provided
	if input.PriceRating == nil && input.QualityRating == nil &&
		input.DeliveryRating == nil && input.ServiceRating == nil {
		return nil, &ValidationError{
			Field:   "ratings",
			Message: "at least one rating must be provided",
		}
	}

	// Update fields
	rating.PriceRating = input.PriceRating
	rating.QualityRating = input.QualityRating
	rating.DeliveryRating = input.DeliveryRating
	rating.ServiceRating = input.ServiceRating
	rating.Comments = input.Comments
	if input.RatedBy != "" {
		rating.RatedBy = input.RatedBy
	}

	if err := s.db.Save(&rating).Error; err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Vendor").Preload("PurchaseOrder").First(&rating, rating.ID).Error; err != nil {
		return nil, err
	}

	return &rating, nil
}

// Delete deletes a vendor rating
func (s *VendorRatingService) Delete(id uint) error {
	result := s.db.Delete(&models.VendorRating{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "VendorRating", ID: id}
	}
	return nil
}

// GetAverageRatings calculates average ratings for a vendor
func (s *VendorRatingService) GetAverageRatings(vendorID uint) (map[string]float64, error) {
	var ratings []*models.VendorRating
	if err := s.db.Where("vendor_id = ?", vendorID).Find(&ratings).Error; err != nil {
		return nil, err
	}

	if len(ratings) == 0 {
		return map[string]float64{
			"price":    0,
			"quality":  0,
			"delivery": 0,
			"service":  0,
			"overall":  0,
			"count":    0,
		}, nil
	}

	var priceSum, qualitySum, deliverySum, serviceSum float64
	var priceCount, qualityCount, deliveryCount, serviceCount int

	for _, r := range ratings {
		if r.PriceRating != nil {
			priceSum += float64(*r.PriceRating)
			priceCount++
		}
		if r.QualityRating != nil {
			qualitySum += float64(*r.QualityRating)
			qualityCount++
		}
		if r.DeliveryRating != nil {
			deliverySum += float64(*r.DeliveryRating)
			deliveryCount++
		}
		if r.ServiceRating != nil {
			serviceSum += float64(*r.ServiceRating)
			serviceCount++
		}
	}

	result := make(map[string]float64)
	if priceCount > 0 {
		result["price"] = priceSum / float64(priceCount)
	}
	if qualityCount > 0 {
		result["quality"] = qualitySum / float64(qualityCount)
	}
	if deliveryCount > 0 {
		result["delivery"] = deliverySum / float64(deliveryCount)
	}
	if serviceCount > 0 {
		result["service"] = serviceSum / float64(serviceCount)
	}

	// Calculate overall average
	totalSum := priceSum + qualitySum + deliverySum + serviceSum
	totalCount := priceCount + qualityCount + deliveryCount + serviceCount
	if totalCount > 0 {
		result["overall"] = totalSum / float64(totalCount)
	}
	result["count"] = float64(len(ratings))

	return result, nil
}

// Count returns the total number of vendor ratings
func (s *VendorRatingService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.VendorRating{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
