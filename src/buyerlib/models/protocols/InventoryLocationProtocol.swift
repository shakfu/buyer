import Foundation

public protocol InventoryLocationProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype ProjectType: ProjectProtocol
    var code: String { get }
    var name: String { get }
    var siteType: String { get }
    var project: ProjectType? { get }
    var address: String? { get }
}

