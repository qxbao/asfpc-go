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

type ScoringResult map[string] float64

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

	pythonService := python.PythonService{
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

	logger.Info(fmt.Sprintf("Scored %d profiles, %d successful", len(profiles), successCount))
}