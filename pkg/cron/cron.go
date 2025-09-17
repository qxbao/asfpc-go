package cron

import (
	"fmt"
	"github.com/qxbao/asfpc/infras"
	"github.com/go-co-op/gocron/v2"
)

type CronService struct {
	Scheduler gocron.Scheduler
	Server    *infras.Server
}

type Task struct {
	Def gocron.JobDefinition
	Fn  gocron.Task
}

func (c *CronService) Setup() {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(fmt.Errorf("failed to create cron scheduler: %w", err))
	}
	for _, task := range CollectTasks(c.Server) {
		s.NewJob(
			task.Def,
			task.Fn,
		)
	}
	c.Scheduler = s
}

func (c *CronService) Start() {
	c.Scheduler.Start()
}