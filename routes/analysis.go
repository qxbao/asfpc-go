package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitAnalysisRoutes(s infras.Server) {
	service := services.AnalysisService{Server: s}
	e := s.Echo

	e.GET("/analysis/profile/list", service.GetProfiles)
	e.GET("/analysis/key/list", service.GetGeminiKeys)
	e.POST("/analysis/key/add", service.AddGeminiKey)
	e.POST("/analysis/profile/analyze", service.AnalyzeProfileWithGemini)
}