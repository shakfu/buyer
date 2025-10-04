import Foundation

public protocol RFxEventProtocol: ModelIdentifiable, TenantScoped, Timestamped {
    associatedtype ProjectType: ProjectProtocol
    var project: ProjectType { get }
    var eventNumber: String { get }
    var type: String { get }
    var status: String { get }
    var submissionDeadline: Date { get }
}

