import Foundation

public protocol SupplierQualificationProtocol: ModelIdentifiable, Timestamped {
    associatedtype SupplierType: SupplierProtocol
    var supplier: SupplierType { get }
    var qualificationType: String { get }
    var validFrom: Date { get }
    var validTo: Date? { get }
    var status: String { get }
    var documentURI: URL? { get }
}

