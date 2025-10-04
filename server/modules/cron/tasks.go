package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/qxbao/asfpc/infras"
	analysis "github.com/qxbao/asfpc/server/modules/cron/tasks/analysis"
	"github.com/qxbao/asfpc/server/modules/cron/tasks/ml"
	scan "github.com/qxbao/asfpc/server/modules/cron/tasks/scan"
)

type Task struct {
	Name string
	Def  gocron.JobDefinition
	Fn   gocron.Task
}

// Any new tasks must be added here
var TaskFuncs = map[string]func(s *infras.Server, name string) Task{
	"Scan Groups":  scanGroups,
	"Scan Profiles": scanProfiles,
	"Gemini Scoring": geminiScoring,
	"Embed Profiles": embedProfiles,
	"Score Profiles": scoreProfiles,
}

func scanGroups(s *infras.Server, name string) Task {
	return Task{
		Name: name,
		Def: gocron.DurationJob(
			10 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			scanService := &scan.ScanService{
				Server: *server,
			}
			scanService.ScanAllGroups()
		}, s),
	}
}

func scanProfiles(s *infras.Server, name string) Task {
	return Task{
		Name: name,
		Def: gocron.DurationJob(
			10 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			scanService := &scan.ScanService{
				Server: *server,
			}
			scanService.ScanAllProfiles()
		}, s),
	}
}

func geminiScoring(s *infras.Server, name string) Task {
	return Task{
		Name: name,
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			analysisService := &analysis.AnalysisService{
				Server: server,
			}
			analysisService.GeminiScoringCronjob()
		}, s),
	}
}

func embedProfiles(s *infras.Server, name string) Task {
	return Task{
		Name: name,
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			analysisService := &analysis.AnalysisService{
				Server: server,
			}
			analysisService.SelfEmbeddingCronjob()
		}, s),
	}
}

func scoreProfiles(s *infras.Server, name string) Task {
	return Task{
		Name: name,
		Def: gocron.DurationJob(
			1 * time.Minute,
		),
		Fn: gocron.NewTask(func(server *infras.Server) {
			mlService := &ml.MLService{
				Server: server,
			}
			mlService.ScoreProfilesCronjob()
		}, s),
	}
}
