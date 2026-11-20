param(
    [string]$Icon = (Join-Path $PSScriptRoot '..\assets\neomind.ico'),
    [string]$OutDir = (Join-Path $PSScriptRoot 'installer')
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

function New-WizardImage {
    param(
        [int]$Width,
        [int]$Height,
        [int]$IconSize,
        [int]$TextSize,
        [string]$Path
    )

    $bmp = New-Object System.Drawing.Bitmap($Width, $Height, [System.Drawing.Imaging.PixelFormat]::Format24bppRgb)
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    try {
        $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
        $g.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::ClearTypeGridFit
        $g.Clear([System.Drawing.Color]::White)

        $icon = New-Object System.Drawing.Icon($Icon, $IconSize, $IconSize)
        try {
            $x = [int](($Width - $IconSize) / 2)
            $y = [int](($Height - $IconSize) / 2)
            if ($TextSize -gt 0) { $y = [Math]::Max(24, [int](($Height - $IconSize) / 2) - 26) }
            $g.DrawIcon($icon, $x, $y)
        } finally {
            $icon.Dispose()
        }

        if ($TextSize -gt 0) {
            $font = New-Object System.Drawing.Font('Segoe UI Semibold', $TextSize, [System.Drawing.FontStyle]::Regular, [System.Drawing.GraphicsUnit]::Pixel)
            $brush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::FromArgb(31, 35, 40))
            $format = New-Object System.Drawing.StringFormat
            $format.Alignment = [System.Drawing.StringAlignment]::Center
            $rect = New-Object System.Drawing.RectangleF(0, ($y + $IconSize + 18), $Width, ($TextSize + 18))
            try {
                $g.DrawString('NeoMind', $font, $brush, $rect, $format)
            } finally {
                $font.Dispose(); $brush.Dispose(); $format.Dispose()
            }
        }

        $bmp.Save($Path, [System.Drawing.Imaging.ImageFormat]::Bmp)
    } finally {
        $g.Dispose(); $bmp.Dispose()
    }
}

New-WizardImage -Width 164 -Height 314 -IconSize 96 -TextSize 20 -Path (Join-Path $OutDir 'wizard.bmp')
New-WizardImage -Width 55 -Height 55 -IconSize 40 -TextSize 0 -Path (Join-Path $OutDir 'wizard-small.bmp')

Get-Item (Join-Path $OutDir 'wizard.bmp'), (Join-Path $OutDir 'wizard-small.bmp') |
    Select-Object Name, Length
