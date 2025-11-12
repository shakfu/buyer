# Configuration Guide

This document provides a comprehensive guide to configuring the Buyer application, including the configuration loading sequence, defaults, and behavior when values are not found.

---

## Table of Contents

1. [Configuration Loading Sequence](#configuration-loading-sequence)
2. [Configuration Sources](#configuration-sources)
3. [Environment Variables Reference](#environment-variables-reference)
4. [Default Values and Behavior](#default-values-and-behavior)
5. [Security Configuration](#security-configuration)
6. [Configuration Examples](#configuration-examples)
7. [Troubleshooting](#troubleshooting)

---

## Configuration Loading Sequence

The application loads configuration in the following order (later sources override earlier ones):

```
1. Application defaults (hardcoded in code)
   ↓
2. .env file (if present)
   ↓
3. Environment variables (highest priority)
   ↓
4. Command-line flags (for specific options like --port)
```

### Detailed Loading Process

When the application starts (`cmd/buyer/main.go`):

```go
func main() {
    // STEP 1: Load .env file if it exists (silently ignore if not present)
    _ = godotenv.Load()

    // STEP 2: Parse command-line arguments (cobra/viper)
    rootCmd.Execute()

    // STEP 3: Initialize configuration (reads env vars)
    initConfig()  // Called in PersistentPreRun
}
```

**What happens:**

1. **`.env` file check**:
   - Looks for `.env` in current working directory
   - If found: Loads all `KEY=value` pairs into environment
   - If not found: Silently continues (no error)
   - Already-set environment variables are NOT overwritten

2. **Environment variable resolution**:
   - Reads `BUYER_*` environment variables using `os.Getenv()`
   - Uses defaults if variable not set
   - Validates values (especially for security settings)

3. **Configuration validation**:
   - Checks required values (e.g., username/password when auth enabled)
   - Validates password strength if authentication enabled
   - Creates necessary directories (e.g., `~/.buyer/`)

---

## Configuration Sources

### 1. .env File

**Location**: Current working directory (`.env`)

**Format**:
```bash
# Comments start with #
BUYER_ENV=production
BUYER_WEB_PORT=8080
BUYER_ENABLE_AUTH=true
BUYER_USERNAME=admin
BUYER_PASSWORD=SecureP@ssw0rd123!
```

**Behavior**:
- ✅ **Optional** - Application works without it
- ✅ **Gitignored** - Safe for local development (`.env` in `.gitignore`)
- ✅ **Overridable** - Environment variables take precedence
- ✅ **Silent failure** - No error if file missing

**Creating .env file**:
```bash
# Copy example and customize
cp .env.example .env
nano .env
```

### 2. Environment Variables

**Setting environment variables**:

```bash
# Temporary (current shell session)
export BUYER_WEB_PORT=3000
buyer web

# Inline (single command)
BUYER_WEB_PORT=3000 buyer web

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export BUYER_WEB_PORT=3000' >> ~/.bashrc
```

**Priority**: Highest (overrides `.env` file and defaults)

### 3. Command-Line Flags

Some settings have command-line flag alternatives:

```bash
buyer web --port 3000           # Override BUYER_WEB_PORT
buyer --verbose list brands     # Enable verbose logging
```

**Priority**: Flags override environment variables for specific options

---

## Environment Variables Reference

### General Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUYER_ENV` | string | `development` | Environment mode: `development`, `production`, or `testing` |
| `BUYER_DB_PATH` | string | `~/.buyer/buyer.db` | Path to SQLite database file (`:memory:` in testing mode) |
| `BUYER_WEB_PORT` | integer | `8080` | Web server listening port |

### Security Configuration

| Variable | Type | Default | Required When | Description |
|----------|------|---------|---------------|-------------|
| `BUYER_ENABLE_AUTH` | boolean | `false` | - | Enable HTTP basic authentication |
| `BUYER_USERNAME` | string | *none* | Auth enabled | Username for authentication (**no default**) |
| `BUYER_PASSWORD` | string | *none* | Auth enabled | Password for authentication (**no default**) |
| `BUYER_ENABLE_CSRF` | boolean | `false` | - | Enable CSRF protection |

**Important**: When `BUYER_ENABLE_AUTH=true`, both `BUYER_USERNAME` and `BUYER_PASSWORD` are **required**. The application will exit with an error if either is missing.

---

## Default Values and Behavior

### What Happens When Values Are Not Found

#### 1. BUYER_ENV

**Not set or .env missing:**
```go
// internal/config/config.go
func GetEnv() Environment {
    env := os.Getenv("BUYER_ENV")
    switch env {
    case "production", "prod":
        return Production
    case "testing", "test":
        return Testing
    default:
        return Development  // ← DEFAULT
    }
}
```

**Result**: Application runs in `development` mode
- Database: `~/.buyer/buyer.db`
- Logging: Text format with source location
- Log level: INFO (DEBUG if `--verbose`)

#### 2. BUYER_DB_PATH

**Not set or .env missing:**
```go
// Default path: ~/.buyer/buyer.db
homeDir, err := os.UserHomeDir()
buyerDir := filepath.Join(homeDir, ".buyer")
os.MkdirAll(buyerDir, 0755)  // Creates directory if needed
config.DatabasePath = filepath.Join(buyerDir, "buyer.db")
```

**Result**:
- Uses `~/.buyer/buyer.db`
- Creates `~/.buyer/` directory automatically
- On first run, creates empty database and runs migrations

**Special case - Testing mode:**
```go
if env == Testing {
    config.DatabasePath = ":memory:"  // In-memory database
}
```

#### 3. BUYER_WEB_PORT

**Not set or .env missing:**
```go
// internal/config/config.go
config.WebPort = getEnvInt("BUYER_WEB_PORT", 8080)  // ← DEFAULT: 8080
```

**Result**: Web server listens on port `8080`

**Command-line flag override:**
```bash
buyer web --port 3000  # Uses 3000, ignores env var
```

#### 4. BUYER_ENABLE_AUTH

**Not set or .env missing:**
```go
// cmd/buyer/web.go
enableAuth := os.Getenv("BUYER_ENABLE_AUTH") == "true"  // ← DEFAULT: false
```

**Result**: Authentication is **disabled**
- No credentials required
- Web interface accessible without login
- Suitable for local development

**Security note**: Always enable authentication in production!

#### 5. BUYER_USERNAME and BUYER_PASSWORD

**Not set when BUYER_ENABLE_AUTH=true:**

```go
username := os.Getenv("BUYER_USERNAME")
if username == "" {
    slog.Error("BUYER_USERNAME is required when BUYER_ENABLE_AUTH=true")
    fmt.Fprintln(os.Stderr, "Error: BUYER_USERNAME environment variable is required...")
    os.Exit(1)  // ← APPLICATION EXITS
}
```

**Result**: Application **exits with error message**

**Error message:**
```
Error: BUYER_USERNAME environment variable is required when authentication is enabled
```

**No defaults** - This is intentional for security. Never had dangerous `admin/admin` defaults.

#### 6. BUYER_ENABLE_CSRF

**Not set or .env missing:**
```go
EnableCSRF: os.Getenv("BUYER_ENABLE_CSRF") == "true",  // ← DEFAULT: false
```

**Result**: CSRF protection is **disabled**
- Suitable for local development
- Should be enabled in production

---

## Security Configuration

### Authentication and Password Requirements

When `BUYER_ENABLE_AUTH=true`, passwords are validated against strict requirements:

#### Password Requirements (Enforced)

```go
// cmd/buyer/web_security.go - ValidatePassword()
```

- ✅ **Minimum length**: 12 characters
- ✅ **Uppercase letter**: At least one (A-Z)
- ✅ **Lowercase letter**: At least one (a-z)
- ✅ **Digit**: At least one (0-9)
- ✅ **Special character**: At least one (punctuation/symbol)

**Valid password examples:**
- `SecureP@ssw0rd123!`
- `MyV3ry$ecurePass`
- `Admin2024#Secure!`

**Invalid passwords:**
```
admin123         ❌ Too short, no uppercase, no special char
Password123      ❌ No special character
Password123!     ✅ Valid
```

**What happens on invalid password:**

Application exits with detailed error message:

```
Error: Invalid password - password must contain at least one uppercase letter
Password requirements:
  - At least 12 characters long
  - Contains at least one uppercase letter
  - Contains at least one lowercase letter
  - Contains at least one digit
  - Contains at least one special character
```

### Password Hashing

Passwords are **never stored in plain text**:

```go
// Password is hashed at startup using bcrypt
passwordHash, err := HashPassword(password)
// Uses bcrypt.GenerateFromPassword() with DefaultCost (10)

// During authentication, bcrypt compares hash
err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
```

**Security features:**
- bcrypt with cost factor 10 (2^10 = 1024 rounds)
- Automatic salt generation
- Timing-attack resistant comparison

### CSRF Token Generation

When `BUYER_ENABLE_CSRF=true`:

```go
// Cryptographically secure token generation
func generateCSRFToken() string {
    b := make([]byte, 32)  // 32 bytes = 256 bits entropy
    rand.Read(b)           // crypto/rand (NOT math/rand)
    return base64.URLEncoding.EncodeToString(b)
}
```

**Security features:**
- Uses `crypto/rand` for cryptographic randomness
- 256 bits of entropy (not predictable)
- Base64 URL-safe encoding
- 1-hour expiration

### Rate Limiting

**Always enabled** (cannot be disabled):

#### General Request Rate Limiting
- **Limit**: 100 requests per minute per IP
- **Applies to**: All endpoints except static files
- **Implementation**: Fiber limiter middleware

#### Authentication Rate Limiting (when auth enabled)
- **Limit**: 5 authentication attempts per minute per IP
- **Applies to**: Authentication challenges only
- **Implementation**: Separate auth-specific limiter
- **Behavior**:
  - Failed attempts count against limit
  - Successful auth doesn't count
  - Returns 429 (Too Many Requests) when exceeded

**Error message when limit exceeded:**
```
Too many authentication attempts. Please try again later.
```

---

## Configuration Examples

### Example 1: Local Development (Default)

**No configuration needed!**

```bash
buyer web
```

**What happens:**
- Environment: `development`
- Database: `~/.buyer/buyer.db`
- Port: `8080`
- Authentication: **disabled**
- CSRF: **disabled**
- Logging: Text format, INFO level

**Access**: http://localhost:8080 (no login required)

---

### Example 2: Local Development with Custom Port

**Using .env file:**

```bash
# .env
BUYER_WEB_PORT=3000
```

```bash
buyer web
```

**Or using environment variable:**

```bash
BUYER_WEB_PORT=3000 buyer web
```

**Or using command-line flag:**

```bash
buyer web --port 3000
```

**Access**: http://localhost:3000

---

### Example 3: Production with Security Enabled

**Using .env file:**

```bash
# .env
BUYER_ENV=production
BUYER_DB_PATH=/var/lib/buyer/production.db
BUYER_WEB_PORT=8080
BUYER_ENABLE_AUTH=true
BUYER_USERNAME=admin
BUYER_PASSWORD=SecureP@ssw0rd2024!
BUYER_ENABLE_CSRF=true
```

```bash
buyer web
```

**What happens:**
- Environment: `production`
- Database: `/var/lib/buyer/production.db`
- Port: `8080`
- Authentication: **enabled** (HTTP Basic Auth)
- CSRF: **enabled**
- Logging: JSON format, WARN level
- Password: Hashed with bcrypt
- Rate limiting: 5 auth attempts/min

**Access**: http://localhost:8080 (requires username/password)

---

### Example 4: Testing Environment

```bash
BUYER_ENV=testing go test ./...
```

**What happens:**
- Environment: `testing`
- Database: `:memory:` (in-memory, temporary)
- Logging: ERROR level only
- Each test gets fresh database
- No file I/O for database

---

### Example 5: Custom Database Location

```bash
# Use custom database path
BUYER_DB_PATH=/tmp/buyer-test.db buyer web
```

**Use case**:
- Testing with different databases
- Network-attached storage
- Backup/restore scenarios

---

### Example 6: Verbose Mode for Debugging

```bash
buyer --verbose web
```

**What happens:**
- Shows SQL queries in output
- Detailed logging
- Useful for debugging database issues

**Example output:**
```
2024-01-15T10:30:45.123 INFO initializing buyer application environment=development verbose=true
2024-01-15T10:30:45.125 DEBUG database configured path=/Users/user/.buyer/buyer.db

[2024-01-15 10:30:45] [3.12ms] [rows:5] SELECT * FROM `brands` ORDER BY `id`
```

---

### Example 7: Override .env with Environment Variables

```bash
# .env file says:
# BUYER_WEB_PORT=8080

# But you want to temporarily use different port:
BUYER_WEB_PORT=9000 buyer web
```

**Result**: Uses port `9000` (environment variable overrides `.env`)

**Why this is useful:**
- Temporary overrides without editing `.env`
- Different settings per developer
- CI/CD pipeline configurations

---

## Troubleshooting

### Problem: "BUYER_USERNAME is required when BUYER_ENABLE_AUTH=true"

**Cause**: Authentication is enabled but credentials not provided.

**Solution**:

```bash
# Option 1: Disable authentication
BUYER_ENABLE_AUTH=false buyer web

# Option 2: Provide credentials
export BUYER_ENABLE_AUTH=true
export BUYER_USERNAME=admin
export BUYER_PASSWORD="SecureP@ss123!"
buyer web

# Option 3: Use .env file
cat > .env << EOF
BUYER_ENABLE_AUTH=true
BUYER_USERNAME=admin
BUYER_PASSWORD=SecureP@ss123!
EOF
buyer web
```

---

### Problem: "Invalid password - password must contain..."

**Cause**: Password doesn't meet security requirements.

**Solution**: Use a password that meets all requirements:

```bash
# Bad passwords:
BUYER_PASSWORD=admin123          # Too short, no uppercase, no special
BUYER_PASSWORD=Password123       # No special character

# Good passwords:
BUYER_PASSWORD="SecureP@ss123!"
BUYER_PASSWORD="MyV3ry$ecurePass"
BUYER_PASSWORD="Admin2024#Secure!"
```

**Requirements**:
- Minimum 12 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one digit (0-9)
- At least one special character (!@#$%^&*)

---

### Problem: Can't connect to database

**Possible causes:**

1. **Permission denied**
   ```
   Error: failed to connect to database: unable to open database file
   ```
   **Solution**: Check directory permissions
   ```bash
   mkdir -p ~/.buyer
   chmod 755 ~/.buyer
   ```

2. **Custom path doesn't exist**
   ```bash
   BUYER_DB_PATH=/nonexistent/path/db.sqlite buyer web
   ```
   **Solution**: Ensure parent directory exists
   ```bash
   mkdir -p /path/to/directory
   BUYER_DB_PATH=/path/to/directory/buyer.db buyer web
   ```

3. **Database locked** (another instance running)
   ```
   Error: database is locked
   ```
   **Solution**: Stop other buyer processes
   ```bash
   pkill -f buyer
   ```

---

### Problem: Port already in use

```
Error: failed to start server: listen tcp :8080: bind: address already in use
```

**Solution**: Use different port

```bash
# Check what's using the port
lsof -i :8080

# Use different port
buyer web --port 3000
# or
BUYER_WEB_PORT=3000 buyer web
```

---

### Problem: .env file not being loaded

**Symptoms**: Settings in `.env` file are ignored

**Causes and solutions:**

1. **Running from different directory**
   ```bash
   # .env must be in current directory
   pwd  # Check where you are
   ls .env  # Verify .env exists here
   ```

2. **Syntax error in .env file**
   ```bash
   # Bad:
   BUYER_WEB_PORT = 8080  # Spaces around =

   # Good:
   BUYER_WEB_PORT=8080    # No spaces
   ```

3. **Environment variable already set**
   ```bash
   # Check if variable is set
   echo $BUYER_WEB_PORT

   # Unset it to use .env value
   unset BUYER_WEB_PORT
   ```

---

### Problem: "Too many authentication attempts"

```
Error: Too many authentication attempts. Please try again later.
```

**Cause**: Exceeded 5 authentication attempts per minute (rate limiting)

**Solution**: Wait 1 minute and try again, or check credentials:

```bash
# Wait 60 seconds
sleep 60

# Verify credentials
echo $BUYER_USERNAME
echo $BUYER_PASSWORD  # (or check .env file)
```

---

## Configuration Precedence Summary

When the same setting is defined in multiple places:

```
Command-line flags  (--port 3000)
    ↓ overrides
Environment variables  (export BUYER_WEB_PORT=8080)
    ↓ overrides
.env file  (BUYER_WEB_PORT=7000)
    ↓ overrides
Application defaults  (8080)
```

**Example:**

```bash
# .env file
BUYER_WEB_PORT=7000

# Environment variable
export BUYER_WEB_PORT=8080

# Command-line flag
buyer web --port 3000
```

**Result**: Uses port **3000** (flag wins)

---

## Environment-Specific Defaults

### Development Mode (`BUYER_ENV=development` or not set)

```
Database:    ~/.buyer/buyer.db
Log format:  Text with source location
Log level:   INFO (DEBUG if --verbose)
Auth:        Disabled by default
CSRF:        Disabled by default
```

### Production Mode (`BUYER_ENV=production`)

```
Database:    ~/.buyer/buyer.db (or custom via BUYER_DB_PATH)
Log format:  JSON (structured logging)
Log level:   WARN (INFO if --verbose)
Auth:        Should be enabled (BUYER_ENABLE_AUTH=true)
CSRF:        Should be enabled (BUYER_ENABLE_CSRF=true)
```

### Testing Mode (`BUYER_ENV=testing`)

```
Database:    :memory: (in-memory, temporary)
Log format:  Text
Log level:   ERROR only
Auth:        Disabled by default
CSRF:        Disabled by default
```

---

## Security Best Practices

### Development
- ✅ Authentication disabled (default)
- ✅ Use `.env` file for local settings
- ✅ Add `.env` to `.gitignore` (already done)
- ✅ Never commit credentials

### Production
- ✅ Enable authentication: `BUYER_ENABLE_AUTH=true`
- ✅ Enable CSRF: `BUYER_ENABLE_CSRF=true`
- ✅ Use strong password (12+ chars, complexity)
- ✅ Use environment variables, not `.env` file
- ✅ Set `BUYER_ENV=production`
- ✅ Restrict file permissions: `chmod 600 .env`
- ✅ Use HTTPS reverse proxy (nginx, Caddy)
- ✅ Regular database backups

### Multi-User Environments
- ⚠️ Current limitation: Single user only
- ⚠️ Future: Multi-user RBAC support needed
- 💡 Workaround: Use reverse proxy with auth (nginx, Caddy)

---

## Quick Reference Card

```bash
# Minimal development setup
buyer web

# Production with security
BUYER_ENV=production \
BUYER_ENABLE_AUTH=true \
BUYER_USERNAME=admin \
BUYER_PASSWORD="SecureP@ss123!" \
BUYER_ENABLE_CSRF=true \
buyer web

# Custom database and port
BUYER_DB_PATH=/custom/path/db.sqlite \
BUYER_WEB_PORT=3000 \
buyer web

# Verbose mode for debugging
buyer --verbose web

# Using .env file
cp .env.example .env
# Edit .env with your settings
buyer web
```

---

## Additional Resources

- **CLAUDE.md**: Architecture and development guide
- **CODE_REVIEW.md**: Security audit and best practices
- **README.md**: Quick start and feature overview
- **.env.example**: Template configuration file

---

## Change Log

### 2024-01 - Security Hardening
- ✅ Removed default credentials (admin/admin)
- ✅ Added password strength validation
- ✅ Implemented bcrypt password hashing
- ✅ Added cryptographically secure CSRF tokens
- ✅ Added authentication-specific rate limiting
- ✅ Added graceful shutdown on signals

### 2024-01 - Configuration Improvements
- ✅ Added godotenv support for `.env` files
- ✅ Documented complete configuration sequence
- ✅ Added comprehensive error messages
- ✅ Environment variable precedence documented
