#!/bin/bash

set -e

export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH

if ! command -v air &> /dev/null; then
    echo "Error: air is not installed"
    echo "Please install air using: go install github.com/air-verse/air@latest"
    exit 1
fi

source ./scripts/env.sh

echo "Starting development server with air..."

air -c .air.toml