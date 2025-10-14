package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/ml"
)

func InitMLRoutes(s *infras.Server) {
	service := ml.MLRoutingService{Server: s}
	e := s.Echo

	e.GET("/ml/list", service.ListModels)
	e.GET("/ml/export", service.ExportModel)
	e.POST("/ml/train", service.Train)
	e.DELETE("/ml/delete", service.DeleteModel)

	// ML Config routes
	e.GET("/ml/config/all", s.GetAllCategoryMLConfigs)
	e.GET("/ml/config/model/:category_id", s.GetMLModelConfig)
	e.POST("/ml/config/model", s.SetMLModelConfig)
	e.GET("/ml/config/embedding/:category_id", s.GetEmbeddingModelConfig)
	e.POST("/ml/config/embedding", s.SetEmbeddingModelConfig)
}
