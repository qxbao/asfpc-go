package data

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type DataRoutingService infras.RoutingService

// Helper function for consistent error responses
func errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]any{
		"error": message,
	})
}

// Helper function for consistent success responses
func successResponse(c echo.Context, data any) error {
	return c.JSON(200, map[string]any{
		"data": data,
	})
}

func (ds *DataRoutingService) GetDataStats(c echo.Context) error {
	queries := ds.Server.Queries
	stats, err := queries.GetStats(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get data stats: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": stats,
	})
}

func (ds *DataRoutingService) TraceRequest(c echo.Context) error {
	queries := ds.Server.Queries
	dto := new(infras.TraceRequestDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	trace, err := queries.GetRequestById(c.Request().Context(), dto.RequestID)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to trace request: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": trace,
	})
}

func (ds *DataRoutingService) GetAllPrompts(c echo.Context) error {
	queries := ds.Server.Queries
	dto := new(infras.QueryWithPageDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
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
			"error": "failed to get prompts: " + err.Error(),
		})
	}

	if prompts == nil {
		prompts = make([]db.GetAllPromptsRow, 0)
	}

	count, err := queries.CountPrompts(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to count prompts: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data":  prompts,
		"total": count,
	})
}

func (ds *DataRoutingService) CreatePrompt(c echo.Context) error {
	queries := ds.Server.Queries

	dto := new(infras.CreatePromptRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	prompt, err := queries.CreatePrompt(c.Request().Context(), db.CreatePromptParams{
		ServiceName: dto.ServiceName,
		Content:     dto.Content,
		CreatedBy:   dto.CreatedBy,
		CategoryID:  int32(dto.CategoryID),
	})

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to create prompt: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": prompt,
	})
}

func (ds *DataRoutingService) DeletePrompt(c echo.Context) error {
	queries := ds.Server.Queries

	dto := new(infras.DeletePromptRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	err := queries.DeletePrompt(c.Request().Context(), dto.ID)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to delete prompt: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": "Prompt deleted successfully",
	})
}

func (ds *DataRoutingService) RollbackPrompt(c echo.Context) error {
	queries := ds.Server.Queries
	dto := new(infras.RollbackPromptRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	err := queries.RollbackPrompt(c.Request().Context(), db.RollbackPromptParams{
		CategoryID:  dto.CategoryID,
		ServiceName: dto.ServiceName,
	})
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to rollback prompt: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": "Prompt rolled back successfully",
	})
}

func (ds *DataRoutingService) GetLogs(c echo.Context) error {
	queries := ds.Server.Queries
	dto := new(infras.QueryWithPageDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	if dto.Page == nil {
		dto.Page = new(int32)
		*dto.Page = 0
	}
	if dto.Limit == nil {
		dto.Limit = new(int32)
		*dto.Limit = 10
	}
	logs, err := queries.GetLogs(c.Request().Context(), db.GetLogsParams{
		Limit:  *dto.Limit,
		Offset: *dto.Page * *dto.Limit,
	})
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get logs: " + err.Error(),
		})
	}
	if logs == nil {
		logs = make([]db.GetLogsRow, 0)
	}
	count, err := queries.CountLogs(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to count logs: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data":  logs,
		"total": count,
	})
}

// Charts API Services

func (ds *DataRoutingService) GetDashboardStats(c echo.Context) error {
	queries := ds.Server.Queries
	stats, err := queries.GetDashboardStats(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get dashboard stats: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": stats,
	})
}

func (ds *DataRoutingService) GetTimeSeriesData(c echo.Context) error {
	queries := ds.Server.Queries
	categoryIDStr := c.QueryParam("category_id")
	
	var data []db.GetTimeSeriesDataRow
	var err error
	
	if categoryIDStr != "" {
		categoryID, parseErr := strconv.Atoi(categoryIDStr)
		if parseErr != nil {
			return errorResponse(c, 400, "Invalid category_id parameter")
		}
		categoryData, err := queries.GetTimeSeriesDataByCategory(c.Request().Context(), int32(categoryID))
		if err != nil {
			return errorResponse(c, 500, "failed to get time series data: "+err.Error())
		}
		
		// Convert category data to standard format
		data = make([]db.GetTimeSeriesDataRow, len(categoryData))
		for i, row := range categoryData {
			data[i] = db.GetTimeSeriesDataRow{
				Date:  row.Date,
				Count: row.Count,
			}
		}
	} else {
		data, err = queries.GetTimeSeriesData(c.Request().Context())
		if err != nil {
			return errorResponse(c, 500, "failed to get time series data: "+err.Error())
		}
	}

	if data == nil {
		data = make([]db.GetTimeSeriesDataRow, 0)
	}

	return successResponse(c, data)
}

type Data struct {
	Range       string  `json:"range"`
	GeminiScore float64 `json:"gemini_score"`
	ModelScore  float64 `json:"model_score"`
}

func (ds *DataRoutingService) GetScoreDistribution(c echo.Context) error {
	queries := ds.Server.Queries
	categoryIDStr := c.QueryParam("category_id")
	
	var scoreDistribution []db.GetScoreDistributionRow
	var err error
	
	if categoryIDStr != "" {
		categoryID, parseErr := strconv.Atoi(categoryIDStr)
		if parseErr != nil {
			return errorResponse(c, 400, "Invalid category_id parameter")
		}
		categoryData, err := queries.GetScoreDistributionByCategory(c.Request().Context(), int32(categoryID))
		if err != nil {
			return errorResponse(c, 500, "failed to get score distribution: "+err.Error())
		}
		
		// Convert category data to standard format
		scoreDistribution = make([]db.GetScoreDistributionRow, len(categoryData))
		for i, row := range categoryData {
			scoreDistribution[i] = db.GetScoreDistributionRow{
				ScoreRange:  row.ScoreRange,
				GeminiCount: row.GeminiCount,
				ModelCount:  row.ModelCount,
			}
		}
	} else {
		scoreDistribution, err = queries.GetScoreDistribution(c.Request().Context())
		if err != nil {
			return errorResponse(c, 500, "failed to get score distribution: "+err.Error())
		}
	}

	if scoreDistribution == nil {
		scoreDistribution = make([]db.GetScoreDistributionRow, 0)
	}

	// Map từ kết quả query
	data := make([]Data, 0)
	for _, row := range scoreDistribution {
		data = append(data, Data{
			Range:       row.ScoreRange,
			GeminiScore: float64(row.GeminiCount),
			ModelScore:  float64(row.ModelCount),
		})
	}

	return successResponse(c, data)
}
