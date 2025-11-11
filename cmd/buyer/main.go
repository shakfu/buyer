package main

import (
	"fmt"
	"os"

	"github.com/shakfu/buyer/internal/config"
	"github.com/shakfu/buyer/internal/models"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
)

func initConfig() {
	// Initialize configuration
	env := config.GetEnv()
	config.SetupLogger(env)

	var err error
	cfg, err = config.NewConfig(env, verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := cfg.AutoMigrate(
		&models.Vendor{},
		&models.Brand{},
		&models.Product{},
		&models.Quote{},
		&models.Forex{},
	); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}
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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (SQL queries)")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(webCmd)
}
