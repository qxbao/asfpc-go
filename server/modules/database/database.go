package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/pkg/logger"
	"go.uber.org/fx"
)

func NewDatabase(lc fx.Lifecycle) (*sql.DB, error) {
	logger := logger.GetLogger("DatabaseModule")
	pgUser := os.Getenv("POSTGRE_USER")
	pgPassword := os.Getenv("POSTGRE_PASSWORD")
	pgDBName := os.Getenv("POSTGRE_DBNAME")
	pgHost := os.Getenv("POSTGRE_HOST")
	pgPort := os.Getenv("POSTGRE_PORT")
	dataSourceName := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgUser, pgPassword, pgHost, pgPort, pgDBName,
	)
	print(dataSourceName)
	database, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		return nil, err
	}

	err = database.Ping()

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing the database connection...")
			return database.Close()
		},
	})

	logger.Info("Connected to the database successfully!")
	return database, nil
}

func NewQueries(database *sql.DB) *db.Queries {
	return db.New(database)
}

var DatabaseModule = fx.Module("Database",
	fx.Provide(
		NewDatabase,
		NewQueries,
	),
)
