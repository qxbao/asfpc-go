package cron

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/logger"
	"go.uber.org/fx"
)

type CronService struct {
	infras.CronService
}

func NewCronService(server *infras.Server) (*CronService, error) {
	s, err := gocron.NewScheduler()

	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}

	logger := logger.GetLogger("CronModule")

	cronService := &CronService{
		CronService: infras.CronService{
			Scheduler: s,
			Logger:    logger,
			Server:    server,
			Jobs:      make(map[string]*infras.JobDetail),
		},
	}

	logger.Info("Registering cron tasks...")
	cronService.AddTask(server)
	return cronService, nil
}

func NewRawCronService(cronService *CronService) *infras.CronService {
	return &cronService.CronService
}

func (c *CronService) AddTask(s *infras.Server) {
	c.Logger.Infof("Found %d tasks to register", len(TaskFuncs))

	successCount := 0

	for name, taskFunc := range TaskFuncs {
		task := taskFunc(s, name)
		t, err := c.Scheduler.NewJob(
			task.Def,
			task.Fn,
		)
		if err != nil {
			c.Logger.Errorf("Failed to register cron job %s: %v", name, err)
			continue
		}
		c.Jobs[name] = &infras.JobDetail{
			Name: name,
			Job:  &t,
		}
		successCount++
		c.Logger.Infof("Successfully registered cron job %s", name)
	}

	c.Logger.Infof("Successfully registered all %d/%d cron jobs", successCount, len(TaskFuncs))
}

func (c *CronService) ListJobs() (map[string]*infras.JobStatus, error) {
	var jobMap = make(map[string]*infras.JobStatus)
	for name := range TaskFuncs {
		jobRef, exists := c.Jobs[name]
		if !exists || jobRef.Job == nil {
			jobMap[name] = &infras.JobStatus{
				Name:      name,
				NextRun:   nil,
				LastRun:   nil,
				IsRunning: false,
				Tags:      nil,
			}
		} else {
			job := *jobRef.Job
			nextRun, _ := job.NextRun()
			lastRun, _ := job.LastRun()
			jobMap[name] = &infras.JobStatus{
				ID:        job.ID().String(),
				Name:      job.Name(),
				NextRun:   &nextRun,
				LastRun:   &lastRun,
				IsRunning: true,
				Tags:      job.Tags(),
			}
		}
	}
	return jobMap, nil
}

func (c *CronService) StopJob(jobName string) error {
	jobRef, exists := c.Jobs[jobName]
	if !exists || jobRef.Job == nil {
		return fmt.Errorf("cron job %s is not running", jobName)
	}
	job := *jobRef.Job
	err := c.Scheduler.RemoveJob(job.ID())
	if err != nil {
		c.Logger.Errorf("Failed to stop cron job %s: %v", jobName, err)
		return err
	}
	c.Logger.Infof("Successfully stopped cron job %s", jobName)
	c.Jobs[jobName] = &infras.JobDetail{
		Name: jobName,
		Job:  nil,
	}
	return nil
}

func (c *CronService) ResumeJob(jobName string) error {
	jobRef, exists := c.Jobs[jobName]
	if exists && jobRef.Job != nil {
		return fmt.Errorf("cron job %s is running", jobName)
	}
	taskFunc, exists := TaskFuncs[jobName]
	if !exists {
		return fmt.Errorf("cron job %s does not exist", jobName)
	}
	job := taskFunc(c.Server, jobName)
	j, err := c.Scheduler.NewJob(
		job.Def,
		job.Fn,
	)
	if err != nil {
		c.Logger.Errorf("Failed to resume cron job %s: %v", jobName, err)
		return err
	}
	c.Jobs[jobName] = &infras.JobDetail{
		Name: jobName,
		Job:  &j,
	}
	c.Logger.Infof("Successfully resumed cron job %s", jobName)
	return nil
}

func (c *CronService) ForceRun(jobName string) error {
	jobRef, exists := c.Jobs[jobName]
	if !exists || jobRef == nil {
		return fmt.Errorf("cron job %s is not running", jobName)
	}
	job := *jobRef.Job
	err := job.RunNow()
	if err != nil {
		c.Logger.Errorf("Failed to force run cron job %s: %v", jobName, err)
		return err
	}
	c.Logger.Infof("Successfully forced run cron job %s", jobName)
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
	fx.Provide(NewCronService, NewRawCronService),
	fx.Invoke(func(c *CronService, lc fx.Lifecycle) {
		c.RegisterHooks(lc)
	}),
)
