# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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

### Changed
  - Web forms now automatically clear input values after successful submission
  - Improved user experience by resetting forms to default state after adding new items
  - Refactored web handlers to eliminate ~850 lines of duplicated code by consolidating CRUD endpoints into `SetupCRUDHandlers()` function

### Fixed
  - Removed massive code duplication in route handlers (C1 from CODE_REVIEW.md)
  - Replaced inline HTML generation with consistent use of render functions
  - Fixed unsafe string concatenation in HTML generation (C3 from CODE_REVIEW.md)
  - All HTML rendering now uses proper template auto-escaping instead of manual string building
  - **Product SKU field** - Changed from string to pointer type (*string) to properly handle NULL values in unique constraint, preventing UNIQUE constraint violations when multiple products have no SKU

## [0.1.0]

### Added
  - Initial implementation created