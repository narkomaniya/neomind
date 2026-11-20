$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root 'dist'
New-Item -ItemType Directory -Path $dist -Force | Out-Null

# Rust normally uses the MSVC toolchain on Windows, which needs link.exe from
# Visual Studio. When it is missing, fall back to the GNU toolchain and make
# its MinGW linker visible. CI runners have MSVC, so they keep the default.
$vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio\Installer\vswhere.exe'
$msvcReady = (Get-Command link.exe -ErrorAction SilentlyContinue) -or (Test-Path $vswhere)
if (-not $msvcReady) {
    Write-Host '==> MSVC linker not found: using stable-x86_64-pc-windows-gnu'

    $rustup = Join-Path $env:USERPROFILE '.cargo\bin\rustup.exe'
    if (-not (Test-Path $rustup)) {
        throw 'rustup.exe not found — install Rust first (winget install Rustlang.Rustup)'
    }

    $installed = & $rustup toolchain list 2>&1 | Out-String
    if ($installed -notmatch 'windows-gnu') {
        & $rustup toolchain install stable-x86_64-pc-windows-gnu --profile minimal
        if ($LASTEXITCODE -ne 0) { throw 'cannot install the GNU Rust toolchain' }
    }
    $env:RUSTUP_TOOLCHAIN = 'stable-x86_64-pc-windows-gnu'

    $wingetRoot = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages'
    $gcc = Get-ChildItem $wingetRoot -Recurse -Filter gcc.exe -ErrorAction SilentlyContinue |
           Where-Object { $_.FullName -like '*mingw64*' } | Select-Object -First 1
    if (-not $gcc) {
        throw 'MinGW gcc.exe not found — install WinLibs/MinGW-w64 (winget install BrechtSanders.WinLibs.POSIX.UCRT)'
    }
    $env:PATH = "$(Split-Path $gcc.FullName -Parent);$env:PATH"
}

Push-Location $root
try {
    Write-Host '==> building Rust core'
    & cargo build --release --manifest-path (Join-Path $root 'rust\Cargo.toml')
    if ($LASTEXITCODE -ne 0) { throw 'cargo build failed' }

    Write-Host '==> building Go CLI'
    & go build -trimpath -o (Join-Path $dist 'neomind.exe') .\cmd\neomind
    if ($LASTEXITCODE -ne 0) { throw 'go build failed' }

    Copy-Item -Force (Join-Path $root 'rust\target\release\neomind-core.exe') (Join-Path $dist 'neomind-core.exe')
} finally {
    Pop-Location
}

Write-Host "NeoMind is ready: $dist\neomind.exe"
Write-Host 'Add that folder to PATH or run: .\dist\neomind.exe help'
