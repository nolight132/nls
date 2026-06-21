# Changelog

## v0.1.4

- Fixed non-TTY compatibility with GNU `ls` for long output metadata, totals, inode/block columns, locale sorting, unsorted listing, `-f`, `--full-time`, access/ctime sorting, directory grouping, and symlink indicators.
- Added GNU comparison tests covering name, long, recursive, inode, block, sorting, and directory-grouping output.

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
