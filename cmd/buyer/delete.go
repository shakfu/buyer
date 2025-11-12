package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete entities (specification, brand, product, vendor, quote, forex, requisition, project, bom-item, project-requisition)",
	Long:  "Delete entities by ID with confirmation",
}

var deleteSpecificationCmd = &cobra.Command{
	Use:   "specification [id]",
	Short: "Delete a specification",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("specification", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewSpecificationService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Specification ID %d deleted successfully.\n", id)
	},
}

var deleteRequisitionCmd = &cobra.Command{
	Use:   "requisition [id]",
	Short: "Delete a requisition",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("requisition", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewRequisitionService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Requisition ID %d deleted successfully.\n", id)
	},
}

var deleteRequisitionItemCmd = &cobra.Command{
	Use:   "requisition-item [id]",
	Short: "Delete a requisition line item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("requisition item", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewRequisitionService(cfg.DB)
		if err := svc.DeleteItem(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Requisition item ID %d deleted successfully.\n", id)
	},
}

var deleteBrandCmd = &cobra.Command{
	Use:   "brand [id]",
	Short: "Delete a brand",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("brand", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewBrandService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Brand deleted (ID: %d)\n", id)
	},
}

var deleteProductCmd = &cobra.Command{
	Use:   "product [id]",
	Short: "Delete a product",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("product", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewProductService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Product deleted (ID: %d)\n", id)
	},
}

var deleteVendorCmd = &cobra.Command{
	Use:   "vendor [id]",
	Short: "Delete a vendor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("vendor", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewVendorService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Vendor deleted (ID: %d)\n", id)
	},
}

var deleteQuoteCmd = &cobra.Command{
	Use:   "quote [id]",
	Short: "Delete a quote",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("quote", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewQuoteService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Quote deleted (ID: %d)\n", id)
	},
}

var deleteForexCmd = &cobra.Command{
	Use:   "forex [id]",
	Short: "Delete a forex rate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("forex rate", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewForexService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Forex rate deleted (ID: %d)\n", id)
	},
}

var deleteProjectCmd = &cobra.Command{
	Use:   "project [id]",
	Short: "Delete a project (also deletes associated BOM and project requisitions)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("project", uint(id)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewProjectService(cfg.DB)
		if err := svc.Delete(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project deleted (ID: %d)\n", id)
	},
}

var deleteBOMItemCmd = &cobra.Command{
	Use:   "bom-item [item_id]",
	Short: "Delete a Bill of Materials item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		itemID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid item ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("BOM item", uint(itemID)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewProjectService(cfg.DB)
		if err := svc.DeleteBillOfMaterialsItem(uint(itemID)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Bill of Materials item deleted (ID: %d)\n", itemID)
	},
}

var deleteProjectRequisitionCmd = &cobra.Command{
	Use:   "project-requisition [id]",
	Short: "Delete a project requisition",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		reqID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project requisition ID: %v\n", err)
			os.Exit(1)
		}

		if !force && !confirmDelete("project requisition", uint(reqID)) {
			fmt.Println("Deletion cancelled.")
			return
		}

		svc := services.NewProjectRequisitionService(cfg.DB)
		if err := svc.Delete(uint(reqID)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project requisition deleted (ID: %d)\n", reqID)
	},
}

func confirmDelete(entity string, id uint) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to delete %s with ID %d? (y/N): ", entity, id)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func init() {
	deleteCmd.AddCommand(deleteSpecificationCmd)
	deleteCmd.AddCommand(deleteBrandCmd)
	deleteCmd.AddCommand(deleteProductCmd)
	deleteCmd.AddCommand(deleteVendorCmd)
	deleteCmd.AddCommand(deleteQuoteCmd)
	deleteCmd.AddCommand(deleteForexCmd)
	deleteCmd.AddCommand(deleteRequisitionCmd)
	deleteCmd.AddCommand(deleteRequisitionItemCmd)
	deleteCmd.AddCommand(deleteProjectCmd)
	deleteCmd.AddCommand(deleteBOMItemCmd)
	deleteCmd.AddCommand(deleteProjectRequisitionCmd)

	// Add force flag to all delete commands
	for _, cmd := range []*cobra.Command{deleteSpecificationCmd, deleteBrandCmd, deleteProductCmd, deleteVendorCmd, deleteQuoteCmd, deleteForexCmd, deleteRequisitionCmd, deleteRequisitionItemCmd, deleteProjectCmd, deleteBOMItemCmd, deleteProjectRequisitionCmd} {
		cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	}
}
