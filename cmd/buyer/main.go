package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
	// Version is set via build flags: go build -ldflags "-X main.Version=v1.0.0"
	Version = "dev"
)

func initConfig() {
	// Initialize configuration
	env := config.GetEnv()
	logger := config.SetupLogger(env, verbose)

	logger.Info("initializing buyer application",
		slog.String("environment", string(env)),
		slog.Bool("verbose", verbose))

	var err error
	cfg, err = config.NewConfig(env, verbose)
	if err != nil {
		logger.Error("failed to initialize config", slog.String("error", err.Error()))
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	logger.Debug("database configured",
		slog.String("path", cfg.DatabasePath))

	// Run migrations
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Specification{},
		&models.Product{},
		&models.Requisition{},
		&models.RequisitionItem{},
		&models.Quote{},
		&models.Forex{},
		&models.Project{},
		&models.BillOfMaterials{},
		&models.BillOfMaterialsItem{},
		&models.ProjectRequisition{},
		&models.ProjectRequisitionItem{},
	); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	logger.Info("database migrations completed successfully")
}

func main() {
	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	if cfg != nil {
		cfg.Close()
	}
}

var rootCmd = &cobra.Command{
	Use:   "buyer",
	Short: "A purchasing support and vendor quote management tool",
	Long: `buyer is a CLI tool for tracking brands, products, vendors, and price quotes
across multiple vendors with multi-currency support.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version number of buyer",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip config initialization for version command
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("buyer version %s\n", Version)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (SQL queries)")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(versionCmd)
}
