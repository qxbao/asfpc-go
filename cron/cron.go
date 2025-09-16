package cron

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
	// "github.com/qxbao/asfpc/db"
)

type CronScheduler struct {
	S       gocron.Scheduler
	// queries *db.Queries
}

func (c *CronScheduler) Setup() {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(fmt.Errorf("failed to create cron scheduler: %w", err))
	}
	c.S = s
}

func (c *CronScheduler) Start() {
	c.S.Start()
}
