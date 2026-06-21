# Changelog

## v0.1.3

- Fixed colored table headers so ANSI styling does not break header alignment.

## v0.1.2

- Fixed empty path arguments (`nls ""`) so they return a real path error instead of resolving to the current directory.
- Centralized CLI error formatting so returned errors are printed consistently with the `nls:` prefix.
- Updated missing-path error tests to match real filesystem error behavior.
- Changed human-readable size units from decimal labels (`kB`, `MB`, etc.) to binary labels (`KiB`, `MiB`, etc.) by default. Config coming soon.

## v0.1.1

- Fixed modified time formatting so recent timestamps remain relative instead of being treated as absolute dates.

## v0.1.0

Initial release.
