If (-not (Test-Path .env)) {
  Write-Host "Error: .env file not found"
  exit 1
}

Get-Content .env | foreach {
  If (($_ -match '^\s*#') -or ([string]::IsNullOrWhiteSpace($_))) {
    continue
  }
  $name, $value = $_.split('=', 2)
  Set-Item env:\$name $value
}

Set-Item env:\GOOSE_DRIVER "postgres"
Set-Item env:\GOOSE_DBSTRING "host=$env:POSTGRE_HOST port=$env:POSTGRE_PORT user=$env:POSTGRE_USER password=$env:POSTGRE_PASSWORD dbname=$env:POSTGRE_DBNAME sslmode=disable"
Set-Item env:\GOOSE_MIGRATION_DIR "./db/migrations"