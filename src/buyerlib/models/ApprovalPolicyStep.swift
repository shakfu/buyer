import Foundation

public struct ApprovalPolicyStep: ApprovalPolicyStepProtocol {
    public typealias PolicyType = ApprovalPolicy
    public typealias RoleType = Role

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var policy: ApprovalPolicy
    public var sequence: Int
    public var role: Role
    public var ruleExpression: String?
    public var slaHours: Int

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                policy: ApprovalPolicy,
                sequence: Int,
                role: Role,
                ruleExpression: String? = nil,
                slaHours: Int) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.policy = policy
        self.sequence = sequence
        self.role = role
        self.ruleExpression = ruleExpression
        self.slaHours = slaHours
    }
}
