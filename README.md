<div align="center">

# nls

### neo-ls: a modern `ls` with useful tables

A fast, cross-platform file listing tool that feels familiar in scripts
and looks beautiful in your terminal.

<br>

`nls` is heavily inspired by Nushell's `ls`.

The original goal was simple: bring Nushell's excellent table-based file listings
to traditional shells and environments without requiring users to switch to Nushell itself.

</div>

<div align="center">
<img width="820" height="563" alt="image" src="https://github.com/user-attachments/assets/bf3c69f0-7263-46d3-b0cb-ad319d029a84" />
</div>

---

## Why nls?

`nls` is not trying to be a full shell, and it is not just another colorful `ls`.

It is a **neo-ls**: a modern file listing command designed around three ideas:

- beautiful tables when you are looking at files interactively
- practical compatibility when used in pipes and scripts
- useful defaults, like showing directory sizes without making normal usage slow

Nushell already provides an excellent file listing experience, and much of `nls` is inspired by it.

However, not everyone wants a new shell.

Many users are perfectly happy with bash, zsh, fish, PowerShell, or existing terminal workflows. They want the table layouts, metadata presentation, and overall polish of `nu ls`, but they do not necessarily want Nushell's programming model, pipeline semantics, or compatibility tradeoffs.

For those users, switching shells can sometimes reduce usability rather than improve it, especially when working with existing shell scripts, documentation, and POSIX-oriented tooling.

`nls` exists for people who want the visual experience of modern terminal tools while keeping the workflows they already know.

It works in bash, zsh, fish, Nushell, PowerShell, Windows Terminal, Linux terminals, macOS terminals, and other environments where a normal CLI binary can run.

---

## Features

- Nushell-style tables for interactive terminal use
- Directory sizes shown by default
- Fast non-TTY behavior for pipes, redirects, and scripts
- GNU `ls`-like behavior for common workflows
- Helpful suggestions when flags or arguments are mistyped
- Optional icons with `NLS_ICONS=1`
- Colors for files, directories, symlinks, executables, sizes, and timestamps
- JSON output for structured usage
- Works on Linux, macOS, and Windows

---

## Install

### Arch Linux (AUR)

Source package:

```bash
yay -S nls
```

Prebuilt binary:

```bash
yay -S nls-bin
```

Both packages are maintained by the author.

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

Enable icons:

```bash
NLS_ICONS=1 nls
NLS_ICONS=nerd nls
```

---

## Configuration

`nls` reads an optional TOML config file from the OS-specific config directory:

```
$XDG_CONFIG_HOME/nls/config.toml        (Linux/macOS, XDG set)
~/.config/nls/config.toml               (Linux/macOS, XDG unset)
%APPDATA%\nls\config.toml               (Windows)
```

A missing file is fine; built-in defaults apply. See `config.example.toml`
at the repo root for a commented template. The initial schema covers:

```toml
icons = false

[dir_size]
default_depth = 0      # max walk depth when --estimate-depth is not passed (0 = unlimited)
timing = "balanced"    # strict | balanced | relaxed

default_columns = [    # columns shown in interactive table mode, in order
    "id", "name", "type", "size", "modified",
]
```

`timing` picks a named budget preset for bounded directory size estimation on
a TTY. The concrete millisecond/entry caps are derived from the name and are
not exposed in the file:

| Preset     | Trade-off                                     |
| ---------- | --------------------------------------------- |
| `strict`   | Aggressive caps, stays snappy on huge trees   |
| `balanced` | Default, enough headroom for typical repos    |
| `relaxed`  | Generous, prefers accurate sizes over latency |

`default_columns` controls which columns appear in interactive table mode and
their order. Omit a column to hide it; list one to enable it by default.
Available columns: `id`, `name`, `type`, `size`, `modified`, `accessed`,
`changed`, `permissions`, `links`, `owner`, `group`, `inode`, `blocks`.
Flags `-i`, `-s`, `-l` still append their columns if not already listed.

Precedence, highest to lowest: command-line flags, environment variables
(`NLS_ICONS`, `NERD_FONT`, ...), config file, built-in defaults.

---

## Output behavior

| Context                   | Behavior                                      |
| ------------------------- | --------------------------------------------- |
| Interactive terminal      | Pretty table output                           |
| Pipe / redirect / non-TTY | Fast plain output, close to GNU `ls` behavior |
| `--json`                  | Structured JSON                               |
| Common `ls` flags         | GNU-like formatting where supported           |

`nls` aims to behave like GNU `ls` in normal pipe/non-TTY usage, but it does not cover every historical edge case yet.

---

## Flags

Common flags:

| Flag                     | Description                             |
| ------------------------ | --------------------------------------- |
| `-a`, `--all`            | Show hidden entries                     |
| `-A`, `--almost-all`     | Show hidden entries except `.` and `..` |
| `-l`, `--long`           | Long listing format                     |
| `-h`, `--human-readable` | Human-readable sizes                    |
| `-1`, `--one`            | One entry per line                      |
| `-R`, `--recursive`      | Recursive listing                       |
| `-r`, `--reverse`        | Reverse sort                            |
| `-t`, `--time`           | Sort by modified time                   |
| `-S`, `--size`           | Sort by size                            |
| `-X`, `--extension`      | Sort by extension                       |
| `-U`, `--unsorted`       | Do not sort                             |
| `-F`, `--classify`       | Append file type indicators             |
| `-p`, `--slash`          | Append `/` to directories               |
| `-L`, `--dereference`    | Follow symlinks                         |

`nls` specific:

| Flag         | Description    |
| ------------ | -------------- |
| `--json`     | Output JSON    |
| `--no-icons` | Disable icons  |
| `--no-color` | Disable colors |

---

## Icons

Icons are disabled by default.

Enable them with:

```bash
NLS_ICONS=1 nls
```

Future versions will include broader file-type icon coverage.

---

## Colors

`nls` uses terminal-friendly ANSI colors and reads `LS_COLORS` when available.

Default highlights:

- directories
- symlinks
- executables
- sizes
- modified timestamps

---

## Roadmap

Planned features:

- ~~XDG config file support~~
- custom color overrides
- custom themes
- alternative layouts
- more table styles
- wider file-type icon support
- better Windows-specific polish
- more GNU `ls` compatibility where it makes sense
- optional Git status column
- optional tree layout
- more structured output modes

Possible future tools:

- `nfind`
- `ndu`
- `nps`
- `nstat`

The goal is a small suite of modern coreutils-style tools with beautiful interactive output and sane script behavior.

---

## Development

```bash
go fmt ./...
go test ./...
go run ./cmd/nls
```

Build:

```bash
go build -o nls ./cmd/nls
```

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
