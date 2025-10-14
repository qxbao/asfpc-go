package routes

import (
	"context"

	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/ml"
	"go.uber.org/fx"
)

var RoutesModule = fx.Module(
	"RoutesModule",
	fx.Invoke(
		InitAccountRoutes,
		InitDataRoutes,
		InitAnalysisRoutes,
		InitMLRoutes,
		InitSettingRoutes,
		InitCronRoutes,
		InitCategoryRoutes,
		InitModelRoutes,
		SyncModelsOnStartup,
	),
)

func SyncModelsOnStartup(lc fx.Lifecycle, s *infras.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mlService := ml.MLRoutingService{Server: s}
			return mlService.SyncModelsWithDatabase(ctx)
		},
	})
}
