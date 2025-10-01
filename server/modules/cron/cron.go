package cron

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CronService struct {
	Scheduler gocron.Scheduler
	Logger    *zap.SugaredLogger
	Server    *server.Server
}

func NewCronService(server *server.Server) (*CronService, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}

	logger := logger.GetLogger("CronModule")

	cronService := &CronService{
		Scheduler: s,
		Logger:    logger,
		Server:    server,
	}

	logger.Info("Registering cron tasks...")
	if err := cronService.AddTask(server); err != nil {
		return nil, fmt.Errorf("failed to add cron tasks: %w", err)
	}

	return cronService, nil
}

func (c *CronService) AddTask(s *server.Server) error {
	tasks := CollectTasks(s.Server)
	c.Logger.Infof("Found %d tasks to register", len(tasks))

	for i, task := range tasks {
		_, err := c.Scheduler.NewJob(
			task.Def,
			task.Fn,
		)
		if err != nil {
			c.Logger.Errorf("Failed to register cron job %d: %v", i, err)
			return fmt.Errorf("failed to register cron job %d: %w", i, err)
		}
		c.Logger.Infof("Successfully registered cron job %d", i+1)
	}

	c.Logger.Infof("Successfully registered all %d cron jobs", len(tasks))
	return nil
}

func (c *CronService) RegisterHooks(lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			c.Logger.Info("Starting cron scheduler...")
			c.Scheduler.Start()
			c.Logger.Info("Cron scheduler started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.Logger.Info("Stopping cron scheduler...")
			if err := c.Scheduler.Shutdown(); err != nil {
				c.Logger.Errorf("Failed to shutdown cron scheduler: %v", err)
				return err
			}
			c.Logger.Info("Cron scheduler stopped successfully")
			return nil
		},
	})
}

var CronModule = fx.Module("CronModule",
	fx.Provide(NewCronService),
	fx.Invoke(func(c *CronService, lc fx.Lifecycle) {
		c.RegisterHooks(lc)
	}),
)
