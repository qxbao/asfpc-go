@echo off
echo Setting up Python virtual environment...

cd python

REM Create virtual environment
python -m venv venv

REM Install dependencies
venv\Scripts\pip install -r requirements.txt

echo Virtual environment setup complete!
echo To activate manually: python\venv\Scripts\activate

cd ..
