import Foundation

public struct Demand: DemandProtocol {
    public typealias ProjectType = Project

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var project: Project
    public var category: String
    public var demandDescription: String
    public var requiredDate: Date
    public var quantity: Decimal
    public var unitOfMeasure: String
    public var status: String

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                project: Project,
                category: String,
                demandDescription: String,
                requiredDate: Date,
                quantity: Decimal,
                unitOfMeasure: String,
                status: String) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.project = project
        self.category = category
        self.demandDescription = demandDescription
        self.requiredDate = requiredDate
        self.quantity = quantity
        self.unitOfMeasure = unitOfMeasure
        self.status = status
    }
}

