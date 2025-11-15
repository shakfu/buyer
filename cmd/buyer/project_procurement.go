package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rodaine/table"
	"github.com/shakfu/buyer/internal/services"
	"github.com/spf13/cobra"
)

var projectProcurementCmd = &cobra.Command{
	Use:   "procurement",
	Short: "Project procurement analysis and optimization",
	Long:  "Analyze procurement options, compare vendors, manage strategies, and apply recommendations",
}

var procurementAnalyzeCmd = &cobra.Command{
	Use:   "analyze <project-id>",
	Short: "Analyze procurement options for a project",
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

		comparison, err := procurementSvc.GetProjectProcurementComparison(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nProcurement Analysis for Project: %s\n", comparison.Project.Name)
		fmt.Printf("Budget: $%.2f\n", comparison.Project.Budget)
		if comparison.Project.Deadline != nil {
			fmt.Printf("Deadline: %s\n", comparison.Project.Deadline.Format("2006-01-02"))
		}
		fmt.Println()

		// BOM Items Summary
		fmt.Println("Bill of Materials Analysis:")
		tbl := table.New("Item", "Specification", "Qty", "Quotes", "Best Price", "Risk")
		for _, item := range comparison.BOMItemAnalyses {
			bestPrice := "N/A"
			if item.BestQuote != nil {
				bestPrice = fmt.Sprintf("$%.2f", item.BestQuote.ConvertedPrice)
			}
			specName := "N/A"
			if item.Specification != nil {
				specName = item.Specification.Name
			}
			tbl.AddRow(
				item.BOMItem.ID,
				specName,
				item.BOMItem.Quantity,
				len(item.AvailableQuotes),
				bestPrice,
				item.RiskLevel,
			)
		}
		tbl.Print()

		// Summary
		fmt.Printf("\nTotal BOM Items: %d\n", comparison.TotalBOMItems)
		fmt.Printf("Fully Covered: %d, Partially Covered: %d, Uncovered: %d\n",
			comparison.FullyCoveredItems, comparison.PartiallyCoveredItems, comparison.UncoveredItems)
		fmt.Printf("Total Vendors Needed: %d\n", comparison.TotalVendorsNeeded)
	},
}

var procurementRisksCmd = &cobra.Command{
	Use:   "risks <project-id>",
	Short: "Assess procurement risks for a project",
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

		risks, err := procurementSvc.AssessEnhancedProjectRisks(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nRisk Assessment for Project ID %d\n", projectID)
		fmt.Printf("Overall Risk: %s (Score: %d/100)\n\n", risks.OverallRisk, risks.RiskScore)

		// Risk Categories
		fmt.Println("Risk Categories:")
		tbl := table.New("Category", "Level", "Issues")
		for category, catRisk := range risks.CategoryRisks {
			issueCount := len(catRisk.Issues)
			tbl.AddRow(category, catRisk.Level, issueCount)
		}
		tbl.Print()

		// Mitigation Actions
		if len(risks.MitigationActions) > 0 {
			fmt.Println("\nRecommended Mitigation Actions:")
			highPriority := 0
			mediumPriority := 0
			for _, action := range risks.MitigationActions {
				if action.Priority == "high" {
					highPriority++
				} else if action.Priority == "medium" {
					mediumPriority++
				}
			}
			fmt.Printf("  High Priority: %d\n", highPriority)
			fmt.Printf("  Medium Priority: %d\n", mediumPriority)
			fmt.Printf("  Total Actions: %d\n", len(risks.MitigationActions))
		}
	},
}

var procurementDashboardCmd = &cobra.Command{
	Use:   "dashboard <project-id>",
	Short: "Show project procurement dashboard",
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

		dashboard, err := procurementSvc.GetProjectDashboard(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDashboard for Project: %s\n", dashboard.Project.Name)
		fmt.Println("=" + string(make([]byte, 60)))

		// Progress
		fmt.Println("\nProgress:")
		fmt.Printf("  BOM Coverage: %.1f%%\n", dashboard.Progress.BOMCoverage)
		fmt.Printf("  Requisitions: %d/%d complete\n", dashboard.Progress.RequisitionsComplete, dashboard.Progress.RequisitionsTotal)
		fmt.Printf("  Orders: %d placed, %d received\n", dashboard.Progress.OrdersPlaced, dashboard.Progress.OrdersReceived)
		fmt.Printf("  Timeline: %s (%d days to deadline)\n", dashboard.Progress.TimelineStatus, dashboard.Progress.DaysToDeadline)

		// Financial
		fmt.Println("\nFinancial:")
		fmt.Printf("  Budget: $%.2f\n", dashboard.Financial.Budget)
		fmt.Printf("  Committed: $%.2f\n", dashboard.Financial.Committed)
		fmt.Printf("  Estimated: $%.2f\n", dashboard.Financial.Estimated)
		fmt.Printf("  Remaining: $%.2f\n", dashboard.Financial.Remaining)
		fmt.Printf("  Savings: $%.2f (%.1f%%)\n", dashboard.Financial.Savings, dashboard.Financial.SavingsPercent)
		fmt.Printf("  Health: %s\n", dashboard.Financial.BudgetHealth)

		// Procurement Status
		fmt.Println("\nProcurement Status:")
		fmt.Printf("  Total Items: %d\n", dashboard.Procurement.TotalItems)
		fmt.Printf("  Items with Quotes: %d\n", dashboard.Procurement.ItemsWithQuotes)
		fmt.Printf("  Items Ordered: %d\n", dashboard.Procurement.ItemsOrdered)
		fmt.Printf("  Items Received: %d\n", dashboard.Procurement.ItemsReceived)
		fmt.Printf("  Vendors Engaged: %d\n", dashboard.Procurement.VendorsEngaged)
		fmt.Printf("  Quote Freshness: %s\n", dashboard.Procurement.QuoteFreshness)

		// Vendor Performance
		if len(dashboard.VendorPerformance) > 0 {
			fmt.Println("\nVendor Performance:")
			tbl := table.New("Vendor", "Items", "Value", "Rating", "On-Time %")
			for _, vp := range dashboard.VendorPerformance {
				tbl.AddRow(
					vp.VendorName,
					vp.ItemsSupplied,
					fmt.Sprintf("$%.2f", vp.TotalValue),
					fmt.Sprintf("%.1f", vp.AverageRating),
					fmt.Sprintf("%.0f%%", vp.OnTimeDelivery),
				)
			}
			tbl.Print()
		}

		// Risks
		if len(dashboard.Risks) > 0 {
			fmt.Println("\nTop Risks:")
			for i, risk := range dashboard.Risks {
				if i >= 3 {
					break
				}
				fmt.Printf("  %d. [%s] %s: %s\n", i+1, risk.Level, risk.Category, risk.TopIssue)
			}
		}
	},
}

var procurementSavingsCmd = &cobra.Command{
	Use:   "savings <project-id>",
	Short: "Calculate potential savings for a project",
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

		savings, err := procurementSvc.CalculateProjectSavings(projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSavings Analysis for Project ID %d\n", projectID)
		fmt.Println("=" + string(make([]byte, 50)))
		fmt.Printf("\nTotal Savings: $%.2f (%.1f%%)\n", savings.TotalSavingsUSD, savings.SavingsPercent)
		fmt.Printf("Consolidation Savings: $%.2f\n", savings.ConsolidationSavings)

		if len(savings.SavingsByCategory) > 0 {
			fmt.Println("\nSavings by Category:")
			tbl := table.New("Category", "Savings")
			for category, amount := range savings.SavingsByCategory {
				if amount > 0 {
					tbl.AddRow(category, fmt.Sprintf("$%.2f", amount))
				}
			}
			tbl.Print()
		}

		if len(savings.DetailedBreakdown) > 0 {
			fmt.Println("\nDetailed Breakdown:")
			tbl := table.New("Specification", "Qty", "Target", "Best", "Savings")
			for _, item := range savings.DetailedBreakdown {
				if item.TotalSavings > 0 {
					tbl.AddRow(
						truncate(item.SpecificationName, 30),
						item.Quantity,
						fmt.Sprintf("$%.2f", item.TargetPrice),
						fmt.Sprintf("$%.2f", item.BestPrice),
						fmt.Sprintf("$%.2f", item.TotalSavings),
					)
				}
			}
			tbl.Print()
		}
	},
}

func init() {
	// Add procurement subcommands
	projectProcurementCmd.AddCommand(procurementAnalyzeCmd)
	projectProcurementCmd.AddCommand(procurementRisksCmd)
	projectProcurementCmd.AddCommand(procurementDashboardCmd)
	projectProcurementCmd.AddCommand(procurementSavingsCmd)

	// Add to root
	rootCmd.AddCommand(projectProcurementCmd)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
