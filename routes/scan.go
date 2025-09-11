package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitScanRoutes(s infras.Server) {
	e := s.Echo
	
	e.POST("/scan/group/:id", func(c echo.Context) error {
		return services.ScanGroupFeed(s, c)
	})

	e.POST("/scan/post/:id", func(c echo.Context) error {
		return services.ScanPostComments(s, c)
	})
}