# Changelog

## v0.5.2

- Switch to using `git status --porcelain` for git state detection. Rationale for this is that `go-git` is too large for this simple task and used to make up the majority of the package size. I didn't want to rely on `git` being installed on the system, but eventually came to the conclusion that an average user is unlikely to have git repos on their system without having `git` installed. Systems without `git` fallback to uncolored output.

## v0.5.1

- Fixed directory size estimation silently skipping directories beyond the per-listing cap or time budget: skipped directories showed their bare stat size (e.g. `4.0 KiB`) as if it were the content size. They are now marked with the `>` lower-bound prefix, and the default per-listing cap was raised from 6 to 16 directories so typical listings get fully estimated within the existing time budget.

## v0.5.0

- Added `-g`/`--git-status` for a `git` table column showing per-entry status as `staging│worktree` (`?│?` untracked, ` │M` modified, `I│ ` ignored). Directories aggregate the status of their contents. The column can be enabled by default via `default_columns = ["git"]` and hides itself outside git repositories.
- Added git-state coloring of entry names in interactive listings: modified entries render yellow, untracked bright green, ignored gray. Enabled by default; toggle with `color_entries` under `[git]` in the config.
- Added a `git_status` field to `--json` output when git status is computed.
- Added global gitignore support: ignored-entry detection honors `core.excludesFile` and falls back to `$XDG_CONFIG_HOME/git/ignore`, matching git's own behavior instead of only reading per-repo `.gitignore`.
- Changed headings, table headers, and the index column from green to blue so untracked green stays distinct.
- Changed directory modified times to reflect the newest change anywhere in the subtree, so parent directories no longer look stale when only deeply-nested files changed.
- Changed `-R`/`--recursive` to skip descending into `.git` directories.
- Changed unsorted output (`-f`) to always list `.` and `..` first.

## v0.4.1

- Fixed permission strings for special files: FIFOs, sockets, and block/character devices now show their type char (`p`, `s`, `b`, `c`), and setuid/setgid/sticky bits render as `s`/`S`/`t`/`T` (`sudo` shows `-rwsr-xr-x` instead of `-rwxr-xr-x`).
- Fixed `-F`/`--classify` to append `|` for FIFOs and `=` for sockets, as the help text documents.
- Fixed table rendering to fit the terminal width. The name column shrinks and truncates with an ellipsis instead of wrapping every row when a long filename overflows.
- Changed `-f`/`--fast` to imply `-a` again, per POSIX.
- Fixed `--json` to be machine-parseable: RFC 3339 timestamps and a full `path` field in multi-directory or recursive listings.
- Fixed listings to continue past failed operands, entries deleted mid-listing, and unreadable subdirectories under `-R`. Errors go to stderr, readable entries still print, and the exit code is nonzero.
- Fixed control characters in filenames to render as `?` on terminals so crafted names cannot break table layout or inject escape sequences; piped output keeps raw names.
- Fixed future timestamps to show the date instead of "just now".
- Fixed human-readable sizes that round up to exactly 1024.0 to promote to the next unit.
- Fixed `-u`/`-c` to show the accessed/changed column in long output.
- Fixed `LS_COLORS` parsing to reject non-SGR values such as `ln=target`.
- Fixed directory size estimation to skip `.` and `..`, and the raw Linux getdents path to skip deleted-but-present dirents (inode 0).
- Removed the stale `NLS_ICONS` mention from the example config; environment-based icon configuration was removed in v0.4.0.
- Improved sorting performance (`slices.SortStableFunc` instead of insertion sort) and memoized owner/group name lookups.

## v0.4.0

- Removed the GNU `ls` compatibility layer. Plain output is now simpler and follows `nls` behavior for common listing workflows instead of trying to match GNU-specific edge cases.
- Changed plain `-l`/`--long` output to render the configured `nls` columns as aligned plain text without table borders or headers.
- Removed `--full-time`; timestamp output now uses the normal `nls` modified/accessed/changed time formatting.
- Changed `-f`/`--fast` to mean unsorted output only; it no longer implies `-a`/`--all`.
- Removed environment-based icon configuration. `NLS_ICONS`, `NERD_FONT`, and `NLS_NERD_FONT` no longer affect icon selection; use the config file or `--no-icons` instead.
- Fixed CJK and other wide-character table alignment by measuring terminal display width instead of rune count.
- Fixed `LS_COLORS` extension and suffix matching to be case-sensitive.
- Fixed ANSI width calculations so non-SGR CSI sequences and OSC sequences do not break table alignment.
- Fixed `--no-color`/color-disabled rendering so one `nls` render no longer mutates global color state for later output.

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
