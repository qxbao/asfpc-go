package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/data"
)

func InitDataRoutes(s *infras.Server) {
	e := s.Echo
	services := data.DataRoutingService{Server: s}

	e.GET("/data/stats", services.GetDataStats)
	e.GET("/data/prompt/list", services.GetAllPrompts)
	e.GET("/data/log/list", services.GetLogs)
	e.GET("/data/request/:request_id", services.TraceRequest)
	e.POST("/data/prompt/add", services.CreatePrompt)
}