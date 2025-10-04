import Foundation

public struct SupplierRecord: Codable, Equatable, Hashable {
    public let id: UUID
    public let tenantID: UUID
    public let legalName: String
    public let country: String
    public let category: String
    public let isActive: Bool
    public let riskRating: String?
    public let spendYearToDate: Decimal

    public init(id: UUID,
                tenantID: UUID,
                legalName: String,
                country: String,
                category: String,
                isActive: Bool,
                riskRating: String?,
                spendYearToDate: Decimal) {
        self.id = id
        self.tenantID = tenantID
        self.legalName = legalName
        self.country = country
        self.category = category
        self.isActive = isActive
        self.riskRating = riskRating
        self.spendYearToDate = spendYearToDate
    }
}

public enum PurchaseOrderStatus: String, Codable {
    case draft
    case approved
    case released
    case partiallyReceived
    case completed
    case cancelled

    public var isOpen: Bool {
        switch self {
        case .draft, .approved, .released, .partiallyReceived:
            return true
        case .completed, .cancelled:
            return false
        }
    }
}

public struct PurchaseOrderRecord: Codable, Equatable {
    public let id: UUID
    public let tenantID: UUID
    public let number: String
    public let supplierID: UUID
    public let supplierName: String
    public let projectCode: String
    public let projectName: String
    public let status: PurchaseOrderStatus
    public let currency: String
    public let totalValue: Decimal
    public let expectedDelivery: Date
    public let issuedAt: Date

    public init(id: UUID,
                tenantID: UUID,
                number: String,
                supplierID: UUID,
                supplierName: String,
                projectCode: String,
                projectName: String,
                status: PurchaseOrderStatus,
                currency: String,
                totalValue: Decimal,
                expectedDelivery: Date,
                issuedAt: Date) {
        self.id = id
        self.tenantID = tenantID
        self.number = number
        self.supplierID = supplierID
        self.supplierName = supplierName
        self.projectCode = projectCode
        self.projectName = projectName
        self.status = status
        self.currency = currency
        self.totalValue = totalValue
        self.expectedDelivery = expectedDelivery
        self.issuedAt = issuedAt
    }
}

public enum ApprovalStatus: String, Codable {
    case pending
    case escalated
    case approved
    case rejected
}

public struct ApprovalQueueItem: Codable, Equatable {
    public let id: UUID
    public let tenantID: UUID
    public let title: String
    public let requestType: String
    public let requestedBy: String
    public let pendingWith: String
    public let dueDate: Date
    public let status: ApprovalStatus

    public init(id: UUID,
                tenantID: UUID,
                title: String,
                requestType: String,
                requestedBy: String,
                pendingWith: String,
                dueDate: Date,
                status: ApprovalStatus) {
        self.id = id
        self.tenantID = tenantID
        self.title = title
        self.requestType = requestType
        self.requestedBy = requestedBy
        self.pendingWith = pendingWith
        self.dueDate = dueDate
        self.status = status
    }
}

public enum DeliveryStatus: String, Codable {
    case scheduled
    case inTransit
    case received
    case delayed
}

public struct DeliveryMilestoneRecord: Codable, Equatable {
    public let id: UUID
    public let tenantID: UUID
    public let purchaseOrderNumber: String
    public let description: String
    public let expectedOn: Date
    public let status: DeliveryStatus

    public init(id: UUID,
                tenantID: UUID,
                purchaseOrderNumber: String,
                description: String,
                expectedOn: Date,
                status: DeliveryStatus) {
        self.id = id
        self.tenantID = tenantID
        self.purchaseOrderNumber = purchaseOrderNumber
        self.description = description
        self.expectedOn = expectedOn
        self.status = status
    }
}

public enum InvoiceStatus: String, Codable {
    case pending
    case onHold
    case paid
    case disputed
}

public struct InvoiceRecord: Codable, Equatable {
    public let id: UUID
    public let tenantID: UUID
    public let supplierID: UUID
    public let supplierName: String
    public let invoiceNumber: String
    public let amount: Decimal
    public let currency: String
    public let dueDate: Date
    public let status: InvoiceStatus

    public init(id: UUID,
                tenantID: UUID,
                supplierID: UUID,
                supplierName: String,
                invoiceNumber: String,
                amount: Decimal,
                currency: String,
                dueDate: Date,
                status: InvoiceStatus) {
        self.id = id
        self.tenantID = tenantID
        self.supplierID = supplierID
        self.supplierName = supplierName
        self.invoiceNumber = invoiceNumber
        self.amount = amount
        self.currency = currency
        self.dueDate = dueDate
        self.status = status
    }
}

public struct ProcurementDataSet: Codable {
    public var suppliers: [SupplierRecord]
    public var purchaseOrders: [PurchaseOrderRecord]
    public var approvalQueue: [ApprovalQueueItem]
    public var deliveryMilestones: [DeliveryMilestoneRecord]
    public var invoices: [InvoiceRecord]

    public init(suppliers: [SupplierRecord],
                purchaseOrders: [PurchaseOrderRecord],
                approvalQueue: [ApprovalQueueItem],
                deliveryMilestones: [DeliveryMilestoneRecord],
                invoices: [InvoiceRecord]) {
        self.suppliers = suppliers
        self.purchaseOrders = purchaseOrders
        self.approvalQueue = approvalQueue
        self.deliveryMilestones = deliveryMilestones
        self.invoices = invoices
    }
}

public enum ProcurementAlertSeverity: String, Codable {
    case info
    case warning
    case critical
}

public struct ProcurementAlert: Codable, Equatable {
    public let message: String
    public let category: String
    public let severity: ProcurementAlertSeverity
    public let relatedIdentifier: String?

    public init(message: String,
                category: String,
                severity: ProcurementAlertSeverity,
                relatedIdentifier: String?) {
        self.message = message
        self.category = category
        self.severity = severity
        self.relatedIdentifier = relatedIdentifier
    }
}

public struct ProcurementSummary: Codable {
    public let generatedAt: Date
    public let activeSuppliers: Int
    public let openPurchaseOrders: Int
    public let deliveriesDueThisWeek: Int
    public let pendingApprovals: Int
    public let invoicesOnHold: Int
    public let alerts: [ProcurementAlert]

    public init(generatedAt: Date,
                activeSuppliers: Int,
                openPurchaseOrders: Int,
                deliveriesDueThisWeek: Int,
                pendingApprovals: Int,
                invoicesOnHold: Int,
                alerts: [ProcurementAlert]) {
        self.generatedAt = generatedAt
        self.activeSuppliers = activeSuppliers
        self.openPurchaseOrders = openPurchaseOrders
        self.deliveriesDueThisWeek = deliveriesDueThisWeek
        self.pendingApprovals = pendingApprovals
        self.invoicesOnHold = invoicesOnHold
        self.alerts = alerts
    }
}

public struct SupplierSpendSummary: Codable, Equatable {
    public let supplier: SupplierRecord
    public let totalOpenPOValue: Decimal
    public let invoicesOnHold: Int
    public let overdueDeliveries: Int

    public init(supplier: SupplierRecord,
                totalOpenPOValue: Decimal,
                invoicesOnHold: Int,
                overdueDeliveries: Int) {
        self.supplier = supplier
        self.totalOpenPOValue = totalOpenPOValue
        self.invoicesOnHold = invoicesOnHold
        self.overdueDeliveries = overdueDeliveries
    }
}
