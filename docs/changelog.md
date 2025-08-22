# Change Log
All notable changes to this project will be documented in this file.

## [Unreleased] - yyyy-mm-dd

### Added
- CLI
    - Log command for in-depth viewing of error logs from database
    - Auto updating/syncing of user items upon user login
    - Commands for setting/clearing default item/account values, that may be used in place of command arguments

### Changed
- CLI
    - User viewable errors are more generic, with verbose error details stored in local database
    - Certain command arguments can now be ommitted, if default values have been set in settings

### Fixed
- CLI
    - Webhook checking on user login now works properly


## [v1.0.0-beta-3] - 2025-08-16
- Functional deployment of both server & CLI