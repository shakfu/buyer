package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

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

	// Set database path based on environment
	switch env {
	case Testing:
		config.DatabasePath = ":memory:"
		config.LogLevel = logger.Silent
	case Development, Production:
		// Check for custom database path from environment
		if dbPath := os.Getenv("BUYER_DB_PATH"); dbPath != "" {
			config.DatabasePath = dbPath
		} else {
			// Default path: ~/.buyer/buyer.db
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
	db, err := gorm.Open(sqlite.Open(c.DatabasePath), &gorm.Config{
		Logger: logger.Default.LogMode(c.LogLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign key constraints for SQLite
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return fmt.Errorf("failed to enable foreign key constraints: %w", err)
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
