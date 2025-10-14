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

	// Get all categories
	categories, err := as.Server.Queries.GetCategories(ctx)
	if err != nil {
		as.Server.Queries.LogAction(ctx, db.LogActionParams{
			Action: "gemini_scoring_cronjob",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to get categories: %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: 0, Valid: false},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
		logger.Errorf("Failed to get categories: %v", err.Error())
		return
	}

	if len(categories) == 0 {
		logger.Info("No categories found. Skipping...")
		return
	}

	// Collect all profiles across all categories
	var profiles []db.GetProfilesAnalysisCronjobRow
	for _, category := range categories {
		categoryProfiles, err := as.Server.Queries.GetProfilesAnalysisCronjob(ctx, db.GetProfilesAnalysisCronjobParams{
			CategoryID: category.ID,
			Limit:      int32(geminiAPILimitInt),
		})
		if err != nil {
			logger.Errorf("Failed to get profiles for category %s: %v", category.Name, err)
			continue
		}
		profiles = append(profiles, categoryProfiles...)
	}

	if len(profiles) == 0 {
		logger.Info("No profiles to score across all categories. Exiting cronjob.")
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

	semaphore := async.GetSemaphore[infras.GeminiScoringTaskInput, bool](15)
	promptService := prompt.PromptService{Server: as.Server}

	for _, profile := range profiles {
		pr, err := promptService.GetPrompt(ctx, "gemini-preprocess-1", profile.CategoryID)
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
		businessDesc, err := promptService.GetPrompt(ctx, "business-description", profile.CategoryID)
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
		profilePromptContent := promptService.ReplacePrompt(pr.Content,
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
	queries := as.Server.Queries

	limitStr := as.Server.GetConfig(ctx, "GEMINI_EMBEDDING_LIMIT", "100")
	limit, err := strconv.ParseInt(limitStr, 10, 32)

	if err != nil || limit <= 0 {
		logger.Warn("Invalid GEMINI_EMBEDDING_LIMIT, using default %s", defaultEmbeddingLimit)
		limit = int64(defaultEmbeddingLimit)
	}

	// Get all categories
	categories, err := queries.GetCategories(ctx)
	if err != nil {
		logger.Errorf("Failed to get categories: %v", err)
		return
	}

	if len(categories) == 0 {
		logger.Info("No categories found. Skipping...")
		return
	}

	// Process each category
	for _, category := range categories {
		logger.Infof("Processing embedding for category: %s (ID: %d)", category.Name, category.ID)

		// Embedding model is hard-coded (BGEM3) in Python, no need to query from DB
		// Just proceed with embedding process
		embeddingModel := "BAAI/bge-m3" // Hard-coded embedding model

		profiles, err := queries.GetProfileIDForEmbedding(ctx, db.GetProfileIDForEmbeddingParams{
			CategoryID: category.ID,
			Limit:      int32(limit),
		})
		if err != nil {
			logger.Errorf("Failed to get profiles for embedding (category %s): %v", category.Name, err)
			continue
		}

		if len(profiles) == 0 {
			logger.Infof("No profiles to embed for category %s. Skipping...", category.Name)
			continue
		}

		pythonService := python.NewPythonService(os.Getenv("PYTHON_ENV_NAME"), false, true, nil)
		idStrs := make([]string, 0, len(profiles))
		for _, profileId := range profiles {
			idStrs = append(idStrs, fmt.Sprintf("%d", profileId))
		}
		idStr := strings.Join(idStrs, ",")
		output, err := pythonService.RunScript("--task=embed",
			fmt.Sprintf("--targets=%s", idStr),
			fmt.Sprintf("--embedding-model=%s", embeddingModel),
			fmt.Sprintf("--category-id=%d", category.ID),
		)

		if err != nil {
			logger.Errorf("Failed to run embedding script for category %s: %v", category.Name, err)
			continue
		}
		logger.Infof("Category %s embedding output: %s", category.Name, output)
	}

	logger.Info("Completed SelfEmbeddingCronjob for all categories")
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
