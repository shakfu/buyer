import Foundation

public struct ApprovalRequest: ApprovalRequestProtocol {
    public typealias PolicyType = ApprovalPolicy
    public typealias StepType = ApprovalPolicyStep

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var objectType: String
    public var objectID: UUID
    public var policy: ApprovalPolicy
    public var currentStep: ApprovalPolicyStep?
    public var status: String
    public var completedAt: Date?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                objectType: String,
                objectID: UUID,
                policy: ApprovalPolicy,
                currentStep: ApprovalPolicyStep? = nil,
                status: String,
                completedAt: Date? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.objectType = objectType
        self.objectID = objectID
        self.policy = policy
        self.currentStep = currentStep
        self.status = status
        self.completedAt = completedAt
    }
}

