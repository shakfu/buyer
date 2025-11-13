package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/shakfu/buyer/internal/services"
	"github.com/shakfu/buyer/web"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long:  "Start the FastAPI-inspired web server with HTMX support",
	Run: func(cmd *cobra.Command, args []string) {
		// Get port from flag or config (which reads from environment)
		port, _ := cmd.Flags().GetInt("port")
		// If port is still default 8080, check if config has a different value from env
		if port == 8080 && cfg.WebPort != 8080 {
			port = cfg.WebPort
		}

		app := fiber.New(fiber.Config{
			AppName: fmt.Sprintf("Buyer %s", Version),
		})

		// Middleware
		app.Use(recover.New())
		app.Use(logger.New())

		// Security middleware
		enableAuth := os.Getenv("BUYER_ENABLE_AUTH") == "true"
		var securityConfig SecurityConfig

		if enableAuth {
			// If auth is enabled, username and password are required (no defaults)
			username := os.Getenv("BUYER_USERNAME")
			password := os.Getenv("BUYER_PASSWORD")

			if username == "" {
				slog.Error("BUYER_USERNAME is required when BUYER_ENABLE_AUTH=true")
				fmt.Fprintln(os.Stderr, "Error: BUYER_USERNAME environment variable is required when authentication is enabled")
				os.Exit(1)
			}

			if password == "" {
				slog.Error("BUYER_PASSWORD is required when BUYER_ENABLE_AUTH=true")
				fmt.Fprintln(os.Stderr, "Error: BUYER_PASSWORD environment variable is required when authentication is enabled")
				os.Exit(1)
			}

			// Validate password strength
			if err := ValidatePassword(password); err != nil {
				slog.Error("invalid password", slog.String("error", err.Error()))
				fmt.Fprintf(os.Stderr, "Error: Invalid password - %v\n", err)
				fmt.Fprintln(os.Stderr, "Password requirements:")
				fmt.Fprintln(os.Stderr, "  - At least 12 characters long")
				fmt.Fprintln(os.Stderr, "  - Contains at least one uppercase letter")
				fmt.Fprintln(os.Stderr, "  - Contains at least one lowercase letter")
				fmt.Fprintln(os.Stderr, "  - Contains at least one digit")
				fmt.Fprintln(os.Stderr, "  - Contains at least one special character")
				os.Exit(1)
			}

			// Hash the password
			passwordHash, err := HashPassword(password)
			if err != nil {
				slog.Error("failed to hash password", slog.String("error", err.Error()))
				fmt.Fprintf(os.Stderr, "Error: Failed to hash password - %v\n", err)
				os.Exit(1)
			}

			securityConfig = SecurityConfig{
				EnableAuth:        true,
				EnableCSRF:        os.Getenv("BUYER_ENABLE_CSRF") == "true",
				EnableRateLimiter: true, // Always enabled for security
				Username:          username,
				PasswordHash:      passwordHash,
			}
		} else {
			securityConfig = SecurityConfig{
				EnableAuth:        false,
				EnableCSRF:        os.Getenv("BUYER_ENABLE_CSRF") == "true",
				EnableRateLimiter: true, // Always enabled for security
			}
		}

		SetupSecurityMiddleware(app, securityConfig)

		// Static files - extract subdirectory from embedded FS
		staticSubFS, err := fs.Sub(web.StaticFS, "static")
		if err != nil {
			slog.Error("failed to extract static files", slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Failed to extract static files: %v\n", err)
			os.Exit(1)
		}
		app.Use("/static", filesystem.New(filesystem.Config{
			Root: http.FS(staticSubFS),
		}))

		slog.Info("security middleware configured",
			slog.Bool("auth_enabled", securityConfig.EnableAuth),
			slog.Bool("csrf_enabled", securityConfig.EnableCSRF),
			slog.Bool("rate_limiter_enabled", securityConfig.EnableRateLimiter))

		// Services
		specSvc := services.NewSpecificationService(cfg.DB)
		brandSvc := services.NewBrandService(cfg.DB)
		productSvc := services.NewProductService(cfg.DB)
		vendorSvc := services.NewVendorService(cfg.DB)
		requisitionSvc := services.NewRequisitionService(cfg.DB)
		quoteSvc := services.NewQuoteService(cfg.DB)
		forexSvc := services.NewForexService(cfg.DB)
		dashboardSvc := services.NewDashboardService(cfg.DB)
		projectSvc := services.NewProjectService(cfg.DB)
		projectReqSvc := services.NewProjectRequisitionService(cfg.DB)

		slog.Debug("services initialized successfully")

		// Routes
		setupRoutes(app, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc, dashboardSvc, projectSvc, projectReqSvc)

		// Setup graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Start server in a goroutine
		addr := fmt.Sprintf(":%d", port)
		slog.Info("starting web server",
			slog.String("address", addr),
			slog.String("url", fmt.Sprintf("http://localhost%s", addr)))

		go func() {
			if err := app.Listen(addr); err != nil {
				slog.Error("failed to start server", slog.String("error", err.Error()))
				fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
			}
		}()

		// Wait for interrupt signal
		<-c
		slog.Info("shutting down server gracefully...")

		// Shutdown with timeout
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			slog.Error("server shutdown failed", slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		} else {
			slog.Info("server stopped gracefully")
		}
	},
}

func setupRoutes(
	app *fiber.App,
	specSvc *services.SpecificationService,
	brandSvc *services.BrandService,
	productSvc *services.ProductService,
	vendorSvc *services.VendorService,
	requisitionSvc *services.RequisitionService,
	quoteSvc *services.QuoteService,
	forexSvc *services.ForexService,
	dashboardSvc *services.DashboardService,
	projectSvc *services.ProjectService,
	projectReqSvc *services.ProjectRequisitionService,
) {
	// Home page
	app.Get("/", func(c *fiber.Ctx) error {
		return renderTemplate(c, "index.html", fiber.Map{
			"Title": "Home",
		})
	})

	// Dashboard page
	app.Get("/dashboard", func(c *fiber.Ctx) error {
		stats, err := dashboardSvc.GetStats()
		if err != nil {
			return err
		}

		vendorSpending, err := dashboardSvc.GetVendorSpending()
		if err != nil {
			return err
		}

		productPrices, err := dashboardSvc.GetProductPriceComparison()
		if err != nil {
			return err
		}

		expiryStats, err := dashboardSvc.GetExpiryStats()
		if err != nil {
			return err
		}

		recentQuotes, err := dashboardSvc.GetRecentQuotes(10)
		if err != nil {
			return err
		}

		return renderTemplate(c, "dashboard.html", fiber.Map{
			"Title":          "Dashboard",
			"Stats":          stats,
			"VendorSpending": vendorSpending,
			"ProductPrices":  productPrices,
			"ExpiryStats":    expiryStats,
			"RecentQuotes":   recentQuotes,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Dashboard", "Active": true},
			},
		})
	})

	// Help page
	app.Get("/help", func(c *fiber.Ctx) error {
		return renderTemplate(c, "help.html", fiber.Map{
			"Title": "Help & User Guide",
		})
	})

	// Specification routes
	app.Get("/specifications", func(c *fiber.Ctx) error {
		specs, err := specSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "specifications.html", fiber.Map{
			"Title":          "Specifications",
			"Specifications": specs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Specifications", "Active": true},
			},
		})
	})

	// Brand routes
	app.Get("/brands", func(c *fiber.Ctx) error {
		brands, err := brandSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "brands.html", fiber.Map{
			"Title":  "Brands",
			"Brands": brands,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Brands", "Active": true},
			},
		})
	})

	// Product routes
	app.Get("/products", func(c *fiber.Ctx) error {
		products, err := productSvc.List(0, 0)
		if err != nil {
			return err
		}
		brands, err := brandSvc.List(0, 0)
		if err != nil {
			return err
		}
		specs, err := specSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "products.html", fiber.Map{
			"Title":          "Products",
			"Products":       products,
			"Brands":         brands,
			"Specifications": specs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Products", "Active": true},
			},
		})
	})

	// Vendor routes
	app.Get("/vendors", func(c *fiber.Ctx) error {
		vendors, err := vendorSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "vendors.html", fiber.Map{
			"Title":   "Vendors",
			"Vendors": vendors,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Vendors", "Active": true},
			},
		})
	})

	// Quote routes
	app.Get("/quotes", func(c *fiber.Ctx) error {
		quotes, err := quoteSvc.List(0, 0)
		if err != nil {
			return err
		}
		vendors, err := vendorSvc.List(0, 0)
		if err != nil {
			return err
		}
		products, err := productSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "quotes.html", fiber.Map{
			"Title":    "Quotes",
			"Quotes":   quotes,
			"Vendors":  vendors,
			"Products": products,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Quotes", "Active": true},
			},
		})
	})

	// Requisition routes
	app.Get("/requisitions", func(c *fiber.Ctx) error {
		requisitions, err := requisitionSvc.List(0, 0)
		if err != nil {
			return err
		}
		specs, err := specSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "requisitions.html", fiber.Map{
			"Title":          "Requisitions",
			"Requisitions":   requisitions,
			"Specifications": specs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Requisitions", "Active": true},
			},
		})
	})

	// Forex routes
	app.Get("/forex", func(c *fiber.Ctx) error {
		rates, err := forexSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "forex.html", fiber.Map{
			"Title": "Forex Rates",
			"Rates": rates,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Forex Rates", "Active": true},
			},
		})
	})

	// Project routes
	app.Get("/projects", func(c *fiber.Ctx) error {
		projects, err := projectSvc.List(0, 0)
		if err != nil {
			return err
		}
		specs, err := specSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "projects.html", fiber.Map{
			"Title":          "Projects",
			"Projects":       projects,
			"Specifications": specs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Projects", "Active": true},
			},
		})
	})

	app.Get("/projects/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
		}

		project, err := projectSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Project not found")
		}

		specs, err := specSvc.List(0, 0)
		if err != nil {
			return err
		}

		projectReqs, err := projectReqSvc.ListByProject(uint(id))
		if err != nil {
			return err
		}

		return renderTemplate(c, "project-detail.html", fiber.Map{
			"Title":               project.Name,
			"Project":             project,
			"Specifications":      specs,
			"ProjectRequisitions": projectReqs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Projects", "URL": "/projects"},
				{"Name": project.Name, "Active": true},
			},
		})
	})

	app.Get("/projects/:id/dashboard", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
		}

		project, err := projectSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Project not found")
		}

		projectStats, err := dashboardSvc.GetProjectStats(uint(id))
		if err != nil {
			return err
		}

		bomItemQuantities, err := dashboardSvc.GetProjectBOMItemQuantities(uint(id))
		if err != nil {
			return err
		}

		requisitionBudgets, err := dashboardSvc.GetProjectRequisitionBudgets(uint(id))
		if err != nil {
			return err
		}

		return renderTemplate(c, "project-dashboard.html", fiber.Map{
			"Title":              project.Name + " - Dashboard",
			"Project":            project,
			"ProjectStats":       projectStats,
			"BOMItemQuantities":  bomItemQuantities,
			"RequisitionBudgets": requisitionBudgets,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Projects", "URL": "/projects"},
				{"Name": project.Name, "URL": fmt.Sprintf("/projects/%d", project.ID)},
				{"Name": "Dashboard", "Active": true},
			},
		})
	})

	// Requisition quote comparison routes
	app.Get("/requisition-comparison", func(c *fiber.Ctx) error {
		requisitions, err := requisitionSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "requisition-comparison.html", fiber.Map{
			"Title":        "Requisition Quote Comparison",
			"Requisitions": requisitions,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Requisitions", "URL": "/requisitions", "Active": false},
				{"Name": "Quote Comparison", "Active": true},
			},
		})
	})

	app.Get("/requisition-comparison/results", func(c *fiber.Ctx) error {
		requisitionIDStr := c.Query("requisition_id")
		if requisitionIDStr == "" {
			return c.SendString("<article><p class='error'>Please select a requisition</p></article>")
		}

		requisitionID, err := strconv.ParseUint(requisitionIDStr, 10, 32)
		if err != nil {
			return c.SendString("<article><p class='error'>Invalid requisition ID</p></article>")
		}

		comparison, err := requisitionSvc.GetQuoteComparison(uint(requisitionID), quoteSvc)
		if err != nil {
			return c.SendString(fmt.Sprintf("<article><p class='error'>Error: %v</p></article>", err))
		}

		html, err := RenderRequisitionComparison(comparison)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render comparison")
		}
		return c.SendString(html.String())
	})

	// Setup CRUD handlers for all entities
	SetupCRUDHandlers(app, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc)

	// Setup project handlers
	SetupProjectHandlers(app, projectSvc, specSvc, requisitionSvc, projectReqSvc)
}

func renderTemplate(c *fiber.Ctx, templateName string, data fiber.Map) error {
	// Create template with custom functions
	tmpl := template.New("base.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b float64) float64 { return a - b },
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	})

	// Parse base, components, and specific template
	tmpl, err := tmpl.ParseFS(web.TemplateFS, "templates/base.html", "templates/components.html", "templates/"+templateName)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")

	// Execute the base template
	return tmpl.ExecuteTemplate(c.Response().BodyWriter(), "base.html", data)
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
