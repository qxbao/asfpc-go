package services

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

func GetAccountStats(s infras.Server, c echo.Context) error {
	queries := s.Queries
	stats, err := queries.GetAccountStats(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve account stats: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]any{
		"data": stats,
	})
}

func AddAccount(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.CreateAccountDTO)

	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}
	if dto.Username == nil || dto.Password == nil {
		return c.String(http.StatusBadRequest, "Username and password are required")
	}

	ua := GenerateModernChromeUA()

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

	account, err := queries.CreateAccount(c.Request().Context(), params)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to create account: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":    "Account created successfully",
		"account":    account,
		"user_agent": ua,
	})
}

func GetAccounts(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.GetAccountsDTO)
	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	if dto.Page == nil {
		dto.Page = new(int32)
		*dto.Page = 0
	}

	if dto.Limit == nil {
		dto.Limit = new(int32)
		*dto.Limit = 10
	}

	accounts, err := queries.GetAccounts(c.Request().Context(), db.GetAccountsParams{
		Offset: *dto.Page * *dto.Limit,
		Limit:  *dto.Limit,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve accounts: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": accounts,
	})
}

func GenAccountAT(s infras.Server, c echo.Context) error {
	account_id := c.Param("id")
	val, err := strconv.ParseInt(account_id, 10, 32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve accounts: " + err.Error(),
		})
	}
	queries := s.Queries
	account, err := queries.GetAccountById(c.Request().Context(), int32(val))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve account: " + err.Error(),
		})
	}

	fg := FacebookGraph{}
	username := account.Email
	if username == "" {
		username = account.Username
	}

	at, err := fg.GenerateFBAccessToken(username, account.Password)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Failed to generate access token: " + err.Error(),
		})
	}

	queries.UpdateAccountAccessToken(c.Request().Context(), db.UpdateAccountAccessTokenParams{
		ID:          account.ID,
		AccessToken: sql.NullString{String: *at, Valid: true},
	})

	return c.JSON(http.StatusOK, map[string]any{
		"access_token": at,
	})
}

func CreateGroup(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.CreateGroupDTO)
	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	if dto.GroupId == nil || dto.GroupName == nil || dto.AccountId == nil {
		return c.String(http.StatusBadRequest, "GroupID, GroupName and AccountID are required")
	}

	params := db.CreateGroupParams{
		GroupID:   *dto.GroupId,
		GroupName: *dto.GroupName,
		AccountID: sql.NullInt32{Int32: *dto.AccountId, Valid: true},
	}

	group, err := queries.CreateGroup(c.Request().Context(), params)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to create group: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": group,
	})
}
