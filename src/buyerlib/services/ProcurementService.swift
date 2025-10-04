import Foundation

public final class ProcurementService {
    private let repository: ProcurementRepository
    private let calendar: Calendar
    private let dateProvider: () -> Date

    public init(repository: ProcurementRepository,
                calendar: Calendar = Calendar(identifier: .gregorian),
                dateProvider: @escaping () -> Date = Date.init) {
        self.repository = repository
        self.calendar = calendar
        self.dateProvider = dateProvider
    }

    public func statusSummary() throws -> ProcurementSummary {
        let dataSet = try repository.loadData()
        let now = dateProvider()
        let startOfToday = calendar.startOfDay(for: now)
        let endOfWeek = calendar.date(byAdding: .day, value: 7, to: startOfToday) ?? now

        let activeSuppliers = dataSet.suppliers.filter { $0.isActive }.count
        let openPurchaseOrders = dataSet.purchaseOrders.filter { $0.status.isOpen }.count

        let deliveriesDueThisWeek = dataSet.deliveryMilestones.filter { milestone in
            guard milestone.status != .received else { return false }
            return milestone.expectedOn >= startOfToday && milestone.expectedOn <= endOfWeek
        }.count

        let pendingApprovals = dataSet.approvalQueue.filter { $0.status == .pending || $0.status == .escalated }.count
        let invoicesOnHold = dataSet.invoices.filter { $0.status == .onHold }.count

        let alerts = buildAlerts(from: dataSet, referenceDate: now)

        return ProcurementSummary(
            generatedAt: now,
            activeSuppliers: activeSuppliers,
            openPurchaseOrders: openPurchaseOrders,
            deliveriesDueThisWeek: deliveriesDueThisWeek,
            pendingApprovals: pendingApprovals,
            invoicesOnHold: invoicesOnHold,
            alerts: alerts
        )
    }

    public func searchSuppliers(matching rawQuery: String, limit: Int = 10) throws -> [SupplierRecord] {
        let trimmedQuery = rawQuery.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmedQuery.count >= 2 else { return [] }

        let dataSet = try repository.loadData()

        let results = dataSet.suppliers.filter { supplier in
            supplier.legalName.range(of: trimmedQuery, options: .caseInsensitive) != nil ||
                supplier.country.range(of: trimmedQuery, options: .caseInsensitive) != nil ||
                supplier.category.range(of: trimmedQuery, options: .caseInsensitive) != nil
        }
        .sorted { lhs, rhs in
            if lhs.isActive != rhs.isActive {
                return lhs.isActive && !rhs.isActive
            }
            if lhs.riskRating != rhs.riskRating {
                switch (lhs.riskRating, rhs.riskRating) {
                case let (l?, r?):
                    return l < r
                case (.some, .none):
                    return true
                case (.none, .some):
                    return false
                case (.none, .none):
                    break
                }
            }
            return lhs.legalName < rhs.legalName
        }

        if limit <= 0 { return results }
        return Array(results.prefix(limit))
    }

    public func openPurchaseOrders(sortedBy sortKey: PurchaseOrderSortKey = .expectedDelivery) throws -> [PurchaseOrderRecord] {
        let dataSet = try repository.loadData()
        let openOrders = dataSet.purchaseOrders.filter { $0.status.isOpen }
        return openOrders.sorted(by: { lhs, rhs in
            switch sortKey {
            case .expectedDelivery:
                if lhs.expectedDelivery == rhs.expectedDelivery {
                    return lhs.number < rhs.number
                }
                return lhs.expectedDelivery < rhs.expectedDelivery
            case .totalValue:
                if lhs.totalValue == rhs.totalValue {
                    return lhs.number < rhs.number
                }
                return lhs.totalValue > rhs.totalValue
            case .supplier:
                if lhs.supplierName == rhs.supplierName {
                    return lhs.number < rhs.number
                }
                return lhs.supplierName < rhs.supplierName
            }
        })
    }

    public func pendingApprovalQueue() throws -> [ApprovalQueueItem] {
        let dataSet = try repository.loadData()
        return dataSet.approvalQueue.filter { $0.status == .pending || $0.status == .escalated }
            .sorted(by: { lhs, rhs in
                if lhs.dueDate == rhs.dueDate {
                    return lhs.title < rhs.title
                }
                return lhs.dueDate < rhs.dueDate
            })
    }

    public func deliveries(restrictingTo statuses: Set<DeliveryStatus> = [.scheduled, .inTransit, .delayed]) throws -> [DeliveryMilestoneRecord] {
        let dataSet = try repository.loadData()
        return dataSet.deliveryMilestones.filter { statuses.contains($0.status) }
            .sorted(by: { lhs, rhs in
                if lhs.expectedOn == rhs.expectedOn {
                    return lhs.purchaseOrderNumber < rhs.purchaseOrderNumber
                }
                return lhs.expectedOn < rhs.expectedOn
            })
    }

    public func invoicesOnHold() throws -> [InvoiceRecord] {
        let dataSet = try repository.loadData()
        return dataSet.invoices.filter { $0.status == .onHold }
            .sorted(by: { lhs, rhs in
                if lhs.dueDate == rhs.dueDate {
                    return lhs.invoiceNumber < rhs.invoiceNumber
                }
                return lhs.dueDate < rhs.dueDate
            })
    }

    public func dataSet() throws -> ProcurementDataSet {
        try repository.loadData()
    }

    public func supplierSpendSummaries() throws -> [SupplierSpendSummary] {
        let dataSet = try repository.loadData()
        let now = dateProvider()

        let suppliersByID = Dictionary(uniqueKeysWithValues: dataSet.suppliers.map { ($0.id, $0) })
        let purchaseOrders = dataSet.purchaseOrders.filter { $0.status.isOpen }

        var spendBySupplier = [UUID: Decimal]()
        for order in purchaseOrders {
            spendBySupplier[order.supplierID, default: 0] += order.totalValue
        }

        let invoicesOnHoldBySupplier = dataSet.invoices.reduce(into: [UUID: Int]()) { partialResult, invoice in
            guard invoice.status == .onHold else { return }
            partialResult[invoice.supplierID, default: 0] += 1
        }

        let purchaseOrderByNumber = Dictionary(uniqueKeysWithValues: dataSet.purchaseOrders.map { ($0.number, $0) })
        let overdueDeliveries = dataSet.deliveryMilestones.reduce(into: [UUID: Int]()) { partialResult, milestone in
            guard milestone.status != .received, milestone.expectedOn < now else { return }
            guard let owningOrder = purchaseOrderByNumber[milestone.purchaseOrderNumber] else { return }
            partialResult[owningOrder.supplierID, default: 0] += 1
        }

        var summaries = [SupplierSpendSummary]()
        for (supplierID, supplier) in suppliersByID {
            let totalValue = spendBySupplier[supplierID] ?? 0
            let invoicesOnHold = invoicesOnHoldBySupplier[supplierID] ?? 0
            let overdue = overdueDeliveries[supplierID] ?? 0
            summaries.append(
                SupplierSpendSummary(
                    supplier: supplier,
                    totalOpenPOValue: totalValue,
                    invoicesOnHold: invoicesOnHold,
                    overdueDeliveries: overdue
                )
            )
        }

        return summaries.sorted { lhs, rhs in
            if lhs.totalOpenPOValue == rhs.totalOpenPOValue {
                return lhs.supplier.legalName < rhs.supplier.legalName
            }
            return lhs.totalOpenPOValue > rhs.totalOpenPOValue
        }
    }

    private func buildAlerts(from dataSet: ProcurementDataSet, referenceDate: Date) -> [ProcurementAlert] {
        var alerts: [ProcurementAlert] = []

        for order in dataSet.purchaseOrders where order.status.isOpen && order.expectedDelivery < referenceDate {
            let message = "PO \(order.number) for \(order.supplierName) is overdue by \(daysBetween(order.expectedDelivery, referenceDate)) days"
            alerts.append(ProcurementAlert(message: message,
                                           category: "Purchase Orders",
                                           severity: .warning,
                                           relatedIdentifier: order.number))
        }

        for milestone in dataSet.deliveryMilestones where milestone.expectedOn < referenceDate && milestone.status != .received {
            let message = "Delivery \(milestone.description) on PO \(milestone.purchaseOrderNumber) missed the expected date"
            alerts.append(ProcurementAlert(message: message,
                                           category: "Deliveries",
                                           severity: .critical,
                                           relatedIdentifier: milestone.purchaseOrderNumber))
        }

        for approval in dataSet.approvalQueue where (approval.status == .pending || approval.status == .escalated) && approval.dueDate < referenceDate {
            let message = "Approval '\(approval.title)' assigned to \(approval.pendingWith) is past due"
            alerts.append(ProcurementAlert(message: message,
                                           category: "Approvals",
                                           severity: .warning,
                                           relatedIdentifier: approval.title))
        }

        for invoice in dataSet.invoices where invoice.status == .onHold && invoice.dueDate < referenceDate {
            let message = "Invoice \(invoice.invoiceNumber) for \(invoice.supplierName) is on hold past due date"
            alerts.append(ProcurementAlert(message: message,
                                           category: "Invoices",
                                           severity: .critical,
                                           relatedIdentifier: invoice.invoiceNumber))
        }

        return alerts.sorted { lhs, rhs in
            if lhs.severity != rhs.severity {
                return lhs.severity.sortOrder < rhs.severity.sortOrder
            }
            return lhs.message < rhs.message
        }
    }

    private func daysBetween(_ start: Date, _ end: Date) -> Int {
        let startDay = calendar.startOfDay(for: start)
        let endDay = calendar.startOfDay(for: end)
        let components = calendar.dateComponents([.day], from: startDay, to: endDay)
        return components.day ?? 0
    }
}

public enum PurchaseOrderSortKey {
    case expectedDelivery
    case totalValue
    case supplier
}

private extension ProcurementAlertSeverity {
    var sortOrder: Int {
        switch self {
        case .critical:
            return 0
        case .warning:
            return 1
        case .info:
            return 2
        }
    }
}
