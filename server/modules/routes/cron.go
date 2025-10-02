package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/cron"
	"github.com/qxbao/asfpc/services"
)

func InitCronRoutes(s *infras.Server, cronService *cron.CronService) {
	service := services.CronService{
		Server: s,
		Cron:   cronService,
	}

	e := s.Echo

	e.GET("/cron/list", service.ListJobs)
}