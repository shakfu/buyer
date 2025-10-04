import Foundation

public protocol PurchaseOrderLineProtocol: ModelIdentifiable, Timestamped {
    associatedtype PurchaseOrderType: PurchaseOrderProtocol
    associatedtype DemandType: DemandProtocol
    var purchaseOrder: PurchaseOrderType { get }
    var demand: DemandType? { get }
    var lineNumber: Int { get }
    var lineDescription: String { get }
    var quantity: Decimal { get }
    var unitOfMeasure: String { get }
    var unitPrice: Decimal { get }
    var deliveryStart: Date? { get }
    var deliveryEnd: Date? { get }
    var status: String { get }
}
