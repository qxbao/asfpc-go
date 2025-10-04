package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/analysis"
)

func InitAnalysisRoutes(s *infras.Server) {
	service := analysis.AnalysisRoutingService{Server: s}
	e := s.Echo

	e.GET("/analysis/profile/list", service.GetProfiles)
	e.GET("/analysis/profile/stats", service.GetProfileStats)
	e.GET("/analysis/key/list", service.GetGeminiKeys)
	e.GET("/analysis/profile/export", service.ExportProfiles)
	e.POST("/analysis/profile/import", service.ImportProfiles)
	e.POST("/analysis/key/add", service.AddGeminiKey)
	e.POST("/analysis/profile/analyze", service.AnalyzeProfileWithGemini)
	e.DELETE("/analysis/key/delete", service.DeleteGeminiKey)
	e.DELETE("/analysis/profile/delete_scores", service.ResetProfilesModelScore)
	e.DELETE("/analysis/profile/delete_junk", service.DeleteJunkProfiles)
}