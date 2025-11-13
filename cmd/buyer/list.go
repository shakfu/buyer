package main

import (
	"fmt"
	"os"

	"github.com/rodaine/table"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities (specifications, brands, products, vendors, quotes, forex, requisitions, projects)",
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

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		svc := services.NewProjectService(cfg.DB)
		projects, err := svc.List(limit, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return
		}

		tbl := table.New("ID", "Name", "Status", "Budget", "Deadline", "BOM Items", "Requisitions")
		for _, proj := range projects {
			budgetStr := "-"
			if proj.Budget > 0 {
				budgetStr = fmt.Sprintf("$%.2f", proj.Budget)
			}

			deadlineStr := "-"
			if proj.Deadline != nil {
				deadlineStr = proj.Deadline.Format("2006-01-02")
			}

			bomItemCount := 0
			if proj.BillOfMaterials != nil {
				bomItemCount = len(proj.BillOfMaterials.Items)
			}

			tbl.AddRow(proj.ID, proj.Name, proj.Status, budgetStr, deadlineStr, bomItemCount, len(proj.Requisitions))
		}
		tbl.Print()
	},
}

var listBOMCmd = &cobra.Command{
	Use:   "bom [project_id]",
	Short: "List Bill of Materials items for a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var projectID uint
		_, err := fmt.Sscanf(args[0], "%d", &projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid project ID\n")
			os.Exit(1)
		}

		svc := services.NewProjectService(cfg.DB)
		project, err := svc.GetByID(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if project.BillOfMaterials == nil || len(project.BillOfMaterials.Items) == 0 {
			fmt.Printf("No Bill of Materials items for project '%s' (ID: %d)\n", project.Name, project.ID)
			return
		}

		fmt.Printf("Bill of Materials for project: %s (ID: %d)\n\n", project.Name, project.ID)

		tbl := table.New("Item ID", "Specification", "Quantity", "Notes")
		for _, item := range project.BillOfMaterials.Items {
			notes := item.Notes
			if len(notes) > 40 {
				notes = notes[:37] + "..."
			}
			tbl.AddRow(item.ID, item.Specification.Name, item.Quantity, notes)
		}
		tbl.Print()
	},
}

var listProjectRequisitionsCmd = &cobra.Command{
	Use:   "project-requisitions [project_id]",
	Short: "List project requisitions (optionally filtered by project)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := services.NewProjectRequisitionService(cfg.DB)

		var requisitions []models.ProjectRequisition
		var err error

		if len(args) == 1 {
			// List for specific project
			var projectID uint
			_, scanErr := fmt.Sscanf(args[0], "%d", &projectID)
			if scanErr != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid project ID\n")
				os.Exit(1)
			}
			requisitions, err = svc.ListByProject(projectID)
		} else {
			// List all
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")
			requisitions, err = svc.List(limit, offset)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(requisitions) == 0 {
			if len(args) == 1 {
				fmt.Println("No project requisitions found for this project.")
			} else {
				fmt.Println("No project requisitions found.")
			}
			return
		}

		tbl := table.New("ID", "Project ID", "Name", "Budget", "Items", "Created")
		for _, req := range requisitions {
			budgetStr := "-"
			if req.Budget > 0 {
				budgetStr = fmt.Sprintf("$%.2f", req.Budget)
			}

			createdStr := req.CreatedAt.Format("2006-01-02")

			tbl.AddRow(req.ID, req.ProjectID, req.Name, budgetStr, len(req.Items), createdStr)
		}
		tbl.Print()
	},
}

var listPurchaseOrdersCmd = &cobra.Command{
	Use:   "purchase-orders",
	Short: "List all purchase orders",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		status, _ := cmd.Flags().GetString("status")

		svc := services.NewPurchaseOrderService(cfg.DB)
		var orders []*models.PurchaseOrder
		var err error

		if status != "" {
			orders, err = svc.ListByStatus(status, limit, offset)
		} else {
			orders, err = svc.List(limit, offset)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(orders) == 0 {
			fmt.Println("No purchase orders found.")
			return
		}

		tbl := table.New("ID", "PO Number", "Status", "Vendor", "Product", "Qty", "Total", "Order Date", "Expected")
		for _, po := range orders {
			vendorName := ""
			if po.Vendor != nil {
				vendorName = po.Vendor.Name
			}
			productName := ""
			if po.Product != nil {
				productName = po.Product.Name
			}

			totalStr := fmt.Sprintf("%.2f %s", po.GrandTotal, po.Currency)
			orderDateStr := po.OrderDate.Format("2006-01-02")
			expectedStr := "-"
			if po.ExpectedDelivery != nil {
				expectedStr = po.ExpectedDelivery.Format("2006-01-02")
			}

			tbl.AddRow(po.ID, po.PONumber, po.Status, vendorName, productName, po.Quantity, totalStr, orderDateStr, expectedStr)
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
	listCmd.AddCommand(listPurchaseOrdersCmd)
	listCmd.AddCommand(listForexCmd)
	listCmd.AddCommand(listRequisitionsCmd)
	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listBOMCmd)
	listCmd.AddCommand(listProjectRequisitionsCmd)

	// Add common pagination flags
	for _, cmd := range []*cobra.Command{listSpecificationsCmd, listBrandsCmd, listProductsCmd, listVendorsCmd, listQuotesCmd, listPurchaseOrdersCmd, listForexCmd, listRequisitionsCmd, listProjectsCmd, listProjectRequisitionsCmd} {
		cmd.Flags().Int("limit", 0, "Maximum number of results (0 = no limit)")
		cmd.Flags().Int("offset", 0, "Number of results to skip")
	}

	// Purchase order specific flags
	listPurchaseOrdersCmd.Flags().String("status", "", "Filter by status (pending, approved, ordered, shipped, received, cancelled)")
}
