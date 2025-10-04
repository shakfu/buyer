import Foundation

public protocol InvoiceLineProtocol: ModelIdentifiable, Timestamped {
    associatedtype InvoiceType: InvoiceProtocol
    associatedtype LineType: PurchaseOrderLineProtocol
    associatedtype GoodsReceiptType: GoodsReceiptProtocol
    var invoice: InvoiceType { get }
    var purchaseOrderLine: LineType? { get }
    var goodsReceipt: GoodsReceiptType? { get }
    var lineNumber: Int { get }
    var lineDescription: String { get }
    var quantity: Decimal { get }
    var unitPrice: Decimal { get }
    var amount: Decimal { get }
}
