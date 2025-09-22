package services

import (
	"context"
	"database/sql"
	"fmt"
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

	promptContent = promptService.ReplacePrompt(&promptContent,
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

	profiles, err := as.Server.Queries.GetProfilesAnalysisCronjob(ctx, db.GetProfilesAnalysisCronjobParams{
		Limit:  15,
		Offset: 0,
	})

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
		profilePromptContent := promptService.ReplacePrompt(&prompt.Content,
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

	succs, errs := semaphore.Run()

	generativeService.SaveUsage(ctx, as.Server.Queries)

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
		}
	}

	count := 0
	for _, succ := range succs {
		if succ {
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
