if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

export $(grep -v '^#' .env | xargs)
export GOOSE_DRIVER="postgres"
export GOOSE_DBSTRING="host=$POSTGRE_HOST port=$POSTGRE_PORT user=$POSTGRE_USER password=$POSTGRE_PASSWORD dbname=$POSTGRE_DBNAME sslmode=disable"
export GOOSE_MIGRATION_DIR="./db/migrations"

echo $GOOSE_DBSTRING