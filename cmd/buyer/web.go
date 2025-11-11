package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

//go:embed web/templates/*.html
var templateFS embed.FS

//go:embed web/static
var staticFS embed.FS

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long:  "Start the FastAPI-inspired web server with HTMX support",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		app := fiber.New(fiber.Config{
			AppName: "Buyer v0.1.0",
		})

		// Middleware
		app.Use(recover.New())
		app.Use(logger.New())

		// Static files - extract subdirectory from embedded FS
		staticSubFS, err := fs.Sub(staticFS, "web/static")
		if err != nil {
			log.Fatal(err)
		}
		app.Use("/static", filesystem.New(filesystem.Config{
			Root: http.FS(staticSubFS),
		}))

		// Services
		brandSvc := services.NewBrandService(cfg.DB)
		productSvc := services.NewProductService(cfg.DB)
		vendorSvc := services.NewVendorService(cfg.DB)
		quoteSvc := services.NewQuoteService(cfg.DB)
		forexSvc := services.NewForexService(cfg.DB)

		// Routes
		setupRoutes(app, brandSvc, productSvc, vendorSvc, quoteSvc, forexSvc)

		// Start server
		addr := fmt.Sprintf(":%d", port)
		log.Printf("Starting web server on http://localhost%s\n", addr)
		if err := app.Listen(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
			os.Exit(1)
		}
	},
}

func setupRoutes(
	app *fiber.App,
	brandSvc *services.BrandService,
	productSvc *services.ProductService,
	vendorSvc *services.VendorService,
	quoteSvc *services.QuoteService,
	forexSvc *services.ForexService,
) {
	// Home page
	app.Get("/", func(c *fiber.Ctx) error {
		return renderTemplate(c, "index.html", fiber.Map{
			"Title": "Home",
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
		})
	})

	// Product routes
	app.Get("/products", func(c *fiber.Ctx) error {
		products, err := productSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "products.html", fiber.Map{
			"Title":    "Products",
			"Products": products,
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
		})
	})

	// Quote routes
	app.Get("/quotes", func(c *fiber.Ctx) error {
		quotes, err := quoteSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "quotes.html", fiber.Map{
			"Title":  "Quotes",
			"Quotes": quotes,
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
		})
	})
}

func renderTemplate(c *fiber.Ctx, templateName string, data fiber.Map) error {
	// Parse both base and specific template
	tmpl, err := template.ParseFS(templateFS, "web/templates/base.html", "web/templates/"+templateName)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")

	// Execute the base template
	return tmpl.Execute(c.Response().BodyWriter(), data)
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
}
