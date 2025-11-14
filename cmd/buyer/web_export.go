package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shakfu/buyer/internal/services"
	"gorm.io/gorm"
)

// SetupExportHandlers sets up export/import endpoints
func SetupExportHandlers(app *fiber.App, db *gorm.DB) {
	exportSvc := services.NewExportImportService(db)

	// ==================== Export Endpoints ====================

	// Export brands
	app.Get("/export/brands/csv", func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := exportSvc.ExportBrandsCSV(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export brands")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=brands.csv")
		return c.Send(buf.Bytes())
	})

	app.Get("/export/brands/excel", func(c *fiber.Ctx) error {
		f, err := exportSvc.ExportBrandsExcel()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export brands")
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to write Excel file")
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=brands.xlsx")
		return c.Send(buf.Bytes())
	})

	// Export vendors
	app.Get("/export/vendors/csv", func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := exportSvc.ExportVendorsCSV(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export vendors")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=vendors.csv")
		return c.Send(buf.Bytes())
	})

	app.Get("/export/vendors/excel", func(c *fiber.Ctx) error {
		f, err := exportSvc.ExportVendorsExcel()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export vendors")
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to write Excel file")
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=vendors.xlsx")
		return c.Send(buf.Bytes())
	})

	// Export products
	app.Get("/export/products/csv", func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := exportSvc.ExportProductsCSV(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export products")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=products.csv")
		return c.Send(buf.Bytes())
	})

	app.Get("/export/products/excel", func(c *fiber.Ctx) error {
		f, err := exportSvc.ExportProductsExcel()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export products")
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to write Excel file")
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=products.xlsx")
		return c.Send(buf.Bytes())
	})

	// Export quotes
	app.Get("/export/quotes/csv", func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := exportSvc.ExportQuotesCSV(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export quotes")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=quotes.csv")
		return c.Send(buf.Bytes())
	})

	app.Get("/export/quotes/excel", func(c *fiber.Ctx) error {
		f, err := exportSvc.ExportQuotesExcel()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export quotes")
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to write Excel file")
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=quotes.xlsx")
		return c.Send(buf.Bytes())
	})

	// Export forex rates
	app.Get("/export/forex/csv", func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := exportSvc.ExportForexCSV(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export forex rates")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=forex_rates.csv")
		return c.Send(buf.Bytes())
	})

	app.Get("/export/forex/excel", func(c *fiber.Ctx) error {
		f, err := exportSvc.ExportForexExcel()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to export forex rates")
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to write Excel file")
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=forex_rates.xlsx")
		return c.Send(buf.Bytes())
	})

	// ==================== Import Endpoints ====================

	// Import brands
	app.Post("/import/brands", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("No file uploaded")
		}

		// Check file extension
		if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
			return c.Status(fiber.StatusBadRequest).SendString("Only CSV files are supported for import")
		}

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to open uploaded file")
		}
		defer src.Close()

		// Import the data
		result, err := exportSvc.ImportBrandsCSV(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Import failed: %v", err))
		}

		// Return import summary
		return c.JSON(fiber.Map{
			"success":       result.SuccessCount,
			"errors":        result.ErrorCount,
			"error_details": result.Errors,
		})
	})

	// Import vendors
	app.Post("/import/vendors", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("No file uploaded")
		}

		if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
			return c.Status(fiber.StatusBadRequest).SendString("Only CSV files are supported for import")
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to open uploaded file")
		}
		defer src.Close()

		result, err := exportSvc.ImportVendorsCSV(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Import failed: %v", err))
		}

		return c.JSON(fiber.Map{
			"success":       result.SuccessCount,
			"errors":        result.ErrorCount,
			"error_details": result.Errors,
		})
	})

	// Import forex rates
	app.Post("/import/forex", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("No file uploaded")
		}

		if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
			return c.Status(fiber.StatusBadRequest).SendString("Only CSV files are supported for import")
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to open uploaded file")
		}
		defer src.Close()

		result, err := exportSvc.ImportForexCSV(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Import failed: %v", err))
		}

		return c.JSON(fiber.Map{
			"success":       result.SuccessCount,
			"errors":        result.ErrorCount,
			"error_details": result.Errors,
		})
	})
}

// Helper function to read multipart file
func readMultipartFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
