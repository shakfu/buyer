package main

import (
	"fmt"
	"os"

	"github.com/rodaine/table"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities (specifications, brands, products, vendors, quotes, forex, requisitions)",
	Long:  "List all entities with optional pagination",
}

var listSpecificationsCmd = &cobra.Command{
	Use:   "specifications",
	Short: "List all specifications",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewSpecificationService(cfg.DB)
		specs, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(specs) == 0 {
			fmt.Println("No specifications found.")
			return
		}

		tbl := table.New("ID", "Name", "Description", "Products")
		for _, spec := range specs {
			desc := spec.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			tbl.AddRow(spec.ID, spec.Name, desc, len(spec.Products))
		}
		tbl.Print()
	},
}

var listBrandsCmd = &cobra.Command{
	Use:   "brands",
	Short: "List all brands",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewBrandService(cfg.DB)
		brands, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(brands) == 0 {
			fmt.Println("No brands found.")
			return
		}

		tbl := table.New("ID", "Name", "Products", "Vendors")
		for _, brand := range brands {
			tbl.AddRow(brand.ID, brand.Name, len(brand.Products), len(brand.Vendors))
		}
		tbl.Print()
	},
}

var listProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "List all products",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewProductService(cfg.DB)
		products, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(products) == 0 {
			fmt.Println("No products found.")
			return
		}

		tbl := table.New("ID", "Name", "Brand", "Quotes")
		for _, product := range products {
			brandName := ""
			if product.Brand != nil {
				brandName = product.Brand.Name
			}
			tbl.AddRow(product.ID, product.Name, brandName, len(product.Quotes))
		}
		tbl.Print()
	},
}

var listVendorsCmd = &cobra.Command{
	Use:   "vendors",
	Short: "List all vendors",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewVendorService(cfg.DB)
		vendors, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(vendors) == 0 {
			fmt.Println("No vendors found.")
			return
		}

		tbl := table.New("ID", "Name", "Currency", "Discount", "Brands", "Quotes")
		for _, vendor := range vendors {
			tbl.AddRow(vendor.ID, vendor.Name, vendor.Currency, vendor.DiscountCode, len(vendor.Brands), len(vendor.Quotes))
		}
		tbl.Print()
	},
}

var listQuotesCmd = &cobra.Command{
	Use:   "quotes",
	Short: "List all quotes",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewQuoteService(cfg.DB)
		quotes, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(quotes) == 0 {
			fmt.Println("No quotes found.")
			return
		}

		tbl := table.New("ID", "Vendor", "Product", "Price", "USD", "Date")
		for _, quote := range quotes {
			vendorName := ""
			if quote.Vendor != nil {
				vendorName = quote.Vendor.Name
			}
			productName := ""
			if quote.Product != nil {
				productName = quote.Product.Name
			}
			priceStr := fmt.Sprintf("%.2f %s", quote.Price, quote.Currency)
			usdStr := fmt.Sprintf("%.2f", quote.ConvertedPrice)
			tbl.AddRow(quote.ID, vendorName, productName, priceStr, usdStr, quote.QuoteDate.Format("2006-01-02"))
		}
		tbl.Print()
	},
}

var listForexCmd = &cobra.Command{
	Use:   "forex",
	Short: "List all forex rates",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewForexService(cfg.DB)
		rates, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(rates) == 0 {
			fmt.Println("No forex rates found.")
			return
		}

		tbl := table.New("ID", "From", "To", "Rate", "Date")
		for _, rate := range rates {
			rateStr := fmt.Sprintf("%.4f", rate.Rate)
			tbl.AddRow(rate.ID, rate.FromCurrency, rate.ToCurrency, rateStr, rate.EffectiveDate.Format("2006-01-02"))
		}
		tbl.Print()
	},
}

var listRequisitionsCmd = &cobra.Command{
	Use:   "requisitions",
	Short: "List all requisitions",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewRequisitionService(cfg.DB)
		reqs, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(reqs) == 0 {
			fmt.Println("No requisitions found.")
			return
		}

		tbl := table.New("ID", "Name", "Items", "Budget", "Justification")
		for _, req := range reqs {
			budgetStr := "-"
			if req.Budget > 0 {
				budgetStr = fmt.Sprintf("%.2f", req.Budget)
			}
			just := req.Justification
			if len(just) > 40 {
				just = just[:37] + "..."
			}
			tbl.AddRow(req.ID, req.Name, len(req.Items), budgetStr, just)
		}
		tbl.Print()
	},
}

func init() {
	listCmd.AddCommand(listSpecificationsCmd)
	listCmd.AddCommand(listBrandsCmd)
	listCmd.AddCommand(listProductsCmd)
	listCmd.AddCommand(listVendorsCmd)
	listCmd.AddCommand(listQuotesCmd)
	listCmd.AddCommand(listForexCmd)
	listCmd.AddCommand(listRequisitionsCmd)

	// Add common pagination flags
	for _, cmd := range []*cobra.Command{listSpecificationsCmd, listBrandsCmd, listProductsCmd, listVendorsCmd, listQuotesCmd, listForexCmd, listRequisitionsCmd} {
		cmd.Flags().Int("limit", 0, "Maximum number of results (0 = no limit)")
		cmd.Flags().Int("offset", 0, "Number of results to skip")
	}
}
