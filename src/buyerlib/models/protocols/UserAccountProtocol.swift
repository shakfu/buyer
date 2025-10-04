import Foundation

public protocol UserAccountProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    var username: String { get }
    var email: String { get }
    var sourceSystem: String? { get }
}

