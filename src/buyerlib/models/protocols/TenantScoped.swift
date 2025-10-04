import Foundation

public protocol TenantScoped {
    var tenantID: UUID { get }
}

