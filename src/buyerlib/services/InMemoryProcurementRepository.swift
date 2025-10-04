import Foundation

public final class InMemoryProcurementRepository: ProcurementRepository {
    private let dataSet: ProcurementDataSet

    public init(seedDate: Date = Date(), tenantID: UUID = UUID(uuidString: "00000000-0000-0000-0000-000000000001")!) {
        dataSet = ProcurementSeedData.makeDataSet(seedDate: seedDate, tenantID: tenantID)
    }

    public func loadData() throws -> ProcurementDataSet {
        return dataSet
    }
}
