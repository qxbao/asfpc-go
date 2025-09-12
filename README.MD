# ASFPC-Go

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
- Docker (optional, for containerization)

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
go mod tidy
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
HOST=<your_host>
PORT=<your_port>
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
To run the application in development mode, use the provided `dev` script.
1. Run the application in development mode
```powershell
./dev.ps1
```

## Docker
To build and run the application using Docker, follow these steps:
1. Build the Docker image
```bash
docker build -t asfpc-go .
```

2. Run the Docker container
```bash
docker run -d -p <host_port>:<container_port> --env-file .env asfpc-go
```