import Foundation

public struct PurchaseOrder: PurchaseOrderProtocol {
    public typealias SupplierType = Supplier
    public typealias ProjectType = Project
    public typealias ContractType = Contract

    public var id: UUID
    public var tenantID: UUID
    public var createdAt: Date
    public var updatedAt: Date?
    public var supplier: Supplier
    public var project: Project
    public var contract: Contract?
    public var poNumber: String
    public var issuedAt: Date
    public var currency: String
    public var status: String
    public var sapPONumber: String?

    public init(id: UUID,
                tenantID: UUID,
                createdAt: Date,
                updatedAt: Date? = nil,
                supplier: Supplier,
                project: Project,
                contract: Contract? = nil,
                poNumber: String,
                issuedAt: Date,
                currency: String,
                status: String,
                sapPONumber: String? = nil) {
        self.id = id
        self.tenantID = tenantID
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.supplier = supplier
        self.project = project
        self.contract = contract
        self.poNumber = poNumber
        self.issuedAt = issuedAt
        self.currency = currency
        self.status = status
        self.sapPONumber = sapPONumber
    }
}

