package analysis

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/logger"
)

type AnalysisRoutingService infras.RoutingService

// Helper function for consistent error responses
func errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]any{
		"error": message,
	})
}

// Helper function for consistent success responses
func successResponse(c echo.Context, data any) error {
	return c.JSON(200, map[string]any{
		"data": data,
	})
}

func (as *AnalysisRoutingService) GetProfiles(c echo.Context) error {
	log := logger.GetLogger("GetProfiles")
	log.Info("Starting GetProfiles request")

	queries := as.Server.Queries
	dto := new(infras.QueryWithPageDTO)

	if err := c.Bind(dto); err != nil {
		log.Errorf("Failed to bind request: %v", err)
		return errorResponse(c, 400, "Invalid request body")
	}

	if dto.Page == nil {
		dto.Page = new(int32)
		*dto.Page = 0
	}

	if dto.Limit == nil {
		dto.Limit = new(int32)
		*dto.Limit = 10
	}

	// Get category_id from query params if provided
	categoryIDStr := c.QueryParam("category_id")
	var categoryID *int32
	if categoryIDStr != "" {
		if catID, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = new(int32)
			*categoryID = int32(catID)
			log.Infof("Filtering by category ID: %d", *categoryID)
		}
	}

	log.Infof("Fetching profiles with limit=%d, offset=%d", *dto.Limit, *dto.Page**dto.Limit)

	var count int64
	var err error

	if categoryID != nil {
		// Get profiles filtered by category and count
		categoryProfiles, err := queries.GetProfilesAnalysisPageByCategory(c.Request().Context(), db.GetProfilesAnalysisPageByCategoryParams{
			CategoryID: *categoryID,
			Limit:      *dto.Limit,
			Offset:     *dto.Page * *dto.Limit,
		})
		if err != nil {
			log.Errorf("Failed to get profiles by category from DB: %v", err)
			return errorResponse(c, 500, "failed to get profiles: "+err.Error())
		}

		count, err = queries.CountProfilesInCategory(c.Request().Context(), *categoryID)
		if err != nil {
			log.Errorf("Failed to count profiles in category: %v", err)
			return errorResponse(c, 500, "failed to count profiles: "+err.Error())
		}

		log.Infof("Retrieved %d profiles from database", len(categoryProfiles))
		log.Infof("Total profile count: %d", count)

		if categoryProfiles == nil {
			categoryProfiles = make([]db.GetProfilesAnalysisPageByCategoryRow, 0)
		}

		log.Info("Returning response")
		return c.JSON(200, map[string]any{
			"total": count,
			"data":  categoryProfiles,
		})
	}

	// Get all profiles (original behavior)
	profiles, err := queries.GetProfilesAnalysisPage(c.Request().Context(), db.GetProfilesAnalysisPageParams{
		Limit:  *dto.Limit,
		Offset: *dto.Page * *dto.Limit,
	})
	if err != nil {
		log.Errorf("Failed to get profiles from DB: %v", err)
		return errorResponse(c, 500, "failed to get profiles: "+err.Error())
	}

	count, err = queries.CountProfiles(c.Request().Context())
	if err != nil {
		log.Errorf("Failed to count profiles: %v", err)
		return errorResponse(c, 500, "failed to count profiles: "+err.Error())
	}

	log.Infof("Retrieved %d profiles from database", len(profiles))
	log.Infof("Total profile count: %d", count)

	if profiles == nil {
		profiles = make([]db.GetProfilesAnalysisPageRow, 0)
	}

	log.Info("Returning response")

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
	
	// Get category_id from query parameter
	categoryIDStr := c.QueryParam("category_id")
	if categoryIDStr == "" {
		return c.JSON(400, map[string]any{
			"error": "category_id query parameter is required",
		})
	}
	
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 32)
	if err != nil {
		return c.JSON(400, map[string]any{
			"error": "invalid category_id: " + err.Error(),
		})
	}
	
	profiles, err := queries.GetProfilesForExport(c.Request().Context(), int32(categoryID))
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
		})
		if err != nil {
			logger.GetLogger("ARS").Errorf("Failed to import profile for Facebook ID %s: %v", profile.FacebookID, err)
		}
		err = as.Server.Queries.UpsertEmbeddedProfiles(c.Request().Context(), db.UpsertEmbeddedProfilesParams{
			Pid:       p.ID,
			Cid:       profile.CategoryID,
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

	if dto.CategoryID == nil {
		return c.JSON(400, map[string]any{
			"error": "category_id is required",
		})
	}

	similarProfiles, err := as.Server.Queries.FindSimilarProfiles(c.Request().Context(), db.FindSimilarProfilesParams{
		Pid:   *dto.ProfileID,
		Limit: *dto.TopK,
		Cid:   *dto.CategoryID,
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

func (as *AnalysisRoutingService) AddAllProfilesToCategory(c echo.Context) error {
	log := logger.GetLogger("AddAllProfilesToCategory")

	dto := new(infras.AddAllProfilesToCategoryDTO)
	if err := c.Bind(dto); err != nil {
		log.Errorf("Failed to bind request: %v", err)
		return c.JSON(400, map[string]any{
			"error": "Invalid request body",
		})
	}

	if dto.CategoryID == nil {
		return c.JSON(400, map[string]any{
			"error": "category_id is required",
		})
	}

	// Check if category exists
	_, err := as.Server.Queries.GetCategoryByID(c.Request().Context(), *dto.CategoryID)
	if err != nil {
		log.Errorf("Category not found: %v", err)
		return c.JSON(404, map[string]any{
			"error": "Category not found",
		})
	}

	// Add all profiles to the category
	rowsAffected, err := as.Server.Queries.AddAllProfilesToCategory(c.Request().Context(), *dto.CategoryID)
	if err != nil {
		log.Errorf("Failed to add profiles to category: %v", err)
		return c.JSON(500, map[string]any{
			"error": "Failed to add profiles to category: " + err.Error(),
		})
	}

	log.Infof("Successfully added %d profiles to category ID %d", rowsAffected, *dto.CategoryID)

	return c.JSON(200, map[string]any{
		"data": rowsAffected,
	})
}
