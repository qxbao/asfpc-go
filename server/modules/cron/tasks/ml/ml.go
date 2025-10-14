package ml

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/async"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/pkg/utils/python"
)

type MLService struct {
	Server *infras.Server
}

type ScoringResult map[string]float64

var logger = lg.GetLogger("MachineLearningCronService")

func (s *MLService) ScoreProfilesCronjob() {
	logger.Info("Starting cron task [ScoreProfilesCronjob]...")
	queries := s.Server.Queries
	ctx := context.Background()
	limit := s.Server.GetConfig(ctx, "ML_SCORING_PROFILE_LIMIT", "50")
	limitInt, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		logger.Errorf("invalid ML_SCORING_PROFILE_LIMIT: %v", err)
		return
	}

	// Get all categories
	categories, err := queries.GetCategories(ctx)
	if err != nil {
		logger.Errorf("failed to get categories: %v", err)
		return
	}

	if len(categories) == 0 {
		logger.Info("No categories found. Skipping...")
		return
	}

	// Process each category
	for _, category := range categories {
		logger.Infof("Processing category: %s (ID: %d)", category.Name, category.ID)

		// Get model for this category from model table
		model, err := queries.GetModelByCategory(ctx, sql.NullInt32{Int32: category.ID, Valid: true})
		if err != nil {
			logger.Warnf("No model assigned to category %s. Skipping...", category.Name)
			continue
		}

		modelName := model.Name
		if modelName == "" {
			logger.Infof("No model configured for category %s. Skipping...", category.Name)
			continue
		}

		profiles, err := queries.GetProfilesForScoring(ctx, db.GetProfilesForScoringParams{
			CategoryID: category.ID,
			Limit:      int32(limitInt),
		})
		if err != nil {
			logger.Errorf("failed to get profiles for scoring (category %s): %v", category.Name, err)
			continue
		}

		if len(profiles) == 0 {
			logger.Infof("No profiles to score for category %s. Skipping...", category.Name)
			continue
		}

		profileIDs := make([]string, len(profiles))
		for i, p := range profiles {
			profileIDs[i] = strconv.Itoa(int(p))
		}
		profileIDsStr := strings.Join(profileIDs, ",")

		logger.Infof("Scoring %d profiles for category %s using model %s", len(profiles), category.Name, modelName)

		pythonService := python.NewPythonService(os.Getenv("PYTHON_ENV_NAME"), false, true, nil)

		res, err := pythonService.RunScript("--task=predict",
			fmt.Sprintf("--targets=%s", profileIDsStr),
			fmt.Sprintf("--model-name=%s", modelName),
			fmt.Sprintf("--category-id=%d", category.ID),
		)

		if err != nil {
			logger.Errorf("failed to run python script for category %s: %v", category.Name, err)
			continue
		}

		var resData ScoringResult

		if err := json.Unmarshal([]byte(res), &resData); err != nil {
			logger.Errorf("failed to unmarshal scoring result for category %s: %v", category.Name, err)
			logger.Info("Raw response:", res)
			continue
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
				ID: int32(strid),
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

		logger.Infof("Category %s: Scored %d profiles, %d successful", category.Name, len(profiles), successCount)
	} // End category loop

	logger.Info("Completed ScoreProfilesCronjob for all categories")
}
