package infras

import (
	"context"
	"database/sql"

	lg "github.com/qxbao/asfpc/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
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