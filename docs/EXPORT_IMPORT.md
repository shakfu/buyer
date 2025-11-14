# Export/Import Documentation

**Last Updated:** 2025-11-14

## Overview

The Buyer application supports systematic export and import of data in both CSV and Excel (.xlsx) formats. This feature allows you to:

- **Export** data for backup, reporting, or analysis
- **Import** bulk data from spreadsheets
- **Migrate** data between systems
- **Share** data with external stakeholders

## Supported Entities

The following entities support export/import:

| Entity | CSV Export | CSV Import | Excel Export | Excel Import |
|--------|------------|------------|--------------|--------------|
| **Brands** | ✅ | ✅ | ✅ | ❌ |
| **Vendors** | ✅ | ✅ | ✅ | ❌ |
| **Products** | ✅ | ❌ | ✅ | ❌ |
| **Quotes** | ✅ | ❌ | ✅ | ❌ |
| **Forex Rates** | ✅ | ✅ | ✅ | ❌ |

**Note:** Excel import is currently not supported. Use CSV format for importing data.

---

## CLI Usage

### Export Commands

Export data from the command line:

```bash
# Export brands to CSV
buyer export brands brands.csv

# Export brands to Excel
buyer export brands brands.xlsx

# Export vendors to CSV
buyer export vendors vendors.csv

# Export vendors to Excel
buyer export vendors vendors.xlsx

# Export products to CSV
buyer export products products.csv

# Export products to Excel
buyer export products products.xlsx

# Export quotes to CSV
buyer export quotes quotes.csv

# Export quotes to Excel
buyer export quotes quotes.xlsx

# Export forex rates to CSV
buyer export forex forex_rates.csv

# Export forex rates to Excel
buyer export forex forex_rates.xlsx
```

**Format Detection:**
- File extension `.csv` → CSV format
- File extension `.xlsx` → Excel format

### Import Commands

Import data from CSV files:

```bash
# Import brands from CSV
buyer import brands brands.csv

# Import vendors from CSV
buyer import vendors vendors.csv

# Import forex rates from CSV
buyer import forex forex_rates.csv
```

**Import Output:**
```
Import Summary:
  Successfully imported: 15 brands
  Errors: 2

Error Details:
  - Row 5: brand name cannot be empty
  - Row 12: Brand with name 'Apple' already exists
```

---

## Web Interface

### Exporting Data

Each entity page in the web interface has export buttons:

1. Navigate to the entity page (e.g., `/brands`)
2. Click **"Export CSV"** or **"Export Excel"** button
3. File will download automatically

**Export Endpoints:**

```
GET /export/brands/csv      → brands.csv
GET /export/brands/excel    → brands.xlsx
GET /export/vendors/csv     → vendors.csv
GET /export/vendors/excel   → vendors.xlsx
GET /export/products/csv    → products.csv
GET /export/products/excel  → products.xlsx
GET /export/quotes/csv      → quotes.csv
GET /export/quotes/excel    → quotes.xlsx
GET /export/forex/csv       → forex_rates.csv
GET /export/forex/excel     → forex_rates.xlsx
```

### Importing Data

Import data via file upload:

1. Navigate to the entity page
2. Click **"Import"** button
3. Select CSV file
4. Review import summary

**Import Endpoints:**

```
POST /import/brands   → Upload brands.csv
POST /import/vendors  → Upload vendors.csv
POST /import/forex    → Upload forex_rates.csv
```

**Response Format:**
```json
{
  "success": 15,
  "errors": 2,
  "error_details": [
    "Row 5: brand name cannot be empty",
    "Row 12: Brand with name 'Apple' already exists"
  ]
}
```

---

## CSV Format Specifications

### Brands CSV

**Format:**
```csv
ID,Name,CreatedAt,UpdatedAt
1,Apple,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
2,Samsung,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
3,Sony,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
```

**Fields:**
- `ID`: Auto-generated (ignored on import)
- `Name`: Brand name (required, unique)
- `CreatedAt`: Timestamp (auto-generated on import)
- `UpdatedAt`: Timestamp (auto-generated on import)

### Vendors CSV

**Format:**
```csv
ID,Name,Currency,DiscountCode,ContactPerson,Email,Phone,Website,AddressLine1,AddressLine2,City,State,PostalCode,Country,TaxID,PaymentTerms,CreatedAt,UpdatedAt
1,B&H Photo,USD,SAVE10,John Doe,john@bhphoto.com,212-555-0100,https://bhphoto.com,420 9th Ave,,New York,NY,10001,US,12-3456789,Net 30,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
```

**Required Fields:**
- `Name`: Vendor name (required, unique)
- `Currency`: 3-letter ISO 4217 code (required, e.g., USD, EUR, GBP)

**Optional Fields:**
- `DiscountCode`: Vendor-specific discount code
- `ContactPerson`: Primary contact name
- `Email`: Contact email address
- `Phone`: Contact phone number
- `Website`: Vendor website URL
- `AddressLine1`, `AddressLine2`: Street address
- `City`, `State`, `PostalCode`: Location details
- `Country`: 2-letter ISO 3166-1 alpha-2 code
- `TaxID`: Tax identification number
- `PaymentTerms`: Payment terms (e.g., "Net 30", "Net 60")

### Products CSV

**Format:**
```csv
ID,Name,BrandID,BrandName,SpecificationID,SpecificationName,SKU,Description,UnitOfMeasure,MinOrderQty,LeadTimeDays,IsActive,DiscontinuedAt,CreatedBy,UpdatedBy,CreatedAt,UpdatedAt
1,iPhone 15 Pro,1,Apple,2,Smartphone,IPHONE15PRO,Latest flagship phone,each,1,7,true,,,admin,,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
```

**Note:** Products cannot be imported via CSV (requires existing Brand and Specification references). Export is for reporting purposes only.

### Quotes CSV

**Format:**
```csv
ID,VendorID,VendorName,ProductID,ProductName,Price,Currency,ConvertedPrice,ConversionRate,MinQuantity,QuoteDate,ValidUntil,Status,Version,Notes,CreatedBy,UpdatedBy,CreatedAt,UpdatedAt
1,1,B&H Photo,1,iPhone 15 Pro,1199.99,USD,1199.99,1.000000,1,2024-01-01T00:00:00Z,2024-03-01T00:00:00Z,active,1,Best price for bulk orders,sales@example.com,,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
```

**Note:** Quotes cannot be imported via CSV (complex relationships and validation). Export is for reporting and analysis.

### Forex Rates CSV

**Format:**
```csv
ID,FromCurrency,ToCurrency,Rate,EffectiveDate,CreatedAt,UpdatedAt
1,EUR,USD,1.200000,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
2,GBP,USD,1.350000,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
3,JPY,USD,0.006700,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
```

**Required Fields:**
- `FromCurrency`: 3-letter ISO 4217 code (e.g., EUR, GBP, JPY)
- `ToCurrency`: 3-letter ISO 4217 code (e.g., USD)
- `Rate`: Exchange rate (decimal, 6 decimal places)
- `EffectiveDate`: Date when rate becomes effective (RFC3339 format)

---

## Excel Format Specifications

### Excel Features

Excel exports include:

- ✅ **Formatted Headers**: Bold with gray background
- ✅ **Auto-sized Columns**: Optimal column widths
- ✅ **Multiple Sheets**: Each entity in separate sheet
- ✅ **Data Types**: Proper numeric, date, and text formatting
- ✅ **UTF-8 Support**: International characters supported

### Excel Sheet Names

- Brands → **"Brands"** sheet
- Vendors → **"Vendors"** sheet
- Products → **"Products"** sheet
- Quotes → **"Quotes"** sheet
- Forex Rates → **"Forex Rates"** sheet

---

## Import Validation

### Validation Rules

**Brands:**
- ✅ Name cannot be empty
- ✅ Name must be unique
- ✅ Whitespace is trimmed

**Vendors:**
- ✅ Name cannot be empty
- ✅ Name must be unique
- ✅ Currency must be 3-letter code
- ✅ Currency defaults to USD if empty

**Forex Rates:**
- ✅ FromCurrency and ToCurrency cannot be empty
- ✅ Both must be 3-letter codes
- ✅ Rate must be positive number
- ✅ EffectiveDate must be valid RFC3339 timestamp

### Error Handling

**Import Process:**
1. Read CSV file
2. Validate header row
3. Process each data row
4. Track successes and errors
5. Return summary with error details

**Example Error Messages:**
```
Row 3: brand name cannot be empty
Row 5: Brand with name 'Apple' already exists
Row 8: invalid currency code: XY (must be 3 letters)
Row 12: invalid rate: abc (must be a number)
Row 15: invalid date format: 2024-01-01 (use RFC3339)
```

---

## Use Cases

### Bulk Data Entry

Import initial data from spreadsheets:

```bash
# Create brands from spreadsheet
buyer import brands initial_brands.csv

# Create vendors from spreadsheet
buyer import vendors initial_vendors.csv

# Import exchange rates
buyer import forex exchange_rates_2024.csv
```

### Data Backup

Export all data for backup:

```bash
# Export all entities
buyer export brands backup/brands_$(date +%Y%m%d).csv
buyer export vendors backup/vendors_$(date +%Y%m%d).csv
buyer export products backup/products_$(date +%Y%m%d).csv
buyer export quotes backup/quotes_$(date +%Y%m%d).csv
buyer export forex backup/forex_$(date +%Y%m%d).csv
```

### Reporting

Export to Excel for analysis:

```bash
# Generate monthly reports
buyer export quotes reports/quotes_$(date +%Y%m).xlsx
buyer export vendors reports/vendors_$(date +%Y%m).xlsx
```

### Data Migration

Migrate between environments:

```bash
# On production server
buyer export brands prod_brands.csv
buyer export vendors prod_vendors.csv

# Transfer files to staging
scp prod_*.csv staging:/tmp/

# On staging server
buyer import brands /tmp/prod_brands.csv
buyer import vendors /tmp/prod_vendors.csv
```

### External Sharing

Share data with partners:

```bash
# Export vendor list to Excel
buyer export vendors vendor_directory.xlsx

# Email or upload to shared drive
```

---

## Best Practices

### Exporting

1. **Regular Backups**: Export data weekly to CSV for disaster recovery
2. **Excel for Reports**: Use Excel format for stakeholder presentations
3. **Date Stamping**: Include dates in filenames (e.g., `brands_20241114.csv`)
4. **Version Control**: Keep export files in version-controlled directories

### Importing

1. **Validate First**: Review CSV format before importing
2. **Test with Small Batches**: Import a few rows first to test
3. **Check for Duplicates**: Ensure no name conflicts before import
4. **Review Error Messages**: Address all errors before re-importing
5. **Backup Before Import**: Export current data before bulk imports

### Data Quality

1. **Clean Data**: Remove empty rows and trim whitespace
2. **Validate Codes**: Ensure currency codes are valid ISO 4217
3. **Date Format**: Use RFC3339 format (YYYY-MM-DDTHH:MM:SSZ)
4. **UTF-8 Encoding**: Save CSV files with UTF-8 encoding
5. **Consistent Naming**: Use consistent capitalization and spelling

---

## API Integration

### Programmatic Export

Use HTTP GET requests:

```bash
# Export brands as CSV
curl -o brands.csv http://localhost:8080/export/brands/csv

# Export vendors as Excel
curl -o vendors.xlsx http://localhost:8080/export/vendors/excel

# With authentication
curl -u admin:password -o brands.csv http://localhost:8080/export/brands/csv
```

### Programmatic Import

Use HTTP POST with multipart form data:

```bash
# Import brands
curl -X POST \
  -F "file=@brands.csv" \
  http://localhost:8080/import/brands

# Import vendors
curl -X POST \
  -F "file=@vendors.csv" \
  http://localhost:8080/import/vendors

# With authentication
curl -u admin:password \
  -X POST \
  -F "file=@brands.csv" \
  http://localhost:8080/import/brands
```

**Response:**
```json
{
  "success": 15,
  "errors": 0,
  "error_details": []
}
```

---

## Troubleshooting

### Common Issues

**1. "CSV file must contain at least a header and one data row"**
- Ensure CSV has both header row and at least one data row
- Check for empty file

**2. "Only CSV files are supported for import"**
- Excel import not yet supported
- Convert .xlsx to .csv first

**3. "Brand with name 'Apple' already exists"**
- Duplicate names not allowed
- Check existing data or rename in CSV

**4. "invalid currency code: XY (must be 3 letters)"**
- Currency codes must be exactly 3 letters
- Use ISO 4217 codes (USD, EUR, GBP, etc.)

**5. "invalid date format"**
- Dates must be in RFC3339 format
- Format: `2024-01-01T00:00:00Z`
- Use Excel formula: `=TEXT(A1,"yyyy-mm-dd")&"T00:00:00Z"`

### Getting Help

- Check CSV format in this documentation
- Review error messages in import summary
- Test with example data from exports
- Contact support with error details

---

## Technical Implementation

### Library Used

**Excel Support:** [excelize](https://github.com/xuri/excelize) v2.10.0
- Fast and memory-efficient
- Pure Go implementation
- Supports Excel 2007+ (.xlsx)
- Full formatting capabilities

### Service Layer

**Export/Import Service:** `internal/services/export_import.go`
- `ExportBrandsCSV(w io.Writer) error`
- `ExportBrandsExcel() (*excelize.File, error)`
- `ImportBrandsCSV(r io.Reader) (*ImportResult, error)`
- Similar methods for other entities

### CLI Commands

**Export:** `cmd/buyer/export.go`
```go
buyer export brands brands.csv
buyer export brands brands.xlsx
```

**Import:** `cmd/buyer/import.go`
```go
buyer import brands brands.csv
```

### Web Handlers

**Handlers:** `cmd/buyer/web_export.go`
- GET `/export/{entity}/csv` → Download CSV
- GET `/export/{entity}/excel` → Download Excel
- POST `/import/{entity}` → Upload CSV

---

## Future Enhancements

### Planned Features

- [ ] Excel import support
- [ ] Products CSV import (with FK resolution)
- [ ] Quotes CSV import (with validation)
- [ ] Multi-sheet Excel export (all entities in one file)
- [ ] Import templates download
- [ ] Data validation preview before import
- [ ] Incremental imports (update existing + add new)
- [ ] Import history tracking

---

## Version History

### v1.0.0 (2025-11-14)

**Initial Release:**
- CSV export for all entities
- Excel export for all entities
- CSV import for Brands, Vendors, Forex
- CLI commands
- Web API endpoints
- Comprehensive validation
- Error reporting

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/shakfu/buyer/issues
- Documentation: https://github.com/shakfu/buyer/docs
- Email: support@example.com

---

**Last Updated:** 2025-11-14
**Version:** 1.0.0
