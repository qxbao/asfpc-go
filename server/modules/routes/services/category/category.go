package category

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type CategoryRoutingService infras.RoutingService

func (s *CategoryRoutingService) GetCategories(c echo.Context) error {
	categories, err := s.Server.Queries.GetCategories(c.Request().Context())
	
	if err != nil {
		return err
	}

	if categories == nil {
		categories = make([]db.Category, 0)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": categories,
	})
}

func (s *CategoryRoutingService) AddCategory(c echo.Context) error {
	dto := new(infras.AddCategoryRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	
	category, err := s.Server.Queries.CreateCategory(c.Request().Context(), db.CreateCategoryParams{
		Name:        dto.Name,
		Description: sql.NullString{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed to create category: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]any{
		"data": category,
	})
}

func (s *CategoryRoutingService) DeleteCategory(c echo.Context) error {
	categoryId := c.Param("id")
	if categoryId == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid category ID",
		})
	}
	categoryIdInt, err := strconv.ParseInt(categoryId, 32, 10)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid category ID",
		})
	}

	err = s.Server.Queries.DeleteCategory(c.Request().Context(), int32(categoryIdInt))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed to delete category: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Category deleted successfully",
	})
}

func (s *CategoryRoutingService) UpdateCategory(c echo.Context) error {
	dto := new(infras.UpdateCategoryRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	category, err := s.Server.Queries.UpdateCategory(c.Request().Context(), db.UpdateCategoryParams{
		ID:          dto.Id,
		Name:        dto.Name,
		Description: sql.NullString{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed to update category: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]any{
		"data": category,
	})
}
