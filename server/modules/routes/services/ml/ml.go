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

	if dto.CategoryID != nil {
		args = append(args, fmt.Sprintf("--category-id=%d", *dto.CategoryID))
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

	ctx := context.Background()
	categoryID := sql.NullInt32{Valid: false}
	if dto.CategoryID != nil {
		categoryID = sql.NullInt32{Int32: *dto.CategoryID, Valid: true}
	}

	description := fmt.Sprintf("Model trained on %s", time.Now().Format("2006-01-02 15:04:05"))
	_, err = s.Server.Queries.CreateModel(ctx, db.CreateModelParams{
		Name:        *dto.ModelName,
		Description: sql.NullString{String: description, Valid: true},
		CategoryID:  categoryID,
	})
	if err != nil {
		logger.Errorf("Failed to create model row in database for %s: %v", *dto.ModelName, err)
	} else {
		logger.Infof("Model row created in database for %s", *dto.ModelName)
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
	ID          int                `json:"id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	CategoryID  int32              `json:"category_id,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
	Metadata    *ModelMetadata     `json:"metadata,omitempty"`
	Validation  *ModelValidation   `json:"validation,omitempty"`
}

// SyncModelsWithDatabase synchronizes models between filesystem and database
// Called during server startup to ensure consistency
func (s *MLRoutingService) SyncModelsWithDatabase(ctx context.Context) error {
	logger.Info("Starting model sync between filesystem and database...")
	
	modelsDir := path.Join("python", "resources", "models")
	exc, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	modelsPath := path.Join(path.Dir(exc), modelsDir)
	if err := os.MkdirAll(modelsPath, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Get all models from database
	dbModels, err := s.Server.Queries.GetModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get models from database: %w", err)
	}

	// Create maps for easy lookup
	dbModelMap := make(map[string]db.Model)
	for _, model := range dbModels {
		dbModelMap[model.Name] = model
	}

	// Get all folders in models directory
	folders, err := os.ReadDir(modelsPath)
	if err != nil {
		return fmt.Errorf("failed to read models directory: %w", err)
	}

	folderMap := make(map[string]bool)
	addedCount := 0
	
	// Check folders and add missing models to database
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		
		folderName := folder.Name()
		folderMap[folderName] = true
		
		// Validate model folder
		validation, err := s.ValidateModel(folderName)
		if err != nil {
			logger.Warnf("Failed to validate model %s: %v", folderName, err)
			continue
		}
		
		// Only sync valid models
		if !validation.IsValid {
			logger.Infof("Skipping invalid model folder: %s", folderName)
			continue
		}
		
		// Check if model exists in database
		if _, exists := dbModelMap[folderName]; !exists {
			// Create model in database
			description := fmt.Sprintf("Auto-synced model from filesystem on %s", time.Now().Format("2006-01-02 15:04:05"))
			_, err := s.Server.Queries.CreateModel(ctx, db.CreateModelParams{
				Name:        folderName,
				Description: sql.NullString{String: description, Valid: true},
				CategoryID:  sql.NullInt32{Valid: false},
			})
			if err != nil {
				logger.Errorf("Failed to create model %s in database: %v", folderName, err)
			} else {
				logger.Infof("✓ Added model to database: %s", folderName)
				addedCount++
			}
		}
	}

	// Check database models and delete orphaned entries
	deletedCount := 0
	for _, dbModel := range dbModels {
		if !folderMap[dbModel.Name] {
			// Model exists in database but not in filesystem, delete it
			if err := s.Server.Queries.DeleteModel(ctx, dbModel.ID); err != nil {
				logger.Errorf("Failed to delete orphaned model %s from database: %v", dbModel.Name, err)
			} else {
				logger.Infof("✗ Removed orphaned model from database: %s (ID: %d)", dbModel.Name, dbModel.ID)
				deletedCount++
			}
		}
	}

	logger.Infof("Model sync completed: %d added, %d removed", addedCount, deletedCount)
	return nil
}

// SyncModels is the HTTP handler for manual model sync
func (s *MLRoutingService) SyncModels(c echo.Context) error {
	ctx := c.Request().Context()
	
	if err := s.SyncModelsWithDatabase(ctx); err != nil {
		return c.JSON(500, map[string]any{
			"error": fmt.Sprintf("Sync failed: %v", err),
		})
	}
	
	return c.JSON(200, map[string]any{
		"message": "Model sync completed successfully",
	})
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

	// Get all models from database
	dbModels, err := s.Server.Queries.GetModels(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to get models from database: " + err.Error(),
		})
	}

	// Create a map for quick lookup
	dbModelMap := make(map[string]db.Model)
	for _, model := range dbModels {
		dbModelMap[model.Name] = model
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
		
		modelInfo := ModelInfo{Name: folder.Name()}
		
		// Get info from database if exists
		if dbModel, exists := dbModelMap[folder.Name()]; exists {
			modelInfo.ID = int(dbModel.ID)
			modelInfo.Description = dbModel.Description.String
			if dbModel.CategoryID.Valid {
				modelInfo.CategoryID = dbModel.CategoryID.Int32
			}
			modelInfo.CreatedAt = dbModel.CreatedAt
		}
		
		validation, err := s.ValidateModel(folder.Name())
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": "Failed to validate model " + folder.Name() + ": " + err.Error(),
			})
		}
		modelInfo.Validation = &validation
		
		if !validation.IsValid {
			models = append(models, modelInfo)
			continue
		}
		
		// Read metadata.json from folder
		metadataPath := path.Join(path.Dir(exc), modelsDir, folder.Name(), "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			data, err := os.ReadFile(metadataPath)
			var metadata ModelMetadata
			if err == nil {
				if err := json.Unmarshal(data, &metadata); err == nil {
					modelInfo.Metadata = &metadata
				}
			} else {
				metadata = ModelMetadata{}
				modelInfo.Metadata = &metadata
			}
		}
		
		models = append(models, modelInfo)
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

	// Delete folder
	if err := os.RemoveAll(modelPath); err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to delete model folder: " + err.Error(),
		})
	}

	// Delete from database if exists
	ctx := c.Request().Context()
	dbModels, err := s.Server.Queries.GetModels(ctx)
	if err == nil {
		for _, dbModel := range dbModels {
			if dbModel.Name == dto.ModelName {
				if err := s.Server.Queries.DeleteModel(ctx, dbModel.ID); err != nil {
					logger.Errorf("Failed to delete model from database: %v", err)
					// Don't fail the request if DB delete fails, just log it
				} else {
					logger.Infof("Model %s deleted from database (ID: %d)", dto.ModelName, dbModel.ID)
				}
				break
			}
		}
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
