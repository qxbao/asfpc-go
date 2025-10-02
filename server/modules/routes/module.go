package routes

import "go.uber.org/fx"

var RoutesModule = fx.Module(
	"RoutesModule",
	fx.Invoke(
		InitAccountRoutes,
		InitDataRoutes,
		InitAnalysisRoutes,
		InitMLRoutes,
		InitSettingRoutes,
		InitCronRoutes,
	),
)
