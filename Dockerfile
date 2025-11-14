# Multi-stage Dockerfile for buyer application
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
ARG VERSION=dev
RUN go build -ldflags "-X main.Version=${VERSION} -s -w" -o /app/bin/buyer ./cmd/buyer

# Runtime stage
FROM alpine:latest

# Install runtime dependencies (PostgreSQL client for connectivity)
RUN apk --no-cache add ca-certificates postgresql-client

# Create non-root user
RUN addgroup -g 1000 buyer && \
    adduser -D -u 1000 -G buyer buyer

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/buyer .

# Change ownership
RUN chown buyer:buyer /app/buyer

# Switch to non-root user
USER buyer

# Environment variables
ENV BUYER_ENV=production \
    BUYER_WEB_PORT=8080 \
    BUYER_DB_HOST=postgres \
    BUYER_DB_PORT=5432 \
    BUYER_DB_NAME=buyer \
    BUYER_DB_USER=buyer \
    BUYER_DB_SSLMODE=disable

# Expose web server port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./buyer", "web"]
