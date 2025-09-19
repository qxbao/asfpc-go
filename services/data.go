package services

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
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

func (ds *DataService) GetAllPrompts(c echo.Context) error {
	queries := ds.Server.Queries
	dto := new(infras.QueryWithPageDTO)
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
	
	prompts, err := queries.GetAllPrompts(c.Request().Context(), db.GetAllPromptsParams{
		Limit:  *dto.Limit,
		Offset: *dto.Page * *dto.Limit,
	})
	
	if err != nil {
		return c.JSON(500, map[string]any{
			"error":   "failed to get prompts: " + err.Error(),
		})
	}

	if prompts == nil {
		prompts = make([]db.GetAllPromptsRow, 0)
	}

	count , err := queries.CountPrompts(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error":   "failed to count prompts: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": prompts,
		"total": count,
	})
}

func (ds *DataService) CreatePrompt(c echo.Context) error {
	queries := ds.Server.Queries
	
	dto := new(infras.CreatePromptRequest)
	if err := c.Bind(dto); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	prompt, err := queries.CreatePrompt(c.Request().Context(), db.CreatePromptParams{
		ServiceName:     dto.ServiceName,
		Content:   dto.Content,
		CreatedBy: dto.CreatedBy,
	})

	if err != nil {
		return c.JSON(500, map[string]any{
			"error":   "failed to create prompt: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": prompt,
	})
}
