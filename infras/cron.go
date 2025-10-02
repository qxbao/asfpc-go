package infras

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type CronService struct {
	CronServiceInterface
	Scheduler gocron.Scheduler
	Logger    *zap.SugaredLogger
	Server    *Server
	Jobs      map[string] JobDetail
}

type JobDetail struct {
	Name string
	Job  *gocron.Job
}

type CronServiceInterface interface {
	NewCronService(server *Server) (*CronService, error)
	AddTask(s *Server)
	ListJobs() ([]*JobStatus, error)
	StopJob(jobName string) error
	ResumeJob(jobName string) error
	ForceRun(jobName string) error
}

type JobStatus struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	NextRun   *time.Time `json:"next_run,omitempty"`
	LastRun   *time.Time `json:"last_run,omitempty"`
	IsRunning bool       `json:"is_running"`
	Tags      []string   `json:"tags,omitempty"`
}

type JobControlRequest struct {
	JobID string `json:"job_id" validate:"required"`
}

type JobListResponse struct {
	Jobs  map[string]*JobStatus `json:"data"`
	Count int                   `json:"count"`
}