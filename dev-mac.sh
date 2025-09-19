#!/bin/bash

# This script require air to be installed
# If you have not installed air, please visit https://github.com/air-verse/air
# Or use `go install github.com/air-verse/air@latest` to install it <3

set -e  # Exit on any error

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "Error: air is not installed"
    echo "Please install air using: go install github.com/air-verse/air@latest"
    exit 1
fi

# Load environment variables from .env file
# macOS-optimized method that handles quotes and special characters
while IFS= read -r line; do
    # Skip empty lines and comments
    if [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]]; then
        continue
    fi
    
    # Export the variable (handle both KEY=value and KEY="value" formats)
    if [[ "$line" =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=[[:space:]]*(.*)$ ]]; then
        key="${BASH_REMATCH[1]}"
        value="${BASH_REMATCH[2]}"
        
        # Remove surrounding quotes if present
        if [[ "$value" =~ ^\"(.*)\"$ ]] || [[ "$value" =~ ^\'(.*)\'$ ]]; then
            value="${BASH_REMATCH[1]}"
        fi
        
        export "$key=$value"
    fi
done < .env

echo "Starting development server with air on macOS..."

# Run air with configuration
air -c .air.toml