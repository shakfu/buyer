package main

import (
	"fmt"
	"strings"

	"github.com/rodaine/table"
	"github.com/shakfu/buyer/internal/models"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search across all entities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.ToLower(args[0])

		// Search brands
		var brands []models.Brand
		cfg.DB.Where("LOWER(name) LIKE ?", "%"+query+"%").Find(&brands)

		// Search products
		var products []models.Product
		cfg.DB.Preload("Brand").Where("LOWER(name) LIKE ?", "%"+query+"%").Find(&products)

		// Search vendors
		var vendors []models.Vendor
		cfg.DB.Where("LOWER(name) LIKE ?", "%"+query+"%").Find(&vendors)

		// Display results
		totalResults := len(brands) + len(products) + len(vendors)
		if totalResults == 0 {
			fmt.Printf("No results found for '%s'\n", args[0])
			return
		}

		fmt.Printf("Search results for '%s' (%d total)\n\n", args[0], totalResults)

		if len(brands) > 0 {
			fmt.Println("Brands:")
			tbl := table.New("ID", "Name")
			for _, brand := range brands {
				tbl.AddRow(brand.ID, brand.Name)
			}
			tbl.Print()
			fmt.Println()
		}

		if len(products) > 0 {
			fmt.Println("Products:")
			tbl := table.New("ID", "Name", "Brand")
			for _, product := range products {
				brandName := ""
				if product.Brand != nil {
					brandName = product.Brand.Name
				}
				tbl.AddRow(product.ID, product.Name, brandName)
			}
			tbl.Print()
			fmt.Println()
		}

		if len(vendors) > 0 {
			fmt.Println("Vendors:")
			tbl := table.New("ID", "Name", "Currency")
			for _, vendor := range vendors {
				tbl.AddRow(vendor.ID, vendor.Name, vendor.Currency)
			}
			tbl.Print()
		}
	},
}
