package infras

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
)

// MLModelConfigDTO represents the request/response for ML model configuration
type MLModelConfigDTO struct {
	CategoryID int    `json:"category_id"`
	ModelPath  string `json:"model_path"`
}

// GetMLModelConfig retrieves the ML model path for a category
func (s *Server) GetMLModelConfig(c echo.Context) error {
	categoryID := c.Param("category_id")
	if categoryID == "" {
		return c.JSON(400, map[string]string{"error": "category_id is required"})
	}

	config, err := s.Queries.GetMLModelConfig(c.Request().Context(), categoryID)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "ML model config not found for this category"})
	}

	catID, _ := strconv.Atoi(categoryID)
	return c.JSON(200, MLModelConfigDTO{
		CategoryID: catID,
		ModelPath:  config.Value,
	})
}

// SetMLModelConfig sets the ML model path for a category
func (s *Server) SetMLModelConfig(c echo.Context) error {
	var dto MLModelConfigDTO
	if err := c.Bind(&dto); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request body"})
	}

	if dto.CategoryID == 0 {
		return c.JSON(400, map[string]string{"error": "category_id is required"})
	}

	if dto.ModelPath == "" {
		return c.JSON(400, map[string]string{"error": "model_path is required"})
	}

	err := s.Queries.SetMLModelConfig(c.Request().Context(), db.SetMLModelConfigParams{
		Column1: strconv.Itoa(dto.CategoryID),
		Value:   dto.ModelPath,
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to set ML model config"})
	}

	return c.JSON(200, dto)
}

// GetEmbeddingModelConfig retrieves the embedding model path for a category
func (s *Server) GetEmbeddingModelConfig(c echo.Context) error {
	categoryID := c.Param("category_id")
	if categoryID == "" {
		return c.JSON(400, map[string]string{"error": "category_id is required"})
	}

	config, err := s.Queries.GetEmbeddingModelConfig(c.Request().Context(), categoryID)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "Embedding model config not found for this category"})
	}

	catID, _ := strconv.Atoi(categoryID)
	return c.JSON(200, MLModelConfigDTO{
		CategoryID: catID,
		ModelPath:  config.Value,
	})
}

// SetEmbeddingModelConfig sets the embedding model path for a category
func (s *Server) SetEmbeddingModelConfig(c echo.Context) error {
	var dto MLModelConfigDTO
	if err := c.Bind(&dto); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request body"})
	}

	if dto.CategoryID == 0 {
		return c.JSON(400, map[string]string{"error": "category_id is required"})
	}

	if dto.ModelPath == "" {
		return c.JSON(400, map[string]string{"error": "model_path is required"})
	}

	err := s.Queries.SetEmbeddingModelConfig(c.Request().Context(), db.SetEmbeddingModelConfigParams{
		Column1: strconv.Itoa(dto.CategoryID),
		Value:   dto.ModelPath,
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to set embedding model config"})
	}

	return c.JSON(200, dto)
}

// GetAllCategoryMLConfigs retrieves all ML model configurations
func (s *Server) GetAllCategoryMLConfigs(c echo.Context) error {
	configs, err := s.Queries.GetCategoryMLConfigs(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to retrieve ML configs"})
	}

	return c.JSON(200, map[string]interface{}{
		"data": configs,
	})
}
