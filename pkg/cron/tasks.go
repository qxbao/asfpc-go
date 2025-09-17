package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func CollectTasks(s *infras.Server) []Task {
	return []Task{
		scanGroups(s),
		scanProfiles(s),
	}
}

func scanGroups(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			30 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			scanService := &services.ScanService{
				Server: *server,
			}
			scanService.ScanAllGroups()
		}, s),
	}
}

func scanProfiles(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			30 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			scanService := &services.ScanService{
				Server: *server,
			}
			scanService.ScanAllProfiles()
		}, s),
	}
}