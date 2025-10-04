import Foundation

public struct SupplierQualification: SupplierQualificationProtocol {
    public typealias SupplierType = Supplier

    public var id: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var supplier: Supplier
    public var qualificationType: String
    public var validFrom: Date
    public var validTo: Date?
    public var status: String
    public var documentURI: URL?

    public init(id: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                supplier: Supplier,
                qualificationType: String,
                validFrom: Date,
                validTo: Date? = nil,
                status: String,
                documentURI: URL? = nil) {
        self.id = id
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.supplier = supplier
        self.qualificationType = qualificationType
        self.validFrom = validFrom
        self.validTo = validTo
        self.status = status
        self.documentURI = documentURI
    }
}
