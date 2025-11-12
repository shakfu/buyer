package services

import (
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// DashboardService provides analytics and reporting functionality
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// Stats holds general statistics
type Stats struct {
	TotalQuotes        int64
	ActiveQuotes       int64
	TotalRequisitions  int64
	TotalVendors       int64
	TotalProducts      int64
	TotalBrands        int64
	TotalSpecifications int64
}

// VendorSpending holds vendor spending statistics
type VendorSpending struct {
	VendorName string
	VendorID   uint
	Currency   string
	QuoteCount int64
	TotalValue float64
	AvgValue   float64
}

// ProductPriceComparison holds product price comparison data
type ProductPriceComparison struct {
	ProductID   uint
	ProductName string
	BrandName   string
	QuoteCount  int64
	MinPrice    float64
	MaxPrice    float64
	AvgPrice    float64
}

// ExpiryStats holds quote expiration statistics
type ExpiryStats struct {
	ExpiringSoon  int64 // < 7 days
	ExpiringMonth int64 // < 30 days
	Expired       int64
	Valid         int64 // 30+ days
	NoExpiry      int64
}

// GetStats returns general system statistics
func (s *DashboardService) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Count quotes
	if err := s.db.Model(&models.Quote{}).Count(&stats.TotalQuotes).Error; err != nil {
		return nil, err
	}

	// Count active quotes (not expired)
	now := time.Now()
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NULL OR valid_until > ?", now).
		Count(&stats.ActiveQuotes).Error; err != nil {
		return nil, err
	}

	// Count requisitions
	if err := s.db.Model(&models.Requisition{}).Count(&stats.TotalRequisitions).Error; err != nil {
		return nil, err
	}

	// Count vendors
	if err := s.db.Model(&models.Vendor{}).Count(&stats.TotalVendors).Error; err != nil {
		return nil, err
	}

	// Count products
	if err := s.db.Model(&models.Product{}).Count(&stats.TotalProducts).Error; err != nil {
		return nil, err
	}

	// Count brands
	if err := s.db.Model(&models.Brand{}).Count(&stats.TotalBrands).Error; err != nil {
		return nil, err
	}

	// Count specifications
	if err := s.db.Model(&models.Specification{}).Count(&stats.TotalSpecifications).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetVendorSpending returns spending statistics by vendor
func (s *DashboardService) GetVendorSpending() ([]VendorSpending, error) {
	var results []VendorSpending

	err := s.db.Model(&models.Quote{}).
		Select("vendors.id as vendor_id, vendors.name as vendor_name, vendors.currency as currency, COUNT(quotes.id) as quote_count, SUM(quotes.converted_price) as total_value, AVG(quotes.converted_price) as avg_value").
		Joins("JOIN vendors ON vendors.id = quotes.vendor_id").
		Group("vendors.id, vendors.name, vendors.currency").
		Order("total_value DESC").
		Scan(&results).Error

	return results, err
}

// GetProductPriceComparison returns products with multiple quotes and price ranges
func (s *DashboardService) GetProductPriceComparison() ([]ProductPriceComparison, error) {
	var results []ProductPriceComparison

	err := s.db.Model(&models.Quote{}).
		Select("products.id as product_id, products.name as product_name, brands.name as brand_name, COUNT(quotes.id) as quote_count, MIN(quotes.converted_price) as min_price, MAX(quotes.converted_price) as max_price, AVG(quotes.converted_price) as avg_price").
		Joins("JOIN products ON products.id = quotes.product_id").
		Joins("LEFT JOIN brands ON brands.id = products.brand_id").
		Group("products.id, products.name, brands.name").
		Having("COUNT(quotes.id) > 1").
		Order("quote_count DESC, products.name ASC").
		Scan(&results).Error

	return results, err
}

// GetExpiryStats returns quote expiration statistics
func (s *DashboardService) GetExpiryStats() (*ExpiryStats, error) {
	stats := &ExpiryStats{}
	now := time.Now()

	// Expiring soon (< 7 days)
	sevenDays := now.AddDate(0, 0, 7)
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NOT NULL AND valid_until > ? AND valid_until <= ?", now, sevenDays).
		Count(&stats.ExpiringSoon).Error; err != nil {
		return nil, err
	}

	// Expiring this month (< 30 days)
	thirtyDays := now.AddDate(0, 0, 30)
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NOT NULL AND valid_until > ? AND valid_until <= ?", now, thirtyDays).
		Count(&stats.ExpiringMonth).Error; err != nil {
		return nil, err
	}

	// Already expired
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NOT NULL AND valid_until <= ?", now).
		Count(&stats.Expired).Error; err != nil {
		return nil, err
	}

	// Valid (30+ days)
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NOT NULL AND valid_until > ?", thirtyDays).
		Count(&stats.Valid).Error; err != nil {
		return nil, err
	}

	// No expiry set
	if err := s.db.Model(&models.Quote{}).
		Where("valid_until IS NULL").
		Count(&stats.NoExpiry).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentQuotes returns the most recent quotes
func (s *DashboardService) GetRecentQuotes(limit int) ([]models.Quote, error) {
	var quotes []models.Quote

	if limit <= 0 {
		limit = 10
	}

	err := s.db.Preload("Vendor").Preload("Product").
		Order("created_at DESC").
		Limit(limit).
		Find(&quotes).Error

	return quotes, err
}

// ProjectStats holds project-level statistics
type ProjectStats struct {
	TotalBOMItems        int64
	TotalRequisitions    int64
	TotalBudget          float64
	AllocatedBudget      float64
	BudgetUtilization    float64 // percentage
	TotalBOMQuantity     int
}

// BOMItemQuantity holds BOM item quantity data for visualization
type BOMItemQuantity struct {
	SpecificationName string
	Quantity          int
}

// RequisitionBudgetData holds requisition budget data for visualization
type RequisitionBudgetData struct {
	RequisitionName string
	Budget          float64
}

// GetProjectStats returns statistics for a specific project
func (s *DashboardService) GetProjectStats(projectID uint) (*ProjectStats, error) {
	stats := &ProjectStats{}

	// Get project
	var project models.Project
	if err := s.db.Preload("BillOfMaterials.Items").Preload("Requisitions").First(&project, projectID).Error; err != nil {
		return nil, err
	}

	stats.TotalBudget = project.Budget

	// Count BOM items
	if project.BillOfMaterials != nil {
		stats.TotalBOMItems = int64(len(project.BillOfMaterials.Items))

		// Calculate total quantity across all BOM items
		for _, item := range project.BillOfMaterials.Items {
			stats.TotalBOMQuantity += item.Quantity
		}
	}

	// Count requisitions and sum their budgets
	stats.TotalRequisitions = int64(len(project.Requisitions))
	for _, req := range project.Requisitions {
		stats.AllocatedBudget += req.Budget
	}

	// Calculate budget utilization percentage
	if stats.TotalBudget > 0 {
		stats.BudgetUtilization = (stats.AllocatedBudget / stats.TotalBudget) * 100
	}

	return stats, nil
}

// GetProjectBOMItemQuantities returns BOM item quantities for a project
func (s *DashboardService) GetProjectBOMItemQuantities(projectID uint) ([]BOMItemQuantity, error) {
	var results []BOMItemQuantity

	err := s.db.Model(&models.BillOfMaterialsItem{}).
		Select("specifications.name as specification_name, bill_of_materials_items.quantity as quantity").
		Joins("JOIN specifications ON specifications.id = bill_of_materials_items.specification_id").
		Joins("JOIN bills_of_materials ON bills_of_materials.id = bill_of_materials_items.bill_of_materials_id").
		Where("bills_of_materials.project_id = ?", projectID).
		Order("quantity DESC").
		Scan(&results).Error

	return results, err
}

// GetProjectRequisitionBudgets returns requisition budget data for a project
func (s *DashboardService) GetProjectRequisitionBudgets(projectID uint) ([]RequisitionBudgetData, error) {
	var results []RequisitionBudgetData

	err := s.db.Model(&models.ProjectRequisition{}).
		Select("name as requisition_name, budget as budget").
		Where("project_id = ? AND budget > 0", projectID).
		Order("budget DESC").
		Scan(&results).Error

	return results, err
}
