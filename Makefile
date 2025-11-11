.PHONY: build test clean install run web coverage lint

# Build the binary
build:
	@go build -o bin/buyer ./cmd/buyer

# Install the binary to $GOPATH/bin
install:
	@go install ./cmd/buyer

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

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install the binary to GOPATH/bin"
	@echo "  run        - Run the CLI"
	@echo "  web        - Start the web server"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage report"
	@echo "  test-race  - Run tests with race detection"
	@echo "  lint       - Lint the code"
	@echo "  fmt        - Format code"
	@echo "  tidy       - Tidy dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  check      - Run fmt, lint, and test"
