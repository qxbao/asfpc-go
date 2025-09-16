# This script require air to be installed
# If you have not installed air, please visit https://github.com/air-verse/air
# Or use `go install github.com/air-verse/air@latest` to install it <3

Get-Content .env | foreach {
    $name, $value = $_.split('=')
    Set-Content env:\$name $value
}
air -c .air.toml