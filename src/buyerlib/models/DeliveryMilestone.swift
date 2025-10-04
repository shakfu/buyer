import Foundation

public struct DeliveryMilestone: DeliveryMilestoneProtocol {
    public typealias LineType = PurchaseOrderLine

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var line: PurchaseOrderLine
    public var expectedDate: Date
    public var expectedQuantity: Decimal
    public var actualDate: Date?
    public var actualQuantity: Decimal?
    public var status: String

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                line: PurchaseOrderLine,
                expectedDate: Date,
                expectedQuantity: Decimal,
                actualDate: Date? = nil,
                actualQuantity: Decimal? = nil,
                status: String) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.line = line
        self.expectedDate = expectedDate
        self.expectedQuantity = expectedQuantity
        self.actualDate = actualDate
        self.actualQuantity = actualQuantity
        self.status = status
    }
}

