# NeoMind
![Uploading image.png…]()

A small, fast note keeper built as **Rust core + Go CLI + native desktop GUI**.

Version **0.5.0** ships three binaries: a Rust core with all storage logic, a Go
CLI for scripts and diagnostics, and a Go/WebView desktop window with a proper
interface, a language/toolchain installer and a VS Code-style Extensions catalog.
The CLI and GUI never link against Rust directly — they call
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
- **Extensions** view (third activity-bar icon): a VS Code-style catalog of the
  100 most installed extensions, live Marketplace search, and local UI tweaks —
  see below

Notes live in `~/.neomind/notes.txt` (`%USERPROFILE%\.neomind\notes.txt`) as
`id<TAB>escaped text` lines — easy to back up. `NEOMIND_DATA_DIR` redirects the
storage location.

## Mods & themes

The third activity-bar icon opens an **Extensions** view built like VS Code's:
a searchable list on the left and a detail pane on the right with an icon,
publisher, install count, rating, an identifier/key-value block and links. One
rule still applies: **NeoMind never installs third-party code**. Nothing is
downloaded, executed or linked into the app, so an extension can never break
your notes or the Rust core.

- **Top 100** — the 100 most installed VS Code extensions, generated from the
  Marketplace by `scripts/gen-top100.ps1` into `internal/mods/top_data.go`.
  Cards show desaturated author icons, install counts and ratings. Sorting by
  installs/rating/name works locally.
- **Live search** — typing in the search box queries the VS Code Marketplace
  through `nm_mods_search` (up to 50 results). Search is read-only; no VSIX is
  downloaded. The bundled top-100 is used when the box is empty or offline.
- **Cosmetic** filter — switches to the original 17 safe cosmetic themes/icon
  themes with a SAFE badge and license info.
- **NeoMind tweaks** tab — safe local CSS classes, stored in `localStorage`
  under `nm_mods`: compact UI, no line numbers, wide sidebar, monospace UI,
  high contrast, hide status bar and focus mode (with an **Exit focus mode**
  button). They only restyle the interface; they add no external code.
- Extensions that execute code in VS Code get an `EXECUTES CODE` marker —
  informative only, since nothing is installed here.

All external links go through `mods.Open`, which allow-lists GitHub and the
Visual Studio Marketplace over HTTPS and rejects every other host. The catalog
lives in `internal/mods` with tests that reject duplicate IDs, missing links and
links outside the allowlist. Regenerate it with:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\gen-top100.ps1
```

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
