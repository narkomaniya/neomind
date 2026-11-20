param([switch]$Run)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root 'dist'
$dest = Join-Path $env:LOCALAPPDATA 'NeoMind'
$files = @('neomind.exe', 'neomind-gui.exe', 'neomind-core.exe')
if (Test-Path (Join-Path $root 'assets\neomind.ico')) {
    Copy-Item -Force (Join-Path $root 'assets\neomind.ico') (Join-Path $src 'neomind.ico')
    $files += 'neomind.ico'
}

foreach ($f in $files) {
    if (-not (Test-Path (Join-Path $src $f))) {
        throw "$f is missing - run scripts\build.ps1 first"
    }
}

New-Item -ItemType Directory -Path $dest -Force | Out-Null
foreach ($f in $files) {
    Copy-Item -Force (Join-Path $src $f) (Join-Path $dest $f)
}

$ws = New-Object -ComObject WScript.Shell
$startMenu = [Environment]::GetFolderPath('Programs')
New-Item -ItemType Directory -Path (Join-Path $startMenu 'NeoMind') -Force | Out-Null
$places = @((Join-Path $startMenu 'NeoMind'), [Environment]::GetFolderPath('Desktop'))
foreach ($where in $places) {
    if (-not $where -or -not (Test-Path $where)) { continue }
    $sc = $ws.CreateShortcut((Join-Path $where 'NeoMind.lnk'))
    $sc.TargetPath = Join-Path $dest 'neomind-gui.exe'
    $sc.WorkingDirectory = $dest
    $sc.Description = 'NeoMind - quick notes'
    if ($files -contains 'neomind.ico') {
        $sc.IconLocation = (Join-Path $dest 'neomind.ico') + ',0'
    }
    $sc.Save()
}

# Force Explorer to drop cached icons for the shortcut/exe paths.
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class ShellIcon {
    [DllImport("shell32.dll")]
    public static extern void SHChangeNotify(int eventId, uint flags, IntPtr item1, IntPtr item2);
}
'@
[ShellIcon]::SHChangeNotify(0x08000000, 0x0000, [IntPtr]::Zero, [IntPtr]::Zero)

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$dest*") {
    [Environment]::SetEnvironmentVariable('Path', $userPath.TrimEnd(';') + ';' + $dest, 'User')
}

Write-Host "NeoMind installed: $dest\neomind-gui.exe"
Write-Host 'Shortcuts: Start Menu -> NeoMind, and Desktop.'
Write-Host 'CLI available in new terminals: neomind help'
if ($Run) { Start-Process -FilePath (Join-Path $dest 'neomind-gui.exe') }
