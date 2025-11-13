# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
  - Web forms now automatically clear input values after successful submission
  - Improved user experience by resetting forms to default state after adding new items
  - Refactored web handlers to eliminate ~850 lines of duplicated code by consolidating CRUD endpoints into `SetupCRUDHandlers()` function

### Fixed
  - Removed massive code duplication in route handlers (C1 from CODE_REVIEW.md)
  - Replaced inline HTML generation with consistent use of render functions

## [0.1.0]

### Added
  - Initial implementation created