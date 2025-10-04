import Foundation

public struct GoodsReceipt: GoodsReceiptProtocol {
    public typealias LineType = PurchaseOrderLine
    public typealias ReleaseType = ReleaseOrder
    public typealias LocationType = InventoryLocation
    public typealias Receiver = UserAccount

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var line: PurchaseOrderLine
    public var release: ReleaseOrder?
    public var location: InventoryLocation
    public var grnNumber: String
    public var receivedDate: Date
    public var receivedQuantity: Decimal
    public var acceptedQuantity: Decimal
    public var receivedBy: UserAccount
    public var status: String

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                line: PurchaseOrderLine,
                release: ReleaseOrder? = nil,
                location: InventoryLocation,
                grnNumber: String,
                receivedDate: Date,
                receivedQuantity: Decimal,
                acceptedQuantity: Decimal,
                receivedBy: UserAccount,
                status: String) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.line = line
        self.release = release
        self.location = location
        self.grnNumber = grnNumber
        self.receivedDate = receivedDate
        self.receivedQuantity = receivedQuantity
        self.acceptedQuantity = acceptedQuantity
        self.receivedBy = receivedBy
        self.status = status
    }
}

