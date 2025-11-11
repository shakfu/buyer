package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add entities (brand, product, vendor, quote, forex)",
	Long:  "Add brands, products, vendors, quotes, or forex rates to the database",
}

var addBrandCmd = &cobra.Command{
	Use:   "brand [name]",
	Short: "Add a new brand",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := services.NewBrandService(cfg.DB)
		brand, err := svc.Create(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Brand created: %s (ID: %d)\n", brand.Name, brand.ID)
	},
}

var addProductCmd = &cobra.Command{
	Use:   "product [name] --brand [brand_name]",
	Short: "Add a new product",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		brandName, _ := cmd.Flags().GetString("brand")
		if brandName == "" {
			fmt.Fprintln(os.Stderr, "Error: --brand flag is required")
			os.Exit(1)
		}

		// Get brand by name
		brandSvc := services.NewBrandService(cfg.DB)
		brand, err := brandSvc.GetByName(brandName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		productSvc := services.NewProductService(cfg.DB)
		product, err := productSvc.Create(args[0], brand.ID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Product created: %s (ID: %d, Brand: %s)\n", product.Name, product.ID, product.Brand.Name)
	},
}

var addVendorCmd = &cobra.Command{
	Use:   "vendor [name] --currency [code] --discount [code]",
	Short: "Add a new vendor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		currency, _ := cmd.Flags().GetString("currency")
		discountCode, _ := cmd.Flags().GetString("discount")

		svc := services.NewVendorService(cfg.DB)
		vendor, err := svc.Create(args[0], currency, discountCode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Vendor created: %s (ID: %d, Currency: %s)\n", vendor.Name, vendor.ID, vendor.Currency)
	},
}

var addQuoteCmd = &cobra.Command{
	Use:   "quote --vendor [name] --product [name] --price [amount] --currency [code]",
	Short: "Add a new quote",
	Run: func(cmd *cobra.Command, args []string) {
		vendorName, _ := cmd.Flags().GetString("vendor")
		productName, _ := cmd.Flags().GetString("product")
		price, _ := cmd.Flags().GetFloat64("price")
		currency, _ := cmd.Flags().GetString("currency")
		notes, _ := cmd.Flags().GetString("notes")

		if vendorName == "" || productName == "" || price == 0 {
			fmt.Fprintln(os.Stderr, "Error: --vendor, --product, and --price are required")
			os.Exit(1)
		}

		// Get vendor and product
		vendorSvc := services.NewVendorService(cfg.DB)
		vendor, err := vendorSvc.GetByName(vendorName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		productSvc := services.NewProductService(cfg.DB)
		product, err := productSvc.GetByName(productName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		quoteSvc := services.NewQuoteService(cfg.DB)
		quote, err := quoteSvc.Create(services.CreateQuoteInput{
			VendorID:  vendor.ID,
			ProductID: product.ID,
			Price:     price,
			Currency:  currency,
			QuoteDate: time.Now(),
			Notes:     notes,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Quote created: ID %d\n", quote.ID)
		fmt.Printf("  Vendor: %s\n", quote.Vendor.Name)
		fmt.Printf("  Product: %s\n", quote.Product.Name)
		fmt.Printf("  Price: %.2f %s (%.2f USD)\n", quote.Price, quote.Currency, quote.ConvertedPrice)
	},
}

var addForexCmd = &cobra.Command{
	Use:   "forex --from [code] --to [code] --rate [rate]",
	Short: "Add a forex exchange rate",
	Run: func(cmd *cobra.Command, args []string) {
		fromCurrency, _ := cmd.Flags().GetString("from")
		toCurrency, _ := cmd.Flags().GetString("to")
		rate, _ := cmd.Flags().GetFloat64("rate")

		if fromCurrency == "" || toCurrency == "" || rate == 0 {
			fmt.Fprintln(os.Stderr, "Error: --from, --to, and --rate are required")
			os.Exit(1)
		}

		svc := services.NewForexService(cfg.DB)
		forex, err := svc.Create(fromCurrency, toCurrency, rate, time.Now())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Forex rate created: %s/%s = %.4f (ID: %d)\n",
			forex.FromCurrency, forex.ToCurrency, forex.Rate, forex.ID)
	},
}

func init() {
	addCmd.AddCommand(addBrandCmd)
	addCmd.AddCommand(addProductCmd)
	addCmd.AddCommand(addVendorCmd)
	addCmd.AddCommand(addQuoteCmd)
	addCmd.AddCommand(addForexCmd)

	// Product flags
	addProductCmd.Flags().String("brand", "", "Brand name (required)")

	// Vendor flags
	addVendorCmd.Flags().String("currency", "USD", "Currency code (default: USD)")
	addVendorCmd.Flags().String("discount", "", "Discount code")

	// Quote flags
	addQuoteCmd.Flags().String("vendor", "", "Vendor name (required)")
	addQuoteCmd.Flags().String("product", "", "Product name (required)")
	addQuoteCmd.Flags().Float64("price", 0, "Price (required)")
	addQuoteCmd.Flags().String("currency", "", "Currency code (defaults to vendor's currency)")
	addQuoteCmd.Flags().String("notes", "", "Additional notes")

	// Forex flags
	addForexCmd.Flags().String("from", "", "From currency code (required)")
	addForexCmd.Flags().String("to", "", "To currency code (required)")
	addForexCmd.Flags().Float64("rate", 0, "Exchange rate (required)")
}
