package models

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable foreign key constraints for SQLite
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("Failed to enable foreign key constraints: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(
		&Vendor{},
		&Brand{},
		&Specification{},
		&Product{},
		&Quote{},
		&Forex{},
		&Requisition{},
		&RequisitionItem{},
		&Project{},
		&BillOfMaterials{},
		&BillOfMaterialsItem{},
		&ProjectRequisition{},
		&ProjectRequisitionItem{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func TestProjectModelCreation(t *testing.T) {
	db := setupTestDB(t)

	deadline := time.Now().Add(30 * 24 * time.Hour)
	project := &Project{
		Name:        "Test Project",
		Description: "A test project",
		Budget:      50000.00,
		Deadline:    &deadline,
	}

	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID == 0 {
		t.Error("Project ID should be set after creation")
	}

	if project.Status != "planning" {
		t.Errorf("Expected status 'planning', got '%s'", project.Status)
	}
}

func TestBillOfMaterialsCreation(t *testing.T) {
	db := setupTestDB(t)

	// Create a project
	project := &Project{
		Name:   "Office Renovation",
		Budget: 100000.00,
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a Bill of Materials
	bom := &BillOfMaterials{
		ProjectID: project.ID,
		Notes:     "Master BOM for office renovation",
	}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	if bom.ID == 0 {
		t.Error("BillOfMaterials ID should be set after creation")
	}
}

func TestBillOfMaterialsOneToOneConstraint(t *testing.T) {
	db := setupTestDB(t)

	// Create a project
	project := &Project{
		Name:   "Test Project",
		Budget: 50000.00,
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create first BOM
	bom1 := &BillOfMaterials{
		ProjectID: project.ID,
		Notes:     "First BOM",
	}
	if err := db.Create(bom1).Error; err != nil {
		t.Fatalf("Failed to create first BillOfMaterials: %v", err)
	}

	// Try to create second BOM for same project - should fail due to unique constraint
	bom2 := &BillOfMaterials{
		ProjectID: project.ID,
		Notes:     "Second BOM",
	}
	err := db.Create(bom2).Error
	if err == nil {
		t.Error("Expected error when creating second BillOfMaterials for same project, but got nil")
	}
}

func TestBillOfMaterialsItemCreation(t *testing.T) {
	db := setupTestDB(t)

	// Create project
	project := &Project{Name: "Test Project"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create BOM
	bom := &BillOfMaterials{ProjectID: project.ID}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	// Create specification
	spec := &Specification{
		Name:        "Laptop - Intel i7",
		Description: "High-performance laptop",
	}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Create BOM item
	bomItem := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          10,
		Notes:             "For developers",
	}
	if err := db.Create(bomItem).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterialsItem: %v", err)
	}

	if bomItem.ID == 0 {
		t.Error("BillOfMaterialsItem ID should be set after creation")
	}
}

func TestBillOfMaterialsItemUniqueConstraint(t *testing.T) {
	db := setupTestDB(t)

	// Create project and BOM
	project := &Project{Name: "Test Project"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	bom := &BillOfMaterials{ProjectID: project.ID}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	// Create specification
	spec := &Specification{Name: "Laptop"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Create first BOM item
	bomItem1 := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          5,
	}
	if err := db.Create(bomItem1).Error; err != nil {
		t.Fatalf("Failed to create first BillOfMaterialsItem: %v", err)
	}

	// Try to create duplicate - should fail
	bomItem2 := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          10,
	}
	err := db.Create(bomItem2).Error
	if err == nil {
		t.Error("Expected error when creating duplicate BillOfMaterialsItem, but got nil")
	}
}

func TestProjectRequisitionCreation(t *testing.T) {
	db := setupTestDB(t)

	// Create project with BOM
	project := &Project{Name: "Test Project"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	bom := &BillOfMaterials{ProjectID: project.ID}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	spec := &Specification{Name: "Laptop"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	bomItem := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          10,
		Notes:             "Laptops for team",
	}
	if err := db.Create(bomItem).Error; err != nil {
		t.Fatalf("Failed to create BOM item: %v", err)
	}

	// Create project requisition from BOM
	projectReq := &ProjectRequisition{
		ProjectID:     project.ID,
		Name:          "Phase 1 Purchase",
		Justification: "Initial equipment procurement",
		Budget:        15000.00,
	}
	if err := db.Create(projectReq).Error; err != nil {
		t.Fatalf("Failed to create ProjectRequisition: %v", err)
	}

	if projectReq.ID == 0 {
		t.Error("ProjectRequisition ID should be set after creation")
	}

	// Add item to project requisition
	projectReqItem := &ProjectRequisitionItem{
		ProjectRequisitionID:  projectReq.ID,
		BillOfMaterialsItemID: bomItem.ID,
		QuantityRequested:     5, // Requesting 5 out of 10 from BOM
		Notes:                 "First batch",
	}
	if err := db.Create(projectReqItem).Error; err != nil {
		t.Fatalf("Failed to create ProjectRequisitionItem: %v", err)
	}

	if projectReqItem.ID == 0 {
		t.Error("ProjectRequisitionItem ID should be set after creation")
	}

	if projectReqItem.QuantityRequested != 5 {
		t.Errorf("Expected quantity 5, got %d", projectReqItem.QuantityRequested)
	}
}

func TestStandaloneRequisition(t *testing.T) {
	db := setupTestDB(t)

	// Create standalone requisition (no project)
	req := &Requisition{
		Name:          "Ad-hoc Purchase",
		Justification: "Quick procurement",
		Budget:        5000.00,
	}
	if err := db.Create(req).Error; err != nil {
		t.Fatalf("Failed to create standalone requisition: %v", err)
	}

	if req.ID == 0 {
		t.Error("Standalone requisition ID should be set after creation")
	}

	// Add items to standalone requisition (linked to specifications, not BOM)
	spec := &Specification{Name: "Monitor"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	reqItem := &RequisitionItem{
		RequisitionID:   req.ID,
		SpecificationID: spec.ID,
		Quantity:        3,
		BudgetPerUnit:   500.00,
		Description:     "Extra monitors needed",
	}
	if err := db.Create(reqItem).Error; err != nil {
		t.Fatalf("Failed to create RequisitionItem: %v", err)
	}

	if reqItem.ID == 0 {
		t.Error("RequisitionItem ID should be set after creation")
	}
}

func TestCascadeDeletes(t *testing.T) {
	db := setupTestDB(t)

	// Create project with BOM and items
	project := &Project{Name: "Test Project"}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	bom := &BillOfMaterials{ProjectID: project.ID}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	spec := &Specification{Name: "Test Spec"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	bomItem := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          5,
	}
	if err := db.Create(bomItem).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterialsItem: %v", err)
	}

	// Delete project - should cascade delete BOM and BOM items
	if err := db.Delete(project).Error; err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify BOM is deleted
	var bomCount int64
	db.Model(&BillOfMaterials{}).Where("id = ?", bom.ID).Count(&bomCount)
	if bomCount != 0 {
		t.Error("BillOfMaterials should be cascade deleted")
	}

	// Verify BOM item is deleted
	var bomItemCount int64
	db.Model(&BillOfMaterialsItem{}).Where("id = ?", bomItem.ID).Count(&bomItemCount)
	if bomItemCount != 0 {
		t.Error("BillOfMaterialsItem should be cascade deleted")
	}

	// Verify specification is NOT deleted (RESTRICT)
	var specCount int64
	db.Model(&Specification{}).Where("id = ?", spec.ID).Count(&specCount)
	if specCount == 0 {
		t.Error("Specification should NOT be deleted (RESTRICT constraint)")
	}
}

func TestPreloadRelationships(t *testing.T) {
	db := setupTestDB(t)

	// Create project with BOM and items
	project := &Project{Name: "Office Renovation", Budget: 50000}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	bom := &BillOfMaterials{ProjectID: project.ID, Notes: "Main BOM"}
	if err := db.Create(bom).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterials: %v", err)
	}

	spec := &Specification{Name: "Laptop"}
	if err := db.Create(spec).Error; err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	bomItem := &BillOfMaterialsItem{
		BillOfMaterialsID: bom.ID,
		SpecificationID:   spec.ID,
		Quantity:          10,
	}
	if err := db.Create(bomItem).Error; err != nil {
		t.Fatalf("Failed to create BillOfMaterialsItem: %v", err)
	}

	// Preload relationships
	var loadedProject Project
	if err := db.Preload("BillOfMaterials.Items.Specification").First(&loadedProject, project.ID).Error; err != nil {
		t.Fatalf("Failed to preload project: %v", err)
	}

	if loadedProject.BillOfMaterials == nil {
		t.Error("BillOfMaterials should be preloaded")
	}

	if len(loadedProject.BillOfMaterials.Items) == 0 {
		t.Error("BillOfMaterialsItems should be preloaded")
	}

	if loadedProject.BillOfMaterials.Items[0].Specification == nil {
		t.Error("Specification should be preloaded")
	}

	if loadedProject.BillOfMaterials.Items[0].Specification.Name != "Laptop" {
		t.Errorf("Expected specification name 'Laptop', got '%s'", loadedProject.BillOfMaterials.Items[0].Specification.Name)
	}
}
