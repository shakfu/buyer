package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long:  "Start the FastAPI-inspired web server with HTMX support",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		app := fiber.New(fiber.Config{
			AppName: "GoBuy v1.0.0",
		})

		// Middleware
		app.Use(recover.New())
		app.Use(logger.New())

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
		return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>GoBuy - Purchasing Management</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #333; }
        nav { margin: 20px 0; }
        nav a { margin-right: 15px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
    </style>
</head>
<body>
    <h1>GoBuy - Purchasing Management</h1>
    <nav>
        <a href="/brands">Brands</a>
        <a href="/products">Products</a>
        <a href="/vendors">Vendors</a>
        <a href="/quotes">Quotes</a>
        <a href="/forex">Forex Rates</a>
    </nav>
    <p>Welcome to GoBuy, your purchasing support and vendor quote management tool.</p>
</body>
</html>
		`)
	})

	// Brand routes
	app.Get("/brands", func(c *fiber.Ctx) error {
		brands, err := brandSvc.List(0, 0)
		if err != nil {
			return err
		}
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Brands - GoBuy</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Brands</h1>
    <a href="/">Back to Home</a>
    <table>
        <tr><th>ID</th><th>Name</th></tr>
`
		for _, brand := range brands {
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td></tr>", brand.ID, brand.Name)
		}
		html += `
    </table>
</body>
</html>
`
		return c.Type("html").SendString(html)
	})

	// Product routes
	app.Get("/products", func(c *fiber.Ctx) error {
		products, err := productSvc.List(0, 0)
		if err != nil {
			return err
		}
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Products - GoBuy</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Products</h1>
    <a href="/">Back to Home</a>
    <table>
        <tr><th>ID</th><th>Name</th><th>Brand</th></tr>
`
		for _, product := range products {
			brandName := ""
			if product.Brand != nil {
				brandName = product.Brand.Name
			}
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td></tr>", product.ID, product.Name, brandName)
		}
		html += `
    </table>
</body>
</html>
`
		return c.Type("html").SendString(html)
	})

	// Vendor routes
	app.Get("/vendors", func(c *fiber.Ctx) error {
		vendors, err := vendorSvc.List(0, 0)
		if err != nil {
			return err
		}
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Vendors - GoBuy</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Vendors</h1>
    <a href="/">Back to Home</a>
    <table>
        <tr><th>ID</th><th>Name</th><th>Currency</th><th>Discount Code</th></tr>
`
		for _, vendor := range vendors {
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>",
				vendor.ID, vendor.Name, vendor.Currency, vendor.DiscountCode)
		}
		html += `
    </table>
</body>
</html>
`
		return c.Type("html").SendString(html)
	})

	// Quote routes
	app.Get("/quotes", func(c *fiber.Ctx) error {
		quotes, err := quoteSvc.List(0, 0)
		if err != nil {
			return err
		}
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Quotes - GoBuy</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Quotes</h1>
    <a href="/">Back to Home</a>
    <table>
        <tr><th>ID</th><th>Vendor</th><th>Product</th><th>Price</th><th>USD</th><th>Date</th></tr>
`
		for _, quote := range quotes {
			vendorName := ""
			if quote.Vendor != nil {
				vendorName = quote.Vendor.Name
			}
			productName := ""
			if quote.Product != nil {
				productName = quote.Product.Name
			}
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%.2f %s</td><td>%.2f</td><td>%s</td></tr>",
				quote.ID, vendorName, productName, quote.Price, quote.Currency, quote.ConvertedPrice, quote.QuoteDate.Format("2006-01-02"))
		}
		html += `
    </table>
</body>
</html>
`
		return c.Type("html").SendString(html)
	})

	// Forex routes
	app.Get("/forex", func(c *fiber.Ctx) error {
		rates, err := forexSvc.List(0, 0)
		if err != nil {
			return err
		}
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Forex Rates - GoBuy</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Forex Rates</h1>
    <a href="/">Back to Home</a>
    <table>
        <tr><th>ID</th><th>From</th><th>To</th><th>Rate</th><th>Date</th></tr>
`
		for _, rate := range rates {
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%.4f</td><td>%s</td></tr>",
				rate.ID, rate.FromCurrency, rate.ToCurrency, rate.Rate, rate.EffectiveDate.Format("2006-01-02"))
		}
		html += `
    </table>
</body>
</html>
`
		return c.Type("html").SendString(html)
	})
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
}
