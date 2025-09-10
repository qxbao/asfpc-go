package infras

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
)

type Server struct {
	Port     *string
	Host     *string
	Database *sql.DB
	Queries  *db.Queries
	Echo     *echo.Echo
}
