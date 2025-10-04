import Foundation

public struct NonConformance: NonConformanceProtocol {
    public typealias GoodsReceiptType = GoodsReceipt

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var goodsReceipt: GoodsReceipt
    public var reportedDate: Date
    public var severity: String
    public var issueType: String
    public var quantityAffected: Decimal
    public var status: String
    public var resolutionNotes: String?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                goodsReceipt: GoodsReceipt,
                reportedDate: Date,
                severity: String,
                issueType: String,
                quantityAffected: Decimal,
                status: String,
                resolutionNotes: String? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.goodsReceipt = goodsReceipt
        self.reportedDate = reportedDate
        self.severity = severity
        self.issueType = issueType
        self.quantityAffected = quantityAffected
        self.status = status
        self.resolutionNotes = resolutionNotes
    }
}

