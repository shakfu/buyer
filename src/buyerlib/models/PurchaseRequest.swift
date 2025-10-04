import Foundation

public struct PurchaseRequest: PurchaseRequestProtocol {
    public typealias ProjectType = Project
    public typealias RequesterType = UserAccount
    public typealias DemandType = Demand

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var requestNumber: String
    public var project: Project
    public var requester: UserAccount
    public var demands: [Demand]
    public var justification: String?
    public var neededBy: Date?
    public var status: String

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                requestNumber: String,
                project: Project,
                requester: UserAccount,
                demands: [Demand],
                justification: String? = nil,
                neededBy: Date? = nil,
                status: String) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.requestNumber = requestNumber
        self.project = project
        self.requester = requester
        self.demands = demands
        self.justification = justification
        self.neededBy = neededBy
        self.status = status
    }
}

