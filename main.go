package main

import (
	"github.com/qxbao/asfpc/server"
	"github.com/qxbao/asfpc/server/modules/cron"
	"github.com/qxbao/asfpc/server/modules/database"
	"github.com/qxbao/asfpc/server/modules/routes"
	"github.com/qxbao/asfpc/server/modules/seeding"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		database.DatabaseModule,
		server.ServerModule,
		seeding.SeedModule,
		cron.CronModule,
		routes.RoutesModule,
	)

	app.Run()
}
