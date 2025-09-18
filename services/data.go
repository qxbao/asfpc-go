package services

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
)

type DataService struct {
	Server infras.Server
}

func (ds *DataService) GetDataStats(c echo.Context) error {
	queries := ds.Server.Queries
	stats, err := queries.GetStats(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error":   "failed to get data stats: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": stats,
	})
}