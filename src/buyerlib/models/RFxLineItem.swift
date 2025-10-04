import Foundation

public struct RFxLineItem: RFxLineItemProtocol {
    public typealias EventType = RFxEvent
    public typealias DemandType = Demand

    public var id: UUID
    public var event: RFxEvent
    public var demand: Demand?
    public var itemNumber: Int
    public var itemDescription: String
    public var quantity: Decimal
    public var unitOfMeasure: String
    public var evaluationWeight: Decimal?

    public init(id: UUID,
                event: RFxEvent,
                demand: Demand? = nil,
                itemNumber: Int,
                itemDescription: String,
                quantity: Decimal,
                unitOfMeasure: String,
                evaluationWeight: Decimal? = nil) {
        self.id = id
        self.event = event
        self.demand = demand
        self.itemNumber = itemNumber
        self.itemDescription = itemDescription
        self.quantity = quantity
        self.unitOfMeasure = unitOfMeasure
        self.evaluationWeight = evaluationWeight
    }
}

