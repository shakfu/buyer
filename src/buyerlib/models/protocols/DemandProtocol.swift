import Foundation

public protocol DemandProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype ProjectType: ProjectProtocol
    var project: ProjectType { get }
    var category: String { get }
    var demandDescription: String { get }
    var requiredDate: Date { get }
    var quantity: Decimal { get }
    var unitOfMeasure: String { get }
    var status: String { get }
}

