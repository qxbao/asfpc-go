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
}