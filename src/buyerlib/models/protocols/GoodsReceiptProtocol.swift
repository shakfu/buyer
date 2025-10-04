import Foundation

public protocol GoodsReceiptProtocol: ModelIdentifiable, Timestamped {
    associatedtype LineType: PurchaseOrderLineProtocol
    associatedtype ReleaseType: ReleaseOrderProtocol
    associatedtype LocationType: InventoryLocationProtocol
    associatedtype Receiver: UserAccountProtocol
    var line: LineType { get }
    var release: ReleaseType? { get }
    var location: LocationType { get }
    var grnNumber: String { get }
    var receivedDate: Date { get }
    var receivedQuantity: Decimal { get }
    var acceptedQuantity: Decimal { get }
    var receivedBy: Receiver { get }
    var status: String { get }
}
