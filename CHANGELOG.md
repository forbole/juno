# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed

- Update SDK version to v0.36.0
  - The `transaction` schema now uses `events` instead of `tags`
- Improved error messages and logs

### Added

- Add support for manual PostgreSQL SSL mode configuration

### Fixed

- Fixed `db.OpenDB` to handle no password correctly

## [0.0.3] - 2019-06-23

### Fixed

- Fixed logging in `Worker#Start`.

## [0.0.2] - 2019-06-23

### Added

- Added additional indexes to the `pre_commit` table.

### Changed

- Updated `Database` to not check if a validator exists; use `ON CONFLICT DO NOTHING`
instead.
- Use `zerolog` logger with `--log-level` and `--log-format` CLI options over the
stdlib `log` package.

## [0.0.1] - 2019-06-21

### Added

- Initial release

[Unreleased]: https://github.com/alexanderbez/juno/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/alexanderbez/juno/releases/tag/v0.0.3
[0.0.2]: https://github.com/alexanderbez/juno/releases/tag/v0.0.2
[0.0.1]: https://github.com/alexanderbez/juno/releases/tag/v0.0.1

<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Types of changes: 

"Added" for new features.
"Changed" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Removed" for now removed features.
"Fixed" for any bug fixes.
"Security" in case of vulnerabilities.
-->
