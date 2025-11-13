package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"gorm.io/gorm"
)

// SetupCRUDHandlers sets up all CRUD endpoints with XSS protection
func SetupCRUDHandlers(
	app *fiber.App,
	db *gorm.DB,
	specSvc *services.SpecificationService,
	brandSvc *services.BrandService,
	productSvc *services.ProductService,
	vendorSvc *services.VendorService,
	requisitionSvc *services.RequisitionService,
	quoteSvc *services.QuoteService,
	forexSvc *services.ForexService,
) {
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
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderProductRow(product)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/products/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := productSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Product attribute values endpoint
	app.Post("/products/:id/attributes", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid product ID")
		}

		// Get product to verify it exists
		product, err := productSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Product not found")
		}

		if product.SpecificationID == nil {
			return c.Status(fiber.StatusBadRequest).SendString("Product has no specification")
		}

		// Get all specification attributes
		var specAttrs []models.SpecificationAttribute
		db.Where("specification_id = ?", *product.SpecificationID).Find(&specAttrs)

		// Delete existing product attributes
		db.Where("product_id = ?", id).Delete(&models.ProductAttribute{})

		// Process form data and create new attributes
		for _, attr := range specAttrs {
			formKey := fmt.Sprintf("attr_%d", attr.ID)
			value := strings.TrimSpace(c.FormValue(formKey))

			// Skip empty values for non-required attributes
			if value == "" && !attr.IsRequired {
				continue
			}

			prodAttr := &models.ProductAttribute{
				ProductID:                uint(id),
				SpecificationAttributeID: attr.ID,
			}

			// Parse value based on data type
			switch attr.DataType {
			case "number":
				if value != "" {
					num, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return c.Status(fiber.StatusBadRequest).
							SendString(fmt.Sprintf("Invalid number for %s", attr.Name))
					}
					// Validate min/max
					if attr.MinValue != nil && num < *attr.MinValue {
						return c.Status(fiber.StatusBadRequest).
							SendString(fmt.Sprintf("%s must be at least %.2f", attr.Name, *attr.MinValue))
					}
					if attr.MaxValue != nil && num > *attr.MaxValue {
						return c.Status(fiber.StatusBadRequest).
							SendString(fmt.Sprintf("%s must be at most %.2f", attr.Name, *attr.MaxValue))
					}
					prodAttr.ValueNumber = &num
				} else if attr.IsRequired {
					return c.Status(fiber.StatusBadRequest).
						SendString(fmt.Sprintf("%s is required", attr.Name))
				}
			case "text":
				if value != "" {
					prodAttr.ValueText = &value
				} else if attr.IsRequired {
					return c.Status(fiber.StatusBadRequest).
						SendString(fmt.Sprintf("%s is required", attr.Name))
				}
			case "boolean":
				if value != "" {
					boolVal := value == "true"
					prodAttr.ValueBoolean = &boolVal
				} else if attr.IsRequired {
					return c.Status(fiber.StatusBadRequest).
						SendString(fmt.Sprintf("%s is required", attr.Name))
				}
			}

			// Only create if we have a value
			if prodAttr.ValueNumber != nil || prodAttr.ValueText != nil || prodAttr.ValueBoolean != nil {
				if err := db.Create(prodAttr).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).
						SendString(fmt.Sprintf("Failed to save %s: %s", attr.Name, err.Error()))
				}
			}
		}

		// Reload product with updated attributes
		product, err = productSvc.GetByID(uint(id))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to reload product")
		}

		// Render the updated attributes display
		var buf bytes.Buffer
		buf.WriteString(`<dl id="attrs-display">`)
		if len(product.Attributes) > 0 {
			for _, attr := range product.Attributes {
				if attr.SpecificationAttribute == nil {
					continue
				}
				buf.WriteString(fmt.Sprintf(`<dt>%s`, escapeHTML(attr.SpecificationAttribute.Name)))
				if attr.SpecificationAttribute.Unit != "" {
					buf.WriteString(fmt.Sprintf(` <small>(%s)</small>`, escapeHTML(attr.SpecificationAttribute.Unit)))
				}
				buf.WriteString(`</dt><dd>`)
				if attr.ValueNumber != nil {
					buf.WriteString(fmt.Sprintf("%.2f", *attr.ValueNumber))
				} else if attr.ValueText != nil {
					buf.WriteString(escapeHTML(*attr.ValueText))
				} else if attr.ValueBoolean != nil {
					if *attr.ValueBoolean {
						buf.WriteString("Yes")
					} else {
						buf.WriteString("No")
					}
				}
				buf.WriteString(`</dd>`)
			}
		} else {
			buf.WriteString(`<p>No attribute values set yet.</p>`)
		}
		buf.WriteString(`</dl>`)

		return c.SendString(buf.String())
	})

	// CRUD endpoints for Vendors
	app.Post("/vendors", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		currency := c.FormValue("currency")
		discountCode := c.FormValue("discount_code")
		vendor, err := vendorSvc.Create(name, currency, discountCode)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderVendorRow(vendor)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/vendors/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		name := c.FormValue("name")
		vendor, err := vendorSvc.Update(uint(id), name)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderVendorRow(vendor)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/vendors/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := vendorSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
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
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderForexRow(forex)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/forex/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := forexSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// CRUD endpoints for Specifications
	app.Post("/specifications", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		description := c.FormValue("description")
		spec, err := specSvc.Create(name, description)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderSpecificationRow(spec)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
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
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderSpecificationRow(spec)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/specifications/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := specSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
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
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderRequisitionRow(req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/requisitions/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := requisitionSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
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
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderQuoteRow(quote)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/quotes/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := quoteSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})
}

// SetupProjectHandlers sets up all project-related CRUD endpoints
func SetupProjectHandlers(
	app *fiber.App,
	projectSvc *services.ProjectService,
	specSvc *services.SpecificationService,
	reqSvc *services.RequisitionService,
	projectReqSvc *services.ProjectRequisitionService,
) {
	// CRUD endpoints for Projects
	app.Post("/projects", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		description := c.FormValue("description")

		var budget float64
		budgetStr := c.FormValue("budget")
		if budgetStr != "" {
			var err error
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		var deadline *time.Time
		deadlineStr := c.FormValue("deadline")
		if deadlineStr != "" {
			parsed, err := time.Parse("2006-01-02", deadlineStr)
			if err == nil {
				deadline = &parsed
			}
		}

		project, err := projectSvc.Create(name, description, budget, deadline)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderProjectRow(project)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/projects/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		name := c.FormValue("name")
		description := c.FormValue("description")
		status := c.FormValue("status")

		var budget float64
		budgetStr := c.FormValue("budget")
		if budgetStr != "" {
			budget, err = strconv.ParseFloat(budgetStr, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid budget")
			}
		}

		var deadline *time.Time
		deadlineStr := c.FormValue("deadline")
		if deadlineStr != "" {
			parsed, err := time.Parse("2006-01-02", deadlineStr)
			if err == nil {
				deadline = &parsed
			}
		}

		project, err := projectSvc.Update(uint(id), name, description, budget, deadline, status)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderProjectRow(project)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/projects/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := projectSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// BOM item endpoints
	app.Post("/projects/:id/bom-items", func(c *fiber.Ctx) error {
		projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
		}

		specID, err := strconv.ParseUint(c.FormValue("specification_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid specification ID")
		}

		quantity, err := strconv.Atoi(c.FormValue("quantity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid quantity")
		}

		notes := c.FormValue("notes")

		bomItem, err := projectSvc.AddBillOfMaterialsItem(uint(projectID), uint(specID), quantity, notes)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderBOMItemRow(bomItem)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/bom-items/:id", func(c *fiber.Ctx) error {
		itemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid item ID")
		}

		quantity, err := strconv.Atoi(c.FormValue("quantity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid quantity")
		}

		notes := c.FormValue("notes")

		bomItem, err := projectSvc.UpdateBillOfMaterialsItem(uint(itemID), quantity, notes)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		html, err := RenderBOMItemRow(bomItem)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/bom-items/:id", func(c *fiber.Ctx) error {
		itemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid item ID")
		}
		if err := projectSvc.DeleteBillOfMaterialsItem(uint(itemID)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Project Requisition endpoints
	app.Post("/project-requisitions", func(c *fiber.Ctx) error {
		projectID, err := strconv.ParseUint(c.FormValue("project_id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
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

		// Parse multiple BOM items from form data
		items := []services.ProjectRequisitionItemInput{}
		for i := 0; ; i++ {
			bomItemIDStr := c.FormValue(fmt.Sprintf("items[%d][bom_item_id]", i))
			if bomItemIDStr == "" {
				break // No more items
			}

			bomItemID, err := strconv.ParseUint(bomItemIDStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid BOM item ID for item %d", i))
			}

			quantityStr := c.FormValue(fmt.Sprintf("items[%d][quantity_requested]", i))
			quantity, err := strconv.Atoi(quantityStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid quantity for item %d", i))
			}

			notes := c.FormValue(fmt.Sprintf("items[%d][notes]", i))

			items = append(items, services.ProjectRequisitionItemInput{
				BOMItemID:         uint(bomItemID),
				QuantityRequested: quantity,
				Notes:             notes,
			})
		}

		if len(items) == 0 {
			return c.Status(fiber.StatusBadRequest).SendString("At least one BOM item is required")
		}

		projectReq, err := projectReqSvc.Create(uint(projectID), name, justification, budget, items)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderProjectRequisitionRow(projectReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Put("/project-requisitions/:id", func(c *fiber.Ctx) error {
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

		projectReq, err := projectReqSvc.Update(uint(id), name, justification, budget)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderProjectRequisitionRow(projectReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	app.Delete("/project-requisitions/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}
		if err := projectReqSvc.Delete(uint(id)); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}
		return c.SendString("")
	})

	// Get project requisitions for a project (for HTMX partial loading)
	app.Get("/projects/:id/project-requisitions", func(c *fiber.Ctx) error {
		projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
		}

		projectReqs, err := projectReqSvc.ListByProject(uint(projectID))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderProjectRequisitionsList(projectReqs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})

	// Get BOM items for a project (for HTMX partial loading)
	app.Get("/projects/:id/bom-items", func(c *fiber.Ctx) error {
		projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid project ID")
		}

		project, err := projectSvc.GetByID(uint(projectID))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(escapeHTML(err.Error()))
		}

		html, err := RenderBOMItemsList(project.BillOfMaterials, specSvc)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to render response")
		}
		return c.SendString(html.String())
	})
}
