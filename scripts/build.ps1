$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root 'dist'
New-Item -ItemType Directory -Path $dist -Force | Out-Null

# --- C/C++ toolchain for the native WebView GUI (cgo) ---
$gccCmd = Get-Command gcc.exe -ErrorAction SilentlyContinue
if ($gccCmd) {
    $gcc = $gccCmd.Source
} else {
    $wingetRoot = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages'
    $found = Get-ChildItem $wingetRoot -Recurse -Filter gcc.exe -ErrorAction SilentlyContinue |
             Where-Object { $_.FullName -like '*mingw64*' } | Select-Object -First 1
    $gcc = if ($found) { $found.FullName } else { $null }
}
if (-not $gcc) { throw 'MinGW gcc.exe not found — it is required to build the desktop GUI.' }

$binDir = Split-Path $gcc -Parent
$env:PATH = "$binDir;$env:PATH"
$env:CGO_ENABLED = '1'
$env:CC = $gcc
$gpp = Join-Path $binDir 'g++.exe'
if (Test-Path $gpp) { $env:CXX = $gpp }

# --- Rust linker: use GNU toolchain when Visual Studio is absent ---
$vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio\Installer\vswhere.exe'
$msvcReady = (Get-Command link.exe -ErrorAction SilentlyContinue) -or (Test-Path $vswhere)
if (-not $msvcReady) {
    Write-Host '==> MSVC linker not found: using stable-x86_64-pc-windows-gnu'
    $rustup = Join-Path $env:USERPROFILE '.cargo\bin\rustup.exe'
    if (-not (Test-Path $rustup)) { throw 'rustup.exe not found — install Rust first.' }
    if (((& $rustup toolchain list 2>&1 | Out-String) -notmatch 'windows-gnu')) {
        & $rustup toolchain install stable-x86_64-pc-windows-gnu --profile minimal
        if ($LASTEXITCODE -ne 0) { throw 'cannot install the GNU Rust toolchain' }
    }
    $env:RUSTUP_TOOLCHAIN = 'stable-x86_64-pc-windows-gnu'
}

Push-Location $root
try {
    Write-Host '==> building Rust core'
    & cargo build --release --manifest-path (Join-Path $root 'rust\Cargo.toml')
    if ($LASTEXITCODE -ne 0) { throw 'cargo build failed' }

    Write-Host '==> building Go CLI'
    & go build -trimpath -o (Join-Path $dist 'neomind.exe') .\cmd\neomind
    if ($LASTEXITCODE -ne 0) { throw 'go build failed' }

    Write-Host '==> building desktop GUI'
    & go build -trimpath -o (Join-Path $dist 'neomind-gui.exe') .\cmd\neomind-gui
    if ($LASTEXITCODE -ne 0) { throw 'GUI build failed' }

    Copy-Item -Force (Join-Path $root 'rust\target\release\neomind-core.exe') (Join-Path $dist 'neomind-core.exe')
} finally {
    Pop-Location
}

Write-Host "NeoMind is ready: $dist\neomind.exe + neomind-gui.exe"
Write-Host 'Run: .\dist\neomind.exe gui   (or double-click neomind-gui.exe)'
