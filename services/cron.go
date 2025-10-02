package services

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/cron"
)

type CronService struct {
	Server *infras.Server
	Cron   *cron.CronService
}

func (s *CronService) ListJobs(c echo.Context) error {
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