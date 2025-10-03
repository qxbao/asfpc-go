package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/setting"
)

func InitSettingRoutes(s *infras.Server) {
	service := setting.SettingRoutingService{Server: s}
	e := s.Echo

	e.GET("/setting/list", service.GetSettings)
	e.POST("/setting/update", service.UpdateSettings)
}