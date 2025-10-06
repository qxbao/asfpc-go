package analysis

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/async"
	"github.com/qxbao/asfpc/pkg/generative"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/pkg/utils/prompt"
	"github.com/qxbao/asfpc/pkg/utils/python"
)

type AnalysisService struct {
	Server *infras.Server
}

var logger = lg.GetLogger("AnalysisService")

func (as *AnalysisService) GeminiScoringCronjob() {
	logger.Info("Starting Gemini scoring cronjob")
	ctx := context.Background()
	defer ctx.Done()

	enableGemini := as.Server.GetConfig(ctx, "USE_GEMINI_ANALYSIS_BOOL", "TRUE")
	if strings.ToLower(enableGemini) != "true" {
		logger.Info("Gemini analysis is disabled. Exiting cronjob.")
		return
	}

	geminiAPILimit := as.Server.GetConfig(ctx, "GEMINI_API_LIMIT", "15")
	geminiAPILimitInt, err := strconv.ParseInt(geminiAPILimit, 10, 32)

	if err != nil || geminiAPILimitInt <= 0 {
		logger.Warn("Invalid GEMINI_API_LIMIT, using default 15")
		geminiAPILimitInt = 15
	}

	profiles, err := as.Server.Queries.GetProfilesAnalysisCronjob(ctx, int32(geminiAPILimitInt))

	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to get profiles: %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get profiles: %v", err.Error())
		return
	}

	apiKey, err := as.Server.Queries.GetGeminiKeyForUse(ctx)

	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to get gemini key: %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get profiles: %v", err.Error())
		return
	}

	generativeService := generative.GetGenerativeService(apiKey.ApiKey, "gemini-2.5-flash-lite")

	err = generativeService.Init()

	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to initialize generative service: %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get profiles: %v", err.Error())
		return
	}

	promptService := prompt.PromptService{Server: as.Server}

	prompt, err := promptService.GetPrompt(ctx, "gemini-preprocess-1")

	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to get prompt (gemini-preprocess-1): %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get profiles: %v", err.Error())
		return
	}

	businessDesc, err := promptService.GetPrompt(ctx, "business-description")

	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to get prompt (business-description): %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get profiles: %v", err.Error())
		return
	}

	semaphore := async.GetSemaphore[infras.GeminiScoringTaskInput, bool](15)

	for _, profile := range profiles {
		profilePromptContent := promptService.ReplacePrompt(prompt.Content,
			businessDesc.Content,
			profile.Name.String,
			profile.Location.String,
			profile.Work.String,
			profile.Bio.String,
			profile.Education.String,
			profile.RelationshipStatus.String,
			profile.Hometown.String,
			profile.Locale,
			profile.Gender.String,
			profile.Birthday.String,
		)
		semaphore.Assign(as.geminiScoringTask, infras.GeminiScoringTaskInput{
			Ctx:     ctx,
			Gs:      generativeService,
			Prompt:  profilePromptContent,
			Profile: &profile,
		})
	}

	_, errs := semaphore.Run()

	generativeService.SaveUsage(ctx, as.Server.Queries)
	count := 0

	for i, err := range errs {
		if err != nil {
			as.Server.Queries.LogAction(ctx, db.LogActionParams{
				Action: "gemini_scoring_cronjob",
				Description: sql.NullString{
					String: fmt.Sprintf("Failed to process profile %d: %v", profiles[i].ID, err.Error()),
					Valid:  true,
				},
				TargetID:  sql.NullInt32{Int32: profiles[i].ID, Valid: true},
				AccountID: sql.NullInt32{Int32: 0, Valid: false},
			})
			logger.Errorf("Failed to process profile %d: %v", profiles[i].ID, err.Error())
		} else {
			count++
		}
	}

	logger.Infof("Gemini scoring cronjob completed: %d/%d profiles processed successfully", count, len(profiles))
}

func (as *AnalysisService) geminiScoringTask(input infras.GeminiScoringTaskInput) bool {
	response, err := input.Gs.GenerateText(input.Prompt)
	if err != nil {
		panic(fmt.Errorf("failed to generate text: %v", err))
	}

	score, err := strconv.ParseFloat(response, 64)

	if err != nil {
		panic(fmt.Errorf("failed to parse score: %v", err))
	}

	_, err = as.Server.Queries.UpdateGeminiAnalysisProfile(input.Ctx, db.UpdateGeminiAnalysisProfileParams{
		ID:          input.Profile.ID,
		GeminiScore: sql.NullFloat64{Float64: score, Valid: true},
	})

	if err != nil {
		panic(fmt.Errorf("failed to update profile: %v", err))
	}

	return true
}

func (as *AnalysisService) GetGeminiKeys(c echo.Context) error {
	queries := as.Server.Queries
	keys, err := queries.GetGeminiKeys(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get gemini keys: " + err.Error(),
		})
	}

	count, err := queries.CountGeminiKeys(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to count gemini keys: " + err.Error(),
		})
	}

	if keys == nil {
		keys = make([]db.GeminiKey, 0)
	}
	return c.JSON(200, map[string]any{
		"total": count,
		"data":  keys,
	})
}

var defaultEmbeddingLimit int32 = 100

func (as *AnalysisService) SelfEmbeddingCronjob() {
	logger.Info("Starting Self-embedding cronjob")
	ctx := context.Background()
	defer ctx.Done()

	limitStr := as.Server.GetConfig(ctx, "GEMINI_EMBEDDING_LIMIT", "100")
	limit, err := strconv.ParseInt(limitStr, 10, 32)

	if err != nil || limit <= 0 {
		logger.Warn("Invalid GEMINI_EMBEDDING_LIMIT, using default %s", defaultEmbeddingLimit)
		limit = int64(defaultEmbeddingLimit)
	}

	profiles, err := as.Server.Queries.GetProfileIDForEmbedding(ctx, int32(limit))
	if err != nil {
		logger.Errorf("Failed to get gemini key: %v", err)
		return
	}

	if len(profiles) == 0 {
		logger.Info("No profiles to embed. Exiting cronjob.")
		return
	}

	ps := prompt.PromptService{Server: as.Server}
	_, err = ps.GetPrompt(ctx, "self-embedding")

	if err != nil {
		logger.Errorf("Failed to get prompt (self-embedding): %v", err)
		return
	}

	pythonService := python.PythonService{
		EnvName: os.Getenv("PYTHON_ENV_NAME"),
		Log:     false,
		Silent:  true,
	}
	idStrs := make([]string, 0, len(profiles))
	for _, profileId := range profiles {
		idStrs = append(idStrs, fmt.Sprintf("%d", profileId))
	}
	idStr := strings.Join(idStrs, ",")
	output, err := pythonService.RunScript("--task=embed",
		fmt.Sprintf("--targets=%s", idStr),
	)

	if err != nil {
		logger.Errorf("Failed to run embedding script: %v", err)
		return
	}
	logger.Info("Embedding script output: " + output)
}

func (as *AnalysisService) AddGeminiKey(c echo.Context) error {
	queries := as.Server.Queries
	dto := new(infras.AddGeminiKeyDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}
	key, err := queries.CreateGeminiKey(c.Request().Context(), dto.APIKey)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to add gemini key: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": key,
	})
}

func (as *AnalysisService) DeleteGeminiKey(c echo.Context) error {
	queries := as.Server.Queries
	dto := new(infras.DeleteGeminiKeyDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}
	err := queries.DeleteGeminiKey(c.Request().Context(), dto.KeyID)
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to delete gemini key: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": "success",
	})
}

func (as *AnalysisService) DeleteJunkProfiles(c echo.Context) error {
	queries := as.Server.Queries
	count, err := queries.DeleteJunkProfiles(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to delete junk profiles: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": count,
	})
}