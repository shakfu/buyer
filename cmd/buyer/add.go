package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add entities (specification, brand, product, vendor, quote, forex, requisition, project, document, vendor-rating)",
	Long:  "Add specifications, brands, products, vendors, quotes, forex rates, requisitions, projects, documents, or vendor ratings to the database",
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
			slog.Error("failed to create brand",
				slog.String("name", args[0]),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		slog.Info("brand created successfully",
			slog.String("name", brand.Name),
			slog.Uint64("id", uint64(brand.ID)))
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

var addProjectCmd = &cobra.Command{
	Use:   "project [name] --budget [amount] --deadline [date] --description [text]",
	Short: "Add a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description, _ := cmd.Flags().GetString("description")
		budget, _ := cmd.Flags().GetFloat64("budget")
		deadlineStr, _ := cmd.Flags().GetString("deadline")

		var deadline *time.Time
		if deadlineStr != "" {
			t, err := time.Parse("2006-01-02", deadlineStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid deadline format. Use YYYY-MM-DD\n")
				os.Exit(1)
			}
			deadline = &t
		}

		svc := services.NewProjectService(cfg.DB)
		project, err := svc.Create(args[0], description, budget, deadline)
		if err != nil {
			slog.Error("failed to create project",
				slog.String("name", args[0]),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("project created successfully",
			slog.String("name", project.Name),
			slog.Uint64("id", uint64(project.ID)))

		fmt.Printf("Project created: %s (ID: %d)\n", project.Name, project.ID)
		fmt.Printf("  Status: %s\n", project.Status)
		if project.Budget > 0 {
			fmt.Printf("  Budget: $%.2f\n", project.Budget)
		}
		if project.Deadline != nil {
			fmt.Printf("  Deadline: %s\n", project.Deadline.Format("2006-01-02"))
		}
		if project.Description != "" {
			fmt.Printf("  Description: %s\n", project.Description)
		}
		fmt.Printf("  Bill of Materials: Created (ID: %d)\n", project.BillOfMaterials.ID)
	},
}

var addBOMItemCmd = &cobra.Command{
	Use:   "bom-item --project [project_id] --spec [specification_name] --quantity [num] --notes [text]",
	Short: "Add an item to a project's Bill of Materials",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.Flags().GetUint("project")
		specName, _ := cmd.Flags().GetString("spec")
		quantity, _ := cmd.Flags().GetInt("quantity")
		notes, _ := cmd.Flags().GetString("notes")

		if projectID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --project flag is required")
			os.Exit(1)
		}
		if specName == "" {
			fmt.Fprintln(os.Stderr, "Error: --spec flag is required")
			os.Exit(1)
		}
		if quantity <= 0 {
			fmt.Fprintln(os.Stderr, "Error: --quantity must be greater than 0")
			os.Exit(1)
		}

		// Get specification by name
		specSvc := services.NewSpecificationService(cfg.DB)
		spec, err := specSvc.GetByName(specName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		projectSvc := services.NewProjectService(cfg.DB)
		bomItem, err := projectSvc.AddBillOfMaterialsItem(projectID, spec.ID, quantity, notes)
		if err != nil {
			slog.Error("failed to add BOM item",
				slog.Uint64("project_id", uint64(projectID)),
				slog.String("specification", specName),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("BOM item added successfully",
			slog.Uint64("project_id", uint64(projectID)),
			slog.Uint64("item_id", uint64(bomItem.ID)))

		fmt.Printf("Bill of Materials item added to project ID %d:\n", projectID)
		fmt.Printf("  Specification: %s (ID: %d)\n", bomItem.Specification.Name, bomItem.Specification.ID)
		fmt.Printf("  Quantity: %d\n", bomItem.Quantity)
		if bomItem.Notes != "" {
			fmt.Printf("  Notes: %s\n", bomItem.Notes)
		}
		fmt.Printf("  Item ID: %d\n", bomItem.ID)
	},
}

var addProjectRequisitionCmd = &cobra.Command{
	Use:   "project-requisition [name] --project [id] --justification [text] --budget [amount] --bom-items [id:qty,id:qty,...]",
	Short: "Create a project requisition from BOM items",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		projectID, _ := cmd.Flags().GetUint("project")
		justification, _ := cmd.Flags().GetString("justification")
		budget, _ := cmd.Flags().GetFloat64("budget")
		bomItemsStr, _ := cmd.Flags().GetString("bom-items")
		notesStr, _ := cmd.Flags().GetString("notes")

		if projectID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --project flag is required")
			os.Exit(1)
		}

		if bomItemsStr == "" {
			fmt.Fprintln(os.Stderr, "Error: --bom-items flag is required (format: id:qty,id:qty,...)")
			os.Exit(1)
		}

		// Parse BOM items (format: "id:qty,id:qty,...")
		items := []services.ProjectRequisitionItemInput{}
		itemPairs := strings.Split(bomItemsStr, ",")
		notesList := []string{}
		if notesStr != "" {
			notesList = strings.Split(notesStr, "|")
		}

		for i, pair := range itemPairs {
			parts := strings.Split(strings.TrimSpace(pair), ":")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Error: invalid bom-item format '%s' (expected id:qty)\n", pair)
				os.Exit(1)
			}

			bomItemID, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid BOM item ID '%s': %v\n", parts[0], err)
				os.Exit(1)
			}

			quantity, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid quantity '%s': %v\n", parts[1], err)
				os.Exit(1)
			}

			notes := ""
			if i < len(notesList) {
				notes = notesList[i]
			}

			items = append(items, services.ProjectRequisitionItemInput{
				BOMItemID:         uint(bomItemID),
				QuantityRequested: quantity,
				Notes:             notes,
			})
		}

		projectReqSvc := services.NewProjectRequisitionService(cfg.DB)
		projectReq, err := projectReqSvc.Create(projectID, name, justification, budget, items)
		if err != nil {
			slog.Error("failed to create project requisition",
				slog.String("name", name),
				slog.Uint64("project_id", uint64(projectID)),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("project requisition created successfully",
			slog.Uint64("id", uint64(projectReq.ID)),
			slog.String("name", projectReq.Name))

		fmt.Printf("Project Requisition created successfully:\n")
		fmt.Printf("  ID: %d\n", projectReq.ID)
		fmt.Printf("  Name: %s\n", projectReq.Name)
		fmt.Printf("  Project ID: %d\n", projectReq.ProjectID)
		if projectReq.Justification != "" {
			fmt.Printf("  Justification: %s\n", projectReq.Justification)
		}
		if projectReq.Budget > 0 {
			fmt.Printf("  Budget: $%.2f\n", projectReq.Budget)
		}
		fmt.Printf("  Items: %d\n", len(projectReq.Items))
		for _, item := range projectReq.Items {
			fmt.Printf("    - BOM Item ID %d: %d units", item.BillOfMaterialsItemID, item.QuantityRequested)
			if item.BOMItem != nil && item.BOMItem.Specification != nil {
				fmt.Printf(" (%s)", item.BOMItem.Specification.Name)
			}
			if item.Notes != "" {
				fmt.Printf(" - %s", item.Notes)
			}
			fmt.Println()
		}
	},
}

var addPurchaseOrderCmd = &cobra.Command{
	Use:   "purchase-order --quote-id [id] --po-number [number] --quantity [qty]",
	Short: "Add a new purchase order from a quote",
	Run: func(cmd *cobra.Command, args []string) {
		quoteID, _ := cmd.Flags().GetUint("quote-id")
		poNumber, _ := cmd.Flags().GetString("po-number")
		quantity, _ := cmd.Flags().GetInt("quantity")
		requisitionID, _ := cmd.Flags().GetUint("requisition-id")
		expectedDeliveryStr, _ := cmd.Flags().GetString("expected-delivery")
		shippingCost, _ := cmd.Flags().GetFloat64("shipping-cost")
		tax, _ := cmd.Flags().GetFloat64("tax")
		notes, _ := cmd.Flags().GetString("notes")

		if quoteID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --quote-id flag is required")
			os.Exit(1)
		}
		if poNumber == "" {
			fmt.Fprintln(os.Stderr, "Error: --po-number flag is required")
			os.Exit(1)
		}
		if quantity == 0 {
			fmt.Fprintln(os.Stderr, "Error: --quantity flag is required")
			os.Exit(1)
		}

		var expectedDelivery *time.Time
		if expectedDeliveryStr != "" {
			parsed, err := time.Parse("2006-01-02", expectedDeliveryStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing expected-delivery: %v\n", err)
				os.Exit(1)
			}
			expectedDelivery = &parsed
		}

		var reqIDPtr *uint
		if requisitionID != 0 {
			reqIDPtr = &requisitionID
		}

		svc := services.NewPurchaseOrderService(cfg.DB)
		po, err := svc.Create(services.CreatePurchaseOrderInput{
			QuoteID:          quoteID,
			RequisitionID:    reqIDPtr,
			PONumber:         poNumber,
			Quantity:         quantity,
			ExpectedDelivery: expectedDelivery,
			ShippingCost:     shippingCost,
			Tax:              tax,
			Notes:            notes,
		})
		if err != nil {
			slog.Error("failed to create purchase order",
				slog.String("po_number", poNumber),
				slog.Uint64("quote_id", uint64(quoteID)),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("purchase order created successfully",
			slog.Uint64("id", uint64(po.ID)),
			slog.String("po_number", po.PONumber))

		fmt.Printf("Purchase Order created successfully:\n")
		fmt.Printf("  ID: %d\n", po.ID)
		fmt.Printf("  PO Number: %s\n", po.PONumber)
		fmt.Printf("  Status: %s\n", po.Status)
		if po.Vendor != nil {
			fmt.Printf("  Vendor: %s\n", po.Vendor.Name)
		}
		if po.Product != nil {
			fmt.Printf("  Product: %s\n", po.Product.Name)
		}
		fmt.Printf("  Quantity: %d\n", po.Quantity)
		fmt.Printf("  Unit Price: %.2f %s\n", po.UnitPrice, po.Currency)
		fmt.Printf("  Total Amount: %.2f %s\n", po.TotalAmount, po.Currency)
		if po.ShippingCost > 0 {
			fmt.Printf("  Shipping: %.2f %s\n", po.ShippingCost, po.Currency)
		}
		if po.Tax > 0 {
			fmt.Printf("  Tax: %.2f %s\n", po.Tax, po.Currency)
		}
		fmt.Printf("  Grand Total: %.2f %s\n", po.GrandTotal, po.Currency)
		fmt.Printf("  Order Date: %s\n", po.OrderDate.Format("2006-01-02"))
		if po.ExpectedDelivery != nil {
			fmt.Printf("  Expected Delivery: %s\n", po.ExpectedDelivery.Format("2006-01-02"))
		}
	},
}

var addDocumentCmd = &cobra.Command{
	Use:   "document --entity-type [type] --entity-id [id] --file-name [name] --file-path [path]",
	Short: "Add a new document to an entity",
	Run: func(cmd *cobra.Command, args []string) {
		entityType, _ := cmd.Flags().GetString("entity-type")
		entityID, _ := cmd.Flags().GetUint("entity-id")
		fileName, _ := cmd.Flags().GetString("file-name")
		filePath, _ := cmd.Flags().GetString("file-path")
		fileType, _ := cmd.Flags().GetString("file-type")
		fileSize, _ := cmd.Flags().GetInt64("file-size")
		description, _ := cmd.Flags().GetString("description")
		uploadedBy, _ := cmd.Flags().GetString("uploaded-by")

		if entityType == "" || entityID == 0 || fileName == "" || filePath == "" {
			fmt.Fprintln(os.Stderr, "Error: --entity-type, --entity-id, --file-name, and --file-path are required")
			os.Exit(1)
		}

		svc := services.NewDocumentService(cfg.DB)
		doc, err := svc.Create(services.CreateDocumentInput{
			EntityType:  entityType,
			EntityID:    entityID,
			FileName:    fileName,
			FileType:    fileType,
			FileSize:    fileSize,
			FilePath:    filePath,
			Description: description,
			UploadedBy:  uploadedBy,
		})
		if err != nil {
			slog.Error("failed to create document",
				slog.String("entity_type", entityType),
				slog.Uint64("entity_id", uint64(entityID)),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("document created successfully",
			slog.Uint64("id", uint64(doc.ID)),
			slog.String("file_name", doc.FileName))

		fmt.Printf("Document created successfully:\n")
		fmt.Printf("  ID: %d\n", doc.ID)
		fmt.Printf("  Entity Type: %s\n", doc.EntityType)
		fmt.Printf("  Entity ID: %d\n", doc.EntityID)
		fmt.Printf("  File Name: %s\n", doc.FileName)
		fmt.Printf("  File Path: %s\n", doc.FilePath)
		if doc.FileType != "" {
			fmt.Printf("  File Type: %s\n", doc.FileType)
		}
		if doc.FileSize > 0 {
			fmt.Printf("  File Size: %.1f KB\n", float64(doc.FileSize)/1024)
		}
		if doc.Description != "" {
			fmt.Printf("  Description: %s\n", doc.Description)
		}
		if doc.UploadedBy != "" {
			fmt.Printf("  Uploaded By: %s\n", doc.UploadedBy)
		}
	},
}

var addVendorRatingCmd = &cobra.Command{
	Use:   "vendor-rating --vendor [name_or_id] --price [1-5] --quality [1-5] --delivery [1-5] --service [1-5]",
	Short: "Add a vendor rating",
	Run: func(cmd *cobra.Command, args []string) {
		vendorInput, _ := cmd.Flags().GetString("vendor")
		poID, _ := cmd.Flags().GetUint("po-id")
		priceRating, _ := cmd.Flags().GetInt("price")
		qualityRating, _ := cmd.Flags().GetInt("quality")
		deliveryRating, _ := cmd.Flags().GetInt("delivery")
		serviceRating, _ := cmd.Flags().GetInt("service")
		comments, _ := cmd.Flags().GetString("comments")
		ratedBy, _ := cmd.Flags().GetString("rated-by")

		if vendorInput == "" {
			fmt.Fprintln(os.Stderr, "Error: --vendor flag is required")
			os.Exit(1)
		}

		// Get vendor by name or ID
		vendorSvc := services.NewVendorService(cfg.DB)
		var vendor *services.Vendor
		var err error

		// Try parsing as ID first
		var vendorID uint
		_, parseErr := fmt.Sscanf(vendorInput, "%d", &vendorID)
		if parseErr == nil {
			vendor, err = vendorSvc.GetByID(vendorID)
		} else {
			vendor, err = vendorSvc.GetByName(vendorInput)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding vendor: %v\n", err)
			os.Exit(1)
		}

		// Build input
		input := services.CreateVendorRatingInput{
			VendorID: vendor.ID,
			Comments: comments,
			RatedBy:  ratedBy,
		}

		if poID > 0 {
			input.PurchaseOrderID = &poID
		}
		if priceRating > 0 {
			input.PriceRating = &priceRating
		}
		if qualityRating > 0 {
			input.QualityRating = &qualityRating
		}
		if deliveryRating > 0 {
			input.DeliveryRating = &deliveryRating
		}
		if serviceRating > 0 {
			input.ServiceRating = &serviceRating
		}

		ratingSvc := services.NewVendorRatingService(cfg.DB)
		rating, err := ratingSvc.Create(input)
		if err != nil {
			slog.Error("failed to create vendor rating",
				slog.Uint64("vendor_id", uint64(vendor.ID)),
				slog.String("error", err.Error()))
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		slog.Info("vendor rating created successfully",
			slog.Uint64("id", uint64(rating.ID)),
			slog.Uint64("vendor_id", uint64(vendor.ID)))

		fmt.Printf("Vendor Rating created successfully:\n")
		fmt.Printf("  ID: %d\n", rating.ID)
		fmt.Printf("  Vendor: %s (ID: %d)\n", rating.Vendor.Name, rating.VendorID)
		if rating.PurchaseOrderID != nil {
			fmt.Printf("  Purchase Order ID: %d\n", *rating.PurchaseOrderID)
		}
		if rating.PriceRating != nil {
			fmt.Printf("  Price Rating: %d/5\n", *rating.PriceRating)
		}
		if rating.QualityRating != nil {
			fmt.Printf("  Quality Rating: %d/5\n", *rating.QualityRating)
		}
		if rating.DeliveryRating != nil {
			fmt.Printf("  Delivery Rating: %d/5\n", *rating.DeliveryRating)
		}
		if rating.ServiceRating != nil {
			fmt.Printf("  Service Rating: %d/5\n", *rating.ServiceRating)
		}
		if rating.Comments != "" {
			fmt.Printf("  Comments: %s\n", rating.Comments)
		}
		if rating.RatedBy != "" {
			fmt.Printf("  Rated By: %s\n", rating.RatedBy)
		}
	},
}

func init() {
	addCmd.AddCommand(addSpecificationCmd)
	addCmd.AddCommand(addBrandCmd)
	addCmd.AddCommand(addProductCmd)
	addCmd.AddCommand(addVendorCmd)
	addCmd.AddCommand(addQuoteCmd)
	addCmd.AddCommand(addPurchaseOrderCmd)
	addCmd.AddCommand(addForexCmd)
	addCmd.AddCommand(addRequisitionCmd)
	addCmd.AddCommand(addRequisitionItemCmd)
	addCmd.AddCommand(addProjectCmd)
	addCmd.AddCommand(addBOMItemCmd)
	addCmd.AddCommand(addProjectRequisitionCmd)
	addCmd.AddCommand(addDocumentCmd)
	addCmd.AddCommand(addVendorRatingCmd)

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

	// Purchase Order flags
	addPurchaseOrderCmd.Flags().Uint("quote-id", 0, "Quote ID (required)")
	addPurchaseOrderCmd.Flags().String("po-number", "", "PO number (required)")
	addPurchaseOrderCmd.Flags().Int("quantity", 0, "Quantity (required)")
	addPurchaseOrderCmd.Flags().Uint("requisition-id", 0, "Requisition ID (optional)")
	addPurchaseOrderCmd.Flags().String("expected-delivery", "", "Expected delivery date (YYYY-MM-DD)")
	addPurchaseOrderCmd.Flags().Float64("shipping-cost", 0, "Shipping cost")
	addPurchaseOrderCmd.Flags().Float64("tax", 0, "Tax amount")
	addPurchaseOrderCmd.Flags().String("notes", "", "Additional notes")

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

	// Project flags
	addProjectCmd.Flags().String("description", "", "Project description")
	addProjectCmd.Flags().Float64("budget", 0, "Overall project budget")
	addProjectCmd.Flags().String("deadline", "", "Project deadline (YYYY-MM-DD format)")

	// BOM item flags
	addBOMItemCmd.Flags().Uint("project", 0, "Project ID (required)")
	addBOMItemCmd.Flags().String("spec", "", "Specification name (required)")
	addBOMItemCmd.Flags().Int("quantity", 0, "Quantity (required)")
	addBOMItemCmd.Flags().String("notes", "", "Additional notes")

	// Project requisition flags
	addProjectRequisitionCmd.Flags().Uint("project", 0, "Project ID (required)")
	addProjectRequisitionCmd.Flags().String("justification", "", "Justification for the requisition")
	addProjectRequisitionCmd.Flags().Float64("budget", 0, "Budget for this requisition")
	addProjectRequisitionCmd.Flags().String("bom-items", "", "BOM items in format id:qty,id:qty,... (required)")
	addProjectRequisitionCmd.Flags().String("notes", "", "Notes for items, separated by | (optional)")

	// Document flags
	addDocumentCmd.Flags().String("entity-type", "", "Entity type (vendor, brand, product, quote, purchase_order, requisition, project) - required")
	addDocumentCmd.Flags().Uint("entity-id", 0, "Entity ID - required")
	addDocumentCmd.Flags().String("file-name", "", "File name - required")
	addDocumentCmd.Flags().String("file-path", "", "File path - required")
	addDocumentCmd.Flags().String("file-type", "", "File type (pdf, doc, xls, etc.)")
	addDocumentCmd.Flags().Int64("file-size", 0, "File size in bytes")
	addDocumentCmd.Flags().String("description", "", "Description")
	addDocumentCmd.Flags().String("uploaded-by", "", "Uploaded by (user email or name)")

	// Vendor rating flags
	addVendorRatingCmd.Flags().String("vendor", "", "Vendor name or ID - required")
	addVendorRatingCmd.Flags().Uint("po-id", 0, "Purchase Order ID (optional)")
	addVendorRatingCmd.Flags().Int("price", 0, "Price rating (1-5)")
	addVendorRatingCmd.Flags().Int("quality", 0, "Quality rating (1-5)")
	addVendorRatingCmd.Flags().Int("delivery", 0, "Delivery rating (1-5)")
	addVendorRatingCmd.Flags().Int("service", 0, "Service rating (1-5)")
	addVendorRatingCmd.Flags().String("comments", "", "Comments about the rating")
	addVendorRatingCmd.Flags().String("rated-by", "", "Rated by (user email or name)")
}
