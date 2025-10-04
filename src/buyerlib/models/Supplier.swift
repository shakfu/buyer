import Foundation

public struct Supplier: SupplierProtocol {
    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var isActive: Bool
    public var sapVendorID: String
    public var legalName: String
    public var country: String
    public var riskRating: String?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                isActive: Bool,
                sapVendorID: String,
                legalName: String,
                country: String,
                riskRating: String? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.isActive = isActive
        self.sapVendorID = sapVendorID
        self.legalName = legalName
        self.country = country
        self.riskRating = riskRating
    }
}

