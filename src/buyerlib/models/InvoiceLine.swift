import Foundation

public struct InvoiceLine: InvoiceLineProtocol {
    public typealias InvoiceType = Invoice
    public typealias LineType = PurchaseOrderLine
    public typealias GoodsReceiptType = GoodsReceipt

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var invoice: Invoice
    public var purchaseOrderLine: PurchaseOrderLine?
    public var goodsReceipt: GoodsReceipt?
    public var lineNumber: Int
    public var lineDescription: String
    public var quantity: Decimal
    public var unitPrice: Decimal
    public var amount: Decimal

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                invoice: Invoice,
                purchaseOrderLine: PurchaseOrderLine? = nil,
                goodsReceipt: GoodsReceipt? = nil,
                lineNumber: Int,
                lineDescription: String,
                quantity: Decimal,
                unitPrice: Decimal,
                amount: Decimal) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.invoice = invoice
        self.purchaseOrderLine = purchaseOrderLine
        self.goodsReceipt = goodsReceipt
        self.lineNumber = lineNumber
        self.lineDescription = lineDescription
        self.quantity = quantity
        self.unitPrice = unitPrice
        self.amount = amount
    }
}
