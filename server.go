package main

import (
	"flag"
	"log"

	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/server"
	"github.com/VertexC/log-formatter/util"
)

var options = &struct {
	configFile string
}{}

func init() {
	flag.StringVar(&options.configFile, "c", "config.yml", "config file path")
}

func main() {
	flag.Parse()
	// load config content
	content, err := config.LoadMapStrFromYamlFile(options.configFile)
	if err != nil {
		log.Fatalf("Failed to parse yaml from file %s: %s", options.configFile, err)
	}
	log.Println(content)
	app, err := server.NewApp(content)
	if err != nil {
		log.Fatalf("Failed to create App: %s", err)
	}
	app.Start()
	util.ExitControl()
}
