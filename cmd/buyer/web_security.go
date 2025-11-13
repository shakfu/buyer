package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableAuth        bool
	EnableCSRF        bool
	EnableRateLimiter bool
	Username          string
	PasswordHash      string // bcrypt hash of the password
}

// SetupSecurityMiddleware adds all security middleware to the Fiber app
func SetupSecurityMiddleware(app *fiber.App, config SecurityConfig) {
	// Security headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data:; connect-src 'self';")
		return c.Next()
	})

	// Rate limiting for general requests
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

	// Authentication-specific rate limiting (stricter)
	if config.EnableAuth {
		authLimiter := limiter.New(limiter.Config{
			Max:        5,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP() + ":auth"
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).SendString("Too many authentication attempts. Please try again later.")
			},
			SkipFailedRequests: true,
			SkipSuccessfulRequests: true,
			Next: func(c *fiber.Ctx) bool {
				// Only apply to non-static requests (auth is checked on these)
				return c.Path() == "/static" || len(c.Path()) > 7 && c.Path()[:7] == "/static"
			},
		})
		app.Use(authLimiter)
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

	// Basic authentication with bcrypt password verification
	if config.EnableAuth {
		app.Use(basicauth.New(basicauth.Config{
			Authorizer: func(username, password string) bool {
				if username != config.Username {
					return false
				}
				// Use bcrypt to compare password with stored hash
				err := bcrypt.CompareHashAndPassword([]byte(config.PasswordHash), []byte(password))
				return err == nil
			},
			Realm: "Buyer Application",
			Next: func(c *fiber.Ctx) bool {
				// Skip auth for static files
				return c.Path() == "/static" || len(c.Path()) > 7 && c.Path()[:7] == "/static"
			},
		}))
	}
}

// generateCSRFToken generates a cryptographically secure random CSRF token
func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate CSRF token: %v", err))
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Password validation and hashing functions

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
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

	// Determine expiry status and days - let template handle HTML generation
	var expiryDays *int
	expiryColor := "gray"
	expiryText := "—"
	if quote.ValidUntil != nil {
		days := quote.DaysUntilExpiration()
		expiryDays = &days
		expiryText = fmt.Sprintf("%d", days)
		if days < 0 || days < 7 {
			expiryColor = "red"
		} else if days < 30 {
			expiryColor = "orange"
		} else {
			expiryColor = "green"
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
		<td>
			{{if .ExpiryDays}}
				<span style="color: {{.ExpiryColor}}{{if or (lt .ExpiryDays 7) (lt .ExpiryDays 0)}}; font-weight: bold{{end}}">{{.ExpiryText}}</span>
			{{else}}
				<span style="color: {{.ExpiryColor}};">{{.ExpiryText}}</span>
			{{end}}
		</td>
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
		ExpiryDays     *int
		ExpiryColor    string
		ExpiryText     string
	}{
		ID:             quote.ID,
		VendorName:     vendorName,
		ProductName:    productName,
		Price:          quote.Price,
		Currency:       quote.Currency,
		ConvertedPrice: quote.ConvertedPrice,
		QuoteDate:      quote.QuoteDate,
		ExpiryDays:     expiryDays,
		ExpiryColor:    expiryColor,
		ExpiryText:     expiryText,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderRequisitionRow safely renders a requisition table row
func RenderRequisitionRow(req *models.Requisition) (SafeHTML, error) {
	// Prepare item data for template
	type ItemData struct {
		SpecName      string
		Quantity      int
		BudgetPerUnit float64
		HasBudget     bool
		Description   string
	}

	items := make([]ItemData, 0, len(req.Items))
	for _, item := range req.Items {
		specName := ""
		if item.Specification != nil {
			specName = item.Specification.Name
		}
		items = append(items, ItemData{
			SpecName:      specName,
			Quantity:      item.Quantity,
			BudgetPerUnit: item.BudgetPerUnit,
			HasBudget:     item.BudgetPerUnit > 0,
			Description:   item.Description,
		})
	}

	tmpl := template.Must(template.New("req-row").Parse(`<tr id="req-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			{{.Name}}
			{{if .Justification}}<br><small>{{.Justification}}</small>{{end}}
			{{if gt .Budget 0.0}}<br><strong>Budget: {{printf "%.2f" .Budget}}</strong>{{end}}
		</td>
		<td>
			<ul>
				{{range .Items}}
				<li>
					{{.SpecName}} (Qty: {{.Quantity}}{{if .HasBudget}}, Budget/unit: {{printf "%.2f" .BudgetPerUnit}}{{end}})
					{{if .Description}} - {{.Description}}{{end}}
				</li>
				{{end}}
			</ul>
		</td>
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
		ID            uint
		Name          string
		Justification string
		Budget        float64
		Items         []ItemData
	}{
		ID:            req.ID,
		Name:          req.Name,
		Justification: req.Justification,
		Budget:        req.Budget,
		Items:         items,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderRequisitionComparison safely renders the requisition comparison results
func RenderRequisitionComparison(comparison *services.RequisitionQuoteComparison) (SafeHTML, error) {
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

// RenderProjectRow safely renders a project table row
func RenderProjectRow(project *models.Project) (SafeHTML, error) {
	tmpl := template.Must(template.New("projectRow").Parse(`<tr id="project-{{.ID}}">
		<td>{{.ID}}</td>
		<td>
			<strong>{{.Name}}</strong>
			{{if .Description}}<br><small>{{.Description}}</small>{{end}}
		</td>
		<td>
			<span class="badge {{if eq .Status "planning"}}secondary{{else if eq .Status "active"}}primary{{else if eq .Status "completed"}}success{{else}}contrast{{end}}">
				{{.Status}}
			</span>
		</td>
		<td>{{.BudgetDisplay}}</td>
		<td>{{.DeadlineDisplay}}</td>
		<td>{{.BOMItemCount}}</td>
		<td>{{.RequisitionCount}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm" onclick="viewProject({{.ID}})">View</button>
				<button class="btn-sm secondary" onclick="toggleProjectEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/projects/{{.ID}}"
						hx-target="#project-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Are you sure you want to delete this project and its Bill of Materials?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	budgetDisplay := "-"
	if project.Budget > 0 {
		budgetDisplay = fmt.Sprintf("$%.2f", project.Budget)
	}

	deadlineDisplay := "-"
	if project.Deadline != nil {
		deadlineDisplay = project.Deadline.Format("2006-01-02")
	}

	bomItemCount := 0
	if project.BillOfMaterials != nil {
		bomItemCount = len(project.BillOfMaterials.Items)
	}

	data := struct {
		ID               uint
		Name             string
		Description      string
		Status           string
		BudgetDisplay    string
		DeadlineDisplay  string
		BOMItemCount     int
		RequisitionCount int
	}{
		ID:               project.ID,
		Name:             project.Name,
		Description:      project.Description,
		Status:           project.Status,
		BudgetDisplay:    budgetDisplay,
		DeadlineDisplay:  deadlineDisplay,
		BOMItemCount:     bomItemCount,
		RequisitionCount: len(project.Requisitions),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderBOMItemRow safely renders a BOM item table row
func RenderBOMItemRow(item *models.BillOfMaterialsItem) (SafeHTML, error) {
	tmpl := template.Must(template.New("bomItemRow").Parse(`<tr id="bom-item-{{.ID}}">
		<td>{{.SpecName}}</td>
		<td>{{.Quantity}}</td>
		<td>{{.Notes}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="toggleBOMItemEdit({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/bom-items/{{.ID}}"
						hx-target="#bom-item-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Delete this item from the BOM?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	specName := ""
	if item.Specification != nil {
		specName = item.Specification.Name
	}

	notes := item.Notes
	if notes == "" {
		notes = "-"
	}

	data := struct {
		ID       uint
		SpecName string
		Quantity int
		Notes    string
	}{
		ID:       item.ID,
		SpecName: specName,
		Quantity: item.Quantity,
		Notes:    notes,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderProjectRequisitionRow safely renders a project requisition table row
// TODO: Update this function for new ProjectRequisition schema
func RenderProjectRequisitionRow(projectReq *models.ProjectRequisition) (SafeHTML, error) {
	tmpl := template.Must(template.New("projectReqRow").Parse(`<tr id="project-req-{{.ID}}">
		<td>{{.Name}}</td>
		<td>{{.Budget}}</td>
		<td>{{.ItemCount}}</td>
		<td>
			<div class="actions">
				<button class="btn-sm secondary" onclick="editProjectRequisition({{.ID}})">Edit</button>
				<button class="btn-sm contrast"
						hx-delete="/project-requisitions/{{.ID}}"
						hx-target="#project-req-{{.ID}}"
						hx-swap="outerHTML"
						hx-confirm="Delete this project requisition?">
					Delete
				</button>
			</div>
		</td>
	</tr>`))

	itemCount := 0
	if projectReq.Items != nil {
		itemCount = len(projectReq.Items)
	}

	data := struct {
		ID        uint
		Name      string
		Budget    float64
		ItemCount int
	}{
		ID:        projectReq.ID,
		Name:      projectReq.Name,
		Budget:    projectReq.Budget,
		ItemCount: itemCount,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return SafeHTML{}, err
	}
	return SafeHTML{content: buf.String()}, nil
}

// RenderBOMItemsList safely renders the full BOM items list for HTMX partial loading
func RenderBOMItemsList(bom *models.BillOfMaterials, specSvc *services.SpecificationService) (SafeHTML, error) {
	if bom == nil || len(bom.Items) == 0 {
		return SafeHTML{content: "<p>No items in the Bill of Materials yet.</p>"}, nil
	}

	var buf bytes.Buffer
	buf.WriteString("<table><thead><tr><th>Specification</th><th>Quantity</th><th>Notes</th><th>Actions</th></tr></thead><tbody>")

	for _, item := range bom.Items {
		html, err := RenderBOMItemRow(&item)
		if err != nil {
			return SafeHTML{}, err
		}
		buf.WriteString(html.String())
	}

	buf.WriteString("</tbody></table>")
	return SafeHTML{content: buf.String()}, nil
}

// RenderProjectRequisitionsList safely renders the full project requisitions list
func RenderProjectRequisitionsList(projectReqs []models.ProjectRequisition) (SafeHTML, error) {
	if len(projectReqs) == 0 {
		return SafeHTML{content: "<p>No project requisitions yet.</p>"}, nil
	}

	var buf bytes.Buffer
	buf.WriteString("<table><thead><tr><th>Name</th><th>Budget</th><th>Items</th><th>Actions</th></tr></thead><tbody>")

	for _, projectReq := range projectReqs {
		html, err := RenderProjectRequisitionRow(&projectReq)
		if err != nil {
			return SafeHTML{}, err
		}
		buf.WriteString(html.String())
	}

	buf.WriteString("</tbody></table>")
	return SafeHTML{content: buf.String()}, nil
}

// RenderPurchaseOrderRow safely renders a purchase order table row
func RenderPurchaseOrderRow(po *models.PurchaseOrder) (SafeHTML, error) {
	tmpl := `
<tr id="po-{{.ID}}">
	<td>{{.ID}}</td>
	<td>{{.PONumber}}</td>
	<td><span class="badge badge-{{.Status}}">{{.Status}}</span></td>
	<td>{{if .Vendor}}{{.Vendor.Name}}{{end}}</td>
	<td>{{if .Product}}{{.Product.Name}}{{end}}</td>
	<td>{{.Quantity}}</td>
	<td>{{printf "%.2f" .UnitPrice}} {{.Currency}}</td>
	<td>{{printf "%.2f" .TotalAmount}} {{.Currency}}</td>
	<td>{{printf "%.2f" .GrandTotal}} {{.Currency}}</td>
	<td>{{.OrderDate.Format "2006-01-02"}}</td>
	<td>
		{{if .ExpectedDelivery}}
			{{.ExpectedDelivery.Format "2006-01-02"}}
		{{else}}
			<span style="color: gray;">—</span>
		{{end}}
	</td>
	<td>
		<div class="actions">
			<button class="btn-sm" onclick="showUpdateStatus({{.ID}}, '{{.Status}}')">
				Update
			</button>
			{{if or (eq .Status "pending") (eq .Status "cancelled")}}
			<button class="btn-sm contrast"
					hx-delete="/purchase-orders/{{.ID}}"
					hx-target="#po-{{.ID}}"
					hx-swap="outerHTML"
					hx-confirm="Are you sure you want to delete this purchase order?">
				Delete
			</button>
			{{end}}
		</div>
	</td>
</tr>
`

	t, err := template.New("purchase-order-row").Parse(tmpl)
	if err != nil {
		return SafeHTML{}, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, po); err != nil {
		return SafeHTML{}, err
	}

	return SafeHTML{content: buf.String()}, nil
}
