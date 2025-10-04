import Foundation

public struct UserAccount: UserAccountProtocol {
    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var username: String
    public var email: String
    public var sourceSystem: String?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                username: String,
                email: String,
                sourceSystem: String? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.username = username
        self.email = email
        self.sourceSystem = sourceSystem
    }
}

