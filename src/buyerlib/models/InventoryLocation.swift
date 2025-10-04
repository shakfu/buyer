import Foundation

public struct InventoryLocation: InventoryLocationProtocol {
    public typealias ProjectType = Project

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var code: String
    public var name: String
    public var siteType: String
    public var project: Project?
    public var address: String?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                code: String,
                name: String,
                siteType: String,
                project: Project? = nil,
                address: String? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.code = code
        self.name = name
        self.siteType = siteType
        self.project = project
        self.address = address
    }
}

