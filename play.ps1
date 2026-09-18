# Launch the desktop game. From this folder in PowerShell:
#   .\play.ps1
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$go = "C:\Program Files\Go\bin\go.exe"
if (Get-Command go -ErrorAction SilentlyContinue) {
    $go = (Get-Command go).Source
}
$exe = Join-Path $PSScriptRoot "go-with-go.exe"

if (-not (Test-Path $exe)) {
    if (-not (Test-Path $go)) {
        Write-Error "Go is not installed. Expected $go"
    }
    & $go build -o $exe ./cmd/go-with-go
}

Write-Host "Starting Go with Go..."
$p = Start-Process -FilePath $exe -WorkingDirectory $PSScriptRoot -PassThru
if (-not $p) {
    Write-Error "failed to start $exe"
}
Write-Host "Window should stay open (pid $($p.Id)). Close it from the game, or press Esc then close."
Write-Host "If it vanishes, read: $(Join-Path $PSScriptRoot 'go-with-go.log')"
