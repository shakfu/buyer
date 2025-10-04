import Foundation

public protocol RFxScorecardProtocol: ModelIdentifiable, Timestamped {
    associatedtype ResponseType: RFxResponseProtocol
    associatedtype Evaluator: UserAccountProtocol
    var response: ResponseType { get }
    var evaluator: Evaluator { get }
    var criterion: String { get }
    var score: Decimal { get }
    var comments: String? { get }
}

