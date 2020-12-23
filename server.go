package main

import (
	"log"

	"github.com/VertexC/log-formatter/server"
	"github.com/VertexC/log-formatter/util"
)

func main() {
	app, err := server.NewApp()
	if err != nil {
		log.Fatalf("Failed to create App: %s", err)
	}
	app.Start()
	util.ExitControl()
}