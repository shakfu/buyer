import XCTest
import class Foundation.Bundle

final class buyerTests: XCTestCase {
    func testExample() throws {
        // This is an example of a functional test case.
        // Use XCTAssert and related functions to verify your tests produce the correct
        // results.

        // Some of the APIs that we use below are available in macOS 10.13 and above.
        guard #available(macOS 10.13, *) else {
            return
        }

        let fooBinary = productsDirectory.appendingPathComponent("buyer")

        let process = Process()
        process.executableURL = fooBinary

        let dbURL = URL(fileURLWithPath: NSTemporaryDirectory()).appendingPathComponent("buyer-cli-test.json")
        var environment = ProcessInfo.processInfo.environment
        environment["BUYER_DB_PATH"] = dbURL.path
        process.environment = environment

        let pipe = Pipe()
        process.standardOutput = pipe

        try process.run()
        process.waitUntilExit()

        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        let output = String(data: data, encoding: .utf8)

        guard let output = output else {
            XCTFail("The buyer CLI produced no output")
            return
        }

        XCTAssertTrue(output.hasPrefix("Procurement status as of"), "Status header should be present")
        XCTAssertTrue(output.contains("Active suppliers: 3"), "Active supplier count should be printed")
        XCTAssertTrue(output.contains("Open purchase orders: 3"), "Open PO count should be printed")
        XCTAssertTrue(output.contains("Pending approvals: 2"), "Pending approvals count should be printed")
        XCTAssertTrue(output.contains("Invoices on hold: 2"), "Invoices on hold count should be printed")
        XCTAssertTrue(output.contains("PO-1002"), "Alerts should reference overdue purchase orders")
    }

    /// Returns path to the built products directory.
    var productsDirectory: URL {
      #if os(macOS)
        for bundle in Bundle.allBundles where bundle.bundlePath.hasSuffix(".xctest") {
            return bundle.bundleURL.deletingLastPathComponent()
        }
        fatalError("couldn't find the products directory")
      #else
        return Bundle.main.bundleURL
      #endif
    }

    static var allTests = [
        ("testExample", testExample),
    ]
}
