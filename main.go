package main

import (
	"github.com/qxbao/asfpc/routes"
	"github.com/qxbao/asfpc/server"
	"github.com/qxbao/asfpc/server/modules/cron"
	"github.com/qxbao/asfpc/server/modules/database"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		database.DatabaseModule,
		server.ServerModule,
		cron.CronModule,
		routes.RoutesModule,
	)

	app.Run()
}
