# ASFPC - Go

![Build Status](https://github.com/qxbao/asfpc-go/actions/workflows/go.yml/badge.svg)


**A**utomated **S**ystem for **F**inding **P**otential **C**ustomers

This project aims to create an automated system that can efficiently identify and target potential customers for a business. By leveraging data analysis and machine learning techniques, the system will analyze customer behavior, preferences, and demographics to generate leads and improve marketing strategies.

## Table of Contents
- [Requirements](#requirements)
- [Installation](#installation)
- [Setup](#setup)
- [Usage](#usage)
- [Development](#development)
- [Docker](#docker)

## Requirements
- Go 1.23+
- Python 3.12+
- PostgreSQL
- Docker (optional, functions might be limited)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/qxbao/asfpc-go.git
```
2. Navigate to the project directory:
```powershell
cd asfpc-go
```
3. Install Go dependencies:
```bash
go mod download
```

## Setup
1. Set up PostgreSQL database and create a new database for the project. The table initialization SQL script is located at [`infras/sql/schema.sql`](infras/sql/schema.sql).

2. Create the `.env` file with the following format (replace <> with the real value).

```env
POSTGRE_USER=<your_postgres_user>
POSTGRE_PASSWORD=<your_postgres_password>
POSTGRE_DBNAME=<your_postgres_dbname>
POSTGRE_HOST=<your_postgres_host>
POSTGRE_PORT=<your_postgres_port>
HOST=<server_host>
PORT=<server_port>
PYTHON_ENV_NAME=<your_python_venv_folder_name>
```

3. Set up Python virtual environment and install dependencies

**Universal setup (any OS) - Recommended:**
```bash
# Compile and run the setup tool
go run tools/setup-venv.go
```

**For Windows (PowerShell):**
```powershell
./venv.ps1
```

**For Windows (Command Prompt/Batch):**
```cmd
venv.bat
```

**For macOS/Linux (Bash):**
```bash
chmod +x venv.sh
./venv.sh
```

**Manual setup (any OS):**
```bash
cd python
python -m venv venv           # Windows
# OR
python3 -m venv venv          # macOS/Linux

# Activate virtual environment:
# Windows:
venv\Scripts\activate
# macOS/Linux:
source venv/bin/activate

# Install dependencies:
pip install -r requirements.txt
cd ..
```

## Usage

1. Build application from source code
```powershell
go build -o <output_file_name>
```

2. Import all the environment variables from the `.env` file
```powershell
get-content .env | foreach {
    $name, $value = $_.split('=')
    set-content env:\$name $value
}
```

3. Run the application
```powershell
./<output_file_name>
```

## Development
To run the application in development mode (with hot-reload), use the provided `dev` script.
```powershell
./dev.ps1
```
Note: This script is for Powershell and require `air` package to be installed.

## Docker
*Docker is currently limited and some features may not work (e.g., UI's `login` and `join group` function).*

To build and run the application using Docker, follow these steps:
1. Build the Docker image
```bash
docker build -t asfpc-go .
```

2. Run the Docker container
```bash
docker run -d -p <host_port>:<container_port> --env-file .env asfpc-go
```