package server

import (
	"context"
	"database/sql"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/logger"
	"go.uber.org/fx"
)

type Server struct {
	*infras.Server
}

// NewServer creates a new server instance with Echo already initialized
func NewServer(database *sql.DB, queries *db.Queries) *Server {
	// Create Echo instance immediately during construction
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	// Initialize server with all dependencies
	infraServer := &infras.Server{
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

	infraServer.Host = &HOST
	infraServer.Port = &PORT

	return &Server{
		Server: infraServer,
	}
}

func (s *Server) RegisterHooks(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := logger.InitLogger(true); err != nil {
				return err
			}

			log := logger.GetLogger("SERVER")

			go func() {
				address := *s.Host + ":" + *s.Port
				log.Info("Starting server on", address)

				if err := s.Echo.Start(address); err != nil {
					log.Error("Server failed to start:", err)
				}
			}()

			log.Info("Server startup initiated")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log := logger.GetLogger("SERVER")
			log.Info("Stopping server...")

			if s.Echo != nil {
				if err := s.Echo.Shutdown(ctx); err != nil {
					log.Error("Failed to shutdown server:", err)
				}
			}

			if s.Database != nil {
				if err := s.Database.Close(); err != nil {
					log.Error("Failed to close database:", err)
				}
			}

			return logger.FlushLogger()
		},
	})
}

var ServerModule = fx.Module("Server",
	fx.Provide(NewServer),
	fx.Invoke(func(s *Server, lc fx.Lifecycle) {
		s.RegisterHooks(lc)
	}),
)
