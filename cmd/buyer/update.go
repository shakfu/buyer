package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update entities (specification, brand, product, vendor, project, bom-item, project-requisition)",
	Long:  "Update entity names by ID",
}

var updateSpecificationCmd = &cobra.Command{
	Use:   "specification [id] [new_name] --description [text]",
	Short: "Update a specification's name and/or description",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		description, _ := cmd.Flags().GetString("description")

		svc := services.NewSpecificationService(cfg.DB)
		spec, err := svc.Update(uint(id), args[1], description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Specification updated: %s (ID: %d)\n", spec.Name, spec.ID)
	},
}

var updateBrandCmd = &cobra.Command{
	Use:   "brand [id] [new_name]",
	Short: "Update a brand's name",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		svc := services.NewBrandService(cfg.DB)
		brand, err := svc.Update(uint(id), args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Brand updated: %s (ID: %d)\n", brand.Name, brand.ID)
	},
}

var updateProductCmd = &cobra.Command{
	Use:   "product [id] [new_name]",
	Short: "Update a product's name",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		svc := services.NewProductService(cfg.DB)
		product, err := svc.Update(uint(id), args[1], nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Product updated: %s (ID: %d)\n", product.Name, product.ID)
	},
}

var updateVendorCmd = &cobra.Command{
	Use:   "vendor [id] [new_name]",
	Short: "Update a vendor's name",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		svc := services.NewVendorService(cfg.DB)
		vendor, err := svc.Update(uint(id), args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Vendor updated: %s (ID: %d)\n", vendor.Name, vendor.ID)
	},
}

var updateProjectCmd = &cobra.Command{
	Use:   "project [id] [new_name] --description [text] --budget [amount] --deadline [date] --status [status]",
	Short: "Update a project",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		description, _ := cmd.Flags().GetString("description")
		budget, _ := cmd.Flags().GetFloat64("budget")
		deadlineStr, _ := cmd.Flags().GetString("deadline")
		status, _ := cmd.Flags().GetString("status")

		var deadline *time.Time
		if deadlineStr != "" {
			t, err := time.Parse("2006-01-02", deadlineStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid deadline format (use YYYY-MM-DD): %v\n", err)
				os.Exit(1)
			}
			deadline = &t
		}

		svc := services.NewProjectService(cfg.DB)
		project, err := svc.Update(uint(id), args[1], description, budget, deadline, status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project updated: %s (ID: %d)\n", project.Name, project.ID)
	},
}

var updateBOMItemCmd = &cobra.Command{
	Use:   "bom-item [item_id] --quantity [num] --notes [text]",
	Short: "Update a Bill of Materials item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid item ID: %v\n", err)
			os.Exit(1)
		}

		quantity, _ := cmd.Flags().GetInt("quantity")
		notes, _ := cmd.Flags().GetString("notes")

		if quantity <= 0 {
			fmt.Fprintf(os.Stderr, "Error: --quantity is required and must be greater than 0\n")
			os.Exit(1)
		}

		svc := services.NewProjectService(cfg.DB)
		item, err := svc.UpdateBillOfMaterialsItem(uint(itemID), quantity, notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Bill of Materials item updated: %s (Quantity: %d, Item ID: %d)\n",
			item.Specification.Name, item.Quantity, item.ID)
	},
}

var updateProjectRequisitionCmd = &cobra.Command{
	Use:   "project-requisition [id] --name [name] --justification [text] --budget [amount]",
	Short: "Update a project requisition's details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		reqID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project requisition ID: %v\n", err)
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		justification, _ := cmd.Flags().GetString("justification")
		budget, _ := cmd.Flags().GetFloat64("budget")

		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: --name flag is required")
			os.Exit(1)
		}

		svc := services.NewProjectRequisitionService(cfg.DB)
		projectReq, err := svc.Update(uint(reqID), name, justification, budget)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project requisition updated successfully:\n")
		fmt.Printf("  ID: %d\n", projectReq.ID)
		fmt.Printf("  Name: %s\n", projectReq.Name)
		if projectReq.Justification != "" {
			fmt.Printf("  Justification: %s\n", projectReq.Justification)
		}
		if projectReq.Budget > 0 {
			fmt.Printf("  Budget: $%.2f\n", projectReq.Budget)
		}
		fmt.Printf("  Items: %d\n", len(projectReq.Items))
	},
}

func init() {
	updateCmd.AddCommand(updateSpecificationCmd)
	updateCmd.AddCommand(updateBrandCmd)
	updateCmd.AddCommand(updateProductCmd)
	updateCmd.AddCommand(updateVendorCmd)
	updateCmd.AddCommand(updateProjectCmd)
	updateCmd.AddCommand(updateBOMItemCmd)
	updateCmd.AddCommand(updateProjectRequisitionCmd)

	// Specification flags
	updateSpecificationCmd.Flags().String("description", "", "New description for the specification")

	// Project flags
	updateProjectCmd.Flags().String("description", "", "New description for the project")
	updateProjectCmd.Flags().Float64("budget", 0, "New budget for the project")
	updateProjectCmd.Flags().String("deadline", "", "New deadline (YYYY-MM-DD)")
	updateProjectCmd.Flags().String("status", "", "New status (planning, active, completed, cancelled)")

	// BOM item flags
	updateBOMItemCmd.Flags().Int("quantity", 0, "New quantity (required)")
	updateBOMItemCmd.Flags().String("notes", "", "New notes")

	// Project requisition flags
	updateProjectRequisitionCmd.Flags().String("name", "", "New name (required)")
	updateProjectRequisitionCmd.Flags().String("justification", "", "New justification")
	updateProjectRequisitionCmd.Flags().Float64("budget", 0, "New budget")
}
