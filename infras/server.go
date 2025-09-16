package infras

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/cron"
)

type Server struct {
	Port         *string
	Host         *string
	Database     *sql.DB
	Queries      *db.Queries
	Echo         *echo.Echo
	Scheduler    *cron.CronScheduler
	GlobalConfig *map[string]string
}

func (s *Server) GetConfig(key string, replace string) string {
	value, exists := (*s.GlobalConfig)[key]
	if !exists {
		return replace
	}
	return value
}