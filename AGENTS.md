# Repository Guidelines

## Project Structure & Module Organization
The Swift Package manifest defines `buyer` (CLI) and `buyerlib` (core services) under `src/`. Command-line entry points live in `src/buyer`, reusable workflows in `src/buyerlib`, and C interop helpers in `src/factorial`. Integration data, templates, and workbook samples sit in `xlsx/` and `doc/`. Tests are split between `tests/buyerlibTests` for library units and `tests/buyerTests` for CLI-level coverage.

## Build, Test, and Development Commands
Run `swift build` for a debug build, or `make build` to inject Homebrew include/lib paths automatically. Use `swift run buyer status` to verify the CLI wiring. Release artifacts come from `make release`. Apply formatting with `make format` (wraps `swift-format --configuration .swiftformatrc`). Clean builds via `make clean`.

## Coding Style & Naming Conventions
Use four-space indentation and keep lines under 100 characters, as enforced by `.swiftformatrc`. Favor lower camel case for Swift identifiers and keep module names in lower case (e.g., `buyerlib`). Imports should stay alphabetized; the formatter will reorder them. Avoid block comments and semicolons. Scope declarations `private` unless a wider access level is intentional.

## Testing Guidelines
All tests use XCTest. Name test files `SomethingTests.swift` under the matching `tests` subfolder and mirror the type under test. Execute `swift test --enable-code-coverage` before sending a patch; capture the resulting `.build/debug/codecov` summary when coverage is relevant. Add regression tests for every bug fix and prefer scenario-focused helper methods over shared mutable state.

## Commit & Pull Request Guidelines
Recent history favors concise, imperative titles (e.g., `Improve swift build`). Limit each commit to a single concern, expanding in the body if context or testing notes help reviewers. For pull requests, include: purpose, notable implementation choices, test output (`swift test`), and any screenshots or sample XLSX artifacts touched. Link issues with `Fixes #123` syntax when applicable.

## Agent Tips
When wiring new commands, register them in `Buyer.Configuration.subcommands` and surface a focused `abstract`. For new templates or workbook assets, document usage in `doc/` so the CLI help stays brief.
