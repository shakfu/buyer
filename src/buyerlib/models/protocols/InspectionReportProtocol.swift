import Foundation

public protocol InspectionReportProtocol: ModelIdentifiable, Timestamped {
    associatedtype GoodsReceiptType: GoodsReceiptProtocol
    associatedtype Inspector: UserAccountProtocol
    var goodsReceipt: GoodsReceiptType { get }
    var inspector: Inspector { get }
    var inspectionDate: Date { get }
    var result: String { get }
    var remarks: String? { get }
}

