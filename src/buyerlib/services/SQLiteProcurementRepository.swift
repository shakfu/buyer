import Foundation
import SQLiteSupport

public final class SQLiteProcurementRepository: ProcurementRepository {
    private let db: Connection
    private let dateFormatter: ISO8601DateFormatter
    private let tenantID: UUID

    // suppliers
    private let suppliersTable = Table("suppliers")
    private let supplierID = Expression<String>("id")
    private let supplierTenantID = Expression<String>("tenant_id")
    private let supplierName = Expression<String>("legal_name")
    private let supplierCountry = Expression<String>("country")
    private let supplierCategory = Expression<String>("category")
    private let supplierIsActive = Expression<Bool>("is_active")
    private let supplierRisk = Expression<String?>("risk_rating")
    private let supplierSpendYTD = Expression<Double>("spend_ytd")

    // purchase orders
    private let purchaseOrdersTable = Table("purchase_orders")
    private let purchaseOrderID = Expression<String>("id")
    private let purchaseOrderTenantID = Expression<String>("tenant_id")
    private let purchaseOrderNumber = Expression<String>("number")
    private let purchaseOrderSupplierID = Expression<String>("supplier_id")
    private let purchaseOrderSupplierName = Expression<String>("supplier_name")
    private let purchaseOrderProjectCode = Expression<String>("project_code")
    private let purchaseOrderProjectName = Expression<String>("project_name")
    private let purchaseOrderStatus = Expression<String>("status")
    private let purchaseOrderCurrency = Expression<String>("currency")
    private let purchaseOrderTotal = Expression<Double>("total_value")
    private let purchaseOrderExpectedDelivery = Expression<String>("expected_delivery")
    private let purchaseOrderIssuedAt = Expression<String>("issued_at")

    // approval queue
    private let approvalsTable = Table("approvals")
    private let approvalID = Expression<String>("id")
    private let approvalTenantID = Expression<String>("tenant_id")
    private let approvalTitle = Expression<String>("title")
    private let approvalType = Expression<String>("request_type")
    private let approvalRequestedBy = Expression<String>("requested_by")
    private let approvalPendingWith = Expression<String>("pending_with")
    private let approvalDueDate = Expression<String>("due_date")
    private let approvalStatus = Expression<String>("status")

    // deliveries
    private let deliveriesTable = Table("deliveries")
    private let deliveryID = Expression<String>("id")
    private let deliveryTenantID = Expression<String>("tenant_id")
    private let deliveryPONumber = Expression<String>("po_number")
    private let deliveryDescription = Expression<String>("description")
    private let deliveryExpectedOn = Expression<String>("expected_on")
    private let deliveryStatus = Expression<String>("status")

    // invoices
    private let invoicesTable = Table("invoices")
    private let invoiceID = Expression<String>("id")
    private let invoiceTenantID = Expression<String>("tenant_id")
    private let invoiceSupplierID = Expression<String>("supplier_id")
    private let invoiceSupplierName = Expression<String>("supplier_name")
    private let invoiceNumber = Expression<String>("invoice_number")
    private let invoiceAmount = Expression<Double>("amount")
    private let invoiceCurrency = Expression<String>("currency")
    private let invoiceDueDate = Expression<String>("due_date")
    private let invoiceStatus = Expression<String>("status")

    public init(path: String,
                seedDate: Date = Date(),
                tenantID: UUID = UUID(uuidString: "00000000-0000-0000-0000-000000000001")!) throws {
        self.db = try Connection(path)
        self.db.busyTimeout = 2.0
        self.dateFormatter = ISO8601DateFormatter()
        self.dateFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        self.tenantID = tenantID

        try self.configureDatabase()
        try self.seedIfNeeded(seedDate: seedDate)
    }

    public func loadData() throws -> ProcurementDataSet {
        let suppliers = try db.prepare(suppliersTable).map { row in
            SupplierRecord(
                id: try parseUUID(row[supplierID], label: "supplier id"),
                tenantID: try parseUUID(row[supplierTenantID], label: "supplier tenant id"),
                legalName: row[supplierName],
                country: row[supplierCountry],
                category: row[supplierCategory],
                isActive: row[supplierIsActive],
                riskRating: row[supplierRisk],
                spendYearToDate: Decimal(row[supplierSpendYTD])
            )
        }

        let purchaseOrders = try db.prepare(purchaseOrdersTable).map { row in
            PurchaseOrderRecord(
                id: try parseUUID(row[purchaseOrderID], label: "purchase order id"),
                tenantID: try parseUUID(row[purchaseOrderTenantID], label: "purchase order tenant id"),
                number: row[purchaseOrderNumber],
                supplierID: try parseUUID(row[purchaseOrderSupplierID], label: "po supplier id"),
                supplierName: row[purchaseOrderSupplierName],
                projectCode: row[purchaseOrderProjectCode],
                projectName: row[purchaseOrderProjectName],
                status: PurchaseOrderStatus(rawValue: row[purchaseOrderStatus]) ?? .draft,
                currency: row[purchaseOrderCurrency],
                totalValue: Decimal(row[purchaseOrderTotal]),
                expectedDelivery: try parseDate(row[purchaseOrderExpectedDelivery], label: "expected delivery"),
                issuedAt: try parseDate(row[purchaseOrderIssuedAt], label: "issued at")
            )
        }

        let approvals = try db.prepare(approvalsTable).map { row in
            ApprovalQueueItem(
                id: try parseUUID(row[approvalID], label: "approval id"),
                tenantID: try parseUUID(row[approvalTenantID], label: "approval tenant id"),
                title: row[approvalTitle],
                requestType: row[approvalType],
                requestedBy: row[approvalRequestedBy],
                pendingWith: row[approvalPendingWith],
                dueDate: try parseDate(row[approvalDueDate], label: "approval due date"),
                status: ApprovalStatus(rawValue: row[approvalStatus]) ?? .pending
            )
        }

        let deliveries = try db.prepare(deliveriesTable).map { row in
            DeliveryMilestoneRecord(
                id: try parseUUID(row[deliveryID], label: "delivery id"),
                tenantID: try parseUUID(row[deliveryTenantID], label: "delivery tenant id"),
                purchaseOrderNumber: row[deliveryPONumber],
                description: row[deliveryDescription],
                expectedOn: try parseDate(row[deliveryExpectedOn], label: "delivery expected"),
                status: DeliveryStatus(rawValue: row[deliveryStatus]) ?? .scheduled
            )
        }

        let invoices = try db.prepare(invoicesTable).map { row in
            InvoiceRecord(
                id: try parseUUID(row[invoiceID], label: "invoice id"),
                tenantID: try parseUUID(row[invoiceTenantID], label: "invoice tenant id"),
                supplierID: try parseUUID(row[invoiceSupplierID], label: "invoice supplier id"),
                supplierName: row[invoiceSupplierName],
                invoiceNumber: row[invoiceNumber],
                amount: Decimal(row[invoiceAmount]),
                currency: row[invoiceCurrency],
                dueDate: try parseDate(row[invoiceDueDate], label: "invoice due"),
                status: InvoiceStatus(rawValue: row[invoiceStatus]) ?? .pending
            )
        }

        return ProcurementDataSet(
            suppliers: suppliers,
            purchaseOrders: purchaseOrders,
            approvalQueue: approvals,
            deliveryMilestones: deliveries,
            invoices: invoices
        )
    }

    // MARK: - Private helpers

    private func configureDatabase() throws {
        try db.run("PRAGMA foreign_keys = ON")

        try db.run(suppliersTable.create(ifNotExists: true) { table in
            table.column(supplierID, primaryKey: true)
            table.column(supplierTenantID)
            table.column(supplierName)
            table.column(supplierCountry)
            table.column(supplierCategory)
            table.column(supplierIsActive)
            table.column(supplierRisk)
            table.column(supplierSpendYTD)
        })

        try db.run(purchaseOrdersTable.create(ifNotExists: true) { table in
            table.column(purchaseOrderID, primaryKey: true)
            table.column(purchaseOrderTenantID)
            table.column(purchaseOrderNumber, unique: true)
            table.column(purchaseOrderSupplierID)
            table.column(purchaseOrderSupplierName)
            table.column(purchaseOrderProjectCode)
            table.column(purchaseOrderProjectName)
            table.column(purchaseOrderStatus)
            table.column(purchaseOrderCurrency)
            table.column(purchaseOrderTotal)
            table.column(purchaseOrderExpectedDelivery)
            table.column(purchaseOrderIssuedAt)
        })

        try db.run(approvalsTable.create(ifNotExists: true) { table in
            table.column(approvalID, primaryKey: true)
            table.column(approvalTenantID)
            table.column(approvalTitle)
            table.column(approvalType)
            table.column(approvalRequestedBy)
            table.column(approvalPendingWith)
            table.column(approvalDueDate)
            table.column(approvalStatus)
        })

        try db.run(deliveriesTable.create(ifNotExists: true) { table in
            table.column(deliveryID, primaryKey: true)
            table.column(deliveryTenantID)
            table.column(deliveryPONumber)
            table.column(deliveryDescription)
            table.column(deliveryExpectedOn)
            table.column(deliveryStatus)
        })

        try db.run(invoicesTable.create(ifNotExists: true) { table in
            table.column(invoiceID, primaryKey: true)
            table.column(invoiceTenantID)
            table.column(invoiceSupplierID)
            table.column(invoiceSupplierName)
            table.column(invoiceNumber, unique: true)
            table.column(invoiceAmount)
            table.column(invoiceCurrency)
            table.column(invoiceDueDate)
            table.column(invoiceStatus)
        })
    }

    private func seedIfNeeded(seedDate: Date) throws {
        let supplierCount = try db.scalar(suppliersTable.count)
        guard supplierCount == 0 else { return }

        let dataset = ProcurementSeedData.makeDataSet(seedDate: seedDate, tenantID: tenantID)
        try db.transaction {
            for supplier in dataset.suppliers {
                try db.run(suppliersTable.insert(or: .replace,
                                                 supplierID <- supplier.id.uuidString,
                                                 supplierTenantID <- supplier.tenantID.uuidString,
                                                 supplierName <- supplier.legalName,
                                                 supplierCountry <- supplier.country,
                                                 supplierCategory <- supplier.category,
                                                 supplierIsActive <- supplier.isActive,
                                                 supplierRisk <- supplier.riskRating,
                                                 supplierSpendYTD <- NSDecimalNumber(decimal: supplier.spendYearToDate).doubleValue))
            }

            for order in dataset.purchaseOrders {
                try db.run(purchaseOrdersTable.insert(or: .replace,
                                                      purchaseOrderID <- order.id.uuidString,
                                                      purchaseOrderTenantID <- order.tenantID.uuidString,
                                                      purchaseOrderNumber <- order.number,
                                                      purchaseOrderSupplierID <- order.supplierID.uuidString,
                                                      purchaseOrderSupplierName <- order.supplierName,
                                                      purchaseOrderProjectCode <- order.projectCode,
                                                      purchaseOrderProjectName <- order.projectName,
                                                      purchaseOrderStatus <- order.status.rawValue,
                                                      purchaseOrderCurrency <- order.currency,
                                                      purchaseOrderTotal <- NSDecimalNumber(decimal: order.totalValue).doubleValue,
                                                      purchaseOrderExpectedDelivery <- dateFormatter.string(from: order.expectedDelivery),
                                                      purchaseOrderIssuedAt <- dateFormatter.string(from: order.issuedAt)))
            }

            for approval in dataset.approvalQueue {
                try db.run(approvalsTable.insert(or: .replace,
                                                 approvalID <- approval.id.uuidString,
                                                 approvalTenantID <- approval.tenantID.uuidString,
                                                 approvalTitle <- approval.title,
                                                 approvalType <- approval.requestType,
                                                 approvalRequestedBy <- approval.requestedBy,
                                                 approvalPendingWith <- approval.pendingWith,
                                                 approvalDueDate <- dateFormatter.string(from: approval.dueDate),
                                                 approvalStatus <- approval.status.rawValue))
            }

            for delivery in dataset.deliveryMilestones {
                try db.run(deliveriesTable.insert(or: .replace,
                                                  deliveryID <- delivery.id.uuidString,
                                                  deliveryTenantID <- delivery.tenantID.uuidString,
                                                  deliveryPONumber <- delivery.purchaseOrderNumber,
                                                  deliveryDescription <- delivery.description,
                                                  deliveryExpectedOn <- dateFormatter.string(from: delivery.expectedOn),
                                                  deliveryStatus <- delivery.status.rawValue))
            }

            for invoice in dataset.invoices {
                try db.run(invoicesTable.insert(or: .replace,
                                                 invoiceID <- invoice.id.uuidString,
                                                 invoiceTenantID <- invoice.tenantID.uuidString,
                                                 invoiceSupplierID <- invoice.supplierID.uuidString,
                                                 invoiceSupplierName <- invoice.supplierName,
                                                 invoiceNumber <- invoice.invoiceNumber,
                                                 invoiceAmount <- NSDecimalNumber(decimal: invoice.amount).doubleValue,
                                                 invoiceCurrency <- invoice.currency,
                                                 invoiceDueDate <- dateFormatter.string(from: invoice.dueDate),
                                                 invoiceStatus <- invoice.status.rawValue))
            }
        }
    }

    private func parseUUID(_ string: String, label: String) throws -> UUID {
        guard let value = UUID(uuidString: string) else {
            throw ProcurementRepositoryError.missingRecord("Unable to parse UUID for \(label)")
        }
        return value
    }

    private func parseDate(_ string: String, label: String) throws -> Date {
        guard let date = dateFormatter.date(from: string) else {
            throw ProcurementRepositoryError.missingRecord("Unable to parse date for \(label)")
        }
        return date
    }
}
