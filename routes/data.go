package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitDataRoutes(s infras.Server) {
	e := s.Echo
	services := services.DataService{Server: s}

	e.GET("/data/stats", services.GetDataStats)
}