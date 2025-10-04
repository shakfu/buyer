import Foundation

public struct Project: ProjectProtocol {
    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var code: String
    public var name: String
    public var sapWBSElement: String
    public var status: String
    public var startDate: Date
    public var endDate: Date?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                code: String,
                name: String,
                sapWBSElement: String,
                status: String,
                startDate: Date,
                endDate: Date? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.code = code
        self.name = name
        self.sapWBSElement = sapWBSElement
        self.status = status
        self.startDate = startDate
        self.endDate = endDate
    }
}

