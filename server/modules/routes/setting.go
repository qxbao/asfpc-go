package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitSettingRoutes(s *infras.Server) {
	service := services.SettingService{Server: s}
	e := s.Echo

	e.GET("/setting/list", service.GetSettings)
	e.POST("/setting/update", service.UpdateSettings)
}