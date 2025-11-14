# buyer

A purchasing support and vendor quote management tool written in Go.

*buyer* helps you track brands, products, vendors, and price quotes across multiple vendors with multi-currency support.

## Features

- **Brand Management**: Track manufacturing brands
- **Product Catalog**: Organize products by brand
- **Vendor Management**: Manage vendors with currency and discount codes
- **Quote Tracking**: Record and compare price quotes with automatic currency conversion
- **Multi-Currency Support**: Built-in forex rate management and automatic USD conversion
- **Document Management**: Attach documents to any entity (vendors, quotes, products, etc.)
- **Vendor Rating System**: Rate vendors on price, quality, delivery, and service
- **Performance Dashboard**: Visualize vendor performance with interactive charts and analytics
- **Export/Import**: Export to CSV or Excel (.xlsx), import from CSV with validation
- **CLI Interface**: Full-featured command-line interface with verbose mode
- **Web Interface**: Modern HTMX-powered web UI with CRUD operations
- **Comprehensive Testing**: Full test coverage for all services

## Installation

### Prerequisites

- Go 1.24 or higher

### Build from Source

```bash
# Clone the repository
cd /path/to/buyer

# Install dependencies
go mod download

# Build the binary
make build

# Or install globally
make install
```

## Quick Start

```bash
# Add a brand
buyer add brand Apple

# Add a product
buyer add product "MacBook Pro" --brand Apple

# Add a vendor
buyer add vendor "B&H Photo" --currency USD --discount SAVE10

# Add a forex rate (EUR to USD)
buyer add forex --from EUR --to USD --rate 1.20

# Add a quote
buyer add quote --vendor "B&H Photo" --product "MacBook Pro" --price 2499.99 --currency USD

# List all brands
buyer list brands

# Search across entities
buyer search apple

# Start the web interface
buyer web
```

## Usage

### Brand Commands

```bash
# Add a brand
buyer add brand [name]

# List all brands
buyer list brands [--limit N] [--offset N]

# Update a brand
buyer update brand [id] [new_name]

# Delete a brand
buyer delete brand [id] [-f|--force]
```

### Product Commands

```bash
# Add a product
buyer add product [name] --brand [brand_name]

# List all products
buyer list products [--limit N] [--offset N]

# Update a product
buyer update product [id] [new_name]

# Delete a product
buyer delete product [id] [-f|--force]
```

### Vendor Commands

```bash
# Add a vendor
buyer add vendor [name] --currency [code] --discount [code]

# List all vendors
buyer list vendors [--limit N] [--offset N]

# Update a vendor
buyer update vendor [id] [new_name]

# Delete a vendor
buyer delete vendor [id] [-f|--force]
```

### Quote Commands

```bash
# Add a quote
buyer add quote --vendor [name] --product [name] --price [amount] --currency [code] --notes [text]

# List all quotes
buyer list quotes [--limit N] [--offset N]

# Delete a quote
buyer delete quote [id] [-f|--force]
```

### Forex Commands

```bash
# Add a forex rate
buyer add forex --from [code] --to [code] --rate [rate]

# List forex rates
buyer list forex [--limit N] [--offset N]

# Delete a forex rate
buyer delete forex [id] [-f|--force]
```

### Document Commands

```bash
# Add a document
buyer add document --entity-type [type] --entity-id [id] --file-name [name] --file-path [path] \
  --file-type [type] --file-size [bytes] --description [text] --uploaded-by [email]

# List all documents
buyer list documents [--limit N] [--offset N]

# List documents by entity type
buyer list documents --entity-type vendor [--limit N] [--offset N]

# List documents for specific entity
buyer list documents --entity-type vendor --entity-id 1

# Delete a document
buyer delete document [id] [-f|--force]
```

### Vendor Rating Commands

```bash
# Add a vendor rating
buyer add vendor-rating --vendor-id [id] --po-id [id] \
  --price-rating [1-5] --quality-rating [1-5] \
  --delivery-rating [1-5] --service-rating [1-5] \
  --comments [text] --rated-by [email]

# List all vendor ratings
buyer list vendor-ratings [--limit N] [--offset N]

# List ratings for a specific vendor
buyer list vendor-ratings --vendor-id [id]

# Delete a vendor rating
buyer delete vendor-rating [id] [-f|--force]
```

### Export Commands

```bash
# Export brands to CSV
buyer export brands brands.csv

# Export brands to Excel
buyer export brands brands.xlsx

# Export vendors to CSV or Excel
buyer export vendors vendors.csv
buyer export vendors vendors.xlsx

# Export products, quotes, or forex rates
buyer export products products.csv
buyer export quotes quotes.xlsx
buyer export forex rates.csv
```

### Import Commands

```bash
# Import brands from CSV
buyer import brands brands.csv

# Import vendors from CSV
buyer import vendors vendors.csv

# Import forex rates from CSV
buyer import forex rates.csv
```

**Note:** CSV format is auto-detected by file extension. See [EXPORT_IMPORT.md](docs/EXPORT_IMPORT.md) for detailed format specifications.

### Search

```bash
# Search across all entities
buyer search [query]
```

### Web Interface

```bash
# Start web server (default port: 8080)
buyer web

# Start on custom port
buyer web --port 3000
```

Then visit http://localhost:8080 in your browser.

**Available Pages:**
- `/` - Dashboard with key metrics and recent activity
- `/brands` - Brand management with CRUD operations (export/import available)
- `/products` - Product catalog management (export available)
- `/vendors` - Vendor management (export/import available)
- `/quotes` - Price quote tracking and comparison (export available)
- `/documents` - Document management for all entities
- `/vendor-ratings` - Vendor rating submission and listing
- `/vendor-performance` - Performance analytics dashboard with charts

**Export/Import API:**
- `GET /export/{entity}/csv` - Download CSV file
- `GET /export/{entity}/excel` - Download Excel (.xlsx) file
- `POST /import/{entity}` - Upload and import CSV file

## Configuration

buyer supports configuration through environment variables and `.env` files. See [CONFIG.md](CONFIG.md) for detailed documentation.

### Using .env File (Recommended)

```bash
# Copy the example configuration
cp .env.example .env

# Edit .env with your settings
nano .env

# Run the application (it will automatically load .env)
buyer web
```

**Note:** `.env` file is optional. Environment variables already set take precedence over `.env` file values.

### Quick Configuration Examples

**Custom database path:**
```bash
BUYER_DB_PATH=/var/lib/buyer/buyer.db buyer web
```

**Custom web port:**
```bash
BUYER_WEB_PORT=3000 buyer web
# Or using flag:
buyer web --port 3000
```

**Production deployment with security:**
```bash
export BUYER_ENV=production
export BUYER_ENABLE_AUTH=true
export BUYER_USERNAME=admin
export BUYER_PASSWORD=YourSecureP@ss123!
export BUYER_ENABLE_CSRF=true
buyer web
```

**Security Note:** When `BUYER_ENABLE_AUTH=true`, you **must** provide `BUYER_USERNAME` and `BUYER_PASSWORD` (no defaults). Password must meet requirements: 12+ chars, uppercase, lowercase, digit, special character.

**All available environment variables:**
- `BUYER_ENV` - Environment mode (development/production/testing)
- `BUYER_DB_PATH` - Database file path
- `BUYER_WEB_PORT` - Web server port
- `BUYER_ENABLE_AUTH` - Enable HTTP basic authentication (default: false)
- `BUYER_USERNAME` - Basic auth username (required if auth enabled, no default)
- `BUYER_PASSWORD` - Basic auth password (required if auth enabled, no default)
- `BUYER_ENABLE_CSRF` - Enable CSRF protection (default: false)

See [CONFIG.md](CONFIG.md) for comprehensive configuration guide including defaults, loading sequence, and troubleshooting.

## Project Structure

```
buyer/
├── cmd/buyer/           # CLI application entry point
│   ├── main.go          # Main application
│   ├── add.go           # Add commands
│   ├── list.go          # List commands
│   ├── update.go        # Update commands
│   ├── delete.go        # Delete commands
│   ├── export.go        # Export commands (CSV/Excel)
│   ├── import.go        # Import commands (CSV)
│   ├── search.go        # Search command
│   ├── web.go           # Web server
│   └── web_export.go    # Export/import web handlers
├── internal/
│   ├── models/          # GORM data models
│   ├── services/        # Business logic layer
│   │   ├── brand.go
│   │   ├── product.go
│   │   ├── vendor.go
│   │   ├── quote.go
│   │   ├── forex.go
│   │   ├── export_import.go  # CSV/Excel export/import
│   │   └── errors.go
│   └── config/          # Configuration management
├── Makefile             # Build automation
├── go.mod               # Go module definition
└── README.md            # This file
```

## Architecture

`buyer` follows clean architecture principles:

1. **Models Layer** (`internal/models`): GORM-based ORM models defining database schema
2. **Service Layer** (`internal/services`): Business logic with validation and error handling
3. **Presentation Layer** (`cmd/buyer`): CLI and web interfaces

### Key Design Patterns

- **Service Layer Pattern**: Business logic isolated from data access
- **Repository Pattern**: GORM provides data access abstraction
- **Dependency Injection**: Services receive database connections
- **Error Handling**: Custom error types (ValidationError, DuplicateError, NotFoundError)

## Data Model

### Entities

- **Brand**: Manufacturing entity (e.g., Apple, Sony)
- **Product**: Item associated with a brand (e.g., MacBook Pro)
- **Vendor**: Selling entity with currency info (e.g., B&H Photo)
- **Quote**: Price quote from a vendor for a product
- **Forex**: Currency exchange rate
- **Document**: File attachment with polymorphic entity association
- **VendorRating**: Multi-category vendor performance rating (price, quality, delivery, service)
- **Requisition**: Internal purchase request with multi-item support
- **Project**: Project tracking with requisition management
- **PurchaseOrder**: Formal purchase order linked to quotes and requisitions

### Relationships

- Brands have many Products
- Vendors have many Brands (many-to-many)
- Products have many Quotes
- Vendors have many Quotes
- Vendors have many VendorRatings
- Quotes automatically convert to USD using Forex rates
- Documents use polymorphic associations (can attach to any entity)
- VendorRatings can optionally link to PurchaseOrders
- Projects have many Requisitions (many-to-many)
- PurchaseOrders reference Quotes and Requisitions

## Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run with race detection
make test-race
```

### Building

```bash
# Build binary
make build

# Install globally
make install

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```

### Testing Philosophy

- **Comprehensive Coverage**: All service methods are tested
- **Isolated Tests**: Each test uses in-memory SQLite database
- **Behavior Testing**: Tests verify business logic, not implementation
- **Error Cases**: Validation and error conditions are thoroughly tested

## Configuration

### Environment Variables

- `BUYER_ENV`: Set environment (development, production, testing)

### Command-Line Flags

- `-v, --verbose`: Enable verbose logging (displays SQL queries)
  ```bash
  # Show SQL queries for debugging
  buyer -v list brands
  buyer --verbose add brand Apple
  ```

### Database

- Development/Production: `~/.buyer/buyer.db` (SQLite)
- Testing: In-memory SQLite database

## Technologies Used

- **Language**: Go 1.21+
- **ORM**: GORM v1.31+ with SQLite driver
- **CLI Framework**: Cobra v1.10+
- **Web Framework**: Fiber v2.52+
- **Excel Library**: excelize v2.10+ for .xlsx export/import
- **Table Rendering**: rodaine/table v1.3+
- **Testing**: Go standard testing package

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## License

See LICENSE file for details.

## Support

For issues and feature requests, please open an issue on the repository.
