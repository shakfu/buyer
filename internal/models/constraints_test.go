package models_test

import (
	"testing"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
)

// setupTestDB creates a test database with migrations
func setupTestDB(t *testing.T) *config.Config {
	cfg, err := config.NewConfig(config.Testing, false)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Run migrations for all models
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.SpecificationAttribute{},
		&models.Product{},
		&models.ProductAttribute{},
		&models.Quote{},
		&models.PurchaseOrder{},
		&models.VendorRating{},
		&models.Forex{},
		&models.Requisition{},
		&models.RequisitionItem{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
		&models.Document{},
	); err != nil{
		t.Fatalf("failed to run migrations: %v", err)
	}

	return cfg
}

// TestForeignKeyConstraints verifies that database-level foreign key constraints work
func TestForeignKeyConstraints(t *testing.T) {
	cfg := setupTestDB(t)
	db := cfg.DB

	t.Run("prevent deletion of brand with products", func(t *testing.T) {
		// Create a brand
		brand := &models.Brand{Name: "TestBrand"}
		if err := db.Create(brand).Error; err != nil {
			t.Fatalf("failed to create brand: %v", err)
		}

		// Create a product for this brand
		product := &models.Product{
			Name:    "TestProduct",
			BrandID: brand.ID,
		}
		if err := db.Create(product).Error; err != nil {
			t.Fatalf("failed to create product: %v", err)
		}

		// Try to delete the brand - should fail due to RESTRICT constraint
		err := db.Delete(brand).Error
		if err == nil {
			t.Error("expected error when deleting brand with products, got nil")
		}
	})

	t.Run("prevent deletion of vendor with quotes", func(t *testing.T) {
		// Create a brand and product first
		brand := &models.Brand{Name: "VendorTestBrand"}
		if err := db.Create(brand).Error; err != nil {
			t.Fatalf("failed to create brand: %v", err)
		}

		product := &models.Product{
			Name:    "VendorTestProduct",
			BrandID: brand.ID,
		}
		if err := db.Create(product).Error; err != nil {
			t.Fatalf("failed to create product: %v", err)
		}

		// Create a vendor
		vendor := &models.Vendor{Name: "TestVendor", Currency: "USD"}
		if err := db.Create(vendor).Error; err != nil {
			t.Fatalf("failed to create vendor: %v", err)
		}

		// Create a quote
		quote := &models.Quote{
			VendorID:       vendor.ID,
			ProductID:      product.ID,
			Price:          100.0,
			Currency:       "USD",
			ConvertedPrice: 100.0,
			ConversionRate: 1.0,
		}
		if err := db.Create(quote).Error; err != nil {
			t.Fatalf("failed to create quote: %v", err)
		}

		// Try to delete the vendor - should fail due to RESTRICT constraint
		err := db.Delete(vendor).Error
		if err == nil {
			t.Error("expected error when deleting vendor with quotes, got nil")
		}
	})

	t.Run("cascade delete requisition items when requisition is deleted", func(t *testing.T) {
		// Create a specification
		spec := &models.Specification{Name: "TestSpec"}
		if err := db.Create(spec).Error; err != nil {
			t.Fatalf("failed to create specification: %v", err)
		}

		// Create a requisition
		requisition := &models.Requisition{Name: "TestRequisition"}
		if err := db.Create(requisition).Error; err != nil {
			t.Fatalf("failed to create requisition: %v", err)
		}

		// Create a requisition item
		item := &models.RequisitionItem{
			RequisitionID:   requisition.ID,
			SpecificationID: spec.ID,
			Quantity:        5,
		}
		if err := db.Create(item).Error; err != nil {
			t.Fatalf("failed to create requisition item: %v", err)
		}

		// Delete the requisition - should cascade delete the item
		if err := db.Delete(requisition).Error; err != nil {
			t.Fatalf("failed to delete requisition: %v", err)
		}

		// Verify the item was deleted
		var count int64
		db.Model(&models.RequisitionItem{}).Where("id = ?", item.ID).Count(&count)
		if count != 0 {
			t.Error("expected requisition item to be cascade deleted, but it still exists")
		}
	})

	t.Run("cascade delete quotes when product is deleted", func(t *testing.T) {
		// Create a brand and product
		brand := &models.Brand{Name: "CascadeBrand"}
		if err := db.Create(brand).Error; err != nil {
			t.Fatalf("failed to create brand: %v", err)
		}

		product := &models.Product{
			Name:    "CascadeProduct",
			BrandID: brand.ID,
		}
		if err := db.Create(product).Error; err != nil {
			t.Fatalf("failed to create product: %v", err)
		}

		// Create a vendor
		vendor := &models.Vendor{Name: "CascadeVendor", Currency: "USD"}
		if err := db.Create(vendor).Error; err != nil {
			t.Fatalf("failed to create vendor: %v", err)
		}

		// Create a quote
		quote := &models.Quote{
			VendorID:       vendor.ID,
			ProductID:      product.ID,
			Price:          100.0,
			Currency:       "USD",
			ConvertedPrice: 100.0,
			ConversionRate: 1.0,
		}
		if err := db.Create(quote).Error; err != nil {
			t.Fatalf("failed to create quote: %v", err)
		}

		// Delete the product - should cascade delete the quote
		if err := db.Delete(product).Error; err != nil {
			t.Fatalf("failed to delete product: %v", err)
		}

		// Verify the quote was deleted
		var count int64
		db.Model(&models.Quote{}).Where("id = ?", quote.ID).Count(&count)
		if count != 0 {
			t.Error("expected quote to be cascade deleted, but it still exists")
		}
	})

	t.Run("set null on specification when specification is deleted", func(t *testing.T) {
		// Create a brand
		brand := &models.Brand{Name: "NullBrand"}
		if err := db.Create(brand).Error; err != nil {
			t.Fatalf("failed to create brand: %v", err)
		}

		// Create a specification
		spec := &models.Specification{Name: "NullSpec"}
		if err := db.Create(spec).Error; err != nil {
			t.Fatalf("failed to create specification: %v", err)
		}

		// Create a product with this specification
		product := &models.Product{
			Name:            "NullProduct",
			BrandID:         brand.ID,
			SpecificationID: &spec.ID,
		}
		if err := db.Create(product).Error; err != nil {
			t.Fatalf("failed to create product: %v", err)
		}

		// Delete the specification - should set product.SpecificationID to null
		if err := db.Delete(spec).Error; err != nil {
			t.Fatalf("failed to delete specification: %v", err)
		}

		// Reload product and verify SpecificationID is null
		var reloadedProduct models.Product
		if err := db.First(&reloadedProduct, product.ID).Error; err != nil {
			t.Fatalf("failed to reload product: %v", err)
		}

		if reloadedProduct.SpecificationID != nil {
			t.Errorf("expected SpecificationID to be nil after deleting specification, got %v", *reloadedProduct.SpecificationID)
		}
	})
}
