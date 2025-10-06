package ml

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/pkg/utils/python"
)

type MLRoutingService struct {
	Server *infras.Server
}

var logger = lg.GetLogger("MLRoutingService")

func (s *MLRoutingService) Train(c echo.Context) error {
	dto := new(infras.MLTrainDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{"error": "Invalid request body"})
	}

	if dto.ModelName == nil {
		name := "Model_" + time.Now().Format("20060102150405")
		dto.ModelName = &name
	}

	id, err := s.Server.Queries.CreateRequest(c.Request().Context(), sql.NullString{
		String: fmt.Sprintf("Queueing model %s's training task...", *dto.ModelName),
		Valid:  true,
	})

	if err != nil {
		return c.JSON(500, map[string]any{"error": "Cannot create request"})
	}

	go s.trainingTask(id, dto)

	return c.JSON(200, map[string]any{
		"request_id": id,
		"message":    "Training started",
	})
}

func (s *MLRoutingService) trainingTask(requestId int32, dto *infras.MLTrainDTO) {
	pythonService := python.NewPythonService(os.Getenv("PYTHON_ENV_NAME"), true, true, nil)
	autoTune := "False"

	if *dto.AutoTune {
		autoTune = "True"
	}

	args := []string{
		"--task=train-model",
		fmt.Sprintf("--model-name=%s", *dto.ModelName),
		fmt.Sprintf("--auto-tune=%s", autoTune),
		fmt.Sprintf("--request-id=%d", requestId),
	}

	if dto.Trials != nil {
		args = append(args, fmt.Sprintf("--trials=%d", *dto.Trials))
	}

	_, err := pythonService.RunScript(
		args...,
	)

	if err != nil {
		err := s.Server.Queries.UpdateRequestStatus(context.Background(), db.UpdateRequestStatusParams{
			ID:           requestId,
			Status:       3,
			ErrorMessage: sql.NullString{String: err.Error(), Valid: true},
			Description:  sql.NullString{String: "Training failed.", Valid: true},
		})
		if err != nil {
			logger.Errorf("Failed to update request status for request %d: %v", requestId, err)
		}
		return
	}
	err = s.Server.Queries.UpdateRequestStatus(context.Background(), db.UpdateRequestStatusParams{
		ID:          requestId,
		Status:      2,
		Progress:    1.0,
		Description: sql.NullString{String: "Training completed.", Valid: true},
	})
	if err != nil {
		logger.Errorf("Failed to update request status for request %d: %v", requestId, err)
	}
}

type PredictionStats struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
}

type ResidualStats struct {
	Mean           float64 `json:"mean"`
	Std            float64 `json:"std"`
	BiasLowScores  float64 `json:"bias_low_scores"`
	BiasHighScores float64 `json:"bias_high_scores"`
}

type TopFeatures map[string]float64

type TrainParams map[string]any

type ModelMetadata struct {
	RMSE            float64         `json:"rmse"`
	R2              float64         `json:"r2"`
	MAE             float64         `json:"mae"`
	RMSLE           float64         `json:"rmsle"`
	SMAPE           float64         `json:"smape"`
	PredictionStats PredictionStats `json:"prediction_stats"`
	ResidualStats   ResidualStats   `json:"residual_stats"`
	TopFeatures     TopFeatures     `json:"top_features"`
	SavedAt         string          `json:"saved_at"`
	IsGPU           bool            `json:"is_gpu"`
	TrainParams     TrainParams     `json:"train_params"`
}

type ModelInfo struct {
	Name       string
	Metadata   *ModelMetadata
	Validation *ModelValidation
}

func (s *MLRoutingService) ListModels(c echo.Context) error {
	modelsDir := path.Join("python", "resources", "models")
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
		validation, err := s.ValidateModel(folder.Name())
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": "Failed to validate model " + folder.Name() + ": " + err.Error(),
			})
		}
		models[len(models)-1].Validation = &validation
		if !validation.IsValid {
			continue
		}
		metadataPath := path.Join(path.Dir(exc), modelsDir, folder.Name(), "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			data, err := os.ReadFile(metadataPath)
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

type ModelValidation struct {
	IsExists bool
	IsValid  bool
}

func (s *MLRoutingService) ValidateModel(modelName string) (ModelValidation, error) {
	exc, err := os.Executable()
	modelsDir := path.Join("python", "resources", "models")

	if err != nil {
		return ModelValidation{}, err
	}

	modelsPath := path.Join(path.Dir(exc), modelsDir)

	if _, err := os.Stat(path.Join(modelsPath, modelName)); err != nil {
		if os.IsNotExist(err) {
			return ModelValidation{IsExists: false, IsValid: false}, nil
		}
		return ModelValidation{}, err
	}

	dirs, err := os.ReadDir(path.Join(modelsPath, modelName))
	if err != nil {
		return ModelValidation{IsExists: true, IsValid: false}, nil
	}
	requiredFiles := map[string]bool{
		"model.json":    true,
		"encoders.pkl":  true,
		"scalers.pkl":   true,
		"metadata.json": true,
	}
	validCount := 0
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if requiredFiles[d.Name()] {
			validCount++
		}
	}

	return ModelValidation{IsExists: true, IsValid: validCount == len(requiredFiles)}, nil
}

func (s *MLRoutingService) DeleteModel(c echo.Context) error {
	dto := new(infras.WithModelNameDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	if dto.ModelName == "" {
		return c.JSON(400, map[string]any{
			"error": "Model name is required",
		})
	}

	if dto.ModelName == "." || dto.ModelName == ".." ||
		strings.Contains(dto.ModelName, "/") ||
		strings.Contains(dto.ModelName, "\\") ||
		strings.Contains(dto.ModelName, "*") {
		return c.JSON(400, map[string]any{
			"error": "Invalid model name",
		})
	}

	validation, err := s.ValidateModel(dto.ModelName)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to validate model: " + err.Error(),
		})
	}
	if !validation.IsExists {
		return c.JSON(400, map[string]any{
			"error": "Model does not exist",
		})
	}

	exc, err := os.Executable()
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to get executable path: " + err.Error(),
		})
	}

	modelsDir := path.Join("python", "resources", "models")
	modelPath := path.Join(path.Dir(exc), modelsDir, dto.ModelName)

	// Double-check the path is within the models directory
	modelsBasePath := path.Join(path.Dir(exc), modelsDir)
	if !strings.HasPrefix(modelPath, modelsBasePath) {
		return c.JSON(400, map[string]any{
			"error": "Invalid model path",
		})
	}

	if err := os.RemoveAll(modelPath); err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to delete model: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"message": "Model deleted successfully",
	})
}

func (s *MLRoutingService) ExportModel(c echo.Context) error {
	dto := new(infras.WithModelNameDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}
	validation, err := s.ValidateModel(dto.ModelName)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to validate model: " + err.Error(),
		})
	}
	if !validation.IsExists {
		return c.JSON(400, map[string]any{
			"error": "Model does not exist",
		})
	}

	exc, err := os.Executable()
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to get executable path: " + err.Error(),
		})
	}

	modelsDir := path.Join("python", "resources", "models")
	modelPath := path.Join(path.Dir(exc), modelsDir, dto.ModelName)
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	dirs, err := os.ReadDir(modelPath)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to read model directory: " + err.Error(),
		})
	}

	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		filePath := path.Join(modelPath, d.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": "Failed to read model file: " + err.Error(),
			})
		}

		f, err := zipWriter.Create(d.Name())
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": "Failed to create zip file: " + err.Error(),
			})
		}
		_, err = f.Write(fileData)
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": "Failed to write to zip file: " + err.Error(),
			})
		}
	}

	if err := zipWriter.Close(); err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to close zip file: " + err.Error(),
		})
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/zip")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+dto.ModelName+`.zip"`)
	c.Response().WriteHeader(http.StatusOK)
	_, err = c.Response().Write(buf.Bytes())
	return err
}
