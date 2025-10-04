import Foundation

public protocol ContractProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    associatedtype SupplierType: SupplierProtocol
    associatedtype ProjectType: ProjectProtocol
    associatedtype EventType: RFxEventProtocol
    var contractNumber: String { get }
    var supplier: SupplierType { get }
    var project: ProjectType? { get }
    var sourceEvent: EventType? { get }
    var type: String { get }
    var effectiveDate: Date { get }
    var expiryDate: Date? { get }
    var status: String { get }
}

