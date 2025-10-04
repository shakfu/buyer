import Foundation

public protocol ApprovalPolicyProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    var objectType: String { get }
    var thresholdCurrency: String? { get }
    var thresholdAmount: Decimal? { get }
    var activeFrom: Date { get }
    var activeTo: Date? { get }
}

