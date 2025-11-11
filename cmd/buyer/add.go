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
	Short: "Add entities (specification, brand, product, vendor, quote, forex, requisition)",
	Long:  "Add specifications, brands, products, vendors, quotes, forex rates, or requisitions to the database",
}

var addSpecificationCmd = &cobra.Command{
	Use:   "specification [name] --description [text]",
	Short: "Add a new specification",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description, _ := cmd.Flags().GetString("description")

		svc := services.NewSpecificationService(cfg.DB)
		spec, err := svc.Create(args[0], description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Specification created: %s (ID: %d)\n", spec.Name, spec.ID)
	},
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

var addRequisitionCmd = &cobra.Command{
	Use:   "requisition [name] --justification [text] --budget [amount]",
	Short: "Add a new requisition",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		justification, _ := cmd.Flags().GetString("justification")
		budget, _ := cmd.Flags().GetFloat64("budget")

		svc := services.NewRequisitionService(cfg.DB)
		req, err := svc.Create(args[0], justification, budget, []services.RequisitionItemInput{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Requisition created: %s (ID: %d)\n", req.Name, req.ID)
		if justification != "" {
			fmt.Printf("  Justification: %s\n", justification)
		}
		if budget > 0 {
			fmt.Printf("  Budget: %.2f\n", budget)
		}
		fmt.Println("  Note: Use 'buyer add requisition-item' to add line items to this requisition")
	},
}

var addRequisitionItemCmd = &cobra.Command{
	Use:   "requisition-item --requisition [id] --specification [name] --quantity [num]",
	Short: "Add a line item to a requisition",
	Run: func(cmd *cobra.Command, args []string) {
		requisitionID, _ := cmd.Flags().GetUint("requisition")
		specificationName, _ := cmd.Flags().GetString("specification")
		quantity, _ := cmd.Flags().GetInt("quantity")
		budgetPerUnit, _ := cmd.Flags().GetFloat64("budget-per-unit")
		description, _ := cmd.Flags().GetString("description")

		if requisitionID == 0 || specificationName == "" || quantity == 0 {
			fmt.Fprintln(os.Stderr, "Error: --requisition, --specification, and --quantity are required")
			os.Exit(1)
		}

		// Get specification by name
		specSvc := services.NewSpecificationService(cfg.DB)
		spec, err := specSvc.GetByName(specificationName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		reqSvc := services.NewRequisitionService(cfg.DB)
		item, err := reqSvc.AddItem(requisitionID, spec.ID, quantity, budgetPerUnit, description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Line item added to requisition ID %d:\n", requisitionID)
		fmt.Printf("  Specification: %s (ID: %d)\n", spec.Name, spec.ID)
		fmt.Printf("  Quantity: %d\n", quantity)
		if budgetPerUnit > 0 {
			fmt.Printf("  Budget per unit: %.2f\n", budgetPerUnit)
		}
		if description != "" {
			fmt.Printf("  Description: %s\n", description)
		}
		fmt.Printf("  Item ID: %d\n", item.ID)
	},
}

func init() {
	addCmd.AddCommand(addSpecificationCmd)
	addCmd.AddCommand(addBrandCmd)
	addCmd.AddCommand(addProductCmd)
	addCmd.AddCommand(addVendorCmd)
	addCmd.AddCommand(addQuoteCmd)
	addCmd.AddCommand(addForexCmd)
	addCmd.AddCommand(addRequisitionCmd)
	addCmd.AddCommand(addRequisitionItemCmd)

	// Specification flags
	addSpecificationCmd.Flags().String("description", "", "Description of the specification")

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

	// Requisition flags
	addRequisitionCmd.Flags().String("justification", "", "Justification for the requisition")
	addRequisitionCmd.Flags().Float64("budget", 0, "Overall budget limit for the requisition")

	// Requisition item flags
	addRequisitionItemCmd.Flags().Uint("requisition", 0, "Requisition ID (required)")
	addRequisitionItemCmd.Flags().String("specification", "", "Specification name (required)")
	addRequisitionItemCmd.Flags().Int("quantity", 0, "Quantity (required)")
	addRequisitionItemCmd.Flags().Float64("budget-per-unit", 0, "Budget per unit (optional)")
	addRequisitionItemCmd.Flags().String("description", "", "Additional description (optional)")
}
