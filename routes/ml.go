package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitMLRoutes(s infras.Server) {
	service := services.MLService{Server: s}
	e := s.Echo

	e.GET("/ml/list", service.ListModels)
	e.POST("/ml/train", service.Train)
}