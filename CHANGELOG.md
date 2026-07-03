# Changelog

## v0.3.0

- Added table-mode handling for empty directories, rendering a bordered `no entries` message instead of an empty output.

## v0.2.0

- Replaced `--estimate-depth` with `-P`/`--precise`, which computes exact directory sizes without depth, time, or entry limits.
- Added `dir_size.enabled` config option to enable directory size estimation by default in interactive table mode, plus an `unlimited` timing preset that drops all caps.
- Changed defaults to enable icons and drop the `type` column from the default table.
- Added `examples/config.default.toml` and `examples/config.nushell.toml`, replacing the old `config.example.toml`.
- Made the index column and table headers bold.
- README cleanup; added a link to the Nushell integration example.

## v0.1.8

- Added configurable table columns via `default_columns` in the config file. All 13 columns are available: `id`, `name`, `type`, `size`, `modified`, `accessed`, `changed`, `permissions`, `links`, `owner`, `group`, `inode`, `blocks`. Omit a column to hide it; flags `-i`/`-s`/`-l` still append their columns if not listed.
- Fixed `--estimate-depth max` hanging on huge filesystems like `/`. Full-walk mode now applies an entry-count safety cap (200k entries per directory, 50 directories) with no time limits and unlimited depth. Truncated sizes are marked approximate.

## v0.1.7

- Added XDG config file support (`$XDG_CONFIG_HOME/nls/config.toml` on Linux/macOS, `%APPDATA%\nls\config.toml` on Windows), TOML-encoded. Initial schema covers the icon toggle, bounded directory size estimation defaults (`dir_size.default_depth`, `dir_size.timing`), and table column selection (`default_columns`).
- Added named timing presets (`strict`, `balanced`, `relaxed`) for bounded estimation budgets; raw millisecond values are no longer the only knob.
- Config `icons` now provides the default icon state; `--no-icons` and `NLS_ICONS` still override it.
- Added `config.example.toml` as a commented template.

## v0.1.6

- Added `--estimate-depth` for directory size estimation: bounded by default in table mode, numeric levels, or `max` for full walks with a safety net so huge trees like `/` cannot hang.
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
