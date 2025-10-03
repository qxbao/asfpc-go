package infras

import (
	"context"
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"go.uber.org/fx"
)

type Server struct {
	Port         *string
	Host         *string
	Database     *sql.DB
	Queries      *db.Queries
	Echo         *echo.Echo
}

var logger = lg.GetLogger("Infras")

func (s *Server) GetConfig(ctx context.Context, key string, replace string) string {
	config, err := s.Queries.GetConfigByKey(ctx, key)
	if err != nil {
		conf, err := s.Queries.UpsertConfig(ctx, db.UpsertConfigParams{
			Key:   key,
			Value: replace,
		})
		if err != nil {
			logger.Errorw("Failed to upsert config", "key", key, "error", err)
			return replace
		}
		return conf.Value
	}
	return config.Value
}

func (s *Server) GetConfigs(ctx context.Context) (map[string]string, error) {
	configs, err := s.Queries.GetAllConfigs(ctx)
	if err != nil {
		logger.Errorw("Failed to get configs", "error", err)
		return nil, err
	}
	result := make(map[string]string)
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}

func (s *Server) RegisterHooks(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := lg.InitLogger(true); err != nil {
				return err
			}

			log := lg.GetLogger("ServerInitialization")

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
			log := lg.GetLogger("ServerShutdown")
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

			return lg.FlushLogger()
		},
	})
}