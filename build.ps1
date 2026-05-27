param(
    [switch]$Release,
    [switch]$NoLogo
)

if (-not $NoLogo) {
    Write-Host "=== Generating icons ===" -ForegroundColor Cyan
    go run ./logo/generate.go
    if ($LASTEXITCODE -ne 0) { exit 1 }

    $rsrc = (Get-Command rsrc.exe -ErrorAction SilentlyContinue).Source
    if (-not $rsrc) { $rsrc = "$env:USERPROFILE\go\bin\rsrc.exe" }
    if (Test-Path $rsrc) {
        Write-Host "=== Generating .syso ===" -ForegroundColor Cyan
        & $rsrc -ico logo/app.ico -o app.syso
        if ($LASTEXITCODE -ne 0) { exit 1 }
    } else {
        Write-Host "rsrc not found - EXE will not have embedded icon. Install with: go install github.com/akavel/rsrc@latest" -ForegroundColor Yellow
    }
}

Write-Host "=== Building ytdlp-desktop.exe ===" -ForegroundColor Cyan
$env:CC = "zig cc -target x86_64-windows-gnu"
$ldflags = "-s -w -H windowsgui"
if ($Release) {
    $ldflags = "-s -w -H windowsgui -buildmode=pie"
}

go build -ldflags="$ldflags" -o ytdlp-desktop.exe .
if ($LASTEXITCODE -ne 0) { exit 1 }

$file = Get-Item .\ytdlp-desktop.exe
Write-Host "Build OK: $($file.Length / 1MB -as [int]) MB" -ForegroundColor Green
