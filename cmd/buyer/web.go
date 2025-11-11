package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

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
		brands, err := brandSvc.List(0, 0)
		if err != nil {
			return err
		}
		return renderTemplate(c, "products.html", fiber.Map{
			"Title":    "Products",
			"Products": products,
			"Brands":   brands,
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

	// CRUD endpoints for Brands
	app.Post("/brands", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		brand, err := brandSvc.Create(name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="brand-%d">
			<td>%d</td>
			<td>
				<span class="brand-name">%s</span>
				<form class="hidden edit-form" hx-put="/brands/%d" hx-target="#brand-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/brands/%d"
							hx-target="#brand-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this brand?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, brand.ID, brand.ID, brand.Name, brand.ID, brand.ID, brand.Name, brand.ID, brand.ID, brand.ID))
	})

	app.Put("/brands/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		brand, err := brandSvc.Update(uint(id), name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf(`<tr id="brand-%d">
			<td>%d</td>
			<td>
				<span class="brand-name">%s</span>
				<form class="hidden edit-form" hx-put="/brands/%d" hx-target="#brand-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
			<td>
				<div class="actions">
					<button class="btn-sm secondary" onclick="toggleEdit(%d)">Edit</button>
					<button class="btn-sm contrast"
							hx-delete="/brands/%d"
							hx-target="#brand-%d"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to delete this brand?">
						Delete
					</button>
				</div>
			</td>
		</tr>`, brand.ID, brand.ID, brand.Name, brand.ID, brand.ID, brand.Name, brand.ID, brand.ID, brand.ID))
	})

	app.Delete("/brands/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := brandSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
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
		product, err := productSvc.Create(name, uint(brandID))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		brandName := ""
		if product.Brand != nil {
			brandName = product.Brand.Name
		}
		return c.SendString(fmt.Sprintf(`<tr id="product-%d">
			<td>%d</td>
			<td>
				<span class="product-name">%s</span>
				<form class="hidden edit-form" hx-put="/products/%d" hx-target="#product-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
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
		</tr>`, product.ID, product.ID, product.Name, product.ID, product.ID, product.Name, brandName, product.ID, product.ID, product.ID))
	})

	app.Put("/products/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		product, err := productSvc.Update(uint(id), name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		brandName := ""
		if product.Brand != nil {
			brandName = product.Brand.Name
		}
		return c.SendString(fmt.Sprintf(`<tr id="product-%d">
			<td>%d</td>
			<td>
				<span class="product-name">%s</span>
				<form class="hidden edit-form" hx-put="/products/%d" hx-target="#product-%d" hx-swap="outerHTML">
					<input type="text" name="name" value="%s" required>
				</form>
			</td>
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
		</tr>`, product.ID, product.ID, product.Name, product.ID, product.ID, product.Name, brandName, product.ID, product.ID, product.ID))
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
