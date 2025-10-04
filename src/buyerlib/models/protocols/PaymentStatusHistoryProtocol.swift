import Foundation

public protocol PaymentStatusHistoryProtocol: ModelIdentifiable, Timestamped {
    associatedtype InvoiceType: InvoiceProtocol
    var invoice: InvoiceType { get }
    var status: String { get }
    var statusDate: Date { get }
    var sapReference: String? { get }
    var notes: String? { get }
}

