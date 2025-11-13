package services

import (
	"testing"
	"time"
)

func TestProjectService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	deadline := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name        string
		projectName string
		description string
		budget      float64
		deadline    *time.Time
		wantErr     bool
		errType     interface{}
	}{
		{
			name:        "valid project",
			projectName: "Office Renovation",
			description: "Complete office upgrade",
			budget:      50000.00,
			deadline:    &deadline,
			wantErr:     false,
		},
		{
			name:        "empty name",
			projectName: "",
			wantErr:     true,
			errType:     &ValidationError{},
		},
		{
			name:        "whitespace name",
			projectName: "   ",
			wantErr:     true,
			errType:     &ValidationError{},
		},
		{
			name:        "negative budget",
			projectName: "Test Project",
			budget:      -1000.00,
			wantErr:     true,
			errType:     &ValidationError{},
		},
		{
			name:        "zero budget is valid",
			projectName: "Zero Budget Project",
			budget:      0,
			wantErr:     false,
		},
		{
			name:        "no deadline is valid",
			projectName: "No Deadline Project",
			budget:      10000,
			deadline:    nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := svc.Create(tt.projectName, tt.description, tt.budget, tt.deadline)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errType != nil {
					switch tt.errType.(type) {
					case *ValidationError:
						if _, ok := err.(*ValidationError); !ok {
							t.Errorf("Expected ValidationError, got %T", err)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if project == nil {
					t.Error("Expected project, got nil")
				} else {
					if project.Name != tt.projectName {
						t.Errorf("Expected name '%s', got '%s'", tt.projectName, project.Name)
					}
					if project.Status != "planning" {
						t.Errorf("Expected status 'planning', got '%s'", project.Status)
					}
					if project.BillOfMaterials == nil {
						t.Error("Expected BillOfMaterials to be created automatically")
					}
				}
			}
		})
	}
}

func TestProjectService_CreateDuplicate(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	_, err := svc.Create("Duplicate Project", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	_, err = svc.Create("Duplicate Project", "", 20000, nil)
	if err == nil {
		t.Error("Expected duplicate error, got nil")
	}

	if _, ok := err.(*DuplicateError); !ok {
		t.Errorf("Expected DuplicateError, got %T", err)
	}
}

func TestProjectService_GetByID(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	created, err := svc.Create("Test Project", "Description", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	project, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", project.Name)
	}

	// Test not found
	_, err = svc.GetByID(99999)
	if err == nil {
		t.Error("Expected not found error, got nil")
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestProjectService_GetByName(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	_, err := svc.Create("Named Project", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	project, err := svc.GetByName("Named Project")
	if err != nil {
		t.Fatalf("Failed to get project by name: %v", err)
	}

	if project.Name != "Named Project" {
		t.Errorf("Expected name 'Named Project', got '%s'", project.Name)
	}

	// Test not found
	_, err = svc.GetByName("Nonexistent Project")
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestProjectService_List(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	// Create test projects
	_, err := svc.Create("Project A", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project A: %v", err)
	}
	_, err = svc.Create("Project B", "", 20000, nil)
	if err != nil {
		t.Fatalf("Failed to create project B: %v", err)
	}
	_, err = svc.Create("Project C", "", 30000, nil)
	if err != nil {
		t.Fatalf("Failed to create project C: %v", err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectedFirst string
	}{
		{
			name:          "all projects",
			limit:         0,
			offset:        0,
			expectedCount: 3,
			expectedFirst: "Project A",
		},
		{
			name:          "limited projects",
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedFirst: "Project A",
		},
		{
			name:          "with offset",
			limit:         2,
			offset:        1,
			expectedCount: 2,
			expectedFirst: "Project B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := svc.List(tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("Failed to list projects: %v", err)
			}

			if len(projects) != tt.expectedCount {
				t.Errorf("Expected %d projects, got %d", tt.expectedCount, len(projects))
			}

			if len(projects) > 0 && projects[0].Name != tt.expectedFirst {
				t.Errorf("Expected first project '%s', got '%s'", tt.expectedFirst, projects[0].Name)
			}
		})
	}
}

func TestProjectService_Update(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	project, err := svc.Create("Original Name", "Original description", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	newDeadline := time.Now().Add(60 * 24 * time.Hour)

	updated, err := svc.Update(project.ID, "Updated Name", "Updated description", 20000, &newDeadline, "active")
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updated.Description)
	}
	if updated.Budget != 20000 {
		t.Errorf("Expected budget 20000, got %f", updated.Budget)
	}
	if updated.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", updated.Status)
	}

	// Test invalid status
	_, err = svc.Update(project.ID, "Name", "", 10000, nil, "invalid_status")
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError for invalid status, got %T", err)
	}

	// Test duplicate name
	_, err = svc.Create("Another Project", "", 5000, nil)
	if err != nil {
		t.Fatalf("Failed to create another project: %v", err)
	}

	_, err = svc.Update(project.ID, "Another Project", "", 10000, nil, "")
	if _, ok := err.(*DuplicateError); !ok {
		t.Errorf("Expected DuplicateError, got %T", err)
	}
}

func TestProjectService_Delete(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	project, err := svc.Create("To Delete", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	err = svc.Delete(project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify deletion
	_, err = svc.GetByID(project.ID)
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError after deletion, got %T", err)
	}

	// Test delete non-existent
	err = svc.Delete(99999)
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError for non-existent project, got %T", err)
	}
}

func TestProjectService_Count(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewProjectService(cfg.DB)

	count, err := svc.Count()
	if err != nil {
		t.Fatalf("Failed to count projects: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	_, err = svc.Create("Project 1", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	count, err = svc.Count()
	if err != nil {
		t.Fatalf("Failed to count projects: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestProjectService_AddBillOfMaterialsItem(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	projectSvc := NewProjectService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)

	// Create project and specification
	project, err := projectSvc.Create("Test Project", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	spec, err := specSvc.Create("Laptop - Intel i7", "High-performance laptop")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	// Add BOM item
	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 10, "For developers")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	if bomItem.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", bomItem.Quantity)
	}
	if bomItem.Specification.Name != "Laptop - Intel i7" {
		t.Errorf("Expected spec name 'Laptop - Intel i7', got '%s'", bomItem.Specification.Name)
	}

	// Test duplicate specification
	_, err = projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 5, "Duplicate")
	if _, ok := err.(*DuplicateError); !ok {
		t.Errorf("Expected DuplicateError for duplicate spec, got %T", err)
	}

	// Test invalid quantity
	spec2, err := specSvc.Create("Monitor", "4K Monitor")
	if err != nil {
		t.Fatalf("Failed to create second specification: %v", err)
	}

	_, err = projectSvc.AddBillOfMaterialsItem(project.ID, spec2.ID, 0, "")
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError for zero quantity, got %T", err)
	}

	// Test non-existent project
	_, err = projectSvc.AddBillOfMaterialsItem(99999, spec.ID, 10, "")
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError for non-existent project, got %T", err)
	}

	// Test non-existent specification
	_, err = projectSvc.AddBillOfMaterialsItem(project.ID, 99999, 10, "")
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError for non-existent specification, got %T", err)
	}
}

func TestProjectService_UpdateBillOfMaterialsItem(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	projectSvc := NewProjectService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)

	project, err := projectSvc.Create("Test Project", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	spec, err := specSvc.Create("Laptop", "")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 5, "Original notes")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	// Update BOM item
	updated, err := projectSvc.UpdateBillOfMaterialsItem(bomItem.ID, 15, "Updated notes")
	if err != nil {
		t.Fatalf("Failed to update BOM item: %v", err)
	}

	if updated.Quantity != 15 {
		t.Errorf("Expected quantity 15, got %d", updated.Quantity)
	}
	if updated.Notes != "Updated notes" {
		t.Errorf("Expected notes 'Updated notes', got '%s'", updated.Notes)
	}

	// Test invalid quantity
	_, err = projectSvc.UpdateBillOfMaterialsItem(bomItem.ID, 0, "")
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError for zero quantity, got %T", err)
	}

	// Test non-existent item
	_, err = projectSvc.UpdateBillOfMaterialsItem(99999, 10, "")
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestProjectService_DeleteBillOfMaterialsItem(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	projectSvc := NewProjectService(cfg.DB)
	specSvc := NewSpecificationService(cfg.DB)

	project, err := projectSvc.Create("Test Project", "", 10000, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	spec, err := specSvc.Create("Laptop", "")
	if err != nil {
		t.Fatalf("Failed to create specification: %v", err)
	}

	bomItem, err := projectSvc.AddBillOfMaterialsItem(project.ID, spec.ID, 5, "")
	if err != nil {
		t.Fatalf("Failed to add BOM item: %v", err)
	}

	// Delete BOM item
	err = projectSvc.DeleteBillOfMaterialsItem(bomItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete BOM item: %v", err)
	}

	// Verify deletion
	_, err = projectSvc.UpdateBillOfMaterialsItem(bomItem.ID, 10, "")
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError after deletion, got %T", err)
	}

	// Test delete non-existent
	err = projectSvc.DeleteBillOfMaterialsItem(99999)
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}
