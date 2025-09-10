package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"net/http"
	"encoding/json"
)

func InitAccountRoutes(e *echo.Echo) {
	e.POST("/account", func(c echo.Context) error {
		account := new(infras.AccountRequest)
		if err := c.Bind(account); err != nil {
			return c.String(http.StatusBadRequest, "Invalid request body")
		}

		if account.Username == nil || account.Password == nil {
			return c.String(http.StatusBadRequest, "Username and Password are required")
		}

		account_val, _ := json.Marshal(account)

		return c.String(http.StatusOK, "Account created successfully {account: "+string(account_val)+"}")
	})
}