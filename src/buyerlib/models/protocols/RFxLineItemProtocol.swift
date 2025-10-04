import Foundation

public protocol RFxLineItemProtocol: ModelIdentifiable {
    associatedtype EventType: RFxEventProtocol
    associatedtype DemandType: DemandProtocol
    var event: EventType { get }
    var demand: DemandType? { get }
    var itemNumber: Int { get }
    var itemDescription: String { get }
    var quantity: Decimal { get }
    var unitOfMeasure: String { get }
    var evaluationWeight: Decimal? { get }
}
