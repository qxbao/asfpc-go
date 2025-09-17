package infras

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
)

type Server struct {
	Port         *string
	Host         *string
	Database     *sql.DB
	Queries      *db.Queries
	Echo         *echo.Echo
	GlobalConfig *map[string]string
}

func (s *Server) GetConfig(key string, replace string) string {
	value, exists := (*s.GlobalConfig)[key]
	if !exists {
		return replace
	}
	return value
}