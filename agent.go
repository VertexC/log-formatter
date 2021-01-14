package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/pkg/profile"

	"github.com/VertexC/log-formatter/agent"
	"github.com/VertexC/log-formatter/config"
	_ "github.com/VertexC/log-formatter/include"
	"github.com/VertexC/log-formatter/util"
)

// Options contains the top level options of log-formatter
var options = &struct {
	configFile  string
	monitorAddr string
	rpcPort     string
	logDir      string
	verboseFlag bool
	cpuProfile  bool
	memProfile  bool
}{}

type Config struct {
	Base   *config.ConfigBase
	LogDir string `yaml:"log" default:"logs"`
}

func init() {
	flag.StringVar(&options.configFile, "c", "config.yml", "config file path")
	flag.StringVar(&options.monitorAddr, "monitor", "", "monitor rpc server address")
	flag.StringVar(&options.rpcPort, "rpcp", "2020", "agent's rpc port")
	flag.StringVar(&options.logDir, "l", "logs", "log directory")
	flag.BoolVar(&options.verboseFlag, "v", false, "add TRACE/WARNING logging if enabled")
	flag.BoolVar(&options.cpuProfile, "cpuprof", false, "enable cpu profile")
	flag.BoolVar(&options.memProfile, "memprof", false, "enable mem profile")
}

func main() {
	logger := util.NewLogger("Main")

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

	util.Verbose = options.verboseFlag
	util.LogFile = path.Join(options.logDir, "runtime.log")

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

	util.ExitControl()
}
