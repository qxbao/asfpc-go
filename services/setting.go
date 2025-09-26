package services

import (
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type SettingService struct {
	Server infras.Server
}

func (s *SettingService) GetSettings(c echo.Context) error {
	settings, err := s.Server.GetConfigs(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": settings,
	})
}

func (s *SettingService) UpdateSettings(c echo.Context) error {
	dto := new(infras.UpdateSettingsDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}
	ctx := c.Request().Context()
	for key, value := range dto.Settings {
		_, err := s.Server.Queries.UpsertConfig(ctx, db.UpsertConfigParams{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}
	}
	return c.JSON(200, map[string]any{
		"message": "Settings updated successfully",
	})
}
