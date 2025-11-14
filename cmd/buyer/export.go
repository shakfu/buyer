package main

import (
	"fmt"
	"os"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to CSV or Excel",
	Long:  `Export brands, vendors, products, quotes, or forex rates to CSV or Excel files.`,
}

var exportBrandsCmd = &cobra.Command{
	Use:   "brands [filename]",
	Short: "Export brands to CSV or Excel",
	Long:  `Export all brands to a CSV or Excel file. Format is determined by file extension (.csv or .xlsx).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		if isExcelFile(filename) {
			// Export to Excel
			f, err := exportSvc.ExportBrandsExcel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting brands to Excel: %v\n", err)
				os.Exit(1)
			}

			if err := f.SaveAs(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving Excel file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Brands exported to Excel file: %s\n", filename)
		} else {
			// Export to CSV
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			if err := exportSvc.ExportBrandsCSV(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting brands to CSV: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Brands exported to CSV file: %s\n", filename)
		}
	},
}

var exportVendorsCmd = &cobra.Command{
	Use:   "vendors [filename]",
	Short: "Export vendors to CSV or Excel",
	Long:  `Export all vendors to a CSV or Excel file. Format is determined by file extension (.csv or .xlsx).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		if isExcelFile(filename) {
			f, err := exportSvc.ExportVendorsExcel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting vendors to Excel: %v\n", err)
				os.Exit(1)
			}

			if err := f.SaveAs(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving Excel file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Vendors exported to Excel file: %s\n", filename)
		} else {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			if err := exportSvc.ExportVendorsCSV(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting vendors to CSV: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Vendors exported to CSV file: %s\n", filename)
		}
	},
}

var exportProductsCmd = &cobra.Command{
	Use:   "products [filename]",
	Short: "Export products to CSV or Excel",
	Long:  `Export all products to a CSV or Excel file. Format is determined by file extension (.csv or .xlsx).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		if isExcelFile(filename) {
			f, err := exportSvc.ExportProductsExcel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting products to Excel: %v\n", err)
				os.Exit(1)
			}

			if err := f.SaveAs(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving Excel file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Products exported to Excel file: %s\n", filename)
		} else {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			if err := exportSvc.ExportProductsCSV(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting products to CSV: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Products exported to CSV file: %s\n", filename)
		}
	},
}

var exportQuotesCmd = &cobra.Command{
	Use:   "quotes [filename]",
	Short: "Export quotes to CSV or Excel",
	Long:  `Export all quotes to a CSV or Excel file. Format is determined by file extension (.csv or .xlsx).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		if isExcelFile(filename) {
			f, err := exportSvc.ExportQuotesExcel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting quotes to Excel: %v\n", err)
				os.Exit(1)
			}

			if err := f.SaveAs(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving Excel file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Quotes exported to Excel file: %s\n", filename)
		} else {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			if err := exportSvc.ExportQuotesCSV(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting quotes to CSV: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Quotes exported to CSV file: %s\n", filename)
		}
	},
}

var exportForexCmd = &cobra.Command{
	Use:   "forex [filename]",
	Short: "Export forex rates to CSV or Excel",
	Long:  `Export all forex rates to a CSV or Excel file. Format is determined by file extension (.csv or .xlsx).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		exportSvc := services.NewExportImportService(cfg.DB)

		if isExcelFile(filename) {
			f, err := exportSvc.ExportForexExcel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting forex rates to Excel: %v\n", err)
				os.Exit(1)
			}

			if err := f.SaveAs(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving Excel file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Forex rates exported to Excel file: %s\n", filename)
		} else {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			if err := exportSvc.ExportForexCSV(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting forex rates to CSV: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Forex rates exported to CSV file: %s\n", filename)
		}
	},
}

func init() {
	exportCmd.AddCommand(exportBrandsCmd)
	exportCmd.AddCommand(exportVendorsCmd)
	exportCmd.AddCommand(exportProductsCmd)
	exportCmd.AddCommand(exportQuotesCmd)
	exportCmd.AddCommand(exportForexCmd)
}

// isExcelFile determines if a filename represents an Excel file
func isExcelFile(filename string) bool {
	return len(filename) > 5 && filename[len(filename)-5:] == ".xlsx"
}
