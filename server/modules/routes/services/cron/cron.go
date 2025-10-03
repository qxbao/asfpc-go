package cron

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/cron"
)

type CronRoutingService struct {
	Server *infras.Server
	Cron   *cron.CronService
}

func (s *CronRoutingService) ListJobs(c echo.Context) error {
	jobsMap, err := s.Cron.ListJobs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to list jobs: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, infras.JobListResponse{
		Jobs:  jobsMap,
		Count: len(jobsMap),
	})
}

func (s *CronRoutingService) TriggerJob(c echo.Context) error {
	dto := new(infras.JobRequestWithName)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}
	if err := s.Cron.ForceRun(dto.JobName); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to trigger job: " + err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}

func (s *CronRoutingService) StopJob(c echo.Context) error {
	dto := new(infras.JobRequestWithName)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}
	if err := s.Cron.StopJob(dto.JobName); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to stop job: " + err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}

func (s *CronRoutingService) ResumeJob(c echo.Context) error {
	dto := new(infras.JobRequestWithName)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}
	if err := s.Cron.ResumeJob(dto.JobName); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to resume job: " + err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}