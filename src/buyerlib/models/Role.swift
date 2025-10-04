import Foundation

public enum RoleKind: String, Codable {
    case buyer
    case approver
    case administrator
    case analyst
    case custom
}

public final class Role: RoleProtocol {
    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var code: String
    public var name: String
    public var roleDescription: String?

    public var kind: RoleKind
    public weak var reportsTo: Role?
    public var childRoles: [Role]
    public var groups: [Group]

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                code: String,
                name: String,
                roleDescription: String? = nil,
                kind: RoleKind = .custom,
                reportsTo: Role? = nil,
                childRoles: [Role] = [],
                groups: [Group] = []) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.code = code
        self.name = name
        self.roleDescription = roleDescription
        self.kind = kind
        self.reportsTo = reportsTo
        self.childRoles = childRoles
        self.groups = groups
    }

    public func addChild(_ role: Role) {
        childRoles.append(role)
        role.reportsTo = self
    }
}
