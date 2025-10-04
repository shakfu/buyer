import ArgumentParser
import Foundation
import Logging
import buyerlib

enum StoreKind: String, ExpressibleByArgument {
    case sqlite
    case json

    static func inferred(from path: String) -> StoreKind {
        let trimmed = path.lowercased()
        if trimmed.hasSuffix(".json") { return .json }
        if trimmed.hasSuffix(".sqlite") || trimmed.hasSuffix(".sqlite3") || trimmed.hasSuffix(".db") {
            return .sqlite
        }
        return .sqlite
    }
}

struct StoreOptions: ParsableArguments {
    @Option(name: [.customShort("s"), .long], help: "Datastore backend to use (sqlite or json).")
    var store: StoreKind?

    @Option(name: [.customShort("d"), .long], help: "Override path to the datastore file.")
    var database: String?
}

let logger = Logger(label: "org.me.swiftbuyer.main")

struct Buyer: ParsableCommand {
    static var configuration = CommandConfiguration(
        abstract: "A utility to help buying worthwhile things.",

        version: "0.0.1",

        // subcommands: [Status.self, Search.self, Report.self, Sound.self, Demo.self],
        subcommands: [Status.self, Search.self, Report.self, Demo.self, Fixtures.self],

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

        @OptionGroup var storeOptions: StoreOptions

        mutating func run() throws {
            let service = try procurementService(storeKind: storeOptions.store,
                                                storePath: storeOptions.database)
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

        @OptionGroup var storeOptions: StoreOptions

        @Argument(help: "Text to match supplier name, country, or category.")
        var query: String

        @Option(name: [.short, .long], help: "Maximum search results to display.")
        var limit: Int = 5

        mutating func run() throws {
            guard limit >= 1 else {
                throw ValidationError("Limit must be at least 1.")
            }

            let service = try procurementService(storeKind: storeOptions.store,
                                                storePath: storeOptions.database)
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

        @OptionGroup var storeOptions: StoreOptions

        @Option(name: [.short, .long], help: "Output path for the generated workbook (defaults to temp directory).")
        var output: String?

        mutating func run() throws {
            logger.info("report generation requested")
            let service = try procurementService(storeKind: storeOptions.store,
                                                storePath: storeOptions.database)
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

        @OptionGroup var storeOptions: StoreOptions

        mutating func run() throws {
            logger.info("demonstrating feature now")
            let service = try procurementService(storeKind: storeOptions.store,
                                                storePath: storeOptions.database)
            let supplierSummaries = try service.supplierSpendSummaries().prefix(3)
            print("Top suppliers by open PO value:")
            for summary in supplierSummaries {
                let spend = formatCurrency(amount: summary.totalOpenPOValue, currency: "USD")
                print("- \(summary.supplier.legalName): \(spend) open, \(summary.invoicesOnHold) invoices on hold, \(summary.overdueDeliveries) overdue deliveries")
            }
        }
    }

    struct Fixtures: ParsableCommand {
        static var configuration = CommandConfiguration(
            abstract: "Generate datastore fixtures for demos or tests."
        )

        @OptionGroup var storeOptions: StoreOptions

        @Option(name: [.short, .long], help: "Output path for the generated fixture. Defaults under ./fixtures.")
        var output: String?

        @Option(name: .long, help: "Seed timestamp in ISO-8601 (defaults to 2024-06-01T12:00:00Z).")
        var seedDate: String?

        @Flag(name: [.customShort("f"), .long], help: "Overwrite the fixture if it already exists.")
        var overwrite: Bool = false

        mutating func run() throws {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            let defaultSeedString = "2024-06-01T12:00:00Z"
            let seedString = seedDate ?? defaultSeedString
            guard let seededDate = formatter.date(from: seedString) else {
                throw ValidationError("Unable to parse seed date '\(seedString)'. Use ISO-8601, e.g. 2024-06-01T12:00:00Z")
            }

            let store = storeOptions.store ?? .sqlite
            let resolvedOutput = output ?? defaultFixturePath(for: store)
            let outputURL = URL(fileURLWithPath: resolvedOutput)
            let fileManager = FileManager.default

            if fileManager.fileExists(atPath: outputURL.path), !overwrite {
                throw ValidationError("Fixture already exists at \(outputURL.path). Pass --overwrite to replace it.")
            }

            if overwrite, fileManager.fileExists(atPath: outputURL.path) {
                try fileManager.removeItem(at: outputURL)
            }

            try fileManager.createDirectory(at: outputURL.deletingLastPathComponent(), withIntermediateDirectories: true)

            let tenantID = UUID(uuidString: "00000000-0000-0000-0000-000000000001")!
            let dataset = ProcurementSeedData.makeDataSet(seedDate: seededDate, tenantID: tenantID)

            switch store {
            case .json:
                let repository = FileProcurementRepository(url: outputURL, seedDate: seededDate)
                // Persist seed dataset explicitly to preserve stable order.
                let encoder = JSONEncoder()
                encoder.outputFormatting = [.sortedKeys, .prettyPrinted]
                encoder.dateEncodingStrategy = .iso8601
                let data = try encoder.encode(dataset)
                try data.write(to: outputURL, options: .atomic)
            case .sqlite:
                _ = try SQLiteProcurementRepository(path: outputURL.path, seedDate: seededDate)
            }

            print("Fixture written to \(outputURL.path)")
        }

        private func defaultFixturePath(for store: StoreKind) -> String {
            switch store {
            case .sqlite:
                return "fixtures/procurement.sqlite3"
            case .json:
                return "fixtures/procurement.json"
            }
        }
    }
}

Buyer.main()

// MARK: - Helpers

private func procurementService(storeKind: StoreKind? = nil, storePath: String? = nil) throws -> ProcurementService {
    let referenceDate = Date()
    let environment = ProcessInfo.processInfo.environment

    let providedPath = storePath ?? environment["BUYER_DB_PATH"]

    var resolvedKind = storeKind
    if resolvedKind == nil, let storeEnv = environment["BUYER_STORE"], let envKind = StoreKind(rawValue: storeEnv.lowercased()) {
        resolvedKind = envKind
    }
    if resolvedKind == nil, let path = providedPath {
        resolvedKind = StoreKind.inferred(from: path)
    }

    let selectedKind = resolvedKind ?? .sqlite
    let targetURL = URL(fileURLWithPath: providedPath ?? defaultStorePath(for: selectedKind))

    switch selectedKind {
    case .sqlite:
        do {
            let repository = try SQLiteProcurementRepository(path: targetURL.path, seedDate: referenceDate)
            return ProcurementService(repository: repository)
        } catch {
            logger.warning("Failed to open SQLite store at \(targetURL.path). Error: \(error). Falling back to JSON store.")
            let fallbackURL = providedPath == nil ? defaultStoreURL(for: .json) : targetURL.deletingPathExtension().appendingPathExtension("json")
            let repository = FileProcurementRepository(url: fallbackURL, seedDate: referenceDate)
            return ProcurementService(repository: repository)
        }
    case .json:
        let repository = FileProcurementRepository(url: targetURL, seedDate: referenceDate)
        return ProcurementService(repository: repository)
    }
}

private func defaultStorePath(for store: StoreKind) -> String {
    defaultStoreURL(for: store).path
}

private func defaultStoreURL(for store: StoreKind) -> URL {
    let tempDirectory = FileManager.default.temporaryDirectory
    switch store {
    case .sqlite:
        return tempDirectory.appendingPathComponent("buyer-procurement.sqlite3")
    case .json:
        return tempDirectory.appendingPathComponent("buyer-procurement.json")
    }
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
