package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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
		securityConfig := SecurityConfig{
			EnableAuth:        os.Getenv("BUYER_ENABLE_AUTH") == "true",
			EnableCSRF:        os.Getenv("BUYER_ENABLE_CSRF") == "true",
			EnableRateLimiter: true, // Always enabled for security
			Username:          getEnvOrDefault("BUYER_USERNAME", "admin"),
			Password:          getEnvOrDefault("BUYER_PASSWORD", "admin"),
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

		// Start server
		addr := fmt.Sprintf(":%d", port)
		slog.Info("starting web server",
			slog.String("address", addr),
			slog.String("url", fmt.Sprintf("http://localhost%s", addr)))

		if err := app.Listen(addr); err != nil {
			slog.Error("failed to start server", slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
			os.Exit(1)
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

	// CRUD endpoints for Brands
	app.Post("/brands", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		brand, err := brandSvc.Create(name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderBrandRow(brand)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/brands/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		brand, err := brandSvc.Update(uint(id), name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderBrandRow(brand)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/brands/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := brandSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// CRUD endpoints for Products
	app.Post("/products", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		brandID, err := strconv.ParseUint(c.FormValue("brand_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid brand ID")
		}

		var specIDPtr *uint
		specIDStr := c.FormValue("specification_id")
		if specIDStr != "" {
			specID, err := strconv.ParseUint(specIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid specification ID")
			}
			specIDUint := uint(specID)
			specIDPtr = &specIDUint
		}

		product, err := productSvc.Create(name, uint(brandID), specIDPtr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderProductRow(product)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/products/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")

		var specIDPtr *uint
		specIDStr := c.FormValue("specification_id")
		if specIDStr != "" {
			specID, err := strconv.ParseUint(specIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid specification ID")
			}
			specIDUint := uint(specID)
			specIDPtr = &specIDUint
		}

		product, err := productSvc.Update(uint(id), name, specIDPtr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		brandName := ""
		if product.Brand != nil {
			brandName = product.Brand.Name
		}
		specName := "-"
		if product.Specification != nil {
			specName = product.Specification.Name
		}
		return c.SendString(fmt.Sprintf(`<tr id="product-%d">
			<td>%d</td>
			<td>
				<span class="product-name">%s</span>
				<form class="hidden edit-form" hx-put="/products/%d" hx-target="#product-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
					<select name="specification_id">
						<option value="">None</option>
					</select>
				</form>
			</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleProductEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/products/%d"
							hx-target="#product-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this product?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, product.ID, product.ID, product.Name, product.ID, product.ID, product.Name, brandName, specName, product.ID, product.ID, product.ID))
	})

	app.Delete("/products/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := productSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// CRUD endpoints for Vendors
	app.Post("/vendors", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		currency := c.FormValue("currency")
		discountCode := c.FormValue("discount_code")
		vendor, err := vendorSvc.Create(name, currency, discountCode)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="vendor-%d">
			<td>%d</td>
			<td>
				<span class="vendor-name">%s</span>
				<form class="hidden edit-form" hx-put="/vendors/%d" hx-target="#vendor-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleVendorEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/vendors/%d"
							hx-target="#vendor-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this vendor?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, vendor.ID, vendor.ID, vendor.Name, vendor.ID, vendor.ID, vendor.Name, vendor.Currency, vendor.DiscountCode, vendor.ID, vendor.ID, vendor.ID))
	})

	app.Put("/vendors/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		vendor, err := vendorSvc.Update(uint(id), name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="vendor-%d">
			<td>%d</td>
			<td>
				<span class="vendor-name">%s</span>
				<form class="hidden edit-form" hx-put="/vendors/%d" hx-target="#vendor-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleVendorEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/vendors/%d"
							hx-target="#vendor-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this vendor?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, vendor.ID, vendor.ID, vendor.Name, vendor.ID, vendor.ID, vendor.Name, vendor.Currency, vendor.DiscountCode, vendor.ID, vendor.ID, vendor.ID))
	})

	app.Delete("/vendors/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := vendorSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// CRUD endpoints for Forex
	app.Post("/forex", func(c *fiber.Ctx) error {
		fromCurrency := c.FormValue("from_currency")
		toCurrency := c.FormValue("to_currency")
		rateStr := c.FormValue("rate")
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid rate")
		}
		forex, err := forexSvc.Create(fromCurrency, toCurrency, rate, time.Now())
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="forex-%d">
			<td>%d</td>
			<td>%s</td>
			<td>%s</td>
			<td>%.4f</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm contrast"
							hx-delete="/forex/%d"
							hx-target="#forex-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this forex rate?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, forex.ID, forex.ID, forex.FromCurrency, forex.ToCurrency, forex.Rate, forex.EffectiveDate.Format("2006-01-02"), forex.ID, forex.ID))
	})

	app.Delete("/forex/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := forexSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// CRUD endpoints for Specifications
	app.Post("/specifications", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		description := c.FormValue("description")
		spec, err := specSvc.Create(name, description)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="spec-%d">
			<td>%d</td>
			<td>
				<span class="spec-name">%s</span>
				<form class="hidden edit-form" hx-put="/specifications/%d" hx-target="#spec-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
					<textarea name="description" rows="2">%s</textarea>
				</form>
			</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleSpecEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/specifications/%d"
							hx-target="#spec-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this specification?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, spec.ID, spec.ID, spec.Name, spec.ID, spec.ID, spec.Name, spec.Description, spec.Description, spec.ID, spec.ID, spec.ID))
	})

	app.Put("/specifications/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		description := c.FormValue("description")
		spec, err := specSvc.Update(uint(id), name, description)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="spec-%d">
			<td>%d</td>
			<td>
				<span class="spec-name">%s</span>
				<form class="hidden edit-form" hx-put="/specifications/%d" hx-target="#spec-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
					<textarea name="description" rows="2">%s</textarea>
				</form>
			</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleSpecEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/specifications/%d"
							hx-target="#spec-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this specification?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, spec.ID, spec.ID, spec.Name, spec.ID, spec.ID, spec.Name, spec.Description, spec.Description, spec.ID, spec.ID, spec.ID))
	})

	app.Delete("/specifications/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := specSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// CRUD endpoints for Requisitions
	app.Post("/requisitions", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		justification := c.FormValue("justification")

		// Parse budget
		var budget float64
		budgetStr := c.FormValue("budget")
		if budgetStr != "" {
			var err error
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		// Parse multiple line items from form data
		items := []services.RequisitionItemInput{}
		for i := 0; ; i++ {
			specIDStr := c.FormValue(fmt.Sprintf("items[%d][specification_id]", i))
			if specIDStr == "" {
				break // No more items
			}

			specID, err := strconv.ParseUint(specIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid specification ID for item %d", i))
			}

			quantityStr := c.FormValue(fmt.Sprintf("items[%d][quantity]", i))
			quantity, err := strconv.Atoi(quantityStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid quantity for item %d", i))
			}

			var itemBudget float64
			itemBudgetStr := c.FormValue(fmt.Sprintf("items[%d][budget_per_unit]", i))
			if itemBudgetStr != "" {
				itemBudget, err = strconv.ParseFloat(itemBudgetStr, 64)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid budget for item %d", i))
				}
			}

			description := c.FormValue(fmt.Sprintf("items[%d][description]", i))

			items = append(items, services.RequisitionItemInput{
				SpecificationID: uint(specID),
				Quantity:        quantity,
				BudgetPerUnit:   itemBudget,
				Description:     description,
			})
		}

		if len(items) == 0 {
			return c.Status(fiber.StatusBadRequest).SendString("At least one line item is required")
		}

		req, err := requisitionSvc.Create(name, justification, budget, items)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		itemsHTML := ""
		for _, item := range req.Items {
			specName := ""
			if item.Specification != nil {
				specName = item.Specification.Name
			}
			budgetDisplay := ""
			if item.BudgetPerUnit > 0 {
				budgetDisplay = fmt.Sprintf(", Budget/unit: %.2f", item.BudgetPerUnit)
			}
			descDisplay := ""
			if item.Description != "" {
				descDisplay = fmt.Sprintf(" - %s", item.Description)
			}
			itemsHTML += fmt.Sprintf("<li>%s (Qty: %d%s)%s</li>", specName, item.Quantity, budgetDisplay, descDisplay)
		}

		justificationDisplay := ""
		if req.Justification != "" {
			justificationDisplay = fmt.Sprintf("<br><small>%s</small>", req.Justification)
		}
		budgetDisplay := ""
		if req.Budget > 0 {
			budgetDisplay = fmt.Sprintf("<br><strong>Budget: %.2f</strong>", req.Budget)
		}

		return c.SendString(fmt.Sprintf(`<tr id="req-%d">
			<td>%d</td>
			<td>%s%s%s</td>
			<td><ul>%s</ul></td>
			<td>
				<div class="actions">
					<button class="btn-sm contrast"
							hx-delete="/requisitions/%d"
							hx-target="#req-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this requisition?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, req.ID, req.ID, req.Name, justificationDisplay, budgetDisplay, itemsHTML, req.ID, req.ID))
	})

	app.Put("/requisitions/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		name := c.FormValue("name")
		justification := c.FormValue("justification")

		var budget float64
		budgetStr := c.FormValue("budget")
		if budgetStr != "" {
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		req, err := requisitionSvc.Update(uint(id), name, justification, budget)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		itemsHTML := ""
		for _, item := range req.Items {
			specName := ""
			if item.Specification != nil {
				specName = item.Specification.Name
			}
			budgetDisplay := ""
			if item.BudgetPerUnit > 0 {
				budgetDisplay = fmt.Sprintf(", Budget/unit: %.2f", item.BudgetPerUnit)
			}
			descDisplay := ""
			if item.Description != "" {
				descDisplay = fmt.Sprintf(" - %s", item.Description)
			}
			itemsHTML += fmt.Sprintf("<li>%s (Qty: %d%s)%s</li>", specName, item.Quantity, budgetDisplay, descDisplay)
		}

		justificationDisplay := ""
		if req.Justification != "" {
			justificationDisplay = fmt.Sprintf("<br><small>%s</small>", req.Justification)
		}
		budgetDisplayStr := ""
		if req.Budget > 0 {
			budgetDisplayStr = fmt.Sprintf("<br><strong>Budget: %.2f</strong>", req.Budget)
		}

		return c.SendString(fmt.Sprintf(`<tr id="req-%d">
			<td>%d</td>
			<td>
				<span class="req-name">%s%s%s</span>
				<form class="hidden edit-form" hx-put="/requisitions/%d" hx-target="#req-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
					<textarea name="justification" rows="2" placeholder="Justification...">%s</textarea>
					<input type="number" name="budget" value="%.2f" step="0.01" min="0" placeholder="Budget...">
				</form>
			</td>
			<td><ul>%s</ul></td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleReqEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/requisitions/%d"
							hx-target="#req-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this requisition?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, req.ID, req.ID, req.Name, justificationDisplay, budgetDisplayStr, req.ID, req.ID, req.Name, req.Justification, req.Budget, itemsHTML, req.ID, req.ID, req.ID))
	})

	// Full requisition update (requisition + all items)
	app.Put("/requisitions/:id/full", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		name := c.FormValue("name")
		justification := c.FormValue("justification")

		var budget float64
		budgetStr := c.FormValue("budget")
		if budgetStr != "" {
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		// Update requisition details
		req, err := requisitionSvc.Update(uint(id), name, justification, budget)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		// Process line items - track which ones to keep/update/delete
		existingItemIDs := make(map[uint]bool)
		for _, item := range req.Items {
			existingItemIDs[item.ID] = false // Mark as not seen yet
		}

		// Update or add items
		for i := 0; ; i++ {
			itemIDStr := c.FormValue(fmt.Sprintf("items[%d][id]", i))
			if itemIDStr == "" {
				break
			}

			itemID, _ := strconv.ParseUint(itemIDStr, 10, 32)
			specIDStr := c.FormValue(fmt.Sprintf("items[%d][specification_id]", i))
			if specIDStr == "" {
				continue
			}

			specID, err := strconv.ParseUint(specIDStr, 10, 32)
			if err != nil {
				continue
			}

			quantityStr := c.FormValue(fmt.Sprintf("items[%d][quantity]", i))
			quantity, err := strconv.Atoi(quantityStr)
			if err != nil {
				continue
			}

			var itemBudget float64
			itemBudgetStr := c.FormValue(fmt.Sprintf("items[%d][budget_per_unit]", i))
			if itemBudgetStr != "" {
				itemBudget, _ = strconv.ParseFloat(itemBudgetStr, 64)
			}

			description := c.FormValue(fmt.Sprintf("items[%d][description]", i))

			if itemID == 0 {
				// New item
				_, _ = requisitionSvc.AddItem(uint(id), uint(specID), quantity, itemBudget, description)
			} else {
				// Update existing item
				_, _ = requisitionSvc.UpdateItem(uint(itemID), uint(specID), quantity, itemBudget, description)
				existingItemIDs[uint(itemID)] = true // Mark as seen
			}
		}

		// Delete items that weren't in the form
		for itemID, wasSeen := range existingItemIDs {
			if !wasSeen {
				_ = requisitionSvc.DeleteItem(itemID)
			}
		}

		// Reload requisition with updated items
		req, _ = requisitionSvc.GetByID(uint(id))

		// Return just the main row (details/edit rows will be hidden)
		return c.SendString(fmt.Sprintf(`<tr id="req-%d">
			<td>%d</td>
			<td>
				<strong>%s</strong>
				%s
			</td>
			<td>%d</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm" onclick="toggleDetails(%d)">Details</button>
					<button class="btn-sm secondary" onclick="toggleEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/requisitions/%d"
							hx-target="#req-%d"
							hx-swap="delete"
							hx-confirm="Are you sure you want to delete this requisition?">
						Delete
					</button>
				</div>
			</td>
		</tr>`,
			req.ID,
			req.ID,
			req.Name,
			func() string {
				if req.Justification != "" {
					return fmt.Sprintf("<br><small>%s</small>", req.Justification)
				}
				return ""
			}(),
			len(req.Items),
			func() string {
				if req.Budget > 0 {
					return fmt.Sprintf("%.2f", req.Budget)
				}
				return "-"
			}(),
			req.ID,
			req.ID,
			req.ID,
			req.ID))
	})

	app.Delete("/requisitions/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := requisitionSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// Line item endpoints
	app.Put("/requisition-items/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		specID, err := strconv.ParseUint(c.FormValue("specification_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid specification ID")
		}

		quantity, err := strconv.Atoi(c.FormValue("quantity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid quantity")
		}

		var budget float64
		budgetStr := c.FormValue("budget_per_unit")
		if budgetStr != "" {
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		description := c.FormValue("description")

		item, err := requisitionSvc.UpdateItem(uint(id), uint(specID), quantity, budget, description)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		specName := ""
		if item.Specification != nil {
			specName = item.Specification.Name
		}
		budgetDisplay := ""
		if item.BudgetPerUnit > 0 {
			budgetDisplay = fmt.Sprintf(", Budget/unit: %.2f", item.BudgetPerUnit)
		}
		descDisplay := ""
		if item.Description != "" {
			descDisplay = fmt.Sprintf(" - %s", item.Description)
		}

		// Get specifications for dropdown
		specs, _ := specSvc.List(0, 0)
		specsOptions := ""
		for _, s := range specs {
			selected := ""
			if s.ID == item.SpecificationID {
				selected = "selected"
			}
			specsOptions += fmt.Sprintf(`<option value="%d" %s>%s</option>`, s.ID, selected, s.Name)
		}

		return c.SendString(fmt.Sprintf(`<li id="item-%d">
			<span class="item-display">%s (Qty: %d%s)%s</span>
			<form class="hidden item-edit-form" hx-put="/requisition-items/%d" hx-target="#item-%d" hx-swap="outerHTML">
				<select name="specification_id" required>%s</select>
				<input type="number" name="quantity" value="%d" min="1" required>
				<input type="number" name="budget_per_unit" value="%.2f" step="0.01" min="0" placeholder="Budget/unit">
				<input type="text" name="description" value="%s" placeholder="Description">
			</form>
			<button class="btn-sm secondary" onclick="toggleItemEdit(%d)">Edit</button>
			<button class="btn-sm contrast"
					hx-delete="/requisition-items/%d"
					hx-target="#item-%d"
					hx-swap="outerHTML"
					hx-confirm="Delete this item?">
				Delete
			</button>
		</li>`, item.ID, specName, item.Quantity, budgetDisplay, descDisplay, item.ID, item.ID, specsOptions, item.Quantity, item.BudgetPerUnit, item.Description, item.ID, item.ID, item.ID))
	})

	app.Delete("/requisition-items/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := requisitionSvc.DeleteItem(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

	// CRUD endpoints for Quotes
	app.Post("/quotes", func(c *fiber.Ctx) error {
		vendorID, err := strconv.ParseUint(c.FormValue("vendor_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid vendor ID")
		}
		productID, err := strconv.ParseUint(c.FormValue("product_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid product ID")
		}
		priceStr := c.FormValue("price")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid price")
		}
		currency := c.FormValue("currency")
		notes := c.FormValue("notes")

		// Parse valid_until if provided
		var validUntil *time.Time
		validUntilStr := c.FormValue("valid_until")
		if validUntilStr != "" {
			parsed, err := time.Parse("2006-01-02", validUntilStr)
			if err == nil {
				validUntil = &parsed
			}
		}

		quote, err := quoteSvc.Create(services.CreateQuoteInput{
			VendorID:   uint(vendorID),
			ProductID:  uint(productID),
			Price:      price,
			Currency:   currency,
			ValidUntil: validUntil,
			Notes:      notes,
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		vendorName := ""
		if quote.Vendor != nil {
			vendorName = quote.Vendor.Name
		}
		productName := ""
		if quote.Product != nil {
			productName = quote.Product.Name
		}

		// Format expiry display
		expiryDisplay := `<span style="color: gray;">—</span>`
		if quote.ValidUntil != nil {
			days := quote.DaysUntilExpiration()
			if days < 0 {
				expiryDisplay = fmt.Sprintf(`<span style="color: red; font-weight: bold;">%d</span>`, days)
			} else if days < 7 {
				expiryDisplay = fmt.Sprintf(`<span style="color: red; font-weight: bold;">%d</span>`, days)
			} else if days < 30 {
				expiryDisplay = fmt.Sprintf(`<span style="color: orange; font-weight: bold;">%d</span>`, days)
			} else {
				expiryDisplay = fmt.Sprintf(`<span style="color: green;">%d</span>`, days)
			}
		}

		return c.SendString(fmt.Sprintf(`<tr id="quote-%d">
			<td>%d</td>
			<td>%s</td>
			<td>%s</td>
			<td>%.2f</td>
			<td>%s</td>
			<td>%.2f</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<div class="actions">
					<button class="btn-sm contrast"
							hx-delete="/quotes/%d"
							hx-target="#quote-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this quote?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, quote.ID, quote.ID, vendorName, productName, quote.Price, quote.Currency, quote.ConvertedPrice, quote.QuoteDate.Format("2006-01-02"), expiryDisplay, quote.ID, quote.ID))
	})

	app.Delete("/quotes/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := quoteSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString("")
	})

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
