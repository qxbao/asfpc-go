package analysis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"encoding/csv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/generative"
	"github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/pkg/utils/prompt"

	"github.com/xuri/excelize/v2"
	"github.com/phpdave11/gofpdf"
)



type AnalysisRoutingService infras.RoutingService

func (as *AnalysisRoutingService) GetProfiles(c echo.Context) error {
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

func (as *AnalysisRoutingService) GetProfileStats(c echo.Context) error {
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

func (as *AnalysisRoutingService) AnalyzeProfileWithGemini(c echo.Context) error {
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

	promptService := prompt.PromptService{Server: as.Server}

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

func (as *AnalysisRoutingService) AddGeminiKey(c echo.Context) error {
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

func (as *AnalysisRoutingService) DeleteGeminiKey(c echo.Context) error {
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

func (as *AnalysisRoutingService) DeleteJunkProfiles(c echo.Context) error {
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

func (as *AnalysisRoutingService) ResetProfilesModelScore(c echo.Context) error {
	queries := as.Server.Queries
	err := queries.ResetProfilesModelScore(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "failed to reset profiles model score: " + err.Error(),
		})
	}
	return c.JSON(200, map[string]any{
		"data": "success",
	})
}

func (as *AnalysisRoutingService) ExportProfiles(c echo.Context) error {
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

	format := c.QueryParam("format")
	switch format {
	case "excel":
		// Excel export
		// github.com/xuri/excelize/v2
		importExcelize := func() error {
			f := excelize.NewFile()
			sheet := "Sheet1"
			f.NewSheet(sheet)
			// Header
			headers := []string{"FacebookID", "Name", "Bio", "Location", "Work", "Education", "RelationshipStatus", "CreatedAt", "UpdatedAt", "IsScanned", "Hometown", "Locale", "Gender", "Birthday", "Email", "Phone", "ProfileUrl", "IsAnalyzed", "GeminiScore"}
			for i, h := range headers {
				cell, _ := excelize.CoordinatesToCellName(i+1, 1)
				f.SetCellValue(sheet, cell, h)
			}
			// Data
			for r, p := range profiles {
				vals := []any{p.FacebookID, p.Name, p.Bio, p.Location, p.Work, p.Education, p.RelationshipStatus, p.CreatedAt, p.UpdatedAt, p.IsScanned, p.Hometown, p.Locale, p.Gender, p.Birthday, p.Email, p.Phone, p.ProfileUrl, p.IsAnalyzed, p.GeminiScore}
				for c, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
					f.SetCellValue(sheet, cell, v)
				}
			}
			c.Response().Header().Set(echo.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=data.xlsx")
			c.Response().WriteHeader(http.StatusOK)
			return f.Write(c.Response().Writer)
		}
		return importExcelize()
	case "csv":
		// CSV export
		c.Response().Header().Set(echo.HeaderContentType, "text/csv")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=data.csv")
		c.Response().WriteHeader(http.StatusOK)
		w := csv.NewWriter(c.Response().Writer)
		headers := []string{"FacebookID", "Name", "Bio", "Location", "Work", "Education", "RelationshipStatus", "CreatedAt", "UpdatedAt", "IsScanned", "Hometown", "Locale", "Gender", "Birthday", "Email", "Phone", "ProfileUrl", "IsAnalyzed", "GeminiScore"}
		w.Write(headers)
		for _, p := range profiles {
			vals := []string{
				fmt.Sprint(p.FacebookID), fmt.Sprint(p.Name), fmt.Sprint(p.Bio), fmt.Sprint(p.Location), fmt.Sprint(p.Work), fmt.Sprint(p.Education), fmt.Sprint(p.RelationshipStatus), fmt.Sprint(p.CreatedAt), fmt.Sprint(p.UpdatedAt), fmt.Sprint(p.IsScanned), fmt.Sprint(p.Hometown), fmt.Sprint(p.Locale), fmt.Sprint(p.Gender), fmt.Sprint(p.Birthday), fmt.Sprint(p.Email), fmt.Sprint(p.Phone), fmt.Sprint(p.ProfileUrl), fmt.Sprint(p.IsAnalyzed), fmt.Sprint(p.GeminiScore),
			}
			w.Write(vals)
		}
		w.Flush()
		return nil
	case "pdf":
		// PDF export
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)
		headers := []string{"FacebookID", "Name", "Bio", "Location", "Work", "Education", "RelationshipStatus", "CreatedAt", "UpdatedAt", "IsScanned", "Hometown", "Locale", "Gender", "Birthday", "Email", "Phone", "ProfileUrl", "IsAnalyzed", "GeminiScore"}
		for _, h := range headers {
			pdf.CellFormat(20, 10, h, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		for _, p := range profiles {
			vals := []string{
				fmt.Sprint(p.FacebookID), fmt.Sprint(p.Name), fmt.Sprint(p.Bio), fmt.Sprint(p.Location), fmt.Sprint(p.Work), fmt.Sprint(p.Education), fmt.Sprint(p.RelationshipStatus), fmt.Sprint(p.CreatedAt), fmt.Sprint(p.UpdatedAt), fmt.Sprint(p.IsScanned), fmt.Sprint(p.Hometown), fmt.Sprint(p.Locale), fmt.Sprint(p.Gender), fmt.Sprint(p.Birthday), fmt.Sprint(p.Email), fmt.Sprint(p.Phone), fmt.Sprint(p.ProfileUrl), fmt.Sprint(p.IsAnalyzed), fmt.Sprint(p.GeminiScore),
			}
			for _, v := range vals {
				pdf.CellFormat(20, 10, v, "1", 0, "C", false, 0, "")
			}
			pdf.Ln(-1)
		}
		c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=data.pdf")
		c.Response().WriteHeader(http.StatusOK)
		return pdf.Output(c.Response().Writer)
	default:
		// JSON export (mặc định)
		c.Response().Header().Set(echo.HeaderContentType, "application/json")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=data.json")
		enc := json.NewEncoder(c.Response().Writer)
		c.Response().WriteHeader(http.StatusOK)
		return enc.Encode(profiles)
	}
}

func (as *AnalysisRoutingService) GetGeminiKeys(c echo.Context) error {
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

func (as *AnalysisRoutingService) ImportProfiles(c echo.Context) error {
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
			logger.GetLogger("ARS").Errorf("Failed to import profile for Facebook ID %s: %v", profile.FacebookID, err)
		}
		err = as.Server.Queries.UpsertEmbeddedProfiles(c.Request().Context(), db.UpsertEmbeddedProfilesParams{
			Pid:       p.ID,
			Embedding: profile.Embedding,
		})
		if err != nil {
			logger.GetLogger("ARS").Errorf("Failed to upsert embedded profile for profile ID %d: %v", p.ID, err)
			continue
		}
		successCount++
	}
	return c.JSON(200, map[string]any{
		"data": successCount,
	})
}

func (as *AnalysisRoutingService) FindSimilarProfiles(c echo.Context) error {
	dto := new(infras.FindSimilarProfilesDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	if dto.TopK == nil {
		dto.TopK = new(int32)
		*dto.TopK = 10
	}

	if dto.ProfileID == nil {
		return c.JSON(400, map[string]any{
			"error": "profile_id is required",
		})
	}

	similarProfiles, err := as.Server.Queries.FindSimilarProfiles(c.Request().Context(), db.FindSimilarProfilesParams{
		Pid:      	*dto.ProfileID,
		Limit:      *dto.TopK,
	})

	if err != nil {
		return c.JSON(500, map[string]any{
			"error": "Failed to find similar profiles: " + err.Error(),
		})
	}

	logger.GetLogger("ARS").Infof("Found %d similar profiles for profile ID %d", len(similarProfiles), dto.ProfileID)

	return c.JSON(200, map[string]any{
		"data": similarProfiles,
	})
}