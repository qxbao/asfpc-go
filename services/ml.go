package services

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
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/async"
)

type MLService struct {
	Server infras.Server
}

func (s *MLService) Train(c echo.Context) error {
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

func (s *MLService) trainingTask(requestId int32, dto *infras.MLTrainDTO) {
	pythonService := PythonService{
		EnvName: os.Getenv("PYTHON_ENV_NAME"),
		Log:     true,
	}
	autoTune := "False"

	if *dto.AutoTune {
		autoTune = "True"
	}

	_, err := pythonService.RunScript("--task=train-model",
		fmt.Sprintf("--model-name=%s", *dto.ModelName),
		fmt.Sprintf("--auto-tune=%s", autoTune),
		fmt.Sprintf("--request-id=%d", requestId),
	)

	if err != nil {
		s.Server.Queries.UpdateRequestStatus(context.Background(), db.UpdateRequestStatusParams{
			ID:     requestId,
			Status: 3,
			ErrorMessage:   sql.NullString{String: err.Error(), Valid: true},
			Description: sql.NullString{String: "Training failed.", Valid: true},
		})
		return
	}
	s.Server.Queries.UpdateRequestStatus(context.Background(), db.UpdateRequestStatusParams{
		ID:     requestId,
		Status: 2,
		Progress: 1.0,
		Description: sql.NullString{String: "Training completed.", Valid: true},
	})
}

type ModelMetadata struct {
	RMSE    float64 `json:"rmse"`
	R2      float64 `json:"r2"`
	MAE     float64 `json:"mae"`
	SavedAt string  `json:"saved_at"`
}

type ModelInfo struct {
	Name       string
	Metadata   *ModelMetadata
	Validation *ModelValidation
}

func (s *MLService) ListModels(c echo.Context) error {
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

func (s *MLService) ValidateModel(modelName string) (ModelValidation, error) {
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

func (s *MLService) DeleteModel(c echo.Context) error {
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

	// Log what we're about to delete for debugging
	fmt.Printf("Deleting model path: %s\n", modelPath)

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

func (s *MLService) ExportModel(c echo.Context) error {
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
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="model_`+dto.ModelName+`.zip"`)
	c.Response().WriteHeader(http.StatusOK)
	_, err = c.Response().Write(buf.Bytes())
	return err
}

type ScoringResult map[string] float64


func (s *MLService) ScoreProfilesCronjob() {
	logger.Info("Starting cron task [ScanAllGroups]")
	queries := s.Server.Queries
	ctx := context.Background()
	limit := s.Server.GetConfig(ctx, "ML_SCORING_PROFILE_LIMIT", "50")
	limitInt, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		logger.Errorf("invalid ML_SCORING_PROFILE_LIMIT: %v", err)
		return
	}

	modelName := s.Server.GetConfig(ctx, "ML_SCORING_MODEL_NAME", "No")
	if modelName == "No" {
		logger.Info("No model configured for scoring. Skipping...")
		return
	}

	profiles, err := queries.GetProfilesForScoring(ctx, int32(limitInt))
	if err != nil {
		logger.Errorf("failed to get profiles for scoring: %v", err)
		return
	}
	if len(profiles) == 0 {
		logger.Info("No profiles to score. Skipping...")
		return
	}

	profileIDs := make([]string, len(profiles))
	for i, p := range profiles {
		profileIDs[i] = strconv.Itoa(int(p))
	}
	profileIDsStr := strings.Join(profileIDs, ",")

	pythonService := PythonService{
		EnvName: os.Getenv("PYTHON_ENV_NAME"),
		Log:     false,
	}

	res, err := pythonService.RunScript("--task=predict",
		fmt.Sprintf("--targets=%s", profileIDsStr),
		fmt.Sprintf("--model-name=%s", modelName),
	)

	if err != nil {
		logger.Errorf("failed to run python script: %v", err)
		return
	}

	var resData ScoringResult
	
	if err := json.Unmarshal([]byte(res), &resData); err != nil {
		logger.Errorf("failed to unmarshal scoring result: %v", err)
		return
	}
	sem := async.GetSemaphore[db.UpdateModelScoreParams, bool](5)
	updateScore := func(params db.UpdateModelScoreParams) bool {
		err := queries.UpdateModelScore(ctx, params)
		if err != nil {
			panic(err)
		}
		return true
	}
	for id, score := range resData {
		strid, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			logger.Errorf("invalid profile ID from model: %v", err)
			continue
		}
		sem.Assign(updateScore, db.UpdateModelScoreParams{
			ID:         int32(strid),
			ModelScore: sql.NullFloat64{
				Float64: float64(score),
				Valid:   true,
			},
		})
	}
	_, errs := sem.Run()

	successCount := 0
	for _, e := range errs {
		if e != nil {
			logger.Errorf("failed to update model score: %v", e)
		} else {
			successCount++
		}
	}

	logger.Info(fmt.Sprintf("Scored %d profiles, %d successful", len(profiles), successCount))
}