public final class Group {
    public var name: String
    public var permissions: Set<Permission>

    public init(name: String, permissions: [Permission] = []) {
        self.name = name
        self.permissions = Set(permissions)
    }

    public init(name: String, permissions: Set<Permission>) {
        self.name = name
        self.permissions = permissions
    }

    public func addPermission(_ permission: Permission) {
        permissions.insert(permission)
    }
}
