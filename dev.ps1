# Powershell

If (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host "Error: air is not installed"
    Write-Host "Please install air using: go install github.com/air-verse/air@latest"
    exit 1
}

./scripts/env

air -c .air.toml