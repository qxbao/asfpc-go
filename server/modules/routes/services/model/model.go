package model

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type ModelRoutingService infras.RoutingService

func (s *ModelRoutingService) GetModels(c echo.Context) error {
	models, err := s.Server.Queries.GetModels(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to retrieve models: " + err.Error(),
		})
	}

	if models == nil {
		models = make([]db.Model, 0)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": models,
	})
}

func (s *ModelRoutingService) CreateModel(c echo.Context) error {
	dto := new(infras.CreateModelRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	var categoryID sql.NullInt32
	if dto.CategoryID != nil {
		categoryID = sql.NullInt32{
			Int32: *dto.CategoryID,
			Valid: true,
		}
	}

	model, err := s.Server.Queries.CreateModel(c.Request().Context(), db.CreateModelParams{
		Name: dto.Name,
		Description: sql.NullString{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		CategoryID: categoryID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to create model: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": model,
	})
}

func (s *ModelRoutingService) UpdateModel(c echo.Context) error {
	dto := new(infras.UpdateModelRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	var categoryID sql.NullInt32
	if dto.CategoryID != nil {
		categoryID = sql.NullInt32{
			Int32: *dto.CategoryID,
			Valid: true,
		}
	}

	model, err := s.Server.Queries.UpdateModel(c.Request().Context(), db.UpdateModelParams{
		ID:   dto.ID,
		Name: dto.Name,
		Description: sql.NullString{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		CategoryID: categoryID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to update model: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": model,
	})
}

func (s *ModelRoutingService) DeleteModel(c echo.Context) error {
	modelIDStr := c.Param("id")
	if modelIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid model ID",
		})
	}

	modelID, err := strconv.ParseInt(modelIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid model ID format",
		})
	}

	err = s.Server.Queries.DeleteModel(c.Request().Context(), int32(modelID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to delete model: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Model deleted successfully",
	})
}

func (s *ModelRoutingService) AssignModelToCategory(c echo.Context) error {
	dto := new(infras.AssignModelToCategoryRequest)
	
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	ctx := c.Request().Context()

	// Get existing model first
	existingModel, err := s.Server.Queries.GetModelByID(ctx, dto.ModelID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"error": "Model not found",
		})
	}

	// Check if category already has a model assigned
	existingCategoryModel, err := s.Server.Queries.GetModelByCategory(ctx, sql.NullInt32{
		Int32: dto.CategoryID,
		Valid: true,
	})
	
	// If there's an existing model for this category, unlink it first
	if err == nil && existingCategoryModel.ID != dto.ModelID {
		_, unlinkErr := s.Server.Queries.UpdateModel(ctx, db.UpdateModelParams{
			ID:          existingCategoryModel.ID,
			Name:        existingCategoryModel.Name,
			Description: existingCategoryModel.Description,
			CategoryID:  sql.NullInt32{Valid: false}, // Set to NULL
		})
		if unlinkErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Failed to unlink previous model: " + unlinkErr.Error(),
			})
		}
	}

	// Update model to assign category
	model, err := s.Server.Queries.UpdateModel(ctx, db.UpdateModelParams{
		ID:          dto.ModelID,
		Name:        existingModel.Name,
		Description: existingModel.Description,
		CategoryID: sql.NullInt32{
			Int32: dto.CategoryID,
			Valid: true,
		},
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to assign model to category: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": model,
	})
}
