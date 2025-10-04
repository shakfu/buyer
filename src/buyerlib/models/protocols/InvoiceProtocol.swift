import Foundation

public protocol InvoiceProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype SupplierType: SupplierProtocol
    associatedtype ProjectType: ProjectProtocol
    var invoiceNumber: String { get }
    var supplier: SupplierType { get }
    var project: ProjectType { get }
    var invoiceDate: Date { get }
    var dueDate: Date { get }
    var currency: String { get }
    var totalAmount: Decimal { get }
    var status: String { get }
    var sapInvoiceID: String? { get }
}

