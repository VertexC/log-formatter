package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/VertexC/log-formatter/formatter"
	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
	"github.com/VertexC/log-formatter/util"
)

type Config struct {
	LogDir string           `yaml:"log" default:"logs"`
	OutCfg output.Config    `yaml:"output"`
	InCfg  input.Config     `yaml:"input"`
	FmtCfg formatter.Config `yaml:"formatter"`
}

var configFilePath = flag.String("c", "config.yml", "config file path")
var verboseFlag = flag.Bool("v", false, "add TRACE/WARNING if enabled")

var logger = new(util.Logger)

func loadConfig(configFile string) *Config {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("Failed to open file: ", err)
	}
	config := Config{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalln("Failed to decode yaml: ", err)
	}
	return &config
}

func Init() (config *Config, verbose bool) {
	flag.Parse()

	configFile := *configFilePath
	verbose = *verboseFlag

	config = loadConfig(configFile)

	// create log Dir
	_ = os.Mkdir(config.LogDir, os.ModePerm)

	return
}

func main() {
	config, verbose := Init()

	logFile := path.Join(config.LogDir, "runtime.log")
	logger.Init(logFile, "Main", verbose)

	logger.Info.Printf("Get config %+v\n", *config)

	inputCh := make(chan interface{}, 1000)
	outputCh := make(chan interface{}, 1000)
	doneCh := make(chan struct{})

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigterm:
			logger.Info.Println("terminating: via signal")
			doneCh <- struct{}{}
		}
	}()

	logger.Info.Println("Start Input Routine")
	go input.Execute(config.InCfg, inputCh, logFile, verbose)
	logger.Info.Println("Start Formatter Routine")
	go formatter.Execute(config.FmtCfg, inputCh, outputCh, logFile, verbose)
	logger.Info.Println("Start Output Routine")
	go output.Execute(config.OutCfg, outputCh, logFile, verbose)

	<-doneCh
}
