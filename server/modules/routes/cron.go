package routes

import (
	"github.com/qxbao/asfpc/infras"
	crs "github.com/qxbao/asfpc/server/modules/routes/services/cron"
	"github.com/qxbao/asfpc/server/modules/cron"
)

func InitCronRoutes(s *infras.Server, cronService *cron.CronService) {
	service := crs.CronRoutingService{
		Server: s,
		Cron:   cronService,
	}

	e := s.Echo

	e.GET("/cron/list", service.ListJobs)
	e.POST("/cron/trigger", service.TriggerJob)
	e.POST("/cron/stop", service.StopJob)
	e.POST("/cron/resume", service.ResumeJob)
}