// swift-tools-version:5.5

import PackageDescription


let package = Package(
    name: "swiftbuyer",
    platforms: [
        .macOS(.v10_15)
    ],
    products: [
        .library(name: "buyerlib", targets: ["buyerlib"]),
        .executable(name: "buyer", targets: ["buyer"]),
    ],
    dependencies: [
        .package(url: "https://github.com/michaelnisi/skull", from: "11.0.4"),
        .package(url: "https://github.com/apple/swift-argument-parser", from: "0.3.0"),
        .package(url: "https://github.com/stephencelis/SQLite.swift.git", from: "0.12.0"),
        .package(url: "https://github.com/stencilproject/Stencil.git", from: "0.14.0"),
        .package(url: "https://github.com/weichsel/ZIPFoundation.git", from: "0.9.16"),
        .package(url: "https://github.com/apple/swift-log.git", from: "1.0.0"),
        // .package(url: "https://github.com/AudioKit/AudioKit", .branch("v5.2.0")),
    ],
    targets: [
        .target(
            name: "SQLiteSupport",
            dependencies: [
                .product(name: "SQLite", package: "SQLite.swift"),
            ],
            path: "Sources/SQLiteSupport",
            resources: [
                .copy("Resources/PrivacyInfo.xcprivacy")
            ]),

        .target(
            name: "buyerlib",
            dependencies: [
                "cfactorial",
                .product(name: "Skull", package: "skull"),
                .product(name: "Stencil", package: "Stencil"),
                "SQLiteSupport",
                .product(name: "ZIPFoundation", package: "ZIPFoundation")
            ],
            path: "src/buyerlib"),
            // dependencies: ["cfactorial", "Skull", "Stencil", "AudioKit"]),
        .testTarget(
            name: "buyerlibTests",
            dependencies: ["buyerlib"],
            path: "tests/buyerlibTests"),

        .executableTarget(
            name: "buyer",
            dependencies: ["buyerlib", 
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
                .product(name: "Logging", package: "swift-log"),
            ],
            path: "src/buyer"),
        .testTarget(
            name: "buyerTests",
            dependencies: ["buyer"],
            path: "tests/buyerTests"), 

        .target(
            name: "cfactorial",
            path: "./src/factorial"),
    ]
)
