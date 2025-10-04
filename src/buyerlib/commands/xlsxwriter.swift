import Foundation
import ZIPFoundation

public struct ProcurementReportInput {
    public let summary: ProcurementSummary
    public let supplierSummaries: [SupplierSpendSummary]
    public let dataset: ProcurementDataSet

    public init(summary: ProcurementSummary,
                supplierSummaries: [SupplierSpendSummary],
                dataset: ProcurementDataSet) {
        self.summary = summary
        self.supplierSummaries = supplierSummaries
        self.dataset = dataset
    }
}

public enum ProcurementReportError: Error {
    case archiveCreationFailed
}

public func writeProcurementWorkbook(input: ProcurementReportInput, to destinationURL: URL) throws {
    let worksheets: [XLSXWorksheet] = [
        buildSummaryWorksheet(input: input),
        buildSuppliersWorksheet(input: input),
        buildPurchaseOrdersWorksheet(input: input),
        buildApprovalsWorksheet(input: input),
        buildDeliveriesWorksheet(input: input),
        buildInvoicesWorksheet(input: input)
    ]

    let builder = XLSXBuilder(createdAt: input.summary.generatedAt)
    try builder.write(worksheets: worksheets, to: destinationURL)
}

// MARK: - Worksheet Builders

private func buildSummaryWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    var rows = [[XLSXCell]]()
    rows.append([.string("Metric"), .string("Value")])
    rows.append([.string("Active suppliers"), .number(String(input.summary.activeSuppliers))])
    rows.append([.string("Open purchase orders"), .number(String(input.summary.openPurchaseOrders))])
    rows.append([.string("Deliveries due (7d)"), .number(String(input.summary.deliveriesDueThisWeek))])
    rows.append([.string("Pending approvals"), .number(String(input.summary.pendingApprovals))])
    rows.append([.string("Invoices on hold"), .number(String(input.summary.invoicesOnHold))])

    rows.append([.string("")])
    rows.append([.string("Top Alerts")])

    if input.summary.alerts.isEmpty {
        rows.append([.string("No active alerts")])
    } else {
        for alert in input.summary.alerts.prefix(5) {
            rows.append([.string("[\(alert.severity.rawValue.uppercased())] \(alert.message)")])
        }
    }

    rows.append([.string("")])
    rows.append([.string("Top Suppliers by Open PO")])
    rows.append([.string("Supplier"), .string("Open PO Value"), .string("Invoices on Hold"), .string("Overdue Deliveries")])
    for summary in input.supplierSummaries.prefix(5) {
        rows.append([
            .string(summary.supplier.legalName),
            .number(decimalString(summary.totalOpenPOValue)),
            .number(String(summary.invoicesOnHold)),
            .number(String(summary.overdueDeliveries))
        ])
    }

    return XLSXWorksheet(name: "Summary", rows: rows)
}

private func buildSuppliersWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    var rows = [[XLSXCell]]()
    rows.append([
        .string("Legal Name"),
        .string("Country"),
        .string("Category"),
        .string("Risk"),
        .string("Active"),
        .string("Spend YTD")
    ])

    for supplier in input.dataset.suppliers {
        rows.append([
            .string(supplier.legalName),
            .string(supplier.country),
            .string(supplier.category),
            .string(supplier.riskRating ?? "Unrated"),
            .string(supplier.isActive ? "Yes" : "No"),
            .number(decimalString(supplier.spendYearToDate))
        ])
    }

    return XLSXWorksheet(name: "Suppliers", rows: rows)
}

private func buildPurchaseOrdersWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd"

    var rows = [[XLSXCell]]()
    rows.append([
        .string("PO Number"),
        .string("Supplier"),
        .string("Project"),
        .string("Status"),
        .string("Currency"),
        .string("Total Value"),
        .string("Expected Delivery")
    ])

    for po in input.dataset.purchaseOrders {
        rows.append([
            .string(po.number),
            .string(po.supplierName),
            .string("\(po.projectCode) — \(po.projectName)"),
            .string(po.status.rawValue),
            .string(po.currency),
            .number(decimalString(po.totalValue)),
            .string(formatter.string(from: po.expectedDelivery))
        ])
    }

    return XLSXWorksheet(name: "Purchase Orders", rows: rows)
}

private func buildApprovalsWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd"

    var rows = [[XLSXCell]]()
    rows.append([
        .string("Title"),
        .string("Type"),
        .string("Requested By"),
        .string("Pending With"),
        .string("Due Date"),
        .string("Status")
    ])

    for approval in input.dataset.approvalQueue {
        rows.append([
            .string(approval.title),
            .string(approval.requestType),
            .string(approval.requestedBy),
            .string(approval.pendingWith),
            .string(formatter.string(from: approval.dueDate)),
            .string(approval.status.rawValue)
        ])
    }

    return XLSXWorksheet(name: "Approvals", rows: rows)
}

private func buildDeliveriesWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd"

    var rows = [[XLSXCell]]()
    rows.append([
        .string("PO Number"),
        .string("Description"),
        .string("Expected On"),
        .string("Status")
    ])

    for delivery in input.dataset.deliveryMilestones {
        rows.append([
            .string(delivery.purchaseOrderNumber),
            .string(delivery.description),
            .string(formatter.string(from: delivery.expectedOn)),
            .string(delivery.status.rawValue)
        ])
    }

    return XLSXWorksheet(name: "Deliveries", rows: rows)
}

private func buildInvoicesWorksheet(input: ProcurementReportInput) -> XLSXWorksheet {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd"

    var rows = [[XLSXCell]]()
    rows.append([
        .string("Invoice"),
        .string("Supplier"),
        .string("Amount"),
        .string("Currency"),
        .string("Due Date"),
        .string("Status")
    ])

    for invoice in input.dataset.invoices {
        rows.append([
            .string(invoice.invoiceNumber),
            .string(invoice.supplierName),
            .number(decimalString(invoice.amount)),
            .string(invoice.currency),
            .string(formatter.string(from: invoice.dueDate)),
            .string(invoice.status.rawValue)
        ])
    }

    return XLSXWorksheet(name: "Invoices", rows: rows)
}

// MARK: - XLSX Builder

private struct XLSXWorksheet {
    let name: String
    let rows: [[XLSXCell]]
}

private enum XLSXCell {
    case string(String)
    case number(String)
}

private final class XLSXBuilder {
    private let createdAt: Date

    init(createdAt: Date) {
        self.createdAt = createdAt
    }

    func write(worksheets: [XLSXWorksheet], to destinationURL: URL) throws {
        let fileManager = FileManager.default
        if fileManager.fileExists(atPath: destinationURL.path) {
            try fileManager.removeItem(at: destinationURL)
        }
        try fileManager.createDirectory(at: destinationURL.deletingLastPathComponent(), withIntermediateDirectories: true)

        guard let archive = Archive(url: destinationURL, accessMode: .create) else {
            throw ProcurementReportError.archiveCreationFailed
        }

        let createdTimestamp = iso8601String(from: createdAt)

        try addEntry(on: archive,
                     path: "[Content_Types].xml",
                     data: contentTypesXML(worksheets: worksheets).data(using: .utf8)!)

        try addEntry(on: archive,
                     path: "_rels/.rels",
                     data: rootRelationshipsXML().data(using: .utf8)!)

        try addEntry(on: archive,
                     path: "docProps/app.xml",
                     data: appPropertiesXML().data(using: .utf8)!)

        try addEntry(on: archive,
                     path: "docProps/core.xml",
                     data: corePropertiesXML(createdTimestamp: createdTimestamp).data(using: .utf8)!)

        try addEntry(on: archive,
                     path: "xl/workbook.xml",
                     data: workbookXML(worksheets: worksheets).data(using: .utf8)!)

        try addEntry(on: archive,
                     path: "xl/_rels/workbook.xml.rels",
                     data: workbookRelationshipsXML(worksheets: worksheets).data(using: .utf8)!)

        for (index, worksheet) in worksheets.enumerated() {
            let sheetPath = "xl/worksheets/sheet\(index + 1).xml"
            let data = worksheetXML(worksheet: worksheet).data(using: .utf8)!
            try addEntry(on: archive, path: sheetPath, data: data)
        }
    }

    private func addEntry(on archive: Archive, path: String, data: Data) throws {
        try archive.addEntry(with: path,
                             type: .file,
                             uncompressedSize: UInt32(data.count),
                             compressionMethod: .deflate) { position, size -> Data in
            let start = Int(position)
            let end = start + Int(size)
            return data.subdata(in: start..<end)
        }
    }

    private func contentTypesXML(worksheets: [XLSXWorksheet]) -> String {
        var overrides = ""
        for index in worksheets.indices {
            overrides += "    <Override PartName=\"/xl/worksheets/sheet\(index + 1).xml\" ContentType=\"application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml\"/>\n"
        }
        return """
        <?xml version="1.0" encoding="UTF-8"?>
        <Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
            <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
            <Default Extension="xml" ContentType="application/xml"/>
            <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
        \(overrides)    <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
            <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
        </Types>
        """
    }

    private func rootRelationshipsXML() -> String {
        """
        <?xml version="1.0" encoding="UTF-8"?>
        <Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
            <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
            <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
            <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
        </Relationships>
        """
    }

    private func appPropertiesXML() -> String {
        """
        <?xml version="1.0" encoding="UTF-8"?>
        <Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
            <Application>buyer</Application>
        </Properties>
        """
    }

    private func corePropertiesXML(createdTimestamp: String) -> String {
        """
        <?xml version="1.0" encoding="UTF-8"?>
        <cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <dc:creator>buyer</dc:creator>
            <cp:lastModifiedBy>buyer</cp:lastModifiedBy>
            <dcterms:created xsi:type="dcterms:W3CDTF">\(createdTimestamp)</dcterms:created>
            <dcterms:modified xsi:type="dcterms:W3CDTF">\(createdTimestamp)</dcterms:modified>
        </cp:coreProperties>
        """
    }

    private func workbookXML(worksheets: [XLSXWorksheet]) -> String {
        var sheetEntries = ""
        for (index, worksheet) in worksheets.enumerated() {
            let name = sanitizeSheetName(worksheet.name)
            sheetEntries += "        <sheet name=\"\(escapeXML(name))\" sheetId=\"\(index + 1)\" r:id=\"rId\(index + 1)\"/>\n"
        }
        return """
        <?xml version="1.0" encoding="UTF-8"?>
        <workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
            <sheets>
        \(sheetEntries)    </sheets>
        </workbook>
        """
    }

    private func workbookRelationshipsXML(worksheets: [XLSXWorksheet]) -> String {
        var relationships = ""
        for index in worksheets.indices {
            relationships += "    <Relationship Id=\"rId\(index + 1)\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet\" Target=\"worksheets/sheet\(index + 1).xml\"/>\n"
        }
        return """
        <?xml version="1.0" encoding="UTF-8"?>
        <Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
        \(relationships)</Relationships>
        """
    }

    private func worksheetXML(worksheet: XLSXWorksheet) -> String {
        var rowsXML = ""
        for (rowIndex, row) in worksheet.rows.enumerated() {
            let rowNumber = rowIndex + 1
            var cellsXML = ""
            for (columnIndex, cell) in row.enumerated() {
                let cellReference = columnReference(columnIndex: columnIndex, rowIndex: rowNumber)
                switch cell {
                case .string(let value):
                    cellsXML += "                <c r=\"\(cellReference)\" t=\"inlineStr\"><is><t>\(escapeXML(value))</t></is></c>\n"
                case .number(let value):
                    cellsXML += "                <c r=\"\(cellReference)\"><v>\(escapeXML(value))</v></c>\n"
                }
            }
            rowsXML += "            <row r=\"\(rowNumber)\">\n\(cellsXML)            </row>\n"
        }

        return """
        <?xml version="1.0" encoding="UTF-8"?>
        <worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
            <sheetData>
        \(rowsXML)    </sheetData>
        </worksheet>
        """
    }

    private func columnReference(columnIndex: Int, rowIndex: Int) -> String {
        let columnString = columnLetters(for: columnIndex)
        return "\(columnString)\(rowIndex)"
    }

    private func columnLetters(for index: Int) -> String {
        var value = index
        var letters = ""
        repeat {
            let remainder = value % 26
            let scalar = UnicodeScalar(65 + remainder)!
            letters = String(scalar) + letters
            value = (value / 26) - 1
        } while value >= 0
        return letters
    }

    private func escapeXML(_ value: String) -> String {
        var escaped = value.replacingOccurrences(of: "&", with: "&amp;")
        escaped = escaped.replacingOccurrences(of: "<", with: "&lt;")
        escaped = escaped.replacingOccurrences(of: ">", with: "&gt;")
        escaped = escaped.replacingOccurrences(of: "\"", with: "&quot;")
        escaped = escaped.replacingOccurrences(of: "'", with: "&apos;")
        return escaped
    }

    private func sanitizeSheetName(_ name: String) -> String {
        let invalidCharacters: CharacterSet = CharacterSet(charactersIn: "[]:*?/\\")
        let filtered = name.unicodeScalars.map { invalidCharacters.contains($0) ? "_" : String($0) }.joined()
        if filtered.count > 31 {
            let endIndex = filtered.index(filtered.startIndex, offsetBy: 31)
            return String(filtered[..<endIndex])
        }
        return filtered.isEmpty ? "Sheet" : filtered
    }

    private func iso8601String(from date: Date) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter.string(from: date)
    }
}

private func decimalString(_ value: Decimal) -> String {
    NSDecimalNumber(decimal: value).stringValue
}
