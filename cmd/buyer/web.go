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
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"github.com/shakfu/buyer/web"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
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
		poSvc := services.NewPurchaseOrderService(cfg.DB)
		docSvc := services.NewDocumentService(cfg.DB)
		ratingsSvc := services.NewVendorRatingService(cfg.DB)

		slog.Debug("services initialized successfully")

		// Routes
		setupRoutes(app, cfg.DB, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc, dashboardSvc, projectSvc, projectReqSvc, poSvc, docSvc, ratingsSvc)

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
	db *gorm.DB,
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
	poSvc *services.PurchaseOrderService,
	docSvc *services.DocumentService,
	ratingsSvc *services.VendorRatingService,
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

	// Specification attributes management routes
	app.Get("/specifications/:id/attributes", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid specification ID")
		}

		spec, err := specSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(404).SendString("Specification not found")
		}

		// Get attributes for this specification
		var attrs []models.SpecificationAttribute
		if err := db.Where("specification_id = ?", id).Order("name ASC").Find(&attrs).Error; err != nil {
			return err
		}

		return renderTemplate(c, "specification-attributes.html", fiber.Map{
			"Title":         spec.Name + " - Attributes",
			"Specification": spec,
			"Attributes":    attrs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Specifications", "URL": "/specifications"},
				{"Name": spec.Name, "Active": true},
			},
		})
	})

	app.Post("/specifications/:id/attributes", func(c *fiber.Ctx) error {
		specID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid specification ID")
		}

		// Parse form data
		name := strings.TrimSpace(c.FormValue("name"))
		dataType := c.FormValue("data_type")
		unit := strings.TrimSpace(c.FormValue("unit"))
		description := strings.TrimSpace(c.FormValue("description"))
		isRequired := c.FormValue("is_required") == "on"

		if name == "" {
			return c.Status(400).SendString("Attribute name is required")
		}

		if dataType != "number" && dataType != "text" && dataType != "boolean" {
			return c.Status(400).SendString("Invalid data type")
		}

		// Parse min/max values for number types
		var minValue, maxValue *float64
		if dataType == "number" {
			if minStr := c.FormValue("min_value"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minValue = &val
				}
			}
			if maxStr := c.FormValue("max_value"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxValue = &val
				}
			}
		}

		attr := &models.SpecificationAttribute{
			SpecificationID: uint(specID),
			Name:            name,
			DataType:        dataType,
			Unit:            unit,
			IsRequired:      isRequired,
			MinValue:        minValue,
			MaxValue:        maxValue,
			Description:     description,
		}

		if err := db.Create(attr).Error; err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Render new row
		html := fmt.Sprintf(`
		<tr id="attr-%d">
			<td><strong>%s</strong></td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<button class="btn-sm contrast"
						hx-delete="/specifications/%d/attributes/%d"
						hx-target="#attr-%d"
						hx-swap="outerHTML"
						hx-confirm="Are you sure? Products using this attribute will lose their values.">
					Delete
				</button>
			</td>
		</tr>`,
			attr.ID,
			attr.Name,
			attr.DataType,
			func() string {
				if attr.Unit != "" {
					return attr.Unit
				}
				return "-"
			}(),
			func() string {
				if attr.IsRequired {
					return "Yes"
				}
				return "No"
			}(),
			func() string {
				if attr.DataType == "number" {
					if attr.MinValue != nil && attr.MaxValue != nil {
						return fmt.Sprintf("Min: %.2f Max: %.2f", *attr.MinValue, *attr.MaxValue)
					} else if attr.MinValue != nil {
						return fmt.Sprintf("Min: %.2f", *attr.MinValue)
					} else if attr.MaxValue != nil {
						return fmt.Sprintf("Max: %.2f", *attr.MaxValue)
					}
				}
				return "-"
			}(),
			func() string {
				if attr.Description != "" {
					return attr.Description
				}
				return "-"
			}(),
			specID, attr.ID, attr.ID,
		)

		return c.SendString(html)
	})

	app.Delete("/specifications/:specId/attributes/:attrId", func(c *fiber.Ctx) error {
		attrID, err := c.ParamsInt("attrId")
		if err != nil {
			return c.Status(400).SendString("Invalid attribute ID")
		}

		// Delete the attribute (cascade will delete product attributes)
		if err := db.Delete(&models.SpecificationAttribute{}, attrID).Error; err != nil {
			return c.Status(400).SendString(err.Error())
		}

		return c.SendString("")
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

	app.Get("/products/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid product ID")
		}
		product, err := productSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(404).SendString("Product not found")
		}

		// Load available attributes for this product's specification
		var availableAttrs []models.SpecificationAttribute
		if product.SpecificationID != nil {
			db.Where("specification_id = ?", *product.SpecificationID).
				Order("name ASC").Find(&availableAttrs)
		}

		return renderTemplate(c, "product-detail.html", fiber.Map{
			"Title":               product.Name,
			"Product":             product,
			"AvailableAttributes": availableAttrs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Products", "URL": "/products"},
				{"Name": product.Name, "Active": true},
			},
		})
	})

	// Product comparison route
	app.Get("/products/compare/:specId", func(c *fiber.Ctx) error {
		specID, err := c.ParamsInt("specId")
		if err != nil {
			return c.Status(400).SendString("Invalid specification ID")
		}

		// Get specification
		spec, err := specSvc.GetByID(uint(specID))
		if err != nil {
			return c.Status(404).SendString("Specification not found")
		}

		// Get all products for this specification
		products, err := productSvc.ListBySpecification(uint(specID))
		if err != nil {
			return err
		}

		// Extract unique attribute names from products
		type AttrInfo struct {
			ID   uint
			Name string
			Unit string
		}
		attrMap := make(map[string]AttrInfo)
		for _, product := range products {
			for _, attr := range product.Attributes {
				if attr.SpecificationAttribute != nil {
					name := attr.SpecificationAttribute.Name
					if _, exists := attrMap[name]; !exists {
						attrMap[name] = AttrInfo{
							ID:   attr.SpecificationAttribute.ID,
							Name: name,
							Unit: attr.SpecificationAttribute.Unit,
						}
					}
				}
			}
		}

		// Convert to ordered slice
		var attrNames []AttrInfo
		for _, info := range attrMap {
			attrNames = append(attrNames, info)
		}

		return renderTemplate(c, "product-comparison.html", fiber.Map{
			"Title":          "Compare Products - " + spec.Name,
			"Specification":  spec,
			"Products":       products,
			"AttributeNames": attrNames,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Products", "URL": "/products"},
				{"Name": "Compare " + spec.Name, "Active": true},
			},
		})
	})

	// Quote comparison matrix - specification level
	app.Get("/quotes/compare/specification/:specId", func(c *fiber.Ctx) error {
		specID, err := c.ParamsInt("specId")
		if err != nil {
			return c.Status(400).SendString("Invalid specification ID")
		}

		showExtra := c.Query("show_extra") == "on"

		matrix, err := quoteSvc.GetQuoteComparisonMatrix(uint(specID), showExtra)
		if err != nil {
			return c.Status(404).SendString(err.Error())
		}

		// Build helper maps for template using index-based keys
		attributeCompliance := make(map[int]map[uint]bool)
		productAttrValues := make(map[int]map[uint]*models.ProductAttribute)
		extraAttrsByQuote := make(map[int]map[string]*models.ProductAttribute)
		extraAttrNamesSet := make(map[string]bool)

		for i := range matrix.QuoteComparisons {
			comp := &matrix.QuoteComparisons[i]
			attributeCompliance[i] = comp.AttributeCompliance

			// Build product attribute value map
			prodAttrMap := make(map[uint]*models.ProductAttribute)
			for j := range comp.Quote.Product.Attributes {
				attr := &comp.Quote.Product.Attributes[j]
				prodAttrMap[attr.SpecificationAttributeID] = attr
			}
			productAttrValues[i] = prodAttrMap

			// Build extra attributes map if showing extras
			if showExtra {
				extraMap := make(map[string]*models.ProductAttribute)
				for j := range comp.ExtraAttributes {
					attr := &comp.ExtraAttributes[j]
					if attr.SpecificationAttribute != nil {
						name := attr.SpecificationAttribute.Name
						extraMap[name] = attr
						extraAttrNamesSet[name] = true
					}
				}
				extraAttrsByQuote[i] = extraMap
			}
		}

		// Convert extra attr names to ordered slice
		var extraAttrNames []string
		for name := range extraAttrNamesSet {
			extraAttrNames = append(extraAttrNames, name)
		}

		return renderTemplate(c, "quote-comparison-matrix.html", fiber.Map{
			"Title":               "Quote Comparison - " + matrix.Specification.Name,
			"Matrix":              matrix,
			"AttributeCompliance": attributeCompliance,
			"ProductAttrValues":   productAttrValues,
			"ExtraAttrsByQuote":   extraAttrsByQuote,
			"ExtraAttrNames":      extraAttrNames,
			"ExtraAttrCount":      len(extraAttrNames),
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Specifications", "URL": "/specifications"},
				{"Name": "Quote Comparison", "Active": true},
			},
		})
	})

	// Quote comparison matrix - product level
	app.Get("/quotes/compare/product/:productId", func(c *fiber.Ctx) error {
		productID, err := c.ParamsInt("productId")
		if err != nil {
			return c.Status(400).SendString("Invalid product ID")
		}

		showExtra := c.Query("show_extra") == "on"

		matrix, err := quoteSvc.GetProductQuoteComparisonMatrix(uint(productID), showExtra)
		if err != nil {
			return c.Status(404).SendString(err.Error())
		}

		// Build helper maps for template using index-based keys
		attributeCompliance := make(map[int]map[uint]bool)
		productAttrValues := make(map[int]map[uint]*models.ProductAttribute)
		extraAttrsByQuote := make(map[int]map[string]*models.ProductAttribute)
		extraAttrNamesSet := make(map[string]bool)

		for i := range matrix.QuoteComparisons {
			comp := &matrix.QuoteComparisons[i]
			attributeCompliance[i] = comp.AttributeCompliance

			prodAttrMap := make(map[uint]*models.ProductAttribute)
			for j := range comp.Quote.Product.Attributes {
				attr := &comp.Quote.Product.Attributes[j]
				prodAttrMap[attr.SpecificationAttributeID] = attr
			}
			productAttrValues[i] = prodAttrMap

			if showExtra {
				extraMap := make(map[string]*models.ProductAttribute)
				for j := range comp.ExtraAttributes {
					attr := &comp.ExtraAttributes[j]
					if attr.SpecificationAttribute != nil {
						name := attr.SpecificationAttribute.Name
						extraMap[name] = attr
						extraAttrNamesSet[name] = true
					}
				}
				extraAttrsByQuote[i] = extraMap
			}
		}

		var extraAttrNames []string
		for name := range extraAttrNamesSet {
			extraAttrNames = append(extraAttrNames, name)
		}

		// Get product name for title
		var product models.Product
		db.First(&product, productID)

		return renderTemplate(c, "quote-comparison-matrix.html", fiber.Map{
			"Title":               "Quote Comparison - " + product.Name,
			"Matrix":              matrix,
			"AttributeCompliance": attributeCompliance,
			"ProductAttrValues":   productAttrValues,
			"ExtraAttrsByQuote":   extraAttrsByQuote,
			"ExtraAttrNames":      extraAttrNames,
			"ExtraAttrCount":      len(extraAttrNames),
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Products", "URL": "/products"},
				{"Name": "Quote Comparison", "Active": true},
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

	app.Get("/vendors/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid vendor ID")
		}
		vendor, err := vendorSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(404).SendString("Vendor not found")
		}
		return renderTemplate(c, "vendor-detail.html", fiber.Map{
			"Title":  vendor.Name,
			"Vendor": vendor,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Vendors", "URL": "/vendors"},
				{"Name": vendor.Name, "Active": true},
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

	app.Get("/quotes/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid quote ID")
		}
		quote, err := quoteSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(404).SendString("Quote not found")
		}
		return renderTemplate(c, "quote-detail.html", fiber.Map{
			"Title": fmt.Sprintf("Quote #%d", quote.ID),
			"Quote": quote,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Quotes", "URL": "/quotes"},
				{"Name": fmt.Sprintf("Quote #%d", quote.ID), "Active": true},
			},
		})
	})

	// Purchase Order routes
	app.Get("/purchase-orders", func(c *fiber.Ctx) error {
		orders, err := poSvc.List(0, 0)
		if err != nil {
			return err
		}
		quotes, err := quoteSvc.List(0, 0)
		if err != nil {
			return err
		}
		requisitions, err := requisitionSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "purchase-orders.html", fiber.Map{
			"Title":          "Purchase Orders",
			"PurchaseOrders": orders,
			"Quotes":         quotes,
			"Requisitions":   requisitions,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Purchase Orders", "Active": true},
			},
		})
	})

	app.Get("/purchase-orders/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid purchase order ID")
		}
		po, err := poSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(404).SendString("Purchase order not found")
		}
		return renderTemplate(c, "purchase-order-detail.html", fiber.Map{
			"Title":         po.PONumber,
			"PurchaseOrder": po,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Purchase Orders", "URL": "/purchase-orders"},
				{"Name": po.PONumber, "Active": true},
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

	// Document routes
	app.Get("/documents", func(c *fiber.Ctx) error {
		docs, err := docSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "documents.html", fiber.Map{
			"Title":     "Documents",
			"Documents": docs,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Documents", "Active": true},
			},
		})
	})

	// Vendor Rating routes
	app.Get("/vendor-ratings", func(c *fiber.Ctx) error {
		ratings, err := ratingsSvc.List(0, 0)
		if err != nil {
			return err
		}
		vendors, err := vendorSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "vendor-ratings.html", fiber.Map{
			"Title":   "Vendor Ratings",
			"Ratings": ratings,
			"Vendors": vendors,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Vendor Ratings", "Active": true},
			},
		})
	})

	// Vendor Performance Dashboard
	app.Get("/vendor-performance", func(c *fiber.Ctx) error {
		performance, err := ratingsSvc.GetVendorPerformance()
		if err != nil {
			return err
		}

		categoryAverages, err := ratingsSvc.GetCategoryAverages()
		if err != nil {
			return err
		}

		totalRatings, err := ratingsSvc.Count()
		if err != nil {
			return err
		}

		// Count vendors that have been rated
		ratedVendors := len(performance)

		// Get total vendor count
		allVendors, err := vendorSvc.List(0, 0)
		if err != nil {
			return err
		}

		// Calculate average overall rating
		avgOverall := 0.0
		if len(performance) > 0 {
			sum := 0.0
			for _, p := range performance {
				sum += p.AvgRating
			}
			avgOverall = sum / float64(len(performance))
		}

		// Find top vendor
		var topVendor *services.VendorPerformance
		if len(performance) > 0 {
			topVendor = &performance[0]
		}

		return renderTemplate(c, "vendor-performance.html", fiber.Map{
			"Title":            "Vendor Performance Dashboard",
			"VendorRatings":    performance,
			"CategoryAverages": categoryAverages,
			"TotalRatings":     totalRatings,
			"RatedVendors":     ratedVendors,
			"TotalVendors":     len(allVendors),
			"AvgOverallRating": avgOverall,
			"TopVendor":        topVendor,
			"Breadcrumb": []map[string]interface{}{
				{"Name": "Vendor Performance", "Active": true},
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
	SetupCRUDHandlers(app, db, specSvc, brandSvc, productSvc, vendorSvc, requisitionSvc, quoteSvc, forexSvc)

	// Setup export/import handlers
	SetupExportHandlers(app, db)

	// Purchase Order CRUD handlers
	app.Post("/purchase-orders", func(c *fiber.Ctx) error {
		quoteID, err := strconv.ParseUint(c.FormValue("quote_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid quote ID")
		}

		var reqIDPtr *uint
		reqIDStr := c.FormValue("requisition_id")
		if reqIDStr != "" {
			reqID, err := strconv.ParseUint(reqIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid requisition ID")
			}
			reqIDUint := uint(reqID)
			reqIDPtr = &reqIDUint
		}

		quantity, err := strconv.Atoi(c.FormValue("quantity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid quantity")
		}

		var expectedDelivery *time.Time
		expectedDeliveryStr := c.FormValue("expected_delivery")
		if expectedDeliveryStr != "" {
			parsed, err := time.Parse("2006-01-02", expectedDeliveryStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid expected delivery date")
			}
			expectedDelivery = &parsed
		}

		shippingCost, _ := strconv.ParseFloat(c.FormValue("shipping_cost"), 64)
		tax, _ := strconv.ParseFloat(c.FormValue("tax"), 64)

		po, err := poSvc.Create(services.CreatePurchaseOrderInput{
			QuoteID:          uint(quoteID),
			RequisitionID:    reqIDPtr,
			PONumber:         c.FormValue("po_number"),
			Quantity:         quantity,
			ExpectedDelivery: expectedDelivery,
			ShippingCost:     shippingCost,
			Tax:              tax,
			Notes:            c.FormValue("notes"),
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderPurchaseOrderRow(po)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/purchase-orders/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		status := c.Query("status")
		invoice := c.Query("invoice")
		actualDeliveryStr := c.Query("actual_delivery")

		if status != "" {
			_, err := poSvc.UpdateStatus(uint(id), status)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
			}
		}

		if invoice != "" {
			_, err := poSvc.UpdateInvoiceNumber(uint(id), invoice)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
			}
		}

		if actualDeliveryStr != "" {
			parsed, err := time.Parse("2006-01-02", actualDeliveryStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid actual delivery date")
			}
			_, err = poSvc.UpdateDeliveryDates(uint(id), nil, &parsed)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
			}
		}

		po, err := poSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderPurchaseOrderRow(po)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/purchase-orders/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := poSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Document CRUD handlers
	app.Post("/documents", func(c *fiber.Ctx) error {
		entityType := c.FormValue("entity_type")
		entityID, err := strconv.ParseUint(c.FormValue("entity_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid entity ID")
		}

		fileSize, _ := strconv.ParseInt(c.FormValue("file_size"), 10, 64)

		doc, err := docSvc.Create(services.CreateDocumentInput{
			EntityType:  entityType,
			EntityID:    uint(entityID),
			FileName:    c.FormValue("file_name"),
			FileType:    c.FormValue("file_type"),
			FileSize:    fileSize,
			FilePath:    c.FormValue("file_path"),
			Description: c.FormValue("description"),
			UploadedBy:  c.FormValue("uploaded_by"),
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderDocumentRow(doc)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/documents/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := docSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Vendor Rating CRUD handlers
	app.Post("/vendor-ratings", func(c *fiber.Ctx) error {
		vendorID, err := strconv.ParseUint(c.FormValue("vendor_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid vendor ID")
		}

		var poIDPtr *uint
		poIDStr := c.FormValue("po_id")
		if poIDStr != "" {
			poID, err := strconv.ParseUint(poIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid PO ID")
			}
			poIDUint := uint(poID)
			poIDPtr = &poIDUint
		}

		var pricePtr, qualityPtr, deliveryPtr, servicePtr *int
		if priceStr := c.FormValue("price_rating"); priceStr != "" {
			if price, err := strconv.Atoi(priceStr); err == nil {
				pricePtr = &price
			}
		}
		if qualityStr := c.FormValue("quality_rating"); qualityStr != "" {
			if quality, err := strconv.Atoi(qualityStr); err == nil {
				qualityPtr = &quality
			}
		}
		if deliveryStr := c.FormValue("delivery_rating"); deliveryStr != "" {
			if delivery, err := strconv.Atoi(deliveryStr); err == nil {
				deliveryPtr = &delivery
			}
		}
		if serviceStr := c.FormValue("service_rating"); serviceStr != "" {
			if service, err := strconv.Atoi(serviceStr); err == nil {
				servicePtr = &service
			}
		}

		rating, err := ratingsSvc.Create(services.CreateVendorRatingInput{
			VendorID:        uint(vendorID),
			PurchaseOrderID: poIDPtr,
			PriceRating:     pricePtr,
			QualityRating:   qualityPtr,
			DeliveryRating:  deliveryPtr,
			ServiceRating:   servicePtr,
			Comments:        c.FormValue("comments"),
			RatedBy:         c.FormValue("rated_by"),
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderVendorRatingRow(rating)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/vendor-ratings/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := ratingsSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Setup project handlers
	SetupProjectHandlers(app, projectSvc, specSvc, requisitionSvc, projectReqSvc)

	// Setup procurement handlers
	registerProcurementRoutes(app)
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
		"deref": func(ptr interface{}) interface{} {
			if ptr == nil {
				return nil
			}
			switch v := ptr.(type) {
			case *float64:
				if v == nil {
					return 0.0
				}
				return *v
			case *string:
				if v == nil {
					return ""
				}
				return *v
			case *bool:
				if v == nil {
					return false
				}
				return *v
			case *int:
				if v == nil {
					return 0
				}
				return *v
			default:
				return ptr
			}
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
