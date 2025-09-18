#!/bin/bash

# Setup Python virtual environment for macOS and Linux
echo "Setting up Python virtual environment..."

cd python

# Create virtual environment
python3 -m venv venv

# Activate virtual environment and install dependencies
source venv/bin/activate
pip install -r requirements.txt

echo "Virtual environment setup complete!"
echo "To activate manually: source python/venv/bin/activate"

cd ..
