package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/model"
)

func InitModelRoutes(s *infras.Server) {
	e := s.Echo
	services := model.ModelRoutingService{
		Server: s,
	}

	e.GET("/model/list", services.GetModels)
	e.POST("/model/create", services.CreateModel)
	e.PUT("/model/update", services.UpdateModel)
	e.DELETE("/model/delete/:id", services.DeleteModel)
	e.POST("/model/assign", services.AssignModelToCategory)
}
