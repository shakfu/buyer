import XCTest

import buyerlibTests
import buyerTests

var tests = [XCTestCaseEntry]()
tests += buyerlibTests.allTests()
tests += buyerTests.allTests()
XCTMain(tests)
