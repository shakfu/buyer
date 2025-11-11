package main

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableAuth       bool
	EnableCSRF       bool
	EnableRateLimiter bool
	Username         string
	Password         string
}

// SetupSecurityMiddleware adds all security middleware to the Fiber app
func SetupSecurityMiddleware(app *fiber.App, config SecurityConfig) {
	// Security headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;")
		return c.Next()
	})

	// Rate limiting
	if config.EnableRateLimiter {
		app.Use(limiter.New(limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
			Next: func(c *fiber.Ctx) bool {
				// Skip rate limiting for static files
				return c.Path() == "/static" || len(c.Path()) > 7 && c.Path()[:7] == "/static"
			},
		}))
	}

	// CSRF protection
	if config.EnableCSRF {
		app.Use(csrf.New(csrf.Config{
			KeyLookup:      "header:X-CSRF-Token",
			CookieName:     "csrf_",
			CookieSameSite: "Strict",
			Expiration:     1 * time.Hour,
			KeyGenerator:   func() string { return generateCSRFToken() },
		}))
	}

	// Basic authentication
	if config.EnableAuth {
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				config.Username: config.Password,
			},
			Realm: "Buyer Application",
			Next: func(c *fiber.Ctx) bool {
				// Skip auth for static files
				return c.Path() == "/static" || len(c.Path()) > 7 && c.Path()[:7] == "/static"
			},
		}))
	}
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// HTML escaping helper functions

// escapeHTML safely escapes HTML content
func escapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// SafeHTML represents HTML that has been properly escaped
type SafeHTML struct {
	content string
}

// String returns the safe HTML content
func (s SafeHTML) String() string {
	return s.content
}

// RenderBrandRow safely renders a brand table row
func RenderBrandRow(brand *models.Brand) (SafeHTML, error) {
	tmpl := template.Must(template.New("brand-row").Parse(`<tr id="brand-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			<span class="brand-name">{{.Name}}</span>
			<form class="hidden edit-form" hx-put="/brands/{{.ID}}" hx-target="#brand-{{.ID}}" hx-swap="outerHTML">
				<input type="text" name="name" value="{{.Name}}" required>
			</form>
		</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="toggleEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/brands/{{.ID}}"
						hx-target="#brand-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this brand?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, brand); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderProductRow safely renders a product table row
func RenderProductRow(product *models.Product) (SafeHTML, error) {
	brandName := ""
	if product.Brand != nil {
		brandName = product.Brand.Name
	}
	specName := "-"
	if product.Specification != nil {
		specName = product.Specification.Name
	}

	tmpl := template.Must(template.New("product-row").Parse(`<tr id="product-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			<span class="product-name">{{.Name}}</span>
			<form class="hidden edit-form" hx-put="/products/{{.ID}}" hx-target="#product-{{.ID}}" hx-swap="outerHTML">
				<input type="text" name="name" value="{{.Name}}" required>
			</form>
		</td>
		<td>{{.BrandName}}</td>
		<td>{{.SpecName}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="toggleProductEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/products/{{.ID}}"
						hx-target="#product-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this product?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	data := struct {
		ID        uint
		Name      string
		BrandName string
		SpecName  string
	}{
		ID:        product.ID,
		Name:      product.Name,
		BrandName: brandName,
		SpecName:  specName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderVendorRow safely renders a vendor table row
func RenderVendorRow(vendor *models.Vendor) (SafeHTML, error) {
	tmpl := template.Must(template.New("vendor-row").Parse(`<tr id="vendor-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			<span class="vendor-name">{{.Name}}</span>
			<form class="hidden edit-form" hx-put="/vendors/{{.ID}}" hx-target="#vendor-{{.ID}}" hx-swap="outerHTML">
				<input type="text" name="name" value="{{.Name}}" required>
			</form>
		</td>
		<td>{{.Currency}}</td>
		<td>{{.DiscountCode}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="toggleVendorEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/vendors/{{.ID}}"
						hx-target="#vendor-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this vendor?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vendor); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderSpecificationRow safely renders a specification table row
func RenderSpecificationRow(spec *models.Specification) (SafeHTML, error) {
	tmpl := template.Must(template.New("spec-row").Parse(`<tr id="spec-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			<span class="spec-name">{{.Name}}</span>
			<form class="hidden edit-form" hx-put="/specifications/{{.ID}}" hx-target="#spec-{{.ID}}" hx-swap="outerHTML">
				<input type="text" name="name" value="{{.Name}}" required>
				<textarea name="description" rows="2">{{.Description}}</textarea>
			</form>
		</td>
		<td>{{.Description}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="toggleSpecEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/specifications/{{.ID}}"
						hx-target="#spec-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this specification?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderForexRow safely renders a forex rate table row
func RenderForexRow(forex *models.Forex) (SafeHTML, error) {
	tmpl := template.Must(template.New("forex-row").Parse(`<tr id="forex-{{.ID}}">
		<td>{{.ID}}</td>
		<td>{{.FromCurrency}}</td>
		<td>{{.ToCurrency}}</td>
		<td>{{printf "%.4f" .Rate}}</td>
		<td>{{.EffectiveDate.Format "2006-01-02"}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm contrast"
						hx-delete="/forex/{{.ID}}"
						hx-target="#forex-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this forex rate?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, forex); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderQuoteRow safely renders a quote table row
func RenderQuoteRow(quote *models.Quote) (SafeHTML, error) {
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

	tmpl := template.Must(template.New("quote-row").Parse(`<tr id="quote-{{.ID}}">
		<td>{{.ID}}</td>
		<td>{{.VendorName}}</td>
		<td>{{.ProductName}}</td>
		<td>{{printf "%.2f" .Price}}</td>
		<td>{{.Currency}}</td>
		<td>{{printf "%.2f" .ConvertedPrice}}</td>
		<td>{{.QuoteDate.Format "2006-01-02"}}</td>
		<td>{{.ExpiryDisplay}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm contrast"
						hx-delete="/quotes/{{.ID}}"
						hx-target="#quote-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this quote?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	data := struct {
		ID             uint
		VendorName     string
		ProductName    string
		Price          float64
		Currency       string
		ConvertedPrice float64
		QuoteDate      time.Time
		ExpiryDisplay  template.HTML
	}{
		ID:             quote.ID,
		VendorName:     vendorName,
		ProductName:    productName,
		Price:          quote.Price,
		Currency:       quote.Currency,
		ConvertedPrice: quote.ConvertedPrice,
		QuoteDate:      quote.QuoteDate,
		ExpiryDisplay:  template.HTML(expiryDisplay),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderRequisitionRow safely renders a requisition table row
func RenderRequisitionRow(req *models.Requisition) (SafeHTML, error) {
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
			descDisplay = fmt.Sprintf(" - %s", escapeHTML(item.Description))
		}
		itemsHTML += fmt.Sprintf("<li>%s (Qty: %d%s)%s</li>", escapeHTML(specName), item.Quantity, budgetDisplay, descDisplay)
	}

	justificationDisplay := ""
	if req.Justification != "" {
		justificationDisplay = fmt.Sprintf("<br><small>%s</small>", escapeHTML(req.Justification))
	}
	budgetDisplay := ""
	if req.Budget > 0 {
		budgetDisplay = fmt.Sprintf("<br><strong>Budget: %.2f</strong>", req.Budget)
	}

	tmpl := template.Must(template.New("req-row").Parse(`<tr id="req-{{.ID}}">
		<td>{{.ID}}</td>
		<td>{{.Name}}{{.JustificationDisplay}}{{.BudgetDisplay}}</td>
		<td><ul>{{.ItemsHTML}}</ul></td>
		<td>
			<div class="actions">
				<button class="btn-sm contrast"
						hx-delete="/requisitions/{{.ID}}"
						hx-target="#req-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this requisition?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	data := struct {
		ID                    uint
		Name                  string
		JustificationDisplay  template.HTML
		BudgetDisplay         template.HTML
		ItemsHTML             template.HTML
	}{
		ID:                    req.ID,
		Name:                  req.Name,
		JustificationDisplay:  template.HTML(justificationDisplay),
		BudgetDisplay:         template.HTML(budgetDisplay),
		ItemsHTML:             template.HTML(itemsHTML),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderRequisitionComparison safely renders the requisition comparison results
func RenderRequisitionComparison(comparison *services.QuoteComparison) (SafeHTML, error) {
	// Build comprehensive comparison HTML using template
	tmpl := template.Must(template.New("comparison").Funcs(template.FuncMap{
		"formatPrice": func(price float64) string {
			return fmt.Sprintf("%.2f", price)
		},
	}).Parse(`<article>
		<h2>Quote Comparison for Requisition: {{.Requisition.Name}}</h2>
		{{if .Requisition.Justification}}
		<p><em>{{.Requisition.Justification}}</em></p>
		{{end}}

		<section style="background: #f0f0f0; padding: 1rem; margin: 1rem 0; border-radius: 5px;">
			<h3>Summary</h3>
			<table>
				<tr>
					<td><strong>Total Items:</strong></td>
					<td>{{len .ItemComparisons}}</td>
				</tr>
				<tr>
					<td><strong>Best Quote Total:</strong></td>
					<td style="color: green; font-weight: bold;">${{formatPrice .TotalEstimate}}</td>
				</tr>
				{{if gt .TotalBudget 0.0}}
				<tr>
					<td><strong>Budget:</strong></td>
					<td>${{formatPrice .TotalBudget}}</td>
				</tr>
				<tr>
					<td><strong>Savings:</strong></td>
					<td style="color: {{if lt .TotalSavings 0.0}}red{{else}}green{{end}}; font-weight: bold;">${{formatPrice .TotalSavings}}</td>
				</tr>
				{{end}}
				<tr>
					<td><strong>All Items Have Quotes:</strong></td>
					<td>
						{{if .AllItemsHaveQuotes}}
						<span style="color: green;">✓ Yes</span>
						{{else}}
						<span style="color: red;">✗ No - Some items missing quotes</span>
						{{end}}
					</td>
				</tr>
			</table>
		</section>
	</article>`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, comparison); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}
