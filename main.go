package main

import (
	"github.com/qxbao/asfpc/services"
)

func main() {
	server := new(services.Server)
	server.Run()
}
