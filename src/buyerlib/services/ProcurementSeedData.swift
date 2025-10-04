import Foundation

public struct ProcurementSeedData {
    public static func makeDataSet(seedDate: Date = Date(),
                                   tenantID: UUID = UUID(uuidString: "00000000-0000-0000-0000-000000000001")!) -> ProcurementDataSet {
        let calendar = Calendar(identifier: .gregorian)

        func addDays(_ value: Int) -> Date {
            calendar.date(byAdding: .day, value: value, to: seedDate) ?? seedDate
        }

        let supplierNorthwind = SupplierRecord(
            id: UUID(uuidString: "11111111-1111-1111-1111-111111111111")!,
            tenantID: tenantID,
            legalName: "Northwind Logistics",
            country: "United States",
            category: "Logistics",
            isActive: true,
            riskRating: "Low",
            spendYearToDate: Decimal(string: "1250000.00")!
        )

        let supplierAtlas = SupplierRecord(
            id: UUID(uuidString: "22222222-2222-2222-2222-222222222222")!,
            tenantID: tenantID,
            legalName: "Atlas Steelworks",
            country: "Germany",
            category: "Structural Steel",
            isActive: true,
            riskRating: "Medium",
            spendYearToDate: Decimal(string: "980000.00")!
        )

        let supplierSkyline = SupplierRecord(
            id: UUID(uuidString: "33333333-3333-3333-3333-333333333333")!,
            tenantID: tenantID,
            legalName: "Skyline Electrical",
            country: "United Arab Emirates",
            category: "Electrical",
            isActive: true,
            riskRating: "High",
            spendYearToDate: Decimal(string: "640000.00")!
        )

        let supplierCoastal = SupplierRecord(
            id: UUID(uuidString: "44444444-4444-4444-4444-444444444444")!,
            tenantID: tenantID,
            legalName: "Coastal Finishes",
            country: "United Kingdom",
            category: "Interior Finishes",
            isActive: false,
            riskRating: nil,
            spendYearToDate: Decimal(string: "120000.00")!
        )

        let purchaseOrders = [
            PurchaseOrderRecord(
                id: UUID(uuidString: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1")!,
                tenantID: tenantID,
                number: "PO-1001",
                supplierID: supplierNorthwind.id,
                supplierName: supplierNorthwind.legalName,
                projectCode: "PRJ-ALPHA",
                projectName: "Airport Expansion",
                status: .released,
                currency: "USD",
                totalValue: Decimal(string: "450000.00")!,
                expectedDelivery: addDays(3),
                issuedAt: addDays(-10)
            ),
            PurchaseOrderRecord(
                id: UUID(uuidString: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2")!,
                tenantID: tenantID,
                number: "PO-1002",
                supplierID: supplierAtlas.id,
                supplierName: supplierAtlas.legalName,
                projectCode: "PRJ-ALPHA",
                projectName: "Airport Expansion",
                status: .partiallyReceived,
                currency: "EUR",
                totalValue: Decimal(string: "820000.00")!,
                expectedDelivery: addDays(-2),
                issuedAt: addDays(-25)
            ),
            PurchaseOrderRecord(
                id: UUID(uuidString: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3")!,
                tenantID: tenantID,
                number: "PO-1003",
                supplierID: supplierSkyline.id,
                supplierName: supplierSkyline.legalName,
                projectCode: "PRJ-ORBIT",
                projectName: "Orbital Tower",
                status: .approved,
                currency: "AED",
                totalValue: Decimal(string: "310000.00")!,
                expectedDelivery: addDays(14),
                issuedAt: addDays(-3)
            ),
            PurchaseOrderRecord(
                id: UUID(uuidString: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4")!,
                tenantID: tenantID,
                number: "PO-1004",
                supplierID: supplierNorthwind.id,
                supplierName: supplierNorthwind.legalName,
                projectCode: "PRJ-HARBOR",
                projectName: "Harbor Redevelopment",
                status: .completed,
                currency: "USD",
                totalValue: Decimal(string: "275000.00")!,
                expectedDelivery: addDays(-20),
                issuedAt: addDays(-60)
            )
        ]

        let approvals = [
            ApprovalQueueItem(
                id: UUID(uuidString: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1")!,
                tenantID: tenantID,
                title: "Steel package >$1M",
                requestType: "Purchase Order",
                requestedBy: "Maria Singh",
                pendingWith: "Liam Patel",
                dueDate: addDays(-1),
                status: .pending
            ),
            ApprovalQueueItem(
                id: UUID(uuidString: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2")!,
                tenantID: tenantID,
                title: "Emergency crane rental",
                requestType: "Exception",
                requestedBy: "Noah Martinez",
                pendingWith: "Emma Thompson",
                dueDate: addDays(2),
                status: .escalated
            ),
            ApprovalQueueItem(
                id: UUID(uuidString: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb3")!,
                tenantID: tenantID,
                title: "Contract extension HVAC",
                requestType: "Contract",
                requestedBy: "Oliver Chen",
                pendingWith: "Hannah Wood",
                dueDate: addDays(5),
                status: .approved
            )
        ]

        let deliveries = [
            DeliveryMilestoneRecord(
                id: UUID(uuidString: "cccccccc-cccc-cccc-cccc-ccccccccccc1")!,
                tenantID: tenantID,
                purchaseOrderNumber: "PO-1001",
                description: "Hangar doors shipment",
                expectedOn: addDays(2),
                status: .inTransit
            ),
            DeliveryMilestoneRecord(
                id: UUID(uuidString: "cccccccc-cccc-cccc-cccc-ccccccccccc2")!,
                tenantID: tenantID,
                purchaseOrderNumber: "PO-1002",
                description: "Steel trusses batch 3",
                expectedOn: addDays(-3),
                status: .delayed
            ),
            DeliveryMilestoneRecord(
                id: UUID(uuidString: "cccccccc-cccc-cccc-cccc-ccccccccccc3")!,
                tenantID: tenantID,
                purchaseOrderNumber: "PO-1003",
                description: "Switchgear panels",
                expectedOn: addDays(10),
                status: .scheduled
            ),
            DeliveryMilestoneRecord(
                id: UUID(uuidString: "cccccccc-cccc-cccc-cccc-ccccccccccc4")!,
                tenantID: tenantID,
                purchaseOrderNumber: "PO-1004",
                description: "Marina decking",
                expectedOn: addDays(-25),
                status: .received
            )
        ]

        let invoices = [
            InvoiceRecord(
                id: UUID(uuidString: "dddddddd-dddd-dddd-dddd-ddddddddddd1")!,
                tenantID: tenantID,
                supplierID: supplierAtlas.id,
                supplierName: supplierAtlas.legalName,
                invoiceNumber: "INV-501",
                amount: Decimal(string: "450000.00")!,
                currency: "EUR",
                dueDate: addDays(-5),
                status: .onHold
            ),
            InvoiceRecord(
                id: UUID(uuidString: "dddddddd-dddd-dddd-dddd-ddddddddddd2")!,
                tenantID: tenantID,
                supplierID: supplierSkyline.id,
                supplierName: supplierSkyline.legalName,
                invoiceNumber: "INV-502",
                amount: Decimal(string: "210000.00")!,
                currency: "AED",
                dueDate: addDays(15),
                status: .pending
            ),
            InvoiceRecord(
                id: UUID(uuidString: "dddddddd-dddd-dddd-dddd-ddddddddddd3")!,
                tenantID: tenantID,
                supplierID: supplierNorthwind.id,
                supplierName: supplierNorthwind.legalName,
                invoiceNumber: "INV-503",
                amount: Decimal(string: "125000.00")!,
                currency: "USD",
                dueDate: addDays(-2),
                status: .onHold
            ),
            InvoiceRecord(
                id: UUID(uuidString: "dddddddd-dddd-dddd-dddd-ddddddddddd4")!,
                tenantID: tenantID,
                supplierID: supplierCoastal.id,
                supplierName: supplierCoastal.legalName,
                invoiceNumber: "INV-504",
                amount: Decimal(string: "89000.00")!,
                currency: "GBP",
                dueDate: addDays(20),
                status: .paid
            )
        ]

        return ProcurementDataSet(
            suppliers: [supplierNorthwind, supplierAtlas, supplierSkyline, supplierCoastal],
            purchaseOrders: purchaseOrders,
            approvalQueue: approvals,
            deliveryMilestones: deliveries,
            invoices: invoices
        )
    }
}
