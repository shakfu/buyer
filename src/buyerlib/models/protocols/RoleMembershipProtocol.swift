import Foundation

public protocol RoleMembershipProtocol: ModelIdentifiable, Timestamped {
    associatedtype UserType: UserAccountProtocol
    associatedtype RoleType: RoleProtocol
    associatedtype ProjectType: ProjectProtocol
    var user: UserType { get }
    var role: RoleType { get }
    var project: ProjectType? { get }
    var assignedAt: Date { get }
    var expiresAt: Date? { get }
}

