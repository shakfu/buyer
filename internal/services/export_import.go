package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ExportImportService handles CSV and Excel export/import operations
type ExportImportService struct {
	db *gorm.DB
}

// NewExportImportService creates a new export/import service
func NewExportImportService(db *gorm.DB) *ExportImportService {
	return &ExportImportService{db: db}
}

// ExportFormat represents the export format
type ExportFormat string

const (
	FormatCSV   ExportFormat = "csv"
	FormatExcel ExportFormat = "excel"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	SuccessCount int
	ErrorCount   int
	Errors       []string
}

// ==================== Brand Export/Import ====================

// ExportBrandsCSV exports brands to CSV format
func (s *ExportImportService) ExportBrandsCSV(w io.Writer) error {
	var brands []models.Brand
	if err := s.db.Order("id ASC").Find(&brands).Error; err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"ID", "Name", "CreatedAt", "UpdatedAt"}); err != nil {
		return err
	}

	// Write data
	for _, brand := range brands {
		record := []string{
			fmt.Sprintf("%d", brand.ID),
			brand.Name,
			brand.CreatedAt.Format(time.RFC3339),
			brand.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportBrandsExcel exports brands to Excel format
func (s *ExportImportService) ExportBrandsExcel() (*excelize.File, error) {
	var brands []models.Brand
	if err := s.db.Order("id ASC").Find(&brands).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Brands"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := []string{"ID", "Name", "Created At", "Updated At"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
	}

	// Apply header styling
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err := f.SetCellStyle(sheetName, "A1", "D1", headerStyle); err != nil {
		return nil, err
	}

	// Write data
	for i, brand := range brands {
		row := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), brand.ID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), brand.Name); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), brand.CreatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), brand.UpdatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Auto-fit columns
	if err := f.SetColWidth(sheetName, "A", "A", 10); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "B", "B", 30); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "C", "D", 25); err != nil {
		return nil, err
	}

	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	return f, nil
}

// ImportBrandsCSV imports brands from CSV format
func (s *ExportImportService) ImportBrandsCSV(r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must contain at least a header and one data row")
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	brandSvc := NewBrandService(s.db)

	// Skip header (first row)
	for i, record := range records[1:] {
		if len(record) < 2 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		name := strings.TrimSpace(record[1])
		if name == "" {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: name is empty", i+2))
			continue
		}

		// Try to create the brand
		_, err := brandSvc.Create(name)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+2, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// ==================== Vendor Export/Import ====================

// ExportVendorsCSV exports vendors to CSV format
func (s *ExportImportService) ExportVendorsCSV(w io.Writer) error {
	var vendors []models.Vendor
	if err := s.db.Order("id ASC").Find(&vendors).Error; err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Name", "Currency", "DiscountCode", "ContactPerson", "Email", "Phone",
		"Website", "AddressLine1", "AddressLine2", "City", "State", "PostalCode",
		"Country", "TaxID", "PaymentTerms", "CreatedAt", "UpdatedAt",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, vendor := range vendors {
		record := []string{
			fmt.Sprintf("%d", vendor.ID),
			vendor.Name,
			vendor.Currency,
			vendor.DiscountCode,
			vendor.ContactPerson,
			vendor.Email,
			vendor.Phone,
			vendor.Website,
			vendor.AddressLine1,
			vendor.AddressLine2,
			vendor.City,
			vendor.State,
			vendor.PostalCode,
			vendor.Country,
			vendor.TaxID,
			vendor.PaymentTerms,
			vendor.CreatedAt.Format(time.RFC3339),
			vendor.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportVendorsExcel exports vendors to Excel format
func (s *ExportImportService) ExportVendorsExcel() (*excelize.File, error) {
	var vendors []models.Vendor
	if err := s.db.Order("id ASC").Find(&vendors).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Vendors"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := []string{
		"ID", "Name", "Currency", "Discount Code", "Contact Person", "Email", "Phone",
		"Website", "Address Line 1", "Address Line 2", "City", "State", "Postal Code",
		"Country", "Tax ID", "Payment Terms", "Created At", "Updated At",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
	}

	// Apply header styling
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCol, _ := excelize.ColumnNumberToName(len(headers))
	if err := f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", endCol), headerStyle); err != nil {
		return nil, err
	}

	// Write data
	for i, vendor := range vendors {
		row := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), vendor.ID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), vendor.Name); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), vendor.Currency); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), vendor.DiscountCode); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), vendor.ContactPerson); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), vendor.Email); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), vendor.Phone); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), vendor.Website); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), vendor.AddressLine1); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), vendor.AddressLine2); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), vendor.City); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), vendor.State); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), vendor.PostalCode); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), vendor.Country); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), vendor.TaxID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), vendor.PaymentTerms); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), vendor.CreatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), vendor.UpdatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Auto-fit columns
	if err := f.SetColWidth(sheetName, "A", "A", 10); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "B", "B", 30); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "C", "D", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "E", "H", 20); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "I", "J", 25); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "K", "N", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "O", "P", 20); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "Q", "R", 25); err != nil {
		return nil, err
	}

	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	return f, nil
}

// ImportVendorsCSV imports vendors from CSV format
func (s *ExportImportService) ImportVendorsCSV(r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must contain at least a header and one data row")
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	vendorSvc := NewVendorService(s.db)

	// Skip header (first row)
	for i, record := range records[1:] {
		if len(record) < 3 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		name := strings.TrimSpace(record[1])
		currency := strings.TrimSpace(record[2])
		discountCode := ""
		if len(record) > 3 {
			discountCode = strings.TrimSpace(record[3])
		}

		if name == "" {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: name is empty", i+2))
			continue
		}

		// Try to create the vendor
		_, err := vendorSvc.Create(name, currency, discountCode)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+2, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// ==================== Product Export/Import ====================

// ExportProductsCSV exports products to CSV format
func (s *ExportImportService) ExportProductsCSV(w io.Writer) error {
	var products []models.Product
	if err := s.db.Preload("Brand").Preload("Specification").Order("id ASC").Find(&products).Error; err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Name", "BrandID", "BrandName", "SpecificationID", "SpecificationName",
		"SKU", "Description", "UnitOfMeasure", "MinOrderQty", "LeadTimeDays",
		"IsActive", "DiscontinuedAt", "CreatedBy", "UpdatedBy", "CreatedAt", "UpdatedAt",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, product := range products {
		brandName := ""
		if product.Brand != nil {
			brandName = product.Brand.Name
		}

		specID := ""
		specName := ""
		if product.SpecificationID != nil {
			specID = fmt.Sprintf("%d", *product.SpecificationID)
			if product.Specification != nil {
				specName = product.Specification.Name
			}
		}

		sku := ""
		if product.SKU != nil {
			sku = *product.SKU
		}

		discontinuedAt := ""
		if product.DiscontinuedAt != nil {
			discontinuedAt = product.DiscontinuedAt.Format(time.RFC3339)
		}

		record := []string{
			fmt.Sprintf("%d", product.ID),
			product.Name,
			fmt.Sprintf("%d", product.BrandID),
			brandName,
			specID,
			specName,
			sku,
			product.Description,
			product.UnitOfMeasure,
			fmt.Sprintf("%d", product.MinOrderQty),
			fmt.Sprintf("%d", product.LeadTimeDays),
			fmt.Sprintf("%t", product.IsActive),
			discontinuedAt,
			product.CreatedBy,
			product.UpdatedBy,
			product.CreatedAt.Format(time.RFC3339),
			product.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportProductsExcel exports products to Excel format
func (s *ExportImportService) ExportProductsExcel() (*excelize.File, error) {
	var products []models.Product
	if err := s.db.Preload("Brand").Preload("Specification").Order("id ASC").Find(&products).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Products"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := []string{
		"ID", "Name", "Brand ID", "Brand Name", "Spec ID", "Spec Name",
		"SKU", "Description", "Unit Of Measure", "Min Order Qty", "Lead Time Days",
		"Is Active", "Discontinued At", "Created By", "Updated By", "Created At", "Updated At",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
	}

	// Apply header styling
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCol, _ := excelize.ColumnNumberToName(len(headers))
	if err := f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", endCol), headerStyle); err != nil {
		return nil, err
	}

	// Write data
	for i, product := range products {
		row := i + 2

		brandName := ""
		if product.Brand != nil {
			brandName = product.Brand.Name
		}

		specID := ""
		specName := ""
		if product.SpecificationID != nil {
			specID = fmt.Sprintf("%d", *product.SpecificationID)
			if product.Specification != nil {
				specName = product.Specification.Name
			}
		}

		sku := ""
		if product.SKU != nil {
			sku = *product.SKU
		}

		discontinuedAt := ""
		if product.DiscontinuedAt != nil {
			discontinuedAt = product.DiscontinuedAt.Format(time.RFC3339)
		}

		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), product.ID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), product.Name); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), product.BrandID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), brandName); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), specID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), specName); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), sku); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), product.Description); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), product.UnitOfMeasure); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), product.MinOrderQty); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), product.LeadTimeDays); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), product.IsActive); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), discontinuedAt); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), product.CreatedBy); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), product.UpdatedBy); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), product.CreatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), product.UpdatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Auto-fit columns
	if err := f.SetColWidth(sheetName, "A", "A", 10); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "B", "B", 30); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "C", "F", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "G", "G", 20); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "H", "H", 40); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "I", "K", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "L", "L", 12); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "M", "Q", 25); err != nil {
		return nil, err
	}

	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	return f, nil
}

// ==================== Quote Export/Import ====================

// ExportQuotesCSV exports quotes to CSV format
func (s *ExportImportService) ExportQuotesCSV(w io.Writer) error {
	var quotes []models.Quote
	if err := s.db.Preload("Vendor").Preload("Product").Order("id ASC").Find(&quotes).Error; err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "VendorID", "VendorName", "ProductID", "ProductName",
		"Price", "Currency", "ConvertedPrice", "ConversionRate", "MinQuantity",
		"QuoteDate", "ValidUntil", "Status", "Version", "Notes",
		"CreatedBy", "UpdatedBy", "CreatedAt", "UpdatedAt",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, quote := range quotes {
		vendorName := ""
		if quote.Vendor != nil {
			vendorName = quote.Vendor.Name
		}

		productName := ""
		if quote.Product != nil {
			productName = quote.Product.Name
		}

		validUntil := ""
		if quote.ValidUntil != nil {
			validUntil = quote.ValidUntil.Format(time.RFC3339)
		}

		record := []string{
			fmt.Sprintf("%d", quote.ID),
			fmt.Sprintf("%d", quote.VendorID),
			vendorName,
			fmt.Sprintf("%d", quote.ProductID),
			productName,
			fmt.Sprintf("%.2f", quote.Price),
			quote.Currency,
			fmt.Sprintf("%.2f", quote.ConvertedPrice),
			fmt.Sprintf("%.6f", quote.ConversionRate),
			fmt.Sprintf("%d", quote.MinQuantity),
			quote.QuoteDate.Format(time.RFC3339),
			validUntil,
			quote.Status,
			fmt.Sprintf("%d", quote.Version),
			quote.Notes,
			quote.CreatedBy,
			quote.UpdatedBy,
			quote.CreatedAt.Format(time.RFC3339),
			quote.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportQuotesExcel exports quotes to Excel format
func (s *ExportImportService) ExportQuotesExcel() (*excelize.File, error) {
	var quotes []models.Quote
	if err := s.db.Preload("Vendor").Preload("Product").Order("id ASC").Find(&quotes).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Quotes"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := []string{
		"ID", "Vendor ID", "Vendor Name", "Product ID", "Product Name",
		"Price", "Currency", "Converted Price (USD)", "Conversion Rate", "Min Quantity",
		"Quote Date", "Valid Until", "Status", "Version", "Notes",
		"Created By", "Updated By", "Created At", "Updated At",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
	}

	// Apply header styling
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCol, _ := excelize.ColumnNumberToName(len(headers))
	if err := f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", endCol), headerStyle); err != nil {
		return nil, err
	}

	// Write data
	for i, quote := range quotes {
		row := i + 2

		vendorName := ""
		if quote.Vendor != nil {
			vendorName = quote.Vendor.Name
		}

		productName := ""
		if quote.Product != nil {
			productName = quote.Product.Name
		}

		validUntil := ""
		if quote.ValidUntil != nil {
			validUntil = quote.ValidUntil.Format(time.RFC3339)
		}

		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), quote.ID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), quote.VendorID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), vendorName); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), quote.ProductID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), productName); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), quote.Price); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), quote.Currency); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), quote.ConvertedPrice); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), quote.ConversionRate); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), quote.MinQuantity); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), quote.QuoteDate.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), validUntil); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), quote.Status); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), quote.Version); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), quote.Notes); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), quote.CreatedBy); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), quote.UpdatedBy); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), quote.CreatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("S%d", row), quote.UpdatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Auto-fit columns
	if err := f.SetColWidth(sheetName, "A", "B", 10); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "C", "E", 25); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "F", "J", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "K", "L", 20); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "M", "N", 12); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "O", "O", 40); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "P", "S", 25); err != nil {
		return nil, err
	}

	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	return f, nil
}

// ==================== Forex Export/Import ====================

// ExportForexCSV exports forex rates to CSV format
func (s *ExportImportService) ExportForexCSV(w io.Writer) error {
	var rates []models.Forex
	if err := s.db.Order("id ASC").Find(&rates).Error; err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"ID", "FromCurrency", "ToCurrency", "Rate", "EffectiveDate", "CreatedAt", "UpdatedAt"}); err != nil {
		return err
	}

	// Write data
	for _, rate := range rates {
		record := []string{
			fmt.Sprintf("%d", rate.ID),
			rate.FromCurrency,
			rate.ToCurrency,
			fmt.Sprintf("%.6f", rate.Rate),
			rate.EffectiveDate.Format(time.RFC3339),
			rate.CreatedAt.Format(time.RFC3339),
			rate.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportForexExcel exports forex rates to Excel format
func (s *ExportImportService) ExportForexExcel() (*excelize.File, error) {
	var rates []models.Forex
	if err := s.db.Order("id ASC").Find(&rates).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Forex Rates"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := []string{"ID", "From Currency", "To Currency", "Rate", "Effective Date", "Created At", "Updated At"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
	}

	// Apply header styling
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err := f.SetCellStyle(sheetName, "A1", "G1", headerStyle); err != nil {
		return nil, err
	}

	// Write data
	for i, rate := range rates {
		row := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rate.ID); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rate.FromCurrency); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), rate.ToCurrency); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rate.Rate); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), rate.EffectiveDate.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), rate.CreatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), rate.UpdatedAt.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Auto-fit columns
	if err := f.SetColWidth(sheetName, "A", "A", 10); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "B", "C", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "D", "D", 15); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(sheetName, "E", "G", 25); err != nil {
		return nil, err
	}

	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	return f, nil
}

// ImportForexCSV imports forex rates from CSV format
func (s *ExportImportService) ImportForexCSV(r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must contain at least a header and one data row")
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	forexSvc := NewForexService(s.db)

	// Skip header (first row)
	for i, record := range records[1:] {
		if len(record) < 5 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		fromCurrency := strings.ToUpper(strings.TrimSpace(record[1]))
		toCurrency := strings.ToUpper(strings.TrimSpace(record[2]))
		rateStr := strings.TrimSpace(record[3])
		effectiveDateStr := strings.TrimSpace(record[4])

		if fromCurrency == "" || toCurrency == "" {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: currency codes cannot be empty", i+2))
			continue
		}

		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: invalid rate: %v", i+2, err))
			continue
		}

		effectiveDate, err := time.Parse(time.RFC3339, effectiveDateStr)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: invalid date format: %v", i+2, err))
			continue
		}

		// Try to create the forex rate
		_, err = forexSvc.Create(fromCurrency, toCurrency, rate, effectiveDate)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+2, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}
