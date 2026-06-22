# Changelog

## v0.1.7

- Added XDG config file support (`$XDG_CONFIG_HOME/nls/config.toml` on Linux/macOS, `%APPDATA%\nls\config.toml` on Windows), TOML-encoded. Initial schema covers the icon toggle and bounded directory size estimation defaults (`dir_size.default_depth`, `dir_size.timing`).
- Added named timing presets (`strict`, `balanced`, `relaxed`) for bounded estimation budgets; raw millisecond values are no longer the only knob.
- Config `icons` now provides the default icon state; `--no-icons` and `NLS_ICONS` still override it.
- Added `config.example.toml` as a commented template.

## v0.1.6

- Added `--estimate-depth` for directory size estimation: bounded by default in table mode, numeric levels, or `max` for full walks without time limits.
- Added table-mode guard so bordered output is not skipped when `--estimate-depth` is unset.
- Added disk-usage-based directory size estimation instead of apparent file size, for sane sizes on `/proc`, sparse images, and similar paths.
- Added cross-mount directory walks so overlay-backed directories (e.g. Docker, containerd) are included in estimates.

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
