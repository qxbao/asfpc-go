package server

import (
	"database/sql"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"go.uber.org/fx"
)


// NewServer creates a new server instance with Echo already initialized
func NewServer(database *sql.DB, queries *db.Queries) *infras.Server {
	// Create Echo instance immediately during construction
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	// Initialize server with all dependencies
	server := &infras.Server{
		Database: database,
		Queries:  queries,
		Echo:     e, // Echo is now available immediately
	}

	// Set environment variables with defaults
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")

	if HOST == "" {
		HOST = "localhost"
	}
	if PORT == "" {
		PORT = "8000"
	}

	server.Host = &HOST
	server.Port = &PORT
	return (*infras.Server)(server)
}

var ServerModule = fx.Module("Server",
	fx.Provide(NewServer),
	fx.Invoke(func(s *infras.Server, lc fx.Lifecycle) {
		s.RegisterHooks(lc)
	}),
)
