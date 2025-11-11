package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update entities (specification, brand, product, vendor)",
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

func init() {
	updateCmd.AddCommand(updateSpecificationCmd)
	updateCmd.AddCommand(updateBrandCmd)
	updateCmd.AddCommand(updateProductCmd)
	updateCmd.AddCommand(updateVendorCmd)

	// Specification flags
	updateSpecificationCmd.Flags().String("description", "", "New description for the specification")
}
