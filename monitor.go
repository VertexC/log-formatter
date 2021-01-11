package main

import (
	"flag"

	"github.com/VertexC/log-formatter/monitor"
	"github.com/VertexC/log-formatter/util"
)

var options = &struct {
	rpcport string
	configFile string
	verboseFlag bool
	logDir string
}{}

func init() {
	flag.StringVar(&options.rpcport, "p", "8080", "rpcport to run web server")
	flag.StringVar(&options.logDir, "l", "logs", "log directory")
	flag.BoolVar(&options.verboseFlag, "v", false, "add TRACE/WARNING logging if enabled")
}

func main() {
	flag.Parse()

	util.Verbose = options.verboseFlag
	logger := util.NewLogger("monitor")
	app, err := monitor.NewApp(options.rpcport)
	if err != nil {
		logger.Error.Fatalf("Failed to create App: %s", err)
	}
	app.Start()
	util.ExitControl()
}
