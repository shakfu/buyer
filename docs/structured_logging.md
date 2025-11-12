# Structured Logging Implementation

**Date:** 2025-11-12
**Package:** Go stdlib `log/slog` (Go 1.21+)

## Overview

The buyer application now uses Go's built-in `slog` package for structured logging, providing JSON logs for production and human-readable text logs for development.

## Features

### Environment-Based Configuration

**Development Mode:**
- Text format with source file locations
- INFO level by default, DEBUG with `-v` flag
- Includes source code references for debugging
- Color-coded output from GORM when verbose

**Production Mode:**
- JSON format for log aggregation systems
- WARN level by default, INFO with `-v` flag
- Machine-readable structured logs
- Compatible with ELK, Splunk, CloudWatch, etc.

**Testing Mode:**
- Silent logging (ERROR level only)
- Minimal output during test runs

### Log Levels

- `DEBUG`: Detailed information for debugging (development verbose mode)
- `INFO`: General informational messages (default in development)
- `WARN`: Warning messages (default in production)
- `ERROR`: Error conditions

## Usage Examples

### Development (Text Logging)

```bash
# Normal mode (INFO level)
./bin/buyer list brands
# Output:
# time=2025-11-12T02:38:54.496+03:00 level=INFO source=/path/to/main.go:23 msg="initializing buyer application" environment=development verbose=false

# Verbose mode (DEBUG level + SQL queries)
./bin/buyer -v add brand "Apple"
# Output:
# time=2025-11-12T02:39:53.498+03:00 level=INFO source=/path/to/main.go:23 msg="initializing buyer application" environment=development verbose=true
# time=2025-11-12T02:39:53.498+03:00 level=DEBUG source=/path/to/main.go:35 msg="database configured" path=/Users/sa/.buyer/buyer.db
# ... SQL queries ...
```

### Production (JSON Logging)

```bash
# Normal mode (WARN level)
BUYER_ENV=production ./bin/buyer list brands
# Output:
# {"time":"2025-11-12T02:40:14.049276+03:00","level":"WARN","msg":"..."}

# Verbose mode (INFO level)
BUYER_ENV=production ./bin/buyer -v list brands
# Output:
# {"time":"2025-11-12T02:40:14.049276+03:00","level":"INFO","msg":"initializing buyer application","environment":"production","verbose":true}
```

## Implementation Details

### Configuration (`internal/config/config.go`)

```go
func SetupLogger(env Environment, verbose bool) *slog.Logger {
    var handler slog.Handler
    var level slog.Level

    switch env {
    case Production:
        level = slog.LevelWarn
        if verbose {
            level = slog.LevelInfo
        }
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: level,
        })
    case Development:
        level = slog.LevelInfo
        if verbose {
            level = slog.LevelDebug
        }
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level:     level,
            AddSource: true,
        })
    case Testing:
        level = slog.LevelError
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: level,
        })
    }

    logger := slog.New(handler)
    slog.SetDefault(logger)
    return logger
}
```

### Application Startup (`cmd/buyer/main.go`)

```go
func initConfig() {
    env := config.GetEnv()
    logger := config.SetupLogger(env, verbose)

    logger.Info("initializing buyer application",
        slog.String("environment", string(env)),
        slog.Bool("verbose", verbose))

    // ... configuration and migrations ...

    logger.Info("database migrations completed successfully")
}
```

### Web Server (`cmd/buyer/web.go`)

```go
slog.Info("starting web server",
    slog.String("address", addr),
    slog.String("url", fmt.Sprintf("http://localhost%s", addr)))

slog.Info("security middleware configured",
    slog.Bool("auth_enabled", securityConfig.EnableAuth),
    slog.Bool("csrf_enabled", securityConfig.EnableCSRF),
    slog.Bool("rate_limiter_enabled", securityConfig.EnableRateLimiter))
```

### CLI Commands (`cmd/buyer/add.go`)

```go
svc := services.NewBrandService(cfg.DB)
brand, err := svc.Create(args[0])
if err != nil {
    slog.Error("failed to create brand",
        slog.String("name", args[0]),
        slog.String("error", err.Error()))
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
slog.Info("brand created successfully",
    slog.String("name", brand.Name),
    slog.Uint64("id", uint64(brand.ID)))
```

## Log Aggregation

### Production JSON Output

The JSON format is optimized for log aggregation:

```json
{
  "time": "2025-11-12T02:40:14.049276+03:00",
  "level": "INFO",
  "msg": "brand created successfully",
  "name": "Apple",
  "id": 123
}
```

### Integration with Log Systems

**Elasticsearch/ELK:**
- Parse JSON logs with Logstash
- Index by timestamp, level, environment
- Search by structured fields

**CloudWatch:**
- Stream logs to CloudWatch Logs
- Query with CloudWatch Insights
- Filter by structured attributes

**Splunk:**
- Forward JSON logs to Splunk
- Use structured field extraction
- Create dashboards from log data

## Best Practices

1. **Use Structured Fields:**
   ```go
   // Good
   slog.Info("user action", slog.String("action", "login"), slog.String("user_id", userID))

   // Bad
   slog.Info(fmt.Sprintf("User %s performed login", userID))
   ```

2. **Include Context:**
   ```go
   slog.Error("database query failed",
       slog.String("query", query),
       slog.String("table", tableName),
       slog.String("error", err.Error()))
   ```

3. **Use Appropriate Levels:**
   - DEBUG: Internal state, variable values
   - INFO: Key application events
   - WARN: Recoverable issues
   - ERROR: Failures requiring attention

4. **Consistent Field Names:**
   - Use `error` for error messages
   - Use `id` for entity IDs
   - Use `name` for entity names
   - Use consistent naming across codebase

## Benefits

[x] **Zero External Dependencies:** Uses Go stdlib
[x] **Production Ready:** JSON logs for aggregation
[x] **Developer Friendly:** Text logs with source locations
[x] **Performance:** Minimal overhead
[x] **Structured:** Machine-readable fields
[x] **Contextual:** Rich debugging information
[x] **Configurable:** Environment-based settings

## Migration from Old Logging

**Before (standard log package):**
```go
log.Printf("Starting web server on http://localhost%s\n", addr)
```

**After (structured slog):**
```go
slog.Info("starting web server",
    slog.String("address", addr),
    slog.String("url", fmt.Sprintf("http://localhost%s", addr)))
```

## Testing

Tests run in silent mode (ERROR level only) to keep output clean:

```bash
make test
# No log output during tests unless errors occur
```

## Future Enhancements (Optional)

- Add request ID tracking for distributed tracing
- Log sampling for high-volume endpoints
- Custom log handlers for specific integrations
- Metrics extraction from structured logs
