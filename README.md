# NeoMind

A small, fast note keeper built as **Rust core + Go CLI + native desktop GUI**.

Version **0.2.0** ships three binaries: a Rust core with all storage logic, a Go
CLI for scripts and diagnostics, and a Go/WebView desktop window with a proper
interface. The CLI and GUI never link against Rust directly — they call
`neomind-core <op> [args...]`, which answers with one JSON line. This keeps the
build portable.

## What you get

| Binary | Purpose |
| --- | --- |
| `neomind-core` | Rust storage + JSON protocol (notes, search-ready CRUD) |
| `neomind` | Go CLI: `status`, `note add/list/clear`, `doctor`, `version`, `gui` |
| `neomind-gui` | Native desktop window (WebView2 on Windows, WebKit on macOS, GTK on Linux) |

## Install on a Windows laptop

```powershell
# one-time: build, copy to %LOCALAPPDATA%\NeoMind, create Start Menu + Desktop shortcuts
powershell -ExecutionPolicy Bypass -File scripts\install.ps1 -Run
```

After that just press **Win** and type **NeoMind**, or double-click the desktop
shortcut. The CLI is added to your user PATH, so a new terminal can run `neomind help`.

To rebuild and reinstall:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\build.ps1
powershell -ExecutionPolicy Bypass -File scripts\install.ps1
```

## Build manually

```powershell
# Windows (needs MinGW gcc/g++ for the WebView GUI)
powershell -ExecutionPolicy Bypass -File scripts\build.ps1
```

```sh
# Linux / macOS (GUI built only when GTK + WebKit headers are installed)
sh scripts/build.sh
```

Output in `dist/`. Keep the binaries together: the CLI/GUI find the core next to
themselves. Override with `NEOMIND_CORE=/path/to/neomind-core`.

## GUI

- note list with search and live preview
- multiline editor with **auto-save** (1.2 s) and Ctrl+S
- create (Ctrl+N), delete, open storage folder
- empty state, toasts, keyboard hints

Notes live in `~/.neomind/notes.txt` (`%USERPROFILE%\.neomind\notes.txt`) as
`id<TAB>escaped text` lines — easy to back up. `NEOMIND_DATA_DIR` redirects the
storage location.

## CLI

```text
neomind                       show help
neomind gui                   open the desktop app
neomind status [--json]       core version, note count, data dir
neomind note add <text>       add a note
neomind note list [--json]    list notes
neomind note clear [--yes]    clear all notes
neomind doctor                check installation
neomind version               CLI + core versions
```

`NO_COLOR=1` disables colours; `--json` gives machine-readable output.

## Roadmap

The product behaviour beyond notes is waiting for your spec. The split (Rust
logic, Go UX) is meant to survive a much bigger application.
