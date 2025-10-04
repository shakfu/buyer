import Foundation

public protocol RoleProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    var code: String { get }
    var name: String { get }
    var roleDescription: String? { get }
}

