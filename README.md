# NeoMind

A small, fast note keeper built as **Rust core + Go CLI + native desktop GUI**.

Version **0.4.0** ships three binaries: a Rust core with all storage logic, a Go
CLI for scripts and diagnostics, and a Go/WebView desktop window with a proper
interface, a language/toolchain installer and a curated mods catalog. The CLI
and GUI never link against Rust directly — they call `neomind-core <op> [args...]`,
which answers with one JSON line. This keeps the build portable.

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

Strict monochrome workbench in the spirit of VS Code — no gradients, no accent
colours, just greys and 1px borders.

- explorer sidebar with filter, live preview and sort (date / title)
- resizable sidebar (drag the divider), width is remembered
- tab bar, breadcrumbs, line numbers and a status bar (core version, note count, Ln/Col, save state)
- multiline editor with **auto-save** (0.9 s) and Ctrl+S
- **Ctrl+P** quick-open palette, **Ctrl+F** filter, **Ctrl+N** new note, **Ctrl+D** delete
- arrow-key navigation in the list, Enter to focus the editor
- in-app confirmation dialog instead of the native OS prompt
- **Languages** view (second activity-bar icon): a catalogue of Go, Rust, Python,
  Node.js, Java, C/C++ (MinGW), .NET, Ruby, PHP and Perl with installed status,
  multi-select and **install-all / install-selected** in one background job
  (winget on Windows) with live progress and a log
- **Mods & themes** view (third activity-bar icon): a curated cosmetic catalog
  and local UI tweaks — see below

Notes live in `~/.neomind/notes.txt` (`%USERPROFILE%\.neomind\notes.txt`) as
`id<TAB>escaped text` lines — easy to back up. `NEOMIND_DATA_DIR` redirects the
storage location.

## Mods & themes

The third activity-bar icon opens a marketplace-style catalog, but with one
important rule: **NeoMind never installs third-party code**. Nothing from the
catalog is downloaded, executed, or linked into the app, so a mod can never
break your notes or the Rust core. Two groups live here:

- **Curated VS Code mods** — 17 cosmetic themes and icon themes (Dracula, One
  Dark Pro, GitHub Theme, Material Theme, Nord, Tokyo Night, Catppuccin and
  built-in themes such as Quiet Light, Solarized Dark, Monokai). Each card shows
  the author's GitHub avatar (desaturated to match the monochrome UI), a SAFE
  badge, the license, a **Source** link to the open repository and a
  **Marketplace** link to the VS Code listing. All links are allow-listed to
  GitHub / the Visual Studio Marketplace and open in your browser — `mods.Open`
  rejects every other host and any non-HTTPS URL.
- **NeoMind tweaks** — safe local CSS classes, stored in `localStorage` under
  `nm_mods`: compact UI, no line numbers, wide sidebar, monospace UI, high
  contrast, hide status bar and focus mode (with an **Exit focus mode** button).
  They only change how the existing interface looks; they add no external code.

The catalog lives in `internal/mods` with tests that reject duplicate IDs,
missing links and unsafe entries. Only cosmetic themes/icon themes make it in —
no arbitrary extensions, no code execution, no storage or core changes.

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
