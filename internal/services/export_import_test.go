package services

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shakfu/buyer/internal/models"
)

func TestExportImportService_BrandsCSV(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Create service and sample data
	brandSvc := NewBrandService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test brands
	_, _ = brandSvc.Create("Apple")
	_, _ = brandSvc.Create("Samsung")
	_, _ = brandSvc.Create("Sony")

	// Test CSV export
	t.Run("Export Brands to CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportSvc.ExportBrandsCSV(&buf)
		if err != nil {
			t.Fatalf("Failed to export brands to CSV: %v", err)
		}

		csvContent := buf.String()
		if !strings.Contains(csvContent, "Apple") {
			t.Error("CSV should contain 'Apple'")
		}
		if !strings.Contains(csvContent, "Samsung") {
			t.Error("CSV should contain 'Samsung'")
		}
		if !strings.Contains(csvContent, "Sony") {
			t.Error("CSV should contain 'Sony'")
		}
		if !strings.Contains(csvContent, "ID,Name,CreatedAt,UpdatedAt") {
			t.Error("CSV should contain header row")
		}
	})

	// Test CSV import
	t.Run("Import Brands from CSV", func(t *testing.T) {
		csvData := `ID,Name,CreatedAt,UpdatedAt
0,Microsoft,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
0,Dell,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
0,HP,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z`

		reader := strings.NewReader(csvData)
		result, err := exportSvc.ImportBrandsCSV(reader)
		if err != nil {
			t.Fatalf("Failed to import brands from CSV: %v", err)
		}

		if result.SuccessCount != 3 {
			t.Errorf("Expected 3 successful imports, got %d", result.SuccessCount)
		}
		if result.ErrorCount != 0 {
			t.Errorf("Expected 0 errors, got %d: %v", result.ErrorCount, result.Errors)
		}

		// Verify brands were created
		brand, err := brandSvc.GetByName("Microsoft")
		if err != nil {
			t.Error("Microsoft brand should exist after import")
		}
		if brand.Name != "Microsoft" {
			t.Errorf("Expected brand name 'Microsoft', got '%s'", brand.Name)
		}
	})

	// Test CSV import with errors
	t.Run("Import Brands with Errors", func(t *testing.T) {
		csvData := `ID,Name,CreatedAt,UpdatedAt
0,,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
0,Apple,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z`

		reader := strings.NewReader(csvData)
		result, err := exportSvc.ImportBrandsCSV(reader)
		if err != nil {
			t.Fatalf("Failed to import brands from CSV: %v", err)
		}

		if result.ErrorCount != 2 {
			t.Errorf("Expected 2 errors (empty name + duplicate), got %d", result.ErrorCount)
		}
	})
}

func TestExportImportService_BrandsExcel(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	// Create service and sample data
	brandSvc := NewBrandService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test brands
	_, _ = brandSvc.Create("Apple")
	_, _ = brandSvc.Create("Samsung")

	// Test Excel export
	t.Run("Export Brands to Excel", func(t *testing.T) {
		f, err := exportSvc.ExportBrandsExcel()
		if err != nil {
			t.Fatalf("Failed to export brands to Excel: %v", err)
		}

		// Check sheet exists
		sheetName := "Brands"
		sheets := f.GetSheetList()
		found := false
		for _, s := range sheets {
			if s == sheetName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected sheet '%s' not found in Excel file", sheetName)
		}

		// Check header
		header, err := f.GetCellValue(sheetName, "A1")
		if err != nil {
			t.Errorf("Failed to get cell value: %v", err)
		}
		if header != "ID" {
			t.Errorf("Expected header 'ID', got '%s'", header)
		}

		// Check data
		brandName, err := f.GetCellValue(sheetName, "B2")
		if err != nil {
			t.Errorf("Failed to get brand name: %v", err)
		}
		if brandName != "Apple" && brandName != "Samsung" {
			t.Errorf("Expected 'Apple' or 'Samsung', got '%s'", brandName)
		}
	})
}

func TestExportImportService_VendorsCSV(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test vendors
	_, _ = vendorSvc.Create("B&H Photo", "USD", "SAVE10")
	_, _ = vendorSvc.Create("Adorama", "USD", "SUMMER15")

	t.Run("Export Vendors to CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportSvc.ExportVendorsCSV(&buf)
		if err != nil {
			t.Fatalf("Failed to export vendors to CSV: %v", err)
		}

		csvContent := buf.String()
		if !strings.Contains(csvContent, "B&H Photo") {
			t.Error("CSV should contain 'B&H Photo'")
		}
		if !strings.Contains(csvContent, "SAVE10") {
			t.Error("CSV should contain 'SAVE10'")
		}
	})

	t.Run("Import Vendors from CSV", func(t *testing.T) {
		csvData := `ID,Name,Currency,DiscountCode
0,Amazon,USD,PRIME
0,Newegg,USD,TECH20`

		reader := strings.NewReader(csvData)
		result, err := exportSvc.ImportVendorsCSV(reader)
		if err != nil {
			t.Fatalf("Failed to import vendors from CSV: %v", err)
		}

		if result.SuccessCount != 2 {
			t.Errorf("Expected 2 successful imports, got %d", result.SuccessCount)
		}

		// Verify vendor was created
		vendor, err := vendorSvc.GetByName("Amazon")
		if err != nil {
			t.Error("Amazon vendor should exist after import")
		}
		if vendor.Currency != "USD" {
			t.Errorf("Expected currency 'USD', got '%s'", vendor.Currency)
		}
	})
}

func TestExportImportService_VendorsExcel(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	vendorSvc := NewVendorService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	_, _ = vendorSvc.Create("B&H Photo", "USD", "SAVE10")

	t.Run("Export Vendors to Excel", func(t *testing.T) {
		f, err := exportSvc.ExportVendorsExcel()
		if err != nil {
			t.Fatalf("Failed to export vendors to Excel: %v", err)
		}

		sheetName := "Vendors"
		vendorName, err := f.GetCellValue(sheetName, "B2")
		if err != nil {
			t.Errorf("Failed to get vendor name: %v", err)
		}
		if vendorName != "B&H Photo" {
			t.Errorf("Expected 'B&H Photo', got '%s'", vendorName)
		}

		currency, err := f.GetCellValue(sheetName, "C2")
		if err != nil {
			t.Errorf("Failed to get currency: %v", err)
		}
		if currency != "USD" {
			t.Errorf("Expected 'USD', got '%s'", currency)
		}
	})
}

func TestExportImportService_ProductsCSV(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Apple")
	_, _ = productSvc.Create("iPhone 15 Pro", brand.ID, nil)
	_, _ = productSvc.Create("MacBook Pro", brand.ID, nil)

	t.Run("Export Products to CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportSvc.ExportProductsCSV(&buf)
		if err != nil {
			t.Fatalf("Failed to export products to CSV: %v", err)
		}

		csvContent := buf.String()
		if !strings.Contains(csvContent, "iPhone 15 Pro") {
			t.Error("CSV should contain 'iPhone 15 Pro'")
		}
		if !strings.Contains(csvContent, "MacBook Pro") {
			t.Error("CSV should contain 'MacBook Pro'")
		}
		if !strings.Contains(csvContent, "Apple") {
			t.Error("CSV should contain brand name 'Apple'")
		}
	})
}

func TestExportImportService_ProductsExcel(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	brand, _ := brandSvc.Create("Apple")
	_, _ = productSvc.Create("iPhone 15 Pro", brand.ID, nil)

	t.Run("Export Products to Excel", func(t *testing.T) {
		f, err := exportSvc.ExportProductsExcel()
		if err != nil {
			t.Fatalf("Failed to export products to Excel: %v", err)
		}

		sheetName := "Products"
		productName, err := f.GetCellValue(sheetName, "B2")
		if err != nil {
			t.Errorf("Failed to get product name: %v", err)
		}
		if productName != "iPhone 15 Pro" {
			t.Errorf("Expected 'iPhone 15 Pro', got '%s'", productName)
		}

		brandName, err := f.GetCellValue(sheetName, "D2")
		if err != nil {
			t.Errorf("Failed to get brand name: %v", err)
		}
		if brandName != "Apple" {
			t.Errorf("Expected 'Apple', got '%s'", brandName)
		}
	})
}

func TestExportImportService_QuotesCSV(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Apple")
	product, _ := productSvc.Create("iPhone 15 Pro", brand.ID, nil)
	vendor, _ := vendorSvc.Create("B&H Photo", "USD", "")
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	input := CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     1199.99,
		Currency:  "USD",
		QuoteDate: time.Now(),
	}
	_, _ = quoteSvc.Create(input)

	t.Run("Export Quotes to CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportSvc.ExportQuotesCSV(&buf)
		if err != nil {
			t.Fatalf("Failed to export quotes to CSV: %v", err)
		}

		csvContent := buf.String()
		if !strings.Contains(csvContent, "B&H Photo") {
			t.Error("CSV should contain vendor name 'B&H Photo'")
		}
		if !strings.Contains(csvContent, "iPhone 15 Pro") {
			t.Error("CSV should contain product name 'iPhone 15 Pro'")
		}
		if !strings.Contains(csvContent, "1199.99") {
			t.Error("CSV should contain price '1199.99'")
		}
	})
}

func TestExportImportService_QuotesExcel(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	brandSvc := NewBrandService(cfg.DB)
	productSvc := NewProductService(cfg.DB)
	vendorSvc := NewVendorService(cfg.DB)
	quoteSvc := NewQuoteService(cfg.DB)
	forexSvc := NewForexService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test data
	brand, _ := brandSvc.Create("Apple")
	product, _ := productSvc.Create("iPhone 15 Pro", brand.ID, nil)
	vendor, _ := vendorSvc.Create("B&H Photo", "USD", "")
	_, _ = forexSvc.Create("USD", "USD", 1.0, time.Now())

	input := CreateQuoteInput{
		VendorID:  vendor.ID,
		ProductID: product.ID,
		Price:     1199.99,
		Currency:  "USD",
		QuoteDate: time.Now(),
	}
	_, _ = quoteSvc.Create(input)

	t.Run("Export Quotes to Excel", func(t *testing.T) {
		f, err := exportSvc.ExportQuotesExcel()
		if err != nil {
			t.Fatalf("Failed to export quotes to Excel: %v", err)
		}

		sheetName := "Quotes"
		vendorName, err := f.GetCellValue(sheetName, "C2")
		if err != nil {
			t.Errorf("Failed to get vendor name: %v", err)
		}
		if vendorName != "B&H Photo" {
			t.Errorf("Expected 'B&H Photo', got '%s'", vendorName)
		}

		productName, err := f.GetCellValue(sheetName, "E2")
		if err != nil {
			t.Errorf("Failed to get product name: %v", err)
		}
		if productName != "iPhone 15 Pro" {
			t.Errorf("Expected 'iPhone 15 Pro', got '%s'", productName)
		}
	})
}

func TestExportImportService_ForexCSV(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	forexSvc := NewForexService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	// Create test forex rates
	effectiveDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _ = forexSvc.Create("EUR", "USD", 1.20, effectiveDate)
	_, _ = forexSvc.Create("GBP", "USD", 1.35, effectiveDate)

	t.Run("Export Forex to CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportSvc.ExportForexCSV(&buf)
		if err != nil {
			t.Fatalf("Failed to export forex to CSV: %v", err)
		}

		csvContent := buf.String()
		if !strings.Contains(csvContent, "EUR") {
			t.Error("CSV should contain 'EUR'")
		}
		if !strings.Contains(csvContent, "1.20") {
			t.Error("CSV should contain rate '1.20'")
		}
	})

	t.Run("Import Forex from CSV", func(t *testing.T) {
		csvData := `ID,FromCurrency,ToCurrency,Rate,EffectiveDate
0,JPY,USD,0.0067,2024-01-01T00:00:00Z
0,CAD,USD,0.75,2024-01-01T00:00:00Z`

		reader := strings.NewReader(csvData)
		result, err := exportSvc.ImportForexCSV(reader)
		if err != nil {
			t.Fatalf("Failed to import forex from CSV: %v", err)
		}

		if result.SuccessCount != 2 {
			t.Errorf("Expected 2 successful imports, got %d", result.SuccessCount)
		}

		// Verify forex rate was created
		var rate models.Forex
		err = cfg.DB.Where("from_currency = ? AND to_currency = ?", "JPY", "USD").First(&rate).Error
		if err != nil {
			t.Error("JPY/USD rate should exist after import")
		}
		if rate.Rate != 0.0067 {
			t.Errorf("Expected rate 0.0067, got %.4f", rate.Rate)
		}
	})
}

func TestExportImportService_ForexExcel(t *testing.T) {
	cfg := setupTestDB(t)
	defer func() { _ = cfg.Close() }()

	forexSvc := NewForexService(cfg.DB)
	exportSvc := NewExportImportService(cfg.DB)

	effectiveDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _ = forexSvc.Create("EUR", "USD", 1.20, effectiveDate)

	t.Run("Export Forex to Excel", func(t *testing.T) {
		f, err := exportSvc.ExportForexExcel()
		if err != nil {
			t.Fatalf("Failed to export forex to Excel: %v", err)
		}

		sheetName := "Forex Rates"
		fromCurrency, err := f.GetCellValue(sheetName, "B2")
		if err != nil {
			t.Errorf("Failed to get from currency: %v", err)
		}
		if fromCurrency != "EUR" {
			t.Errorf("Expected 'EUR', got '%s'", fromCurrency)
		}

		toCurrency, err := f.GetCellValue(sheetName, "C2")
		if err != nil {
			t.Errorf("Failed to get to currency: %v", err)
		}
		if toCurrency != "USD" {
			t.Errorf("Expected 'USD', got '%s'", toCurrency)
		}
	})
}
