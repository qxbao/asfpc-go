package services

import (
	"database/sql"
	"net/http"
	"sync"
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

func GetAccount(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.GetAccountDTO)

	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	account, err := queries.GetAccountById(c.Request().Context(), dto.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve account: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": account,
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

func DeleteAccounts(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.DeleteAccountsDTO)
	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}
	if len(dto.IDs) == 0 {
		return c.String(http.StatusBadRequest, "No account IDs provided")
	}
	err := queries.DeleteAccounts(c.Request().Context(), dto.IDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to delete accounts: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":     "Accounts deleted successfully",
		"deleted_ids": dto.IDs,
	})
}

func UpdateAccountCredentials(s infras.Server, c echo.Context) error {
	queries := s.Queries
	dto := new(infras.UpdateAccountCredentialsDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string] string{
			"error": "Invalid request body: " + err.Error(),
		})
	}
	account, err := queries.UpdateAccountCredentials(c.Request().Context(), db.UpdateAccountCredentialsParams{
		ID:       dto.ID,
		Email:    *dto.Email,
		Username: *dto.Username,
		Password: *dto.Password,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string] string{
			"error": "Failed to update account credentials: " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any {
		"message": "Account credentials updated successfully",
		"data": account,
	})
}

func GenAccountsAT(s infras.Server, c echo.Context) error {
	dto := new(infras.GenAccountsATDTO)
	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}
	queries := s.Queries
	if len(dto.IDs) == 0 {
		return c.String(http.StatusBadRequest, "No account IDs provided")
	}

	var wg sync.WaitGroup
	successCount := make(chan int32, len(dto.IDs))
	errorIds := make(chan int32, len(dto.IDs))
	errChan := make(chan error, len(dto.IDs))
	for _, id := range dto.IDs {
		wg.Add(1)
		go func(accountId int32) {
			defer wg.Done()
			account, err := queries.GetAccountById(c.Request().Context(), accountId)
			if err != nil {
				errChan <- err
				errorIds <- accountId
				return
			}
			fg := FacebookGraph{}
			username := account.Email
			if username == "" {
				username = account.Username
			}
			at, err := fg.GenerateFBAccessToken(username, account.Password)
			if err != nil {
				errChan <- err
				errorIds <- accountId
				return
			}
			queries.UpdateAccountAccessToken(c.Request().Context(), db.UpdateAccountAccessTokenParams{
				ID:          account.ID,
				AccessToken: sql.NullString{String: *at, Valid: true},
			})
			successCount <- 1
		}(id)
	}

	wg.Wait()
	close(errChan)
	close(errorIds)
	close(successCount)

	processed := 0
	for range successCount {
		processed++
	}

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	var eIds []int32
	for id := range errorIds {
		eIds = append(eIds, id)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"success_count":  processed,
			"error_accounts": eIds,
			"errors":         errors,
		},
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
