# Setup Python virtual environment for Windows (PowerShell)
Write-Host "Setting up Python virtual environment..." -ForegroundColor Green

Set-Location python

# Check if Python is available
if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Python is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

# Create virtual environment
Write-Host "Creating virtual environment..." -ForegroundColor Yellow
python -m venv venv

if (-not $?) {
    Write-Host "Error: Failed to create virtual environment" -ForegroundColor Red
    exit 1
}

# Install dependencies
Write-Host "Installing dependencies..." -ForegroundColor Yellow
.\venv\Scripts\pip install -r requirements.txt

if (-not $?) {
    Write-Host "Error: Failed to install dependencies" -ForegroundColor Red
    exit 1
}

Write-Host "Virtual environment setup complete!" -ForegroundColor Green
Write-Host "To activate manually: python\venv\Scripts\activate" -ForegroundColor Cyan

Set-Location ..