import Foundation

public struct RFxScorecard: RFxScorecardProtocol {
    public typealias ResponseType = RFxResponse
    public typealias Evaluator = UserAccount

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var response: RFxResponse
    public var evaluator: UserAccount
    public var criterion: String
    public var score: Decimal
    public var comments: String?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                response: RFxResponse,
                evaluator: UserAccount,
                criterion: String,
                score: Decimal,
                comments: String? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.response = response
        self.evaluator = evaluator
        self.criterion = criterion
        self.score = score
        self.comments = comments
    }
}

