import Foundation

public struct Contract: ContractProtocol {
    public typealias SupplierType = Supplier
    public typealias ProjectType = Project
    public typealias EventType = RFxEvent

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var contractNumber: String
    public var supplier: Supplier
    public var project: Project?
    public var sourceEvent: RFxEvent?
    public var type: String
    public var effectiveDate: Date
    public var expiryDate: Date?
    public var status: String

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                contractNumber: String,
                supplier: Supplier,
                project: Project? = nil,
                sourceEvent: RFxEvent? = nil,
                type: String,
                effectiveDate: Date,
                expiryDate: Date? = nil,
                status: String) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.contractNumber = contractNumber
        self.supplier = supplier
        self.project = project
        self.sourceEvent = sourceEvent
        self.type = type
        self.effectiveDate = effectiveDate
        self.expiryDate = expiryDate
        self.status = status
    }
}

