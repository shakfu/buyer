import XCTest
@testable import buyerlib

final class ProcurementServiceTests: XCTestCase {
    private var referenceDate: Date!
    private var service: ProcurementService!

    override func setUpWithError() throws {
        referenceDate = ISO8601DateFormatter().date(from: "2024-06-01T12:00:00Z")
        guard let referenceDate else {
            throw XCTSkip("Unable to create reference date for tests")
        }
        let repository = InMemoryProcurementRepository(seedDate: referenceDate)
        service = ProcurementService(repository: repository, dateProvider: { referenceDate })
    }

    func testStatusSummaryMatchesSeedData() throws {
        let summary = try service.statusSummary()

        XCTAssertEqual(summary.activeSuppliers, 3)
        XCTAssertEqual(summary.openPurchaseOrders, 3)
        XCTAssertEqual(summary.deliveriesDueThisWeek, 1)
        XCTAssertEqual(summary.pendingApprovals, 2)
        XCTAssertEqual(summary.invoicesOnHold, 2)
        XCTAssertEqual(summary.alerts.count, 5)
        XCTAssertTrue(summary.alerts.contains(where: { $0.message.contains("PO-1002") }))
    }

    func testSearchSuppliersByCategory() throws {
        let results = try service.searchSuppliers(matching: "Steel")
        XCTAssertEqual(results.count, 1)
        XCTAssertEqual(results.first?.legalName, "Atlas Steelworks")
    }

    func testSupplierSpendSummariesAreOrdered() throws {
        let summaries = try service.supplierSpendSummaries()
        XCTAssertEqual(summaries.first?.supplier.legalName, "Atlas Steelworks")
        XCTAssertEqual(summaries.first?.totalOpenPOValue, Decimal(string: "820000.00"))
        XCTAssertEqual(summaries.first?.overdueDeliveries, 1)
        XCTAssertEqual(summaries.first?.invoicesOnHold, 1)
    }

    func testFileRepositoryMatchesInMemory() throws {
        let tempURL = FileManager.default
            .temporaryDirectory
            .appendingPathComponent("procurement-tests-\(UUID().uuidString)")
            .appendingPathExtension("json")
        defer { try? FileManager.default.removeItem(at: tempURL) }

        guard let referenceDate else {
            XCTFail("Reference date missing")
            return
        }

        let fileRepository = FileProcurementRepository(url: tempURL, seedDate: referenceDate)
        let fileBackedService = ProcurementService(repository: fileRepository, dateProvider: { referenceDate })

        let summary = try fileBackedService.statusSummary()
        XCTAssertEqual(summary.activeSuppliers, 3)
        XCTAssertEqual(summary.openPurchaseOrders, 3)

        // Rehydrate from disk to ensure persisted dataset matches
        let reloadedRepository = FileProcurementRepository(url: tempURL, seedDate: referenceDate)
        let reloadedService = ProcurementService(repository: reloadedRepository, dateProvider: { referenceDate })
        let reloadedSummary = try reloadedService.statusSummary()
        XCTAssertEqual(reloadedSummary.pendingApprovals, 2)
        let searchResults = try reloadedService.searchSuppliers(matching: "Electrical")
        XCTAssertEqual(searchResults.first?.legalName, "Skyline Electrical")
    }

    func testWorkbookGenerationProducesFile() throws {
        guard let referenceDate else {
            XCTFail("Reference date missing")
            return
        }

        let summary = try service.statusSummary()
        let supplierSummaries = try service.supplierSpendSummaries()
        let dataset = try service.dataSet()
        let input = ProcurementReportInput(summary: summary,
                                           supplierSummaries: supplierSummaries,
                                           dataset: dataset)

        let fileURL = FileManager.default.temporaryDirectory.appendingPathComponent("procurement-report-\(UUID().uuidString).xlsx")
        defer { try? FileManager.default.removeItem(at: fileURL) }

        try writeProcurementWorkbook(input: input, to: fileURL)

        let attributes = try FileManager.default.attributesOfItem(atPath: fileURL.path)
        let fileSize = (attributes[.size] as? NSNumber)?.intValue ?? 0
        XCTAssertGreaterThan(fileSize, 0)
    }

    static var allTests = [
        ("testStatusSummaryMatchesSeedData", testStatusSummaryMatchesSeedData),
        ("testSearchSuppliersByCategory", testSearchSuppliersByCategory),
        ("testSupplierSpendSummariesAreOrdered", testSupplierSpendSummariesAreOrdered),
        ("testFileRepositoryMatchesInMemory", testFileRepositoryMatchesInMemory),
        ("testWorkbookGenerationProducesFile", testWorkbookGenerationProducesFile),
    ]
}
