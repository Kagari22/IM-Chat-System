param(
    [switch]$WithInfra
)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

. (Join-Path $PSScriptRoot "Set-DevEnv.ps1")

if ($WithInfra) {
    docker compose up -d
}

go run .\cmd\server
