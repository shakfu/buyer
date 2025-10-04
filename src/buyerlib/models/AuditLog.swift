import Foundation

public struct AuditLog: AuditLogProtocol {
    public typealias ActorType = UserAccount

    public var id: UUID
    public var tenantID: UUID
    public var actor: UserAccount?
    public var action: String
    public var objectType: String
    public var objectID: UUID?
    public var details: Data?
    public var createdAt: Date
    public var sourceIP: String?

    public init(id: UUID,
                tenantID: UUID,
                actor: UserAccount? = nil,
                action: String,
                objectType: String,
                objectID: UUID? = nil,
                details: Data? = nil,
                createdAt: Date,
                sourceIP: String? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.actor = actor
        self.action = action
        self.objectType = objectType
        self.objectID = objectID
        self.details = details
        self.createdAt = createdAt
        self.sourceIP = sourceIP
    }
}

