# Configuration Guide

**Date:** 2025-11-12
**Application:** buyer - Purchasing Support and Vendor Quote Management Tool

---

## Overview

The buyer application supports configuration through environment variables, providing flexibility for different deployment scenarios without code changes. All configuration values have sensible defaults suitable for local development.

## Configuration Methods

Configuration values are loaded in the following priority order (highest to lowest):

1. **Command-line flags** (e.g., `--port 3000`)
2. **Environment variables** (e.g., `BUYER_WEB_PORT=3000`)
3. **Default values** (hardcoded in the application)

## Environment Variables

### Core Configuration

#### `BUYER_ENV`
- **Description:** Application environment mode
- **Valid Values:** `development`, `production`, `testing`
- **Default:** `development`
- **Example:**
  ```bash
  BUYER_ENV=production ./bin/buyer web
  ```

**Behavior by Environment:**
- **development:** Text logs with source locations, INFO level (DEBUG with `-v`)
- **production:** JSON logs for aggregation, WARN level (INFO with `-v`)
- **testing:** Silent logging (ERROR level only), in-memory database

---

### Database Configuration

#### `BUYER_DB_PATH`
- **Description:** Path to SQLite database file
- **Default:** `~/.buyer/buyer.db`
- **Testing Override:** Always uses `:memory:` when `BUYER_ENV=testing`
- **Example:**
  ```bash
  BUYER_DB_PATH=/var/lib/buyer/production.db ./bin/buyer web
  ```

**Use Cases:**
- Custom database location for production deployments
- Shared database across multiple users (with proper permissions)
- Docker volume mounts

**Notes:**
- Directory must exist and be writable
- Parent directories are NOT created automatically (except for default `~/.buyer/`)
- Testing mode always uses in-memory database regardless of this setting

---

### Web Server Configuration

#### `BUYER_WEB_PORT`
- **Description:** Port number for web server
- **Default:** `8080`
- **Valid Range:** `1024-65535` (recommended to use unprivileged ports)
- **Command-line Override:** `--port` or `-p` flag
- **Example:**
  ```bash
  BUYER_WEB_PORT=3000 ./bin/buyer web

  # Or using flag:
  ./bin/buyer web --port 3000
  ```

**Priority:**
1. Command-line flag (`--port`)
2. Environment variable (`BUYER_WEB_PORT`)
3. Default (`8080`)

---

### Security Configuration

#### `BUYER_ENABLE_AUTH`
- **Description:** Enable HTTP Basic Authentication
- **Valid Values:** `true`, `false`
- **Default:** `false`
- **Recommended:** `true` for production
- **Example:**
  ```bash
  BUYER_ENABLE_AUTH=true \
  BUYER_USERNAME=admin \
  BUYER_PASSWORD=secure-password \
  ./bin/buyer web
  ```

#### `BUYER_USERNAME`
- **Description:** Basic auth username (only used if `BUYER_ENABLE_AUTH=true`)
- **Default:** `admin`
- **Example:** `BUYER_USERNAME=buyer-admin`

#### `BUYER_PASSWORD`
- **Description:** Basic auth password (only used if `BUYER_ENABLE_AUTH=true`)
- **Default:** `admin`
- **Security:** **CHANGE THIS IN PRODUCTION!**
- **Example:** `BUYER_PASSWORD=your-secure-password-here`

#### `BUYER_ENABLE_CSRF`
- **Description:** Enable CSRF protection for web forms
- **Valid Values:** `true`, `false`
- **Default:** `false`
- **Recommended:** `true` for production
- **Example:**
  ```bash
  BUYER_ENABLE_CSRF=true ./bin/buyer web
  ```

**Note:** Rate limiting (100 requests/minute per IP) is always enabled and cannot be disabled.

---

## Configuration Examples

### Local Development

```bash
# Default settings work out of the box
./bin/buyer web

# With verbose logging
./bin/buyer -v web
```

### Production Deployment

```bash
# Minimal production setup
export BUYER_ENV=production
export BUYER_ENABLE_AUTH=true
export BUYER_USERNAME=admin
export BUYER_PASSWORD=your-secure-password
export BUYER_ENABLE_CSRF=true
./bin/buyer web

# Full production with custom database and port
export BUYER_ENV=production
export BUYER_DB_PATH=/var/lib/buyer/buyer.db
export BUYER_WEB_PORT=8080
export BUYER_ENABLE_AUTH=true
export BUYER_USERNAME=buyer-admin
export BUYER_PASSWORD=your-secure-password-here
export BUYER_ENABLE_CSRF=true
./bin/buyer -v web
```

### Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/buyer .

# Environment variables
ENV BUYER_ENV=production \
    BUYER_DB_PATH=/data/buyer.db \
    BUYER_WEB_PORT=8080 \
    BUYER_ENABLE_AUTH=true \
    BUYER_ENABLE_CSRF=true

# Create data directory
RUN mkdir -p /data

# Volume for persistent database
VOLUME /data

EXPOSE 8080

CMD ["./buyer", "web"]
```

**Running the container:**
```bash
docker run -d \
  -p 8080:8080 \
  -v buyer-data:/data \
  -e BUYER_USERNAME=admin \
  -e BUYER_PASSWORD=secure-password \
  buyer:latest
```

### Using .env File

Create a `.env` file (see `.env.example`):

```bash
# .env
BUYER_ENV=production
BUYER_DB_PATH=/var/lib/buyer/buyer.db
BUYER_WEB_PORT=8080
BUYER_ENABLE_AUTH=true
BUYER_USERNAME=admin
BUYER_PASSWORD=your-secure-password
BUYER_ENABLE_CSRF=true
```

Load and run:
```bash
# Using export
export $(cat .env | xargs)
./bin/buyer web

# Or with a tool like direnv
direnv allow
./bin/buyer web
```

---

## Configuration File Location

The application does **not** currently support reading from a configuration file (e.g., `config.yaml` or `config.toml`). All configuration is done via:

1. Environment variables
2. Command-line flags
3. Default values

To add config file support, see the "Future Enhancements" section in CODE_REVIEW.md.

---

## Logging Configuration

Logging behavior is controlled by `BUYER_ENV` and the `-v` (verbose) flag:

| Environment | Default Level | Verbose Level | Format | Source Location |
|-------------|---------------|---------------|--------|-----------------|
| development | INFO          | DEBUG         | Text   | Yes             |
| production  | WARN          | INFO          | JSON   | No              |
| testing     | ERROR         | ERROR         | Text   | No              |

**Examples:**
```bash
# Development with DEBUG logs
./bin/buyer -v list brands

# Production with INFO logs
BUYER_ENV=production ./bin/buyer -v web

# Testing (silent)
BUYER_ENV=testing ./bin/buyer list brands
```

See [STRUCTURED_LOGGING.md](STRUCTURED_LOGGING.md) for detailed logging documentation.

---

## Security Best Practices

### Production Checklist

- [ ] Set `BUYER_ENV=production`
- [ ] Enable authentication: `BUYER_ENABLE_AUTH=true`
- [ ] Use strong password: `BUYER_PASSWORD=...` (not "admin"!)
- [ ] Enable CSRF: `BUYER_ENABLE_CSRF=true`
- [ ] Use custom database path: `BUYER_DB_PATH=/secure/path/buyer.db`
- [ ] Set proper file permissions on database: `chmod 600 /path/to/buyer.db`
- [ ] Run behind reverse proxy (nginx, caddy) with HTTPS
- [ ] Monitor logs for suspicious activity

### Database Security

```bash
# Set restrictive permissions on database file
chmod 600 /var/lib/buyer/buyer.db
chown buyer:buyer /var/lib/buyer/buyer.db

# Ensure directory is not world-readable
chmod 700 /var/lib/buyer
```

### Secrets Management

**DO NOT** commit `.env` files with passwords to version control!

```bash
# Add to .gitignore
echo ".env" >> .gitignore
```

For production, consider:
- **Docker Secrets** (Swarm mode)
- **Kubernetes Secrets**
- **AWS Secrets Manager / Parameter Store**
- **HashiCorp Vault**
- **Azure Key Vault**

---

## Troubleshooting

### Database Permission Errors

**Error:** `failed to create .buyer directory: permission denied`

**Solution:** Ensure the user running buyer has write permissions:
```bash
mkdir -p ~/.buyer
chmod 700 ~/.buyer
```

Or use a custom path:
```bash
BUYER_DB_PATH=/tmp/buyer.db ./bin/buyer web
```

### Port Already in Use

**Error:** `Failed to start server: listen tcp :8080: bind: address already in use`

**Solution:** Change the port:
```bash
BUYER_WEB_PORT=3000 ./bin/buyer web
# Or
./bin/buyer web --port 3000
```

### Authentication Not Working

**Issue:** Still able to access web interface without authentication

**Check:**
1. Ensure `BUYER_ENABLE_AUTH=true` (not just set to empty string)
2. Restart the web server after changing environment variables
3. Clear browser cache/cookies
4. Check logs for security middleware configuration:
   ```
   level=INFO msg="security middleware configured" auth_enabled=true csrf_enabled=true
   ```

---

## Migration from Hardcoded Configuration

**Before (hardcoded):**
```go
// web.go
addr := ":8080"  // Hardcoded port
```

**After (configurable):**
```bash
# Use environment variable
BUYER_WEB_PORT=3000 ./bin/buyer web

# Or command-line flag
./bin/buyer web --port 3000

# Or default
./bin/buyer web  # Uses 8080
```

**Database path migration:**
```bash
# Old: always ~/.buyer/buyer.db
# New: configurable
BUYER_DB_PATH=/var/lib/buyer/production.db ./bin/buyer web
```

---

## Environment Variable Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUYER_ENV` | string | `development` | Environment mode (development/production/testing) |
| `BUYER_DB_PATH` | string | `~/.buyer/buyer.db` | SQLite database file path |
| `BUYER_WEB_PORT` | int | `8080` | Web server port number |
| `BUYER_ENABLE_AUTH` | bool | `false` | Enable HTTP basic authentication |
| `BUYER_USERNAME` | string | `admin` | Basic auth username |
| `BUYER_PASSWORD` | string | `admin` | Basic auth password |
| `BUYER_ENABLE_CSRF` | bool | `false` | Enable CSRF protection |

---

## See Also

- [STRUCTURED_LOGGING.md](STRUCTURED_LOGGING.md) - Logging configuration and usage
- [CODE_REVIEW.md](CODE_REVIEW.md) - Code quality and recommendations
- [docs/security_fixes.md](docs/security_fixes.md) - Security implementation details
- [README.md](README.md) - General usage documentation
