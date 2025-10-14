package category

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

func (s *CategoryRoutingService) AddGroupCategory(c echo.Context) error {
	dto := new(infras.AddGroupCategoryRequest)

	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}

	err := s.Server.Queries.AddGroupCategory(c.Request().Context(), db.AddGroupCategoryParams{
		GroupID:    dto.GroupId,
		CategoryID: dto.CategoryId,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed to create group category: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": "Group category added successfully",
	})
}

func (s *CategoryRoutingService) GetGroupCategories(c echo.Context) error {
	groupId := c.Param("id")
	if groupId == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid group ID",
		})
	}
	groupIdInt, err := strconv.ParseInt(groupId, 32, 10)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid group ID",
		})
	}
	categories, err := s.Server.Queries.GetGroupCategories(c.Request().Context(), int32(groupIdInt))

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

func (s *CategoryRoutingService) DeleteGroupCategory(c echo.Context) error {
	dto := new(infras.DeleteGroupCategoryRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
	}
	err := s.Server.Queries.DeleteGroupCategory(c.Request().Context(), db.DeleteGroupCategoryParams{
		GroupID:    dto.GroupId,
		CategoryID: dto.CategoryId,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed to delete group category: " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"message": "Group category deleted successfully",
	})
}
