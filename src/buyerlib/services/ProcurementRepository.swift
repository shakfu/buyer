import Foundation

public protocol ProcurementRepository {
    func loadData() throws -> ProcurementDataSet
}

public enum ProcurementRepositoryError: Error {
    case missingRecord(String)
    case storageFailure(String)
}
