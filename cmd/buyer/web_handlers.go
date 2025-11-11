package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shakfu/buyer/internal/services"
)

// SetupCRUDHandlers sets up all CRUD endpoints with XSS protection
func SetupCRUDHandlers(
	app *fiber.App,
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
