import Foundation

public protocol ReleaseOrderProtocol: ModelIdentifiable, Timestamped {
    associatedtype PurchaseOrderType: PurchaseOrderProtocol
    associatedtype ContractType: ContractProtocol
    associatedtype LineType: PurchaseOrderLineProtocol
    var purchaseOrder: PurchaseOrderType { get }
    var contract: ContractType? { get }
    var referencedLine: LineType? { get }
    var releaseNumber: String { get }
    var quantity: Decimal { get }
    var releaseDate: Date { get }
    var status: String { get }
}
