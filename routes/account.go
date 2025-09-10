package routes

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"encoding/json"
)

func InitScanRoutes(e *echo.Echo) {
	e.POST("/scan/posts", func(c echo.Context) error {
		account := new(infras.AccountRequest)

		if err := c.Bind(account); err != nil {
			return c.String(http.StatusBadRequest, "Invalid request body")
		}

		if account.Username == nil || account.AccessToken == nil {
			return c.String(http.StatusBadRequest, "Username and AccessToken are required")
		}

		if account.IsBlock == nil {
			account.IsBlock = new(bool)
			*account.IsBlock = false
		}

		account_val, _ := json.Marshal(account)

		return c.String(http.StatusOK, "Account added successfully {account: "+string(account_val)+"}")
	})
}