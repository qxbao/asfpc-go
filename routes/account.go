package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitAccountRoutes(s infras.Server) {
	e := s.Echo

	e.GET("/account/info", func(c echo.Context) error {
		return services.GetAccount(s, c)
	})

	e.POST("/account/update/credentials", func(c echo.Context) error {
		return services.UpdateAccountCredentials(s, c)
	})

	e.GET("/account/list", func(c echo.Context) error {
		return services.GetAccounts(s, c)
	})

	e.GET("/account/stats", func(c echo.Context) error {
		return services.GetAccountStats(s, c)
	})

	e.POST("/account/add", func(c echo.Context) error {
		return services.AddAccount(s, c)
	})

	e.POST("/account/token/gen", func (c echo.Context) error {
		return services.GenAccountsAT(s, c)
	})

	e.POST("/account/group/link", func (c echo.Context) error {
		return services.CreateGroup(s, c)
	})

	e.DELETE("/account/delete", func(c echo.Context) error {
		return services.DeleteAccounts(s, c)
	})
}