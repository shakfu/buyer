import Foundation

public struct PurchaseOrderLine: PurchaseOrderLineProtocol {
    public typealias PurchaseOrderType = PurchaseOrder
    public typealias DemandType = Demand

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var purchaseOrder: PurchaseOrder
    public var demand: Demand?
    public var lineNumber: Int
    public var lineDescription: String
    public var quantity: Decimal
    public var unitOfMeasure: String
    public var unitPrice: Decimal
    public var deliveryStart: Date?
    public var deliveryEnd: Date?
    public var status: String

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                purchaseOrder: PurchaseOrder,
                demand: Demand? = nil,
                lineNumber: Int,
                lineDescription: String,
                quantity: Decimal,
                unitOfMeasure: String,
                unitPrice: Decimal,
                deliveryStart: Date? = nil,
                deliveryEnd: Date? = nil,
                status: String) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.purchaseOrder = purchaseOrder
        self.demand = demand
        self.lineNumber = lineNumber
        self.lineDescription = lineDescription
        self.quantity = quantity
        self.unitOfMeasure = unitOfMeasure
        self.unitPrice = unitPrice
        self.deliveryStart = deliveryStart
        self.deliveryEnd = deliveryEnd
        self.status = status
    }
}
