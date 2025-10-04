import Foundation

public protocol Timestamped {
    var createdAt: Date { get }
    var updatedAt: Date? { get }
}

