package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitAccountRoutes(s infras.Server) {
	e := s.Echo
	
	e.GET("/account/list", func(c echo.Context) error {
		return services.GetAccounts(s, c)
	})

	e.POST("/account/add", func(c echo.Context) error {
		return services.AddAccount(s, c)
	})

	e.POST("/account/token/gen/:id", func (c echo.Context) error {
		return services.GenAccountAT(s, c)
	})

	e.POST("/account/group/link", func (c echo.Context) error {
		return services.CreateGroup(s, c)
	})
}