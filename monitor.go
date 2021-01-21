package main

import (
	"flag"

	"github.com/VertexC/log-formatter/logger"
	"github.com/VertexC/log-formatter/monitor-be"
	"github.com/VertexC/log-formatter/util"
)

var options = &struct {
	rpcport    string
	webport    string
	configFile string
	verbose    int
	logDir     string
}{}

func init() {
	flag.StringVar(&options.rpcport, "rpcp", "8081", "port to run rpc service")
	flag.StringVar(&options.webport, "webp", "8080", "port to run web server")
	flag.StringVar(&options.logDir, "l", "logs", "log directory")
	flag.IntVar(&options.verbose, "v", 0, logger.VerboseDescription)
}

func main() {
	flag.Parse()

	logger.Verbose = options.verbose
	logger := logger.NewLogger("monitor")
	app, err := monitor.NewApp(options.rpcport, options.webport)
	if err != nil {
		logger.Error.Fatalf("Failed to create App: %s", err)
	}
	app.Start()
	handler := func() {}
	util.SigControl(handler)
}
