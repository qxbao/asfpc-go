package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitAccountRoutes(s infras.Server) {
	e := s.Echo
	queries := s.Queries
	e.POST("/account/add", func(c echo.Context) error {
		dto := new(infras.CreateAccountDTO)

		if err := c.Bind(dto); err != nil {
			return c.String(http.StatusBadRequest, "Invalid request body")
		}
		if dto.Username == nil || dto.Password == nil {
			return c.String(http.StatusBadRequest, "Username and password are required")
		}

		ua := services.GenerateModernChromeUA()

		params := db.CreateAccountParams{
			Email:       *dto.Email,
			Username:    *dto.Username,
			Password:    *dto.Password,
			IsBlock:     false,
			Ua:          ua,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			AccessToken: sql.NullString{Valid: false},
			ProxyID:     sql.NullInt32{Valid: false},
		}

		account, err := queries.CreateAccount(context.Background(), params)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Failed to create account: " + err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "Account created successfully",
			"account":    account,
			"user_agent": ua, // Show the generated user agent
		})
	})
}
