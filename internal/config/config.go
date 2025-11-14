package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// Config holds application configuration
type Config struct {
	Environment  Environment
	DatabasePath string
	DatabaseURL  string // PostgreSQL connection string
	WebPort      int
	LogLevel     logger.LogLevel
	DB           *gorm.DB
}

// NewConfig creates a new configuration based on environment
func NewConfig(env Environment, verbose bool) (*Config, error) {
	config := &Config{
		Environment: env,
	}

	// Set web port from environment variable or default
	config.WebPort = getEnvInt("BUYER_WEB_PORT", 8080)

	// Set database path/URL based on environment
	switch env {
	case Testing:
		// Use SQLite in-memory for testing
		config.DatabasePath = ":memory:"
		config.LogLevel = logger.Silent
	case Development, Production:
		// Check if PostgreSQL connection string is provided
		if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
			config.DatabaseURL = dbURL
		} else if dbHost := os.Getenv("BUYER_DB_HOST"); dbHost != "" {
			// Build PostgreSQL connection string from individual env vars
			dbPort := getEnvString("BUYER_DB_PORT", "5432")
			dbName := getEnvString("BUYER_DB_NAME", "buyer")
			dbUser := getEnvString("BUYER_DB_USER", "buyer")
			dbPassword := getEnvString("BUYER_DB_PASSWORD", "")
			dbSSLMode := getEnvString("BUYER_DB_SSLMODE", "disable")

			config.DatabaseURL = fmt.Sprintf(
				"host=%s port=%s user=%s dbname=%s sslmode=%s",
				dbHost, dbPort, dbUser, dbName, dbSSLMode,
			)
			if dbPassword != "" {
				config.DatabaseURL = fmt.Sprintf(
					"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
					dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
				)
			}
		} else if dbPath := os.Getenv("BUYER_DB_PATH"); dbPath != "" {
			// Fallback to SQLite if DB_PATH is explicitly set
			config.DatabasePath = dbPath
		} else {
			// Default: use SQLite in ~/.buyer/buyer.db for local development
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			buyerDir := filepath.Join(homeDir, ".buyer")
			if err := os.MkdirAll(buyerDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create .buyer directory: %w", err)
			}
			config.DatabasePath = filepath.Join(buyerDir, "buyer.db")
		}

		// Set log level
		if verbose {
			config.LogLevel = logger.Info
		} else {
			config.LogLevel = logger.Silent
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	// Initialize database connection
	if err := config.InitDB(); err != nil {
		return nil, err
	}

	return config, nil
}

// InitDB initializes the database connection
func (c *Config) InitDB() error {
	var db *gorm.DB
	var err error

	if c.DatabaseURL != "" {
		// Use PostgreSQL
		db, err = gorm.Open(postgres.Open(c.DatabaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(c.LogLevel),
		})
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	} else if c.DatabasePath != "" {
		// Use SQLite
		db, err = gorm.Open(sqlite.Open(c.DatabasePath), &gorm.Config{
			Logger: logger.Default.LogMode(c.LogLevel),
		})
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite: %w", err)
		}

		// Enable foreign key constraints for SQLite
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return fmt.Errorf("failed to enable foreign key constraints: %w", err)
		}
	} else {
		return fmt.Errorf("no database configuration provided")
	}

	c.DB = db
	return nil
}

// Close closes the database connection
func (c *Config) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}

// AutoMigrate runs database migrations
func (c *Config) AutoMigrate(models ...interface{}) error {
	return c.DB.AutoMigrate(models...)
}

// GetEnv returns the current environment from environment variable or defaults to development
func GetEnv() Environment {
	env := os.Getenv("BUYER_ENV")
	switch env {
	case "production", "prod":
		return Production
	case "testing", "test":
		return Testing
	default:
		return Development
	}
}

// getEnvInt returns an integer from an environment variable or a default value
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvString returns a string from an environment variable or a default value
func getEnvString(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// SetupLogger configures structured logging with slog
func SetupLogger(env Environment, verbose bool) *slog.Logger {
	var handler slog.Handler
	var level slog.Level

	// Set log level based on environment and verbose flag
	switch env {
	case Production:
		if verbose {
			level = slog.LevelInfo
		} else {
			level = slog.LevelWarn
		}
		// Production: JSON format for structured logging
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	case Development:
		if verbose {
			level = slog.LevelDebug
		} else {
			level = slog.LevelInfo
		}
		// Development: Text format with source location for debugging
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	case Testing:
		// Testing: Silent (only errors and above)
		level = slog.LevelError
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	default:
		level = slog.LevelInfo
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
