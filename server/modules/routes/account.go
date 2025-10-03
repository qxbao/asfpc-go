package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/account"
)

func InitAccountRoutes(s *infras.Server) {
	e := s.Echo

	services := account.AccountService{Server: s}

	e.GET("/account/info", services.GetAccount)
	e.GET("/account/list", services.GetAccounts)
	e.GET("/account/stats", services.GetAccountStats)
	e.GET("/account/group/list", services.GetGroupsByAccountID)
	e.POST("/account/update/credentials", services.UpdateAccountCredentials)
	e.POST("/account/login", services.LoginAccount)
	e.POST("/account/add", services.AddAccount)
	e.POST("/account/token/gen", services.GenAccountsAT)
	e.POST("/account/group/join", services.JoinGroup)
	e.POST("/account/group/add", services.CreateGroup)
	e.DELETE("/account/delete", services.DeleteAccounts)
	e.DELETE("/account/group/delete", services.DeleteGroup)
}