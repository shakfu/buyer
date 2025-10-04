import Foundation

public struct PaymentStatusHistory: PaymentStatusHistoryProtocol {
    public typealias InvoiceType = Invoice

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var invoice: Invoice
    public var status: String
    public var statusDate: Date
    public var sapReference: String?
    public var notes: String?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                invoice: Invoice,
                status: String,
                statusDate: Date,
                sapReference: String? = nil,
                notes: String? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.invoice = invoice
        self.status = status
        self.statusDate = statusDate
        self.sapReference = sapReference
        self.notes = notes
    }
}

