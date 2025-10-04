import Foundation

public protocol DeliveryMilestoneProtocol: ModelIdentifiable, Timestamped {
    associatedtype LineType: PurchaseOrderLineProtocol
    var line: LineType { get }
    var expectedDate: Date { get }
    var expectedQuantity: Decimal { get }
    var actualDate: Date? { get }
    var actualQuantity: Decimal? { get }
    var status: String { get }
}
