# NeoMind

A tiny, fast note keeper built as **Rust core + Go CLI**.

Version **0.1.0** is intentionally small and convenient: build once, get two
binaries in `dist/`, and the CLI finds the core automatically. The product
behaviour is a placeholder for what comes next — say the word and we shape it.

## Layout

```text
rust/                 # Rust library + thin JSON command binary (neomind-core)
cmd/neomind/          # Go CLI (user-facing commands)
internal/core/        # Go client: finds and calls the Rust core
scripts/build.*       # one-command build for Windows / POSIX
.github/workflows/ci.yml
```

The CLI never links against Rust. It calls `neomind-core <op> [args...]`, which
answers with one JSON line. This keeps the build portable on all three OSes.

## Build

```powershell
# Windows
powershell -ExecutionPolicy Bypass -File scripts\build.ps1
```

```sh
# Linux / macOS
sh scripts/build.sh
```

Output:

```text
dist/neomind          # or neomind.exe on Windows
dist/neomind-core     # or neomind-core.exe
```

Keep them together. Override discovery with `NEOMIND_CORE=/path/to/neomind-core`.

## Use

```text
neomind                       show help
neomind status [--json]       core version, note count, data dir
neomind note add <text>       add a note
neomind note list [--json]    list notes
neomind note clear [--yes]    clear all notes
neomind doctor                check installation
neomind version               CLI + core versions
```

Notes live in `~/.neomind/notes.txt` (`%USERPROFILE%\.neomind\notes.txt` on
Windows) as `timestamp<TAB>text` lines — easy to inspect or back up.

`NO_COLOR=1` disables colours; `--json` gives machine-readable output.

## Roadmap

Everything after v0.1 is waiting for your spec. The current split (Rust logic,
Go UX) is meant to survive a much bigger product.
