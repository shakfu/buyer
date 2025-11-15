# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
  - **Project procurement dashboard enhancements** - Three new chart calculation functions for comprehensive procurement analysis
    - BOM Items by Value chart: Horizontal bar chart showing aggregate value (quantity × price) for fulfilled requisition items, sorted by highest value
    - Sourcing Performance chart: Tracks quote availability and quality for BOM items (3+ non-stale quotes, fresh quotes only, stale quotes only, no quotes) with savings vs budget calculation and breakdown by requisition
    - Procurement Performance chart: Analyzes PO execution (on-time, delayed, pending), compliance rate (within 5% of best market price), and savings with breakdown by requisition
    - calculateBOMItemsByValue() in project_procurement.go:2480
    - calculateSourcingPerformance() in project_procurement.go:2576
    - calculateProcurementPerformance() in project_procurement.go:2736
    - Added SourcingPerformanceData and ProcurementPerformanceData structs with requisition-level metrics
  - **Tab navigation for procurement dashboard** - CSS-based tab switching for Overview, Analysis, and Charts sections with visual active state indicators
  - **CSV and Excel export/import system** - Comprehensive data export/import functionality using excelize library
    - CSV export for: brands, vendors, products, quotes, and forex rates
    - Excel (.xlsx) export for: brands, vendors, products, quotes, and forex rates
    - CSV import for: brands, vendors, and forex rates with comprehensive validation
    - CLI commands: `buyer export {entity} {filename}` and `buyer import {entity} {filename}`
    - Web API endpoints: GET `/export/{entity}/{csv|excel}` for downloads, POST `/import/{entity}` for uploads
    - ExportImportService with 15+ export/import methods in `internal/services/export_import.go`
    - Auto-format detection by file extension (.csv or .xlsx)
    - Excel exports include formatted headers (bold, gray background) and auto-sized columns
    - Import validation with detailed error reporting including row numbers and specific error messages
    - Import returns success/error counts with detailed error list for troubleshooting
    - Comprehensive test suite with 10 test functions covering all export/import scenarios
    - Complete documentation in `docs/EXPORT_IMPORT.md` with CSV format specifications, usage examples, and best practices
    - Support for UTF-8 and international characters in Excel exports
  - **Document management system** - Full implementation of document management features (D8 from MODEL_ANALYSIS.md)
    - DocumentService with complete CRUD operations
    - CLI commands: `buyer add document` and `buyer list documents` with entity filtering
    - Web CRUD interface at `/documents` with HTMX integration
    - Support for polymorphic document attachments to any entity type
    - Document metadata tracking: file name, type, size, upload info
    - Comprehensive test suite with 8 test functions covering all service operations
  - **Vendor rating system** - Multi-category vendor performance tracking (D6 from MODEL_ANALYSIS.md)
    - VendorRating model with four rating categories: price, quality, delivery, and service
    - VendorRatingService with CRUD operations and analytics functions
    - CLI commands: `buyer add vendor-rating` and `buyer list vendor-ratings` with vendor filtering
    - Web CRUD interface at `/vendor-ratings` with HTMX integration
    - Optional purchase order linkage for context-specific ratings
    - Comprehensive test suite with 8 test functions covering all service operations
  - **Vendor performance dashboard** - Analytics and visualization for vendor ratings
    - Web dashboard at `/vendor-performance` with interactive Vega-Lite charts
    - Overall vendor rankings sorted by average rating
    - Category breakdown visualization (price, quality, delivery, service)
    - Key metrics: total ratings, rated vendors, average overall rating, top performer
    - SQL aggregation with proper null handling for optional rating categories
    - GetVendorPerformance() and GetCategoryAverages() analytics functions
  - **HTMX-powered web interfaces** - Modern interactive web pages with real-time updates
    - Form submissions without page reloads
    - Dynamic table row updates for create/delete operations
    - Improved user experience with instant feedback
  - **Web render functions** - RenderDocumentRow() and RenderVendorRatingRow() for HTMX partial responses
  - MODEL_ANALYSIS.md - Comprehensive analysis of data models with gap analysis and recommendations
  - **PurchaseOrder model** - Track purchase orders from quote acceptance through delivery (D1 from MODEL_ANALYSIS.md)
  - **Vendor contact information** - Added email, phone, website, address fields to Vendor model (D2 from MODEL_ANALYSIS.md)
  - **Vendor business information** - Added TaxID and PaymentTerms fields
  - **PurchaseOrderService** - Complete CRUD service for purchase order management
  - **CLI commands for purchase orders** - add, list, and update purchase-order commands
  - Purchase order status management with workflow validation
  - **Comprehensive test suite for PurchaseOrderService** - 8 test functions covering all CRUD operations, status updates, delivery tracking, and edge cases
  - **Web UI for Purchase Orders** - Full-featured interface for creating, viewing, and managing purchase orders with status tracking, invoice management, and delivery date recording
  - **Purchase order web handlers** - Complete CRUD endpoints for purchase orders with proper validation and error handling
  - **Purchase order fixtures** - Sample data with 6 purchase orders demonstrating various statuses (pending, approved, ordered, shipped, received, cancelled)
  - **Product extended fields** - Added SKU (unique, nullable), Description, UnitOfMeasure, MinOrderQty, LeadTimeDays, IsActive, and DiscontinuedAt fields (D3 from MODEL_ANALYSIS.md)
  - **Quote versioning** - Added Version, PreviousQuoteID, ReplacedBy, MinQuantity, and Status fields for tracking quote negotiations (D4 from MODEL_ANALYSIS.md)
  - **Audit fields** - Added CreatedBy and UpdatedBy fields to Product, Quote, and PurchaseOrder models (D5 from MODEL_ANALYSIS.md)
  - **Document attachment model** - Added polymorphic Document model for attaching files (PDFs, images, spreadsheets) to any entity type (D8 from MODEL_ANALYSIS.md)
  - **Enhanced web UI** - Updated products and quotes templates to display new fields including SKU, description, order quantities, lead times, quote versions, and status
  - **Sample documents in fixtures** - Added 12 example document attachments for vendors, quotes, purchase orders, and products
  - **Detail pages** - Added dedicated detail pages for Products, Quotes, Vendors, and Purchase Orders with comprehensive information display
  - **Simplified table views** - Streamlined list tables to show only essential columns with "View" buttons for accessing full details
  - **Model validation hooks** - Added BeforeSave hooks to validate business constraints at the model level:
    - Quote: Validates positive prices, conversion rates, valid status enums, and non-negative minimum quantities
    - Project: Validates status enum values and non-negative budgets
    - PurchaseOrder: Validates status enums, positive quantities, non-negative amounts (total, shipping, tax)
    - RequisitionItem: Validates positive quantities and non-negative budget per unit
    - Product: Validates non-negative minimum order quantities and lead time days

### Changed
  - Web forms now automatically clear input values after successful submission
  - Improved user experience by resetting forms to default state after adding new items
  - Refactored web handlers to eliminate ~850 lines of duplicated code by consolidating CRUD endpoints into `SetupCRUDHandlers()` function

### Fixed
  - **Procurement dashboard JavaScript bugs** - Fixed multiple data access and display issues
    - Corrected field access path: `data.Progress.ItemsWithQuotes` to `data.Procurement.ItemsWithQuotes` in project-procurement.html:381
    - Added defensive null checks for nested objects: `(item.BestQuote && item.BestQuote.ConvertedPrice)` before calling toFixed()
    - Fixed JSON field naming mismatch: Changed PascalCase JavaScript references to snake_case to match Go's JSON serialization
      - `item.BOMItem.ID` to `item.BOMItem.id`
      - `item.BestQuote.ConvertedPrice` to `item.BestQuote.converted_price`
      - `item.Specification.Name` to `item.BOMItem.specification.name`
      - `item.BOMItem.Quantity` to `item.TotalQuantityNeeded`
  - **Tab navigation styling** - Added CSS to properly hide inactive tabs and highlight active tab with border
  - **Lint errors in project_procurement.go and tests**:
    - Removed unused variable initialization at line 874 (staticcheck SA4006)
    - Added explicit error ignoring (`_, _ =`) to 7 test helper calls in project_procurement_test.go (errcheck)
  - Removed massive code duplication in route handlers (C1 from CODE_REVIEW.md)
  - Replaced inline HTML generation with consistent use of render functions
  - Fixed unsafe string concatenation in HTML generation (C3 from CODE_REVIEW.md)
  - All HTML rendering now uses proper template auto-escaping instead of manual string building
  - **Product SKU field** - Changed from string to pointer type (*string) to properly handle NULL values in unique constraint, preventing UNIQUE constraint violations when multiple products have no SKU
  - **Quote.IsStale() logic** - Fixed incorrect logic that marked long-term valid quotes as stale (D7 from MODEL_ANALYSIS.md)
    - Now correctly handles quotes with ValidUntil set: if not expired, they are not stale regardless of age
    - Only quotes without ValidUntil set are marked stale after 90 days
  - **Linting issues** - Fixed all golangci-lint warnings:
    - Added error checking for all test helper function calls (errcheck)
    - Removed unused `getEnvOrDefault` function in web.go
    - Fixed ineffectual assignment in list.go by using separate variable for fmt.Sscanf error
    - Fixed type mismatch in add.go: changed `*services.Vendor` to `*models.Vendor`
    - Added missing models import in add.go
    - Updated web_test.go to include Document and VendorRating models in AutoMigrate
    - Updated web_test.go setupRoutes call to include docSvc and ratingsSvc parameters

## [0.1.0]

### Added
  - Initial implementation created