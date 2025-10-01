package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

type Task struct {
	Def gocron.JobDefinition
	Fn  gocron.Task
}

func CollectTasks(s *infras.Server) []Task {
	return []Task{
		scanGroups(s),
		scanProfiles(s),
		geminiScoring(s),
		embedProfiles(s),
		scoreProfiles(s),
	}
}

func scanGroups(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			10 * time.Minute,
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
			10 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			scanService := &services.ScanService{
				Server: *server,
			}
			scanService.ScanAllProfiles()
		}, s),
	}
}

func geminiScoring(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			analysisService := &services.AnalysisService{
				Server: server,
			}
			analysisService.GeminiScoringCronjob()
		}, s),
	}
}

func embedProfiles(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			analysisService := &services.AnalysisService{
				Server: server,
			}
			analysisService.GeminiEmbeddingCronjob()
		}, s),
	}
}

func scoreProfiles(s *infras.Server) Task {
	return Task{
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			mlService := &services.MLService{
				Server: server,
			}
			mlService.ScoreProfilesCronjob()
		}, s),
	}
}