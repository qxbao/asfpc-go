package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
)

type MLService struct {
	Server infras.Server
}

func (s *MLService) Train(c echo.Context) error {
	dto := new(infras.MLTrainDTO)

	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	if dto.ModelName == nil {
		dto.ModelName = new(string)
		*dto.ModelName = "Model_" + time.Now().Format("20060102150405")
	}

	pythonService := PythonService{
		EnvName: os.Getenv("PYTHON_ENV_NAME"),
	}

	res, err := pythonService.RunScript("--task=train-model", fmt.Sprintf("--model-name=%s", *dto.ModelName))
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to run python script: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": res,
	})
}

type ModelMetadata struct {
	RMSE     float64 `json:"rmse"`
	R2      float64 `json:"r2"`
	MAE     float64 `json:"mae"`
	SavedAt string  `json:"saved_at"`
}

type ModelInfo struct {
	Name     string
	Metadata *ModelMetadata
}

func (s *MLService) ListModels(c echo.Context) error {
	modelsDir := "python/resources/models/"
	exc, err := os.Executable()

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to get executable path: " + err.Error(),
		})
	}

	modelsPath := path.Join(path.Dir(exc), modelsDir)

	if err := os.MkdirAll(modelsPath, 0755); err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to create models directory: " + err.Error(),
		})
	}

	folders, err := os.ReadDir(modelsPath)

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to read models directory: " + err.Error(),
		})
	}

	models := make([]ModelInfo, 0)
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		models = append(models, ModelInfo{Name: folder.Name()})
		testResultsPath := path.Join(path.Dir(exc), modelsDir, folder.Name(), "metadata.json")
		if _, err := os.Stat(testResultsPath); err == nil {
			data, err := os.ReadFile(testResultsPath)
			var metadata ModelMetadata
			if err == nil {
				if err := json.Unmarshal(data, &metadata); err == nil {
					models[len(models)-1].Metadata = &metadata
				}
			} else {
				metadata = ModelMetadata{}
				models[len(models)-1].Metadata = &metadata
			}
		}
	}

	return c.JSON(200, map[string]any{
		"total": len(models),
		"data":  models,
	})
}
