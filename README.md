# nls

A Nushell-style `ls` for bash, zsh, and fish. `nls` prints directory listings as a clean table with optional icons and colors.

## Features

- Table columns: `#`, name, type, size, modified (permissions with `-l`)
- Sorted alphabetically by name
- Directory sizes estimated by summing contained files when interactive (skipped when piped; prefix `~` if truncated)
- Hidden files with `-a` / `--all`
- Human-readable sizes with `-h` / `--human`
- JSON output with `--json`
- Nerd Font icons when enabled via `NLS_ICONS=1`
- Plain one-name-per-line output when piped
- Colors for directories (blue), symlinks (cyan), executables (green)

## Requirements

- Go 1.22+
- Linux or macOS

## Install

```bash
go install github.com/nolight132/nls/cmd/nls@latest
```

Or build locally:

```bash
git clone https://github.com/nolight132/nls.git
cd nls
go build -o nls ./cmd/nls
sudo mv nls /usr/local/bin/
```

## Output modes

| Context                    | Behavior                                  |
| -------------------------- | ----------------------------------------- |
| **TTY, no ls flags**       | Nushell table, colors, dir size estimates |
| **TTY + `-l`/`-1`/`-F`/…** | Native `ls` formatting (no table)         |
| **Piped / redirected**     | Native `ls` fast path (no stat per file)  |
| **Piped + `-l`/`-F`/…**    | Full `ls` flag support                    |
| **`--json`**               | JSON                                      |

Nushell styling (table, `LS_COLORS`, purple modified, dir walks) is **TTY-only** with default flags.

Clustered short flags work: `nls -la`, `nls -ltr`, `nls -laR`.

## Usage

```bash
# Current directory (table on TTY)
nls

# GNU long listing (works when piped)
nls -la
nls -lah ~/Downloads

# One per line, classify, sort
nls -1
nls -F
nls -lt
nls -lS

# Recursive
nls -R

# Pipe-friendly
nls | wc -l
nls -la | grep '\.go$'
nls -1 ~/bin | xargs -I{} echo {}

# JSON
nls --json | jq .
```

## Flags

POSIX/GNU `ls` flags:

| Flag                        | Short | Description                                  |
| --------------------------- | ----- | -------------------------------------------- |
| `--all`                     | `-a`  | Show hidden entries (including `.` and `..`) |
| `--almost-all`              | `-A`  | Show hidden except `.` and `..`              |
| `--long`                    | `-l`  | Long listing format                          |
| `--human-readable`          | `-h`  | Human sizes with `-l` / table                |
| `--one`                     | `-1`  | One file per line                            |
| `--recursive`               | `-R`  | List subdirectories recursively              |
| `--reverse`                 | `-r`  | Reverse sort order                           |
| `--time`                    | `-t`  | Sort by modification time                    |
| `--access-time`             | `-u`  | Sort by access time                          |
| `--ctime`                   | `-c`  | Sort by change time                          |
| `--size`                    | `-S`  | Sort by size                                 |
| `--extension`               | `-X`  | Sort by extension                            |
| `--unsorted`                | `-U`  | Do not sort                                  |
| `--fast`                    | `-f`  | Do not sort (same as `-U`)                   |
| `--directory`               | `-d`  | List directories themselves                  |
| `--classify`                | `-F`  | Append `*`, `/`, `@`, etc.                   |
| `--slash`                   | `-p`  | Append `/` to directories                    |
| `--ignore-backups`          | `-B`  | Skip `*~` files                              |
| `--dereference`             | `-L`  | Follow symlinks                              |
| `--comma`                   | `-m`  | Comma-separated output                       |
| `--quote-name`              | `-Q`  | Quote names                                  |
| `--full-time`               |       | Full timestamps with `-l`                    |
| `--group-directories-first` |       | Directories before files                     |
| `--inode`                   | `-i`  | Show inode with `-l`                         |
| `--size-blocks`             | `-s`  | Show blocks with `-l`                        |

`nls`-specific:

| Flag         | Description    |
| ------------ | -------------- |
| `--json`     | JSON output    |
| `--no-icons` | Disable icons  |
| `--no-color` | Disable colors |

## Nerd Font icons

Icons are off by default (like Nushell `ls`). Enable with:

```bash
NLS_ICONS=1 nls
NLS_ICONS=nerd nls   # Nerd Font glyphs
NLS_ICONS=emoji nls  # emoji icons
```

## Colors

`nls` uses an `LS_COLORS`-compatible coloring model:

- Reads `LS_COLORS` when set
- Falls back to simple theme-friendly ANSI colors for directories, symlinks, and executables
- Sizes are cyan; modified column is purple

## Development

```bash
go fmt ./...
go test ./...
go run ./cmd/nls
```

## License

MIT
