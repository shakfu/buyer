import Foundation

public struct ApprovalPolicy: ApprovalPolicyProtocol {
    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var objectType: String
    public var thresholdCurrency: String?
    public var thresholdAmount: Decimal?
    public var activeFrom: Date
    public var activeTo: Date?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                objectType: String,
                thresholdCurrency: String? = nil,
                thresholdAmount: Decimal? = nil,
                activeFrom: Date,
                activeTo: Date? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.objectType = objectType
        self.thresholdCurrency = thresholdCurrency
        self.thresholdAmount = thresholdAmount
        self.activeFrom = activeFrom
        self.activeTo = activeTo
    }
}

