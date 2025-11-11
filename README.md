# buyer

A purchasing support and vendor quote management tool written in Go.

*buyer* helps you track brands, products, vendors, and price quotes across multiple vendors with multi-currency support.

## Features

- **Brand Management**: Track manufacturing brands
- **Product Catalog**: Organize products by brand
- **Vendor Management**: Manage vendors with currency and discount codes
- **Quote Tracking**: Record and compare price quotes with automatic currency conversion
- **Multi-Currency Support**: Built-in forex rate management and automatic USD conversion
- **CLI Interface**: Full-featured command-line interface with verbose mode
- **Web Interface**: Simple web UI for viewing data
- **Comprehensive Testing**: Full test coverage for all services

## Installation

### Prerequisites

- Go 1.21 or higher

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

## Project Structure

```
buyer/
├── cmd/buyer/           # CLI application entry point
│   ├── main.go          # Main application
│   ├── add.go           # Add commands
│   ├── list.go          # List commands
│   ├── update.go        # Update commands
│   ├── delete.go        # Delete commands
│   ├── search.go        # Search command
│   └── web.go           # Web server
├── internal/
│   ├── models/          # GORM data models
│   ├── services/        # Business logic layer
│   │   ├── brand.go
│   │   ├── product.go
│   │   ├── vendor.go
│   │   ├── quote.go
│   │   ├── forex.go
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

### Relationships

- Brands have many Products
- Vendors have many Brands (many-to-many)
- Products have many Quotes
- Vendors have many Quotes
- Quotes automatically convert to USD using Forex rates

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
