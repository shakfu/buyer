import Foundation

public struct RoleMembership: RoleMembershipProtocol {
    public typealias UserType = UserAccount
    public typealias RoleType = Role
    public typealias ProjectType = Project

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var user: UserAccount
    public var role: Role
    public var project: Project?
    public var assignedAt: Date
    public var expiresAt: Date?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                user: UserAccount,
                role: Role,
                project: Project? = nil,
                assignedAt: Date,
                expiresAt: Date? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.user = user
        self.role = role
        self.project = project
        self.assignedAt = assignedAt
        self.expiresAt = expiresAt
    }
}
