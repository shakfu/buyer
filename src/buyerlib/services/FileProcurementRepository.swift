import Foundation

public final class FileProcurementRepository: ProcurementRepository {
    private let fileURL: URL
    private let seedDate: Date
    private let tenantID: UUID
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder
    private let accessQueue = DispatchQueue(label: "org.me.swiftbuyer.file-procurement-repository", qos: .userInitiated)

    public init(url: URL,
                seedDate: Date = Date(),
                tenantID: UUID = UUID(uuidString: "00000000-0000-0000-0000-000000000001")!) {
        self.fileURL = url
        self.seedDate = seedDate
        self.tenantID = tenantID

        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys, .prettyPrinted]
        encoder.dateEncodingStrategy = .iso8601
        encoder.dataEncodingStrategy = .base64
        self.encoder = encoder

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.dataDecodingStrategy = .base64
        self.decoder = decoder
    }

    public func loadData() throws -> ProcurementDataSet {
        return try accessQueue.sync {
            let fileManager = FileManager.default
            let directoryURL = fileURL.deletingLastPathComponent()
            if !fileManager.fileExists(atPath: directoryURL.path) {
                do {
                    try fileManager.createDirectory(at: directoryURL, withIntermediateDirectories: true)
                } catch {
                    throw ProcurementRepositoryError.storageFailure("Unable to create directory for procurement store: \(error.localizedDescription)")
                }
            }

            if fileManager.fileExists(atPath: fileURL.path) {
                let data: Data
                do {
                    data = try Data(contentsOf: fileURL)
                } catch {
                    throw ProcurementRepositoryError.storageFailure("Unable to read procurement store: \(error.localizedDescription)")
                }

                do {
                    return try decoder.decode(ProcurementDataSet.self, from: data)
                } catch {
                    throw ProcurementRepositoryError.storageFailure("Unable to decode procurement store: \(error.localizedDescription)")
                }
            }

            let seeded = ProcurementSeedData.makeDataSet(seedDate: seedDate, tenantID: tenantID)
            try persist(dataSet: seeded)
            return seeded
        }
    }

    // MARK: - Private

    private func persist(dataSet: ProcurementDataSet) throws {
        let data: Data
        do {
            data = try encoder.encode(dataSet)
        } catch {
            throw ProcurementRepositoryError.storageFailure("Unable to encode procurement data: \(error.localizedDescription)")
        }
        do {
            try data.write(to: fileURL, options: .atomic)
        } catch {
            throw ProcurementRepositoryError.storageFailure("Unable to write procurement data: \(error.localizedDescription)")
        }
    }
}
