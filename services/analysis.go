package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/async"
	"github.com/qxbao/asfpc/pkg/generative"
	lg "github.com/qxbao/asfpc/pkg/logger"
)

type GeminiScoringTaskInput struct {
	ctx     context.Context
	gs      *generative.GenerativeService
	prompt  string
	profile *db.GetProfilesAnalysisCronjobRow
}

type GeminiEmbeddingTaskInput struct {
	ctx     context.Context
	gs      *generative.GenerativeService
	prompt  string
	profile *db.UserProfile
}

type AnalysisService struct {
	Server infras.Server
}

var anlLoggerName = "AnalysisService"
var anlLogger = lg.GetLogger(&anlLoggerName)

func (as *AnalysisService) GetProfiles(c echo.Context) error {
	queries := as.Server.Queries
	dto := new(infras.QueryWithPageDTO)

	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	if dto.Page == nil {
		dto.Page = new(int32)
		*dto.Page = 0
	}

	if dto.Limit == nil {
		dto.Limit = new(int32)
		*dto.Limit = 10
	}

	profiles, err := queries.GetProfilesAnalysisPage(c.Request().Context(), db.GetProfilesAnalysisPageParams{
		Limit:  *dto.Limit,
		Offset: *dto.Page * *dto.Limit,
	})

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get profiles: " + err.Error(),
		})
	}

	count, err := queries.CountProfiles(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to count profiles: " + err.Error(),
		})
	}

	if profiles == nil {
		profiles = make([]db.GetProfilesAnalysisPageRow, 0)
	}

	return c.JSON(200, map[string]any{
		"total": count,
		"data":  profiles,
	})
}

func (as *AnalysisService) GetProfileStats(c echo.Context) error {
	queries := as.Server.Queries
	stats, err := queries.GetProfileStats(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get profile stats: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": stats,
	})
}

func (as *AnalysisService) AnalyzeProfileWithGemini(c echo.Context) error {
	dto := new(infras.AnalyzeProfileRequest)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	profile, err := as.Server.Queries.GetProfileById(c.Request().Context(), dto.ProfileID)

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get profile: " + err.Error(),
		})
	}

	apiKey, err := as.Server.Queries.GetGeminiKeyForUse(c.Request().Context())

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get gemini key: " + err.Error(),
		})
	}

	generativeService := generative.GetGenerativeService(apiKey.ApiKey, "gemini-2.5-flash-lite")

	err = generativeService.Init()

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to initialize generative service: " + err.Error(),
		})
	}

	promptService := PromptService{Server: as.Server}

	prompt, err := promptService.GetPrompt(c.Request().Context(), "gemini-preprocess-1")

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get prompt (gemini-preprocess-1): " + err.Error(),
		})
	}

	businessDesc, err := promptService.GetPrompt(c.Request().Context(), "business-description")

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get prompt (business-description): " + err.Error(),
		})
	}

	promptContent := prompt.Content

	promptContent = promptService.ReplacePrompt(promptContent,
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

	response, err := generativeService.GenerateText(promptContent)

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to generate text: " + err.Error(),
		})
	}

	err = generativeService.SaveUsage(c.Request().Context(), as.Server.Queries)

	if err != nil {
		as.Server.Queries.LogAction(c.Request().Context(), db.LogActionParams{
			Action: "profile_gemini_analysis",
			Description: sql.NullString{
				String: fmt.Sprintf("Failed to save usage: %v", err.Error()),
				Valid:  true,
			},
			TargetID:  sql.NullInt32{Int32: profile.ID, Valid: true},
			AccountID: sql.NullInt32{Int32: 0, Valid: false},
		})
	}

	score, err := strconv.ParseFloat(response, 64)

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to parse score: " + err.Error(),
		})
	}

	updatedProfile, err := as.Server.Queries.UpdateGeminiAnalysisProfile(c.Request().Context(), db.UpdateGeminiAnalysisProfileParams{
		ID:          profile.ID,
		GeminiScore: sql.NullFloat64{Float64: score, Valid: true},
	})

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to update profile: " + err.Error(),
		})
	}

	return c.JSON(200, map[string]any{
		"data": updatedProfile.Float64,
	})
}

func (as *AnalysisService) GeminiScoringCronjob() {
	anlLogger.Info("Starting Gemini scoring cronjob")
	ctx := context.Background()
	defer ctx.Done()

	geminiAPILimit := as.Server.GetConfig("GEMINI_API_LIMIT", "15")
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
		anlLogger.Error("Failed to get profiles: %v", err.Error())
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
		anlLogger.Error("Failed to get profiles: %v", err.Error())
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
		anlLogger.Error("Failed to get profiles: %v", err.Error())
		return
	}

	promptService := PromptService{Server: as.Server}

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
		anlLogger.Error("Failed to get profiles: %v", err.Error())
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
		anlLogger.Error("Failed to get profiles: %v", err.Error())
		return
	}

	semaphore := async.GetSemaphore[GeminiScoringTaskInput, bool](15)

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
		semaphore.Assign(as.geminiScoringTask, GeminiScoringTaskInput{
			ctx:     ctx,
			gs:      generativeService,
			prompt:  profilePromptContent,
			profile: &profile,
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
			anlLogger.Error("Failed to process profile %d: %v", profiles[i].ID, err.Error())
		} else {
			count++
		}
	}

	anlLogger.Infof("Gemini scoring cronjob completed: %d/%d profiles processed successfully", count, len(profiles))
}

func (as *AnalysisService) geminiScoringTask(input GeminiScoringTaskInput) bool {
	response, err := input.gs.GenerateText(input.prompt)

	if err != nil {
		panic(fmt.Errorf("failed to generate text: %v", err))
	}

	score, err := strconv.ParseFloat(response, 64)

	if err != nil {
		panic(fmt.Errorf("failed to parse score: %v", err))
	}

	_, err = as.Server.Queries.UpdateGeminiAnalysisProfile(input.ctx, db.UpdateGeminiAnalysisProfileParams{
		ID:          input.profile.ID,
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

func (as *AnalysisService) GeminiEmbeddingCronjob() {
	logger.Info("Starting Gemini embedding cronjob")
	ctx := context.Background()
	defer ctx.Done()

	limitStr := as.Server.GetConfig("GEMINI_EMBEDDING_LIMIT", "100")
	limit, err := strconv.ParseInt(limitStr, 10, 32)

	if err != nil || limit <= 0 {
		logger.Warn("Invalid GEMINI_EMBEDDING_LIMIT, using default %s", defaultEmbeddingLimit)
		limit = int64(defaultEmbeddingLimit)
	}

	profiles, err := as.Server.Queries.GetProfileForEmbedding(ctx, int32(limit))
	if err != nil {
		logger.Error("Failed to get profiles for embedding: %v", err)
		return
	}
	apiKey, err := as.Server.Queries.GetGeminiKeyForUse(ctx)

	if err != nil {
		logger.Error("Failed to get gemini key: %v", err)
		return
	}
	generativeService := generative.GetGenerativeService(apiKey.ApiKey, "gemini-embedding-001")
	err = generativeService.Init()

	if err != nil {
		logger.Error("Failed to initialize generative service: %v", err)
		return
	}

	ps := PromptService{Server: as.Server}
	prompt, err := ps.GetPrompt(ctx, "gemini-embedding")

	if err != nil {
		logger.Error("Failed to get prompt (gemini-embedding): %v", err)
		return
	}

	semaphore := async.GetSemaphore[GeminiEmbeddingTaskInput, bool](5)

	for _, profile := range profiles {
		profilePromptContent := ps.ReplacePrompt(prompt.Content,
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
		semaphore.Assign(as.geminiEmbeddingTask, GeminiEmbeddingTaskInput{
			ctx:     ctx,
			profile: &profile,
			prompt:  profilePromptContent,
			gs:      generativeService,
		})
	}
	_, errs := semaphore.Run()
	generativeService.SaveUsage(ctx, as.Server.Queries)
	count := 0
	for i, err := range errs {
		if err != nil {
			logger.Error("Failed to process profile %d: %v", profiles[i].ID, err.Error())
		} else {
			count++
		}
	}
	logger.Infof("Gemini embedding cronjob completed: %d/%d profiles processed successfully", count, len(profiles))
}

func (as *AnalysisService) geminiEmbeddingTask(input GeminiEmbeddingTaskInput) bool {
	response, err := input.gs.GenerateEmbedding(input.prompt)

	if err != nil {
		panic(fmt.Errorf("failed to generate embedding: %v", err))
	}

	_, err = as.Server.Queries.CreateEmbeddedProfile(input.ctx, db.CreateEmbeddedProfileParams{
		Pid:       input.profile.ID,
		Embedding: response,
	})

	if err != nil {
		panic(fmt.Errorf("failed to create embedded profile: %v", err))
	}

	return true
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

func (as *AnalysisService) ExportProfiles(c echo.Context) error {
	queries := as.Server.Queries
	profiles, err := queries.GetProfilesForExport(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to get profiles: " + err.Error(),
		})
	}
	if profiles == nil {
		profiles = make([]db.GetProfilesForExportRow, 0)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=data.json",
	)

	enc := json.NewEncoder(c.Response().Writer)
	c.Response().WriteHeader(http.StatusOK)
	return enc.Encode(profiles)
}

func (as *AnalysisService) ImportProfiles(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid file upload: " + err.Error(),
		})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(400, map[string]any{
			"error": "Failed to open file: " + err.Error(),
		})
	}
	defer src.Close()
	
	var profiles []db.GetProfilesForExportRow

	if err := json.NewDecoder(src).Decode(&profiles); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}
	successCount := 0
	for _, profile := range profiles {
		p, err := as.Server.Queries.ImportProfile(c.Request().Context(), db.ImportProfileParams{
			FacebookID:         profile.FacebookID,
			Name:               profile.Name,
			Bio:                profile.Bio,
			Location:           profile.Location,
			Work:               profile.Work,
			Education:          profile.Education,
			RelationshipStatus: profile.RelationshipStatus,
			CreatedAt:          profile.CreatedAt,
			UpdatedAt:          profile.UpdatedAt,
			IsScanned:          profile.IsScanned,
			Hometown:           profile.Hometown,
			Locale:             profile.Locale,
			Gender:             profile.Gender,
			Birthday:           profile.Birthday,
			Email:              profile.Email,
			Phone:              profile.Phone,
			ProfileUrl:         profile.ProfileUrl,
			IsAnalyzed:         profile.IsAnalyzed,
			GeminiScore:        profile.GeminiScore,
		})
		if err != nil {
			continue
		}
		_, err = as.Server.Queries.CreateEmbeddedProfile(c.Request().Context(), db.CreateEmbeddedProfileParams{
			Pid:       p.ID,
			Embedding: profile.Embedding,
		})

		if err != nil {
			successCount++
		}
	}
	return c.JSON(200, map[string]any{
		"data": successCount,
	})
}
