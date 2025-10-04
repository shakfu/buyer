import Foundation

public protocol ApprovalPolicyStepProtocol: ModelIdentifiable, Timestamped {
    associatedtype PolicyType: ApprovalPolicyProtocol
    associatedtype RoleType: RoleProtocol
    var policy: PolicyType { get }
    var sequence: Int { get }
    var role: RoleType { get }
    var ruleExpression: String? { get }
    var slaHours: Int { get }
}

