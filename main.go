package main

import (
	"github.com/qxbao/asfpc/server"
)

func main() {
	server := new(server.Server)
	server.Run()
}
