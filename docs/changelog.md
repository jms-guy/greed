# Change Log
All notable changes to this project will be documented in this file.

## [Unreleased] - yyyy-mm-dd

## [v1.0.1] - 2025-08-22

### Added
- Enabled authenticated admin sandbox flow for easier testing.
- CLI: Added “test” login mode to quickly log in and initialize a sandbox session.
- CLI: Registration: clearer prompts, option to continue without email verification, and improved resend guidance.
- README: Added Demo section with GIFs (Registration, Account Sync, Transactions/Reports).

### Changed
- Updated Table of Contents label, clarified Plaid usage note and sync call limit, and improved Changelog link wording.

## [v1.0.0] - 2025-08-22

### Added
- CLI: Local error logging with `logs` command
- CLI: Automatic item sync after login
- CLI: Persisted defaults for items/accounts

### Changed
- CLI: Standardized user-friendly error messages
- CLI: Optional command arguments when defaults set

### Fixed
- CLI: Improved webhook reliability during login


## [v1.0.0-beta-3] - 2025-08-16
- Functional deployment of both server & CLI