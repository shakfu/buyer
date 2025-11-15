package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/shakfu/buyer/internal/services"
)

// registerProcurementRoutes adds procurement-related routes to the app
func registerProcurementRoutes(app *fiber.App) {
	// Procurement analysis page
	app.Get("/projects/:id/procurement", handleProjectProcurement)

	// API endpoints for AJAX requests
	app.Get("/api/projects/:id/procurement/analysis", handleProcurementAnalysisAPI)
	app.Get("/api/projects/:id/procurement/dashboard", handleProcurementDashboardAPI)
	app.Get("/api/projects/:id/procurement/risks", handleProcurementRisksAPI)
	app.Get("/api/projects/:id/procurement/savings", handleProcurementSavingsAPI)
	app.Get("/api/projects/:id/procurement/consolidation", handleVendorConsolidationAPI)
	app.Get("/api/projects/:id/procurement/scenarios", handleScenariosComparisonAPI)
	app.Get("/api/projects/:id/procurement/recommendations", handleRecommendationsAPI)

	// Strategy management
	app.Get("/api/projects/:id/procurement/strategy", handleGetStrategyAPI)
	app.Post("/api/projects/:id/procurement/strategy", handleSetStrategyAPI)
}

// handleProjectProcurement renders the main procurement analysis page
func handleProjectProcurement(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid project ID")
	}

	projectSvc := services.NewProjectService(cfg.DB)
	project, err := projectSvc.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("Project not found")
	}

	return renderTemplate(c, "project-procurement.html", fiber.Map{
		"Title":   "Procurement Analysis - " + project.Name,
		"Project": project,
	})
}

// handleProcurementAnalysisAPI returns procurement analysis data as JSON
func handleProcurementAnalysisAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	comparison, err := procurementSvc.GetProjectProcurementComparison(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comparison)
}

// handleProcurementDashboardAPI returns dashboard data as JSON
func handleProcurementDashboardAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	dashboard, err := procurementSvc.GetProjectDashboard(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dashboard)
}

// handleProcurementRisksAPI returns risk assessment data as JSON
func handleProcurementRisksAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	risks, err := procurementSvc.AssessEnhancedProjectRisks(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(risks)
}

// handleProcurementSavingsAPI returns savings analysis as JSON
func handleProcurementSavingsAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	savings, err := procurementSvc.CalculateProjectSavings(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(savings)
}

// handleVendorConsolidationAPI returns vendor consolidation analysis as JSON
func handleVendorConsolidationAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	consolidation, err := procurementSvc.GetVendorConsolidationAnalysis(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(consolidation)
}

// handleScenariosComparisonAPI returns scenario comparison data as JSON
func handleScenariosComparisonAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	scenarios, err := procurementSvc.CompareScenarios(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(scenarios)
}

// handleRecommendationsAPI returns vendor recommendations as JSON
func handleRecommendationsAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	strategyType := c.Query("strategy", "balanced")

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	recommendations, err := procurementSvc.GenerateVendorRecommendations(uint(id), strategyType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(recommendations)
}

// handleGetStrategyAPI returns current procurement strategy as JSON
func handleGetStrategyAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	strategy, err := procurementSvc.GetOrCreateStrategy(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(strategy)
}

// handleSetStrategyAPI updates procurement strategy
func handleSetStrategyAPI(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var input struct {
		Strategy string `json:"strategy"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	quoteSvc := services.NewQuoteService(cfg.DB)
	projectSvc := services.NewProjectService(cfg.DB)
	procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

	strategy, err := procurementSvc.GetOrCreateStrategy(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	strategy.Strategy = input.Strategy
	if err := cfg.DB.Save(strategy).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update strategy"})
	}

	return c.JSON(strategy)
}
