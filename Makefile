.PHONY: build test clean install run web coverage coverage-ci lint snap fixtures reset-db version

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.Version=$(VERSION)

# Build the binary with version info
build:
	@echo "Building buyer $(VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o bin/buyer ./cmd/buyer

# Install the binary to $GOPATH/bin with version info
install:
	@echo "Installing buyer $(VERSION)..."
	@go install -ldflags "$(LDFLAGS)" ./cmd/buyer

# Run the CLI
run:
	@go run ./cmd/buyer

# Start the web server
web:
	@go run ./cmd/buyer web

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with coverage (CI-friendly, no HTML generation)
coverage-ci:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run tests with race detection
test-race:
	@go test -v -race ./...

# Lint the code (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run

# Format code
fmt:
	@go fmt ./...

# Tidy dependencies
tidy:
	@go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run all checks (fmt, lint, test)
check: fmt lint test

# Take git snapshot
snap:
	@git add --all . && git commit -m 'snap' && git push

# Reset database and load fixtures
reset-db:
	@echo "Resetting database..."
	@rm -f ~/.buyer/buyer.db
	@echo "Building application to trigger migrations..."
	@go run ./cmd/buyer list brands > /dev/null 2>&1 || true
	@echo "Loading fixtures..."
	@sqlite3 ~/.buyer/buyer.db < fixtures.sql
	@echo "Database reset complete with fixtures loaded!"

# Load fixtures into existing database (without reset)
fixtures:
	@echo "Loading fixtures..."
	@sqlite3 ~/.buyer/buyer.db < fixtures.sql
	@echo "Fixtures loaded!"

# Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All platform builds complete!"

# Build for Linux (AMD64 and ARM64)
build-linux:
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/buyer-linux-amd64 ./cmd/buyer
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/buyer-linux-arm64 ./cmd/buyer

# Build for macOS (AMD64 and ARM64)
build-darwin:
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/buyer-darwin-amd64 ./cmd/buyer
	@echo "Building for macOS ARM64 (Apple Silicon)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/buyer-darwin-arm64 ./cmd/buyer

# Build for Windows (AMD64)
build-windows:
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/buyer-windows-amd64.exe ./cmd/buyer

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build --build-arg VERSION=$(VERSION) -t buyer:$(VERSION) -t buyer:latest .

# Run Docker container
docker-run:
	@docker-compose up -d

# Stop Docker container
docker-stop:
	@docker-compose down

# Print version
version:
	@echo "Version: $(VERSION)"

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary for current platform"
	@echo "  build-all    - Build for all platforms (Linux, macOS, Windows)"
	@echo "  build-linux  - Build for Linux (AMD64, ARM64)"
	@echo "  build-darwin - Build for macOS (AMD64, ARM64)"
	@echo "  build-windows- Build for Windows (AMD64)"
	@echo "  install      - Install the binary to GOPATH/bin"
	@echo "  run          - Run the CLI"
	@echo "  web          - Start the web server"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  test-race    - Run tests with race detection"
	@echo "  lint         - Lint the code"
	@echo "  fmt          - Format code"
	@echo "  tidy         - Tidy dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  check        - Run fmt, lint, and test"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container with docker-compose"
	@echo "  docker-stop  - Stop Docker container"
	@echo "  version      - Print version information"
	@echo "  snap         - Create and push a git snapshot"
	@echo "  reset-db     - Reset database and load fixtures"
	@echo "  fixtures     - Load fixtures into existing database"
