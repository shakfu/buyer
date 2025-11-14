package main

import (
	"fmt"
	"os"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from CSV files",
	Long:  `Import brands, vendors, or forex rates from CSV files.`,
}

var importBrandsCmd = &cobra.Command{
	Use:   "brands [filename]",
	Short: "Import brands from CSV",
	Long: `Import brands from a CSV file.

CSV Format:
ID,Name,CreatedAt,UpdatedAt
1,Apple,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z
2,Samsung,2024-01-01T00:00:00Z,2024-01-01T00:00:00Z

Note: ID field is ignored during import. New IDs are auto-generated.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		result, err := exportSvc.ImportBrandsCSV(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing brands: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("  Successfully imported: %d brands\n", result.SuccessCount)
		fmt.Printf("  Errors: %d\n", result.ErrorCount)

		if result.ErrorCount > 0 {
			fmt.Printf("\nError Details:\n")
			for _, errMsg := range result.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	},
}

var importVendorsCmd = &cobra.Command{
	Use:   "vendors [filename]",
	Short: "Import vendors from CSV",
	Long: `Import vendors from a CSV file.

CSV Format:
ID,Name,Currency,DiscountCode
1,B&H Photo,USD,SAVE10
2,Adorama,USD,SUMMER15

Note: ID field is ignored during import. New IDs are auto-generated.
Only Name, Currency, and DiscountCode are required. Other fields can be added manually.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		result, err := exportSvc.ImportVendorsCSV(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing vendors: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("  Successfully imported: %d vendors\n", result.SuccessCount)
		fmt.Printf("  Errors: %d\n", result.ErrorCount)

		if result.ErrorCount > 0 {
			fmt.Printf("\nError Details:\n")
			for _, errMsg := range result.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	},
}

var importForexCmd = &cobra.Command{
	Use:   "forex [filename]",
	Short: "Import forex rates from CSV",
	Long: `Import forex exchange rates from a CSV file.

CSV Format:
ID,FromCurrency,ToCurrency,Rate,EffectiveDate
1,EUR,USD,1.20,2024-01-01T00:00:00Z
2,GBP,USD,1.35,2024-01-01T00:00:00Z

Note: ID field is ignored during import. New IDs are auto-generated.
EffectiveDate must be in RFC3339 format (YYYY-MM-DDTHH:MM:SSZ).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		result, err := exportSvc.ImportForexCSV(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing forex rates: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("  Successfully imported: %d forex rates\n", result.SuccessCount)
		fmt.Printf("  Errors: %d\n", result.ErrorCount)

		if result.ErrorCount > 0 {
			fmt.Printf("\nError Details:\n")
			for _, errMsg := range result.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	},
}

func init() {
	importCmd.AddCommand(importBrandsCmd)
	importCmd.AddCommand(importVendorsCmd)
	importCmd.AddCommand(importForexCmd)
}
