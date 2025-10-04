import Foundation

public struct InspectionReport: InspectionReportProtocol {
    public typealias GoodsReceiptType = GoodsReceipt
    public typealias Inspector = UserAccount

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var goodsReceipt: GoodsReceipt
    public var inspector: UserAccount
    public var inspectionDate: Date
    public var result: String
    public var remarks: String?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                goodsReceipt: GoodsReceipt,
                inspector: UserAccount,
                inspectionDate: Date,
                result: String,
                remarks: String? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.goodsReceipt = goodsReceipt
        self.inspector = inspector
        self.inspectionDate = inspectionDate
        self.result = result
        self.remarks = remarks
    }
}

