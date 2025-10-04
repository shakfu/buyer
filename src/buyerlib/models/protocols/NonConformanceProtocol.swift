import Foundation

public protocol NonConformanceProtocol: ModelIdentifiable, Timestamped {
    associatedtype GoodsReceiptType: GoodsReceiptProtocol
    var goodsReceipt: GoodsReceiptType { get }
    var reportedDate: Date { get }
    var severity: String { get }
    var issueType: String { get }
    var quantityAffected: Decimal { get }
    var status: String { get }
    var resolutionNotes: String? { get }
}

