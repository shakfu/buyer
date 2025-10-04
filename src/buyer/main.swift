import ArgumentParser
import Foundation
import Logging
import buyerlib

let logger = Logger(label: "org.me.swiftbuyer.main")

struct Buyer: ParsableCommand {
    static var configuration = CommandConfiguration(
        abstract: "A utility to help buying worthwhile things.",

        version: "0.0.1",

        // subcommands: [Status.self, Search.self, Report.self, Sound.self, Demo.self],
        subcommands: [Status.self, Search.self, Report.self, Demo.self],

        defaultSubcommand: Status.self
    )

    @Flag(name: .shortAndLong, help: "Set debug mode on")
    var debug: Bool = false

    //--------------------------------------------------------
    // SUBCOMMANDS

    struct Status: ParsableCommand {
        static var configuration = CommandConfiguration(
            abstract: "Print the procurement control tower summary."
        )

        mutating func run() throws {
            let service = try procurementService()
            let summary = try service.statusSummary()

            print("Procurement status as of \(formatTimestamp(summary.generatedAt))")
            print("----------------------------------------------")
            print("Active suppliers: \(summary.activeSuppliers)")
            print("Open purchase orders: \(summary.openPurchaseOrders)")
            print("Deliveries due this week: \(summary.deliveriesDueThisWeek)")
            print("Pending approvals: \(summary.pendingApprovals)")
            print("Invoices on hold: \(summary.invoicesOnHold)")
            print("")

            if summary.alerts.isEmpty {
                print("No active alerts. All signals green.")
            } else {
                print("Alerts (top \(min(summary.alerts.count, 5))):")
                for alert in summary.alerts.prefix(5) {
                    print("- [\(alert.severity.rawValue.uppercased())] \(alert.message)")
                }
            }

            logger.info("status summary printed")
        }
    }

    struct Search: ParsableCommand {
        static var configuration = CommandConfiguration(
            abstract: "Search the supplier master."
        )

        @Argument(help: "Text to match supplier name, country, or category.")
        var query: String

        @Option(name: [.short, .long], help: "Maximum search results to display.")
        var limit: Int = 5

        mutating func run() throws {
            guard limit >= 1 else {
                throw ValidationError("Limit must be at least 1.")
            }

            let service = try procurementService()
            let results = try service.searchSuppliers(matching: query, limit: limit)

            if results.isEmpty {
                print("No suppliers found for query '\(query)'.")
                return
            }

            print("Top \(results.count) suppliers matching '\(query)':")
            for supplier in results {
                let risk = supplier.riskRating ?? "Unrated"
                let spend = formatCurrency(amount: supplier.spendYearToDate, currency: "USD")
                let status = supplier.isActive ? "Active" : "Inactive"
                print("- \(supplier.legalName) [\(supplier.country)] — \(supplier.category) — Risk: \(risk) — Spend YTD: \(spend) — \(status)")
            }

            logger.info("search completed for query: \(query)")
        }
    }

    struct Report: ParsableCommand {
        static var configuration = CommandConfiguration(
            abstract: "Generate a procurement operations report."
        )

        @Option(name: [.short, .long], help: "Output path for the generated workbook (defaults to temp directory).")
        var output: String?

        mutating func run() throws {
            logger.info("report generation requested")
            let service = try procurementService()
            let outputURL = try generateOperationsWorkbook(using: service, overridePath: output)
            print("Report written to \(outputURL.path)")
        }
    }

    // struct Sound: ParsableCommand {
    //     static var configuration = CommandConfiguration(
    //         abstract: "Generate a sound."
    //     )

    //     mutating func run() {
    //         logger.info("generating sound now")
    //         demo_audiokit()

    //     }
    // }

    struct Demo: ParsableCommand {
        static var configuration = CommandConfiguration(
            abstract: "Demo a feature."
        )

        mutating func run() throws {
            logger.info("demonstrating feature now")
            let service = try procurementService()
            let supplierSummaries = try service.supplierSpendSummaries().prefix(3)
            print("Top suppliers by open PO value:")
            for summary in supplierSummaries {
                let spend = formatCurrency(amount: summary.totalOpenPOValue, currency: "USD")
                print("- \(summary.supplier.legalName): \(spend) open, \(summary.invoicesOnHold) invoices on hold, \(summary.overdueDeliveries) overdue deliveries")
            }
        }
    }
}

Buyer.main()

// MARK: - Helpers

private func procurementService() throws -> ProcurementService {
    let referenceDate = Date()
    let environment = ProcessInfo.processInfo.environment

    if let path = environment["BUYER_DB_PATH"], !path.isEmpty {
        let url = URL(fileURLWithPath: path)
        let repository = FileProcurementRepository(url: url, seedDate: referenceDate)
        return ProcurementService(repository: repository)
    }

    let cacheURL = FileManager.default
        .temporaryDirectory
        .appendingPathComponent("buyer-procurement")
        .appendingPathExtension("json")

    let repository = FileProcurementRepository(url: cacheURL, seedDate: referenceDate)
    return ProcurementService(repository: repository)
}

private func formatTimestamp(_ date: Date) -> String {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd HH:mm"
    return formatter.string(from: date)
}

private func formatCurrency(amount: Decimal, currency: String) -> String {
    let formatter = NumberFormatter()
    formatter.numberStyle = .currency
    formatter.currencyCode = currency
    formatter.maximumFractionDigits = 2
    formatter.minimumFractionDigits = 2
    let number = NSDecimalNumber(decimal: amount)
    return formatter.string(from: number) ?? "\(currency) \(number.doubleValue)"
}

private func generateOperationsWorkbook(using service: ProcurementService, overridePath: String?) throws -> URL {
    let summary = try service.statusSummary()
    let supplierSummaries = try service.supplierSpendSummaries()
    let dataset = try service.dataSet()

    let input = ProcurementReportInput(summary: summary,
                                       supplierSummaries: supplierSummaries,
                                       dataset: dataset)

    let environment = ProcessInfo.processInfo.environment
    let fileManager = FileManager.default

    let resolvedPath: String
    if let overridePath = overridePath, !overridePath.isEmpty {
        resolvedPath = overridePath
    } else if let envPath = environment["BUYER_REPORT_PATH"], !envPath.isEmpty {
        resolvedPath = envPath
    } else {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyyMMdd-HHmmss"
        let filename = "buyer-report-\(formatter.string(from: summary.generatedAt)).xlsx"
        resolvedPath = fileManager.temporaryDirectory.appendingPathComponent(filename).path
    }

    var outputURL = URL(fileURLWithPath: resolvedPath)
    if outputURL.pathExtension.lowercased() != "xlsx" {
        outputURL.deletePathExtension()
        outputURL.appendPathExtension("xlsx")
    }

    try fileManager.createDirectory(at: outputURL.deletingLastPathComponent(), withIntermediateDirectories: true)

    try writeProcurementWorkbook(input: input, to: outputURL)

    return outputURL
}
