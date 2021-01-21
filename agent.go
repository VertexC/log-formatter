package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/profile"
	"os"
	"path"

	"github.com/VertexC/log-formatter/agent"
	"github.com/VertexC/log-formatter/agent/config"
	_ "github.com/VertexC/log-formatter/agent/include"
	"github.com/VertexC/log-formatter/util"
	mylogger "github.com/VertexC/log-formatter/logger"
)

// Options contains the top level options of log-formatter
var options = &struct {
	configFile  string
	monitorAddr string
	rpcPort     string
	logDir      string
	verbose     int
	cpuProfile  bool
	memProfile  bool
}{}

type Config struct {
	Base   *config.ConfigBase
	LogDir string `yaml:"log" default:"logs"`
}

func init() {
	flag.StringVar(&options.configFile, "c", "config.yml", "config file path")
	flag.StringVar(&options.monitorAddr, "monitor", "", "monitor rpc server address, empty for standalone mode")
	flag.StringVar(&options.rpcPort, "rpcp", "2020", "agent's rpc port")
	flag.StringVar(&options.logDir, "l", "logs", "log directory")
	flag.IntVar(&options.verbose, "v", 0, mylogger.VerboseDescription)
	flag.BoolVar(&options.cpuProfile, "cpuprof", false, "enable cpu profile")
	flag.BoolVar(&options.memProfile, "memprof", false, "enable mem profile")
}

func main() {
	logger := mylogger.NewLogger("Main")

	flag.Parse()

	// prof
	profiles := []func(*profile.Profile){}
	if options.cpuProfile {
		profiles = append(profiles, profile.CPUProfile)
	}

	if options.memProfile {
		profiles = append(profiles, profile.MemProfile)
	}

	if len(profiles) != 0 {
		profiles = append(profiles, profile.NoShutdownHook)
		defer profile.Start(profiles...).Stop()
	}

	if _, err := os.Stat(options.logDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(options.logDir, os.ModePerm); err != nil {
				panic(fmt.Sprintf("Failed to create dir <%s>: %s", options.logDir, err))
			}
		} else {
			panic(fmt.Sprintf("Failed to get stats of dir <%s>: %s", options.logDir, err))
		}
	}

	mylogger.Verbose = options.verbose
	mylogger.LogFile = path.Join(options.logDir, "agent.log")

	// load config content
	content, err := config.LoadMapStrFromYamlFile(options.configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config From File: %s", err))
	}

	configPretty, _ := json.MarshalIndent(content, "", "  ")
	logger.Info.Printf("Get config\n %s\n", configPretty)

	manager, err := agent.NewAgentsManager(options.monitorAddr, options.rpcPort)
	if err != nil {
		panic(err)
	}

	if err := manager.SetConfig(content); err != nil {
		panic(err)
	}
	manager.Run()

	util.SigControl(manager.Stop)
}
