import Foundation

public protocol RFxResponseProtocol: ModelIdentifiable, Timestamped {
    associatedtype EventType: RFxEventProtocol
    associatedtype SupplierType: SupplierProtocol
    var event: EventType { get }
    var supplier: SupplierType { get }
    var submittedAt: Date? { get }
    var commercialScore: Decimal? { get }
    var technicalScore: Decimal? { get }
    var currency: String { get }
    var status: String { get }
}

