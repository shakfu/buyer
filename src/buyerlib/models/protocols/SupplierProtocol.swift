import Foundation

public protocol SupplierProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    var sapVendorID: String { get }
    var legalName: String { get }
    var country: String { get }
    var riskRating: String? { get }
}

