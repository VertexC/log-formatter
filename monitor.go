package main

import (
	"flag"

	"github.com/VertexC/log-formatter/monitor"
	"github.com/VertexC/log-formatter/util"
)

var options = &struct {
	rpcport     string
	webport     string
	configFile  string
	verboseFlag bool
	logDir      string
}{}

func init() {
	flag.StringVar(&options.rpcport, "rpcp", "8081", "port to run rpc service")
	flag.StringVar(&options.webport, "webp", "8080", "port to run web server")
	flag.StringVar(&options.logDir, "l", "logs", "log directory")
	flag.BoolVar(&options.verboseFlag, "v", false, "add TRACE/WARNING logging if enabled")
}

func main() {
	flag.Parse()

	util.Verbose = options.verboseFlag
	logger := util.NewLogger("monitor")
	app, err := monitor.NewApp(options.rpcport, options.webport)
	if err != nil {
		logger.Error.Fatalf("Failed to create App: %s", err)
	}
	app.Start()
	util.ExitControl()
}
