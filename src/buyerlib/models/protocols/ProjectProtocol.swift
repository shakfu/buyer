import Foundation

public protocol ProjectProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    var code: String { get }
    var name: String { get }
    var sapWBSElement: String { get }
    var status: String { get }
    var startDate: Date { get }
    var endDate: Date? { get }
}

