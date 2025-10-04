import Foundation

public struct RFxResponse: RFxResponseProtocol {
    public typealias EventType = RFxEvent
    public typealias SupplierType = Supplier

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var event: RFxEvent
    public var supplier: Supplier
    public var submittedAt: Date?
    public var commercialScore: Decimal?
    public var technicalScore: Decimal?
    public var currency: String
    public var status: String

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                event: RFxEvent,
                supplier: Supplier,
                submittedAt: Date? = nil,
                commercialScore: Decimal? = nil,
                technicalScore: Decimal? = nil,
                currency: String,
                status: String) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.event = event
        self.supplier = supplier
        self.submittedAt = submittedAt
        self.commercialScore = commercialScore
        self.technicalScore = technicalScore
        self.currency = currency
        self.status = status
    }
}

