# Change Log
All notable changes to this project will be documented in this file.

## [Unreleased] - yyyy-mm-dd

## [v1.0.0] - 2025-08-22

### Added
- CLI
    - Local error logging + new logs command to view recent errors.
    - Automatic item update/sync after login.
    - Persisted defaults: set/clear default item and account.

### Changed
- CLI
    - Standardized, user-friendly error messages with detailed errors stored locally.
    - Certain command arguments can now be ommitted, if default values have been set in settings.

### Fixed
- CLI
    - More reliable webhook checking during user login.


## [v1.0.0-beta-3] - 2025-08-16
- Functional deployment of both server & CLI