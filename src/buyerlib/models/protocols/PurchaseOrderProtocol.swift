import Foundation

public protocol PurchaseOrderProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype SupplierType: SupplierProtocol
    associatedtype ProjectType: ProjectProtocol
    associatedtype ContractType: ContractProtocol
    var poNumber: String { get }
    var supplier: SupplierType { get }
    var project: ProjectType { get }
    var contract: ContractType? { get }
    var issuedAt: Date { get }
    var currency: String { get }
    var status: String { get }
    var sapPONumber: String? { get }
}

