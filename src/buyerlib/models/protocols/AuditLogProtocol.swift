import Foundation

public protocol AuditLogProtocol: ModelIdentifiable, TenantScoped {
    associatedtype ActorType: UserAccountProtocol
    var actor: ActorType? { get }
    var action: String { get }
    var objectType: String { get }
    var objectID: UUID? { get }
    var details: Data? { get }
    var createdAt: Date { get }
    var sourceIP: String? { get }
}

