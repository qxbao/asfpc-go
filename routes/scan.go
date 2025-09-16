package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/services"
)

func InitScanRoutes(s infras.Server) {
	e := s.Echo

	scanService := services.ScanService{
		Server: s,
	}
	e.POST("/scan/group/:id", scanService.ScanGroupFeed)

	e.POST("/scan/post/:id", scanService.ScanPostComments)

	e.POST("/scan/profile/:id", scanService.ScanUserProfile)
}
