import Foundation

public protocol PurchaseRequestProtocol: ModelIdentifiable, TenantScoped, Timestamped, SoftDeletable {
    associatedtype ProjectType: ProjectProtocol
    associatedtype RequesterType: UserAccountProtocol
    associatedtype DemandType: DemandProtocol
    var requestNumber: String { get }
    var project: ProjectType { get }
    var requester: RequesterType { get }
    var demands: [DemandType] { get }
    var justification: String? { get }
    var neededBy: Date? { get }
    var status: String { get }
}

