import Foundation

public struct ReleaseOrder: ReleaseOrderProtocol {
    public typealias PurchaseOrderType = PurchaseOrder
    public typealias ContractType = Contract
    public typealias LineType = PurchaseOrderLine

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var purchaseOrder: PurchaseOrder
    public var contract: Contract?
    public var referencedLine: PurchaseOrderLine?
    public var releaseNumber: String
    public var quantity: Decimal
    public var releaseDate: Date
    public var status: String

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                purchaseOrder: PurchaseOrder,
                contract: Contract? = nil,
                referencedLine: PurchaseOrderLine? = nil,
                releaseNumber: String,
                quantity: Decimal,
                releaseDate: Date,
                status: String) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.purchaseOrder = purchaseOrder
        self.contract = contract
        self.referencedLine = referencedLine
        self.releaseNumber = releaseNumber
        self.quantity = quantity
        self.releaseDate = releaseDate
        self.status = status
    }
}

