package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rodaine/table"
	"github.com/shakfu/buyer/internal/models"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var strategyCmd = &cobra.Command{
	Use:   "strategy",
	Short: "Manage procurement strategies",
	Long:  "View and manage project procurement strategies",
}

var strategyShowCmd = &cobra.Command{
	Use:   "show <project-id>",
	Short: "Show current procurement strategy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project ID: %v\n", err)
			os.Exit(1)
		}
		projectID := uint(id)

		var strategy models.ProjectProcurementStrategy
		err = cfg.DB.Where("project_id = ?", projectID).First(&strategy).Error
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: No strategy found for project %d\n", projectID)
			os.Exit(1)
		}

		fmt.Printf("\nProcurement Strategy for Project ID %d\n", projectID)
		fmt.Println("=" + string(make([]byte, 50)))
		fmt.Printf("Strategy: %s\n", strategy.Strategy)
		if strategy.MaxVendors != nil {
			fmt.Printf("Max Vendors: %d\n", *strategy.MaxVendors)
		}
		if strategy.MinVendorRating != nil {
			fmt.Printf("Min Vendor Rating: %.1f\n", *strategy.MinVendorRating)
		}
		fmt.Printf("Allow Partial Fulfill: %v\n", strategy.AllowPartialFulfill)
	},
}

var strategySetCmd = &cobra.Command{
	Use:   "set <project-id> <type>",
	Short: "Set procurement strategy type",
	Long:  "Set strategy type: lowest_cost, fewest_vendors, balanced, quality_focused",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project ID: %v\n", err)
			os.Exit(1)
		}
		projectID := uint(id)

		strategyType := args[1]
		validTypes := map[string]bool{
			"lowest_cost":     true,
			"fewest_vendors":  true,
			"balanced":        true,
			"quality_focused": true,
		}
		if !validTypes[strategyType] {
			fmt.Fprintf(os.Stderr, "Error: Invalid strategy type\n")
			os.Exit(1)
		}

		var strategy models.ProjectProcurementStrategy
		err = cfg.DB.Where("project_id = ?", projectID).First(&strategy).Error
		if err != nil {
			strategy = models.ProjectProcurementStrategy{
				ProjectID: projectID,
				Strategy:  strategyType,
			}
			if err := cfg.DB.Create(&strategy).Error; err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Strategy created: %s\n", strategyType)
		} else {
			strategy.Strategy = strategyType
			if err := cfg.DB.Save(&strategy).Error; err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Strategy updated: %s\n", strategyType)
		}
	},
}

var strategyCompareCmd = &cobra.Command{
	Use:   "compare <project-id>",
	Short: "Compare all strategy scenarios",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project ID: %v\n", err)
			os.Exit(1)
		}
		projectID := uint(id)

		quoteSvc := services.NewQuoteService(cfg.DB)
		projectSvc := services.NewProjectService(cfg.DB)
		procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

		scenarios, err := procurementSvc.CompareScenarios(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nScenario Comparison for Project ID %d\n", projectID)
		fmt.Println("=" + string(make([]byte, 70)))

		tbl := table.New("Strategy", "Cost", "Vendors", "Savings")
		for _, sc := range scenarios {
			tbl.AddRow(
				sc.Name,
				fmt.Sprintf("$%.2f", sc.TotalCost),
				sc.VendorCount,
				fmt.Sprintf("$%.2f", sc.SavingsVsBudget),
			)
		}
		tbl.Print()
	},
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Generate vendor recommendations",
}

var recommendGenerateCmd = &cobra.Command{
	Use:   "generate <project-id>",
	Short: "Generate vendor recommendations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid project ID: %v\n", err)
			os.Exit(1)
		}
		projectID := uint(id)

		strategyType, _ := cmd.Flags().GetString("strategy")
		if strategyType == "" {
			strategyType = "balanced"
		}

		quoteSvc := services.NewQuoteService(cfg.DB)
		projectSvc := services.NewProjectService(cfg.DB)
		procurementSvc := services.NewProjectProcurementService(cfg.DB, quoteSvc, projectSvc)

		recommendations, err := procurementSvc.GenerateVendorRecommendations(projectID, strategyType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nVendor Recommendations (%s strategy)\n", strategyType)
		if len(recommendations) == 0 {
			fmt.Println("No recommendations")
			return
		}

		for _, rec := range recommendations {
			fmt.Printf("\nVendor: %s (ID: %d)\n", rec.VendorName, rec.VendorID)
			fmt.Printf("Total Cost: $%.2f\n", rec.TotalCost)
			fmt.Printf("Item Count: %d\n", rec.ItemCount)
			if rec.Rationale != "" {
				fmt.Printf("Rationale: %s\n", rec.Rationale)
			}
		}
	},
}

func init() {
	strategyCmd.AddCommand(strategyShowCmd)
	strategyCmd.AddCommand(strategySetCmd)
	strategyCmd.AddCommand(strategyCompareCmd)

	recommendCmd.AddCommand(recommendGenerateCmd)
	recommendGenerateCmd.Flags().String("strategy", "balanced", "Strategy type")

	projectProcurementCmd.AddCommand(strategyCmd)
	projectProcurementCmd.AddCommand(recommendCmd)
}
