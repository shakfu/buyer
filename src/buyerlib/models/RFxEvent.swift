import Foundation

public struct RFxEvent: RFxEventProtocol {
    public typealias ProjectType = Project

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var project: Project
    public var eventNumber: String
    public var type: String
    public var status: String
    public var submissionDeadline: Date

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                project: Project,
                eventNumber: String,
                type: String,
                status: String,
                submissionDeadline: Date) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.project = project
        self.eventNumber = eventNumber
        self.type = type
        self.status = status
        self.submissionDeadline = submissionDeadline
    }
}

