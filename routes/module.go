package routes

import (
	"github.com/qxbao/asfpc/server"
	"go.uber.org/fx"
)

type RouteService struct {
	Server *server.Server
}

func NewRouteService(s *server.Server) *RouteService {
	return &RouteService{
		Server: s,
	}
}

func (r *RouteService) InitRoutes() {
	InitAccountRoutes(r.Server.Server)
	InitDataRoutes(r.Server.Server)
	InitAnalysisRoutes(r.Server.Server)
	InitMLRoutes(r.Server.Server)
	InitSettingRoutes(r.Server.Server)
}

var RoutesModule = fx.Module(
	"RoutesModule",
	fx.Provide(
		NewRouteService,
	),
	fx.Invoke(func(rs *RouteService) {
		rs.InitRoutes()
	}),
)
