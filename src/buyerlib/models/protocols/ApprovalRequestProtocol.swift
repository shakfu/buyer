import Foundation

public protocol ApprovalRequestProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype PolicyType: ApprovalPolicyProtocol
    associatedtype StepType: ApprovalPolicyStepProtocol
    var objectType: String { get }
    var objectID: UUID { get }
    var policy: PolicyType { get }
    var currentStep: StepType? { get }
    var status: String { get }
    var completedAt: Date? { get }
}
