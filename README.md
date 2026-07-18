<div align="center">

# nls

[![Build](https://img.shields.io/github/actions/workflow/status/nolight132/nls/ci.yml?branch=main)](https://github.com/nolight132/nls/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/nolight132/nls)](./LICENSE)

### neo-ls: a modern `ls` with useful tables

A fast, cross-platform file listing tool that feels familiar in scripts
and looks beautiful in your terminal.

`nls` is heavily inspired by [Nushell](https://github.com/nushell)'s `ls`.
</div>

<div align="center">
    <img width="837" height="622" alt="image" src="https://github.com/user-attachments/assets/a2d364a4-049c-4815-8cbb-54fa499e8cec" />
</div>
<div align="center">
    <sub>
      A more Nushell-like configuration is available in
      <a href="./examples/config.nushell.toml"><code>config.nushell.toml</code></a>.
    </sub>
</div>

---

## Why nls?

`nls` is a **neo-ls**: a modern file listing command built around beautiful tables, compatibility in pipes and scripts, and useful defaults like directory sizes without slowing normal usage.

Nushell's `ls` already provides this experience, but not everyone wants to switch shells — many users are happy with bash, zsh, fish, or PowerShell and just want the table layouts of `nu ls` without Nushell's programming model and compatibility tradeoffs.

`nls` exists for people who want the visual experience of modern terminal tools while keeping the workflows they already know.

It works in bash, zsh, fish, Nushell, PowerShell, and any terminal on Linux, macOS, or Windows where a normal CLI binary can run.

---

## Features

- Nushell-style tables for interactive terminal use
- Git status per listed entry
- Directory sizes shown by default
- Fast non-TTY behavior for pipes, redirects, and scripts
- `ls`-like behavior for common workflows
- Helpful suggestions when paths are mistyped
- Optional icons (enabled by default)
- Colors for files, directories, symlinks, executables, sizes, and timestamps
- JSON output for structured usage
- Works on Linux, macOS, and Windows, and (probably™) everywhere else you would want to run it

---

## Install

Please note that in order to use icons, you need to have a [Nerd Font](https://www.nerdfonts.com/) installed and configured in your terminal.

### Arch Linux (AUR)

Source package:

```bash
yay -S nls
```

Prebuilt binary:

```bash
yay -S nls-bin
```

### Homebrew

```bash
brew install nolight132/tap/nls
```

### Go

```bash
go install github.com/nolight132/nls/cmd/nls@latest
```

### Build from source

```bash
git clone https://github.com/nolight132/nls.git
cd nls
go build -o nls ./cmd/nls
```

Linux/macOS:

```bash
sudo mv nls /usr/local/bin/
```

Windows PowerShell:

```powershell
go build -o nls.exe ./cmd/nls
```

---

## Usage

```bash
nls
nls ~/Downloads
nls -la
nls -lah
nls -R
nls --json
```

Pipe-friendly:

```bash
nls | wc -l
nls -1 ~/bin | xargs -I{} echo {}
nls --json | jq .
```

Multiple paths are accepted. Each directory is rendered as a separate section:

```bash
nls src tests README.md
```

---

## Configuration

`nls` reads an optional TOML config file from the OS-specific config directory:

```
$XDG_CONFIG_HOME/nls/config.toml        (Linux/macOS, XDG set)
~/.config/nls/config.toml               (Linux/macOS, XDG unset)
%APPDATA%\nls\config.toml               (Windows)
```

See [`examples/config.default.toml`](https://github.com/nolight132/nls/blob/main/examples/config.default.toml) for a commented template.
Precedence, highest to lowest: command-line flags, config file, built-in defaults.

Available settings:

| Setting                  | Default                          | Description                                                                    |
| ------------------------ | -------------------------------- | ------------------------------------------------------------------------------ |
| `default_columns`        | `id`, `name`, `size`, `modified` | Table columns and their order                                                  |
| `icons.enabled`          | `true`                           | Enable Nerd Font icons in table output                                         |
| `icons.special_icons`    | `true`                           | Use filename- and extension-specific icons                                     |
| `dir_size.enabled`       | `true`                           | Estimate directory sizes in interactive table output                           |
| `dir_size.default_depth` | `0`                              | Maximum estimation depth; `0` means unlimited depth within the selected budget |
| `dir_size.timing`        | `balanced`                       | Estimation time budget: `strict` (8ms), `balanced` (20ms), `relaxed` (100ms), or `unlimited` |
| `git.color_entries`      | `true`                           | Color names according to Git state                                             |
| `render.expand_symlinks` | `false`                          | Show symlink targets in table output; `-l` always shows them                   |

Valid column names are `id`, `name`, `type`, `size`, `modified`, `accessed`,
`changed`, `permissions`, `links`, `owner`, `group`, `inode`, `blocks`, and
`git`. The `-i`, `-s`, `-l`, and `-g` flags append their corresponding columns
when they are not already configured.

Unknown settings, invalid column names, and malformed TOML produce a warning
and cause the complete built-in configuration to be used.

### Directory sizes

Directory sizes are estimated by default only for interactive table output.
The timing preset sets one wall-clock budget shared across all directories of
a listing — each directory claims a fair share of the time left, and unused
time flows back to the pool — and the depth bound caps how deep each walk
goes. A `>` prefix marks a size that is only a lower bound because estimation
stopped early. Plain and JSON output use the directory's filesystem-reported
size unless `-P`/`--precise` is supplied.

`--precise` recursively computes directory sizes without the configured depth
and time limits. It can take a while on large directory trees.

---

## Output behavior

| Context                   | Behavior                              |
| ------------------------- | ------------------------------------- |
| Interactive terminal      | Pretty table output                   |
| Pipe / redirect / non-TTY | Fast plain output, one entry per line |
| `--json`                  | Structured JSON                       |
| `--plain`                 | Force plain output                    |
| `--table`                 | Force a bordered table                |

`-1` and `-m` select plain output even in a terminal. When `--table` is used
with piped or redirected output, colors remain disabled. `--no-color` disables
colors in every output context.

### JSON fields

`--json` writes one array containing all listed entries. Each object contains
`name`, `path`, `type`, `size`, and `permissions`. It also contains `modified`
when available, `link_target` for symlinks, `size_human` with `-h`, and
`git_state` when Git status was computed. Timestamps use RFC 3339, sizes are
bytes, and `type` is one of `dir`, `link`, `exec`, or `file`.

---

## Flags

Common flags:

| Flag                        | Description                                            |
| --------------------------- | ------------------------------------------------------ |
| `-a`, `--all`               | Show entries starting with `.`                         |
| `-A`, `--almost-all`        | Show hidden entries except `.` and `..`                |
| `-l`, `--long`              | Show extended metadata                                 |
| `-h`, `--human-readable`    | Print human-readable sizes                             |
| `-1`, `--one`               | List one entry per line                                |
| `-m`, `--comma`             | Print a comma-separated list                           |
| `-R`, `--recursive`         | List subdirectories recursively                        |
| `-d`, `--directory`         | List directories themselves, not their contents        |
| `-r`, `--reverse`           | Reverse the sort order                                 |
| `-t`, `--time`              | Sort by modification time                              |
| `-u`, `--access-time`       | Sort by access time                                    |
| `-c`, `--ctime`             | Sort by status-change time                             |
| `-S`, `--size`              | Sort by size                                           |
| `-X`, `--extension`         | Sort alphabetically by extension                       |
| `-U`, `--unsorted`          | Do not sort                                            |
| `-f`, `--fast`              | Do not sort and show all entries (equivalent to `-aU`) |
| `--group-directories-first` | Group directories before files                         |
| `-F`, `--classify`          | Append a file type indicator                           |
| `-p`, `--slash`             | Append `/` to directories                              |
| `-Q`, `--quote-name`        | Enclose entry names in double quotes                   |
| `-B`, `--ignore-backups`    | Hide entries ending with `~`                           |
| `-L`, `--dereference`       | Follow symlinks                                        |
| `-i`, `--inode`             | Show inode numbers                                     |
| `-s`, `--size-blocks`       | Show allocated block counts                            |

`nls` specific:

| Flag                 | Description                   |
| -------------------- | ----------------------------- |
| `-g`, `--git-status` | Show per-entry Git status     |
| `-P`, `--precise`    | Compute exact directory sizes |
| `--json`             | Output JSON                   |
| `--plain`            | Force plain-text output       |
| `--table`            | Force table output            |
| `--no-icons`         | Disable icons                 |
| `--no-color`         | Disable colors                |
| `--version`          | Print the version             |
| `--help`             | Print command help            |

---

## Icons

Icons are enabled by default in table output and require a Nerd Font. They are
not added to plain or JSON output.

If you want, you can disable them with:

```toml
[icons]
enabled = false
```

---

## Colors

`nls` uses terminal-friendly ANSI colors and reads `LS_COLORS` when available.
Colors are limited to interactive terminals; `--no-color` or a non-empty
`NO_COLOR` environment variable disables them everywhere.

Default highlights:

- directories
- symlinks
- executables
- sizes
- modified timestamps

Git-aware name colors are enabled by default: modified entries are yellow,
untracked entries are bright green, and ignored entries are gray. They can be
disabled independently with `git.color_entries = false`.

---

## Development

```bash
go fmt ./...
go test ./...
go run ./cmd/nls
```

One-time setup — enables the pre-commit hook that gofmt-formats staged files:

```bash
git config core.hooksPath .githooks
```

Build:

```bash
go build -o nls ./cmd/nls
```

---

## Project's future

Possible future tools:

- `nfind`
- `ndu`
- `nps`
- `nstat`

The goal is a small suite of modern coreutils-style tools with beautiful interactive output and sane script behavior.

---

## Special Thanks

### Nushell

`nls` would not exist without Nushell.

The default table layout, metadata presentation, relative timestamps, and much of the overall user experience are directly inspired by Nushell's `ls`.

The goal of this project is not to reinvent that interface, but to bring a similar experience to people who prefer traditional shells and environments such as bash, zsh, fish, PowerShell, and standard terminals on Linux, macOS, and Windows.

If you like `nls`, you probably like Nushell.

https://www.nushell.sh

### bat

`bat` inspired the core philosophy behind this project.

One of the ideas that made `nls` possible was seeing how `bat` provides a significantly better interactive experience while still remaining useful in pipes, scripts, and other non-interactive environments.

https://github.com/sharkdp/bat

---

## License

MIT
