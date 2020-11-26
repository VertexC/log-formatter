package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/pkg/profile"
	"gopkg.in/yaml.v3"

	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
	"github.com/VertexC/log-formatter/pipeline"
	"github.com/VertexC/log-formatter/util"
)

// used for loading included files

type Config struct {
	LogDir      string                  `yaml:"log" default:"logs"`
	OutCfgs     []output.OutputConfig   `yaml:"outputs"`
	InCfg       input.InputConfig       `yaml:"input"`
	PipelineCfg pipeline.PipelineConfig `yaml:"pipeline"`
}

var configFilePath = flag.String("c", "config.yml", "config file path")
var verboseFlag = flag.Bool("v", false, "add TRACE/WARNING logging if enabled")
var cpuProfile = flag.Bool("cpuprof", false, "enable cpu profile")
var memProfile = flag.Bool("memprof", false, "enable mem profile")

var logger = new(util.Logger)

func loadConfig(configFile string) *Config {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("Failed to open file: ", err)
	}
	config := Config{}
	if err := yaml.Unmarshal(yamlFile, &util.IncludeProcessor{&config}); err != nil {
		log.Fatalln("Failed to decode yaml: ", err)
	}
	return &config
}

func Init() (config *Config, verbose bool) {
	configFile := *configFilePath
	verbose = *verboseFlag

	config = loadConfig(configFile)

	// create log Dir
	_ = os.Mkdir(config.LogDir, os.ModePerm)

	return
}

func main() {
	flag.Parse()

	// prof
	profiles := []func(*profile.Profile){}
	if *cpuProfile {
		profiles = append(profiles, profile.CPUProfile)
	}

	if *memProfile {
		profiles = append(profiles, profile.MemProfile)
	}

	if len(profiles) != 0 {
		profiles = append(profiles, profile.NoShutdownHook)
		defer profile.Start(profiles...).Stop()
	}

	config, verbose := Init()
	util.Verbose = verbose
	util.LogFile = path.Join(config.LogDir, "runtime.log")
	logger = util.NewLogger("Main")

	configPretty, _ := json.MarshalIndent(*config, "", "\t")
	logger.Info.Printf("Get config\n %s\n", configPretty)

	// TODO: make it configurable
	inputCh := make(chan util.Doc, 1000)
	outputCh := make(chan util.Doc, 1000)
	doneCh := make(chan struct{})

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// go routine to catch signal interrupt
	go func() {
		select {
		case <-sigterm:
			logger.Info.Println("terminating: via signal")
			doneCh <- struct{}{}
		}
	}()

	pipeline := pipeline.NewPipeline(config.PipelineCfg, inputCh, outputCh)
	input := input.NewInput(config.InCfg, inputCh)
	outputRunner := output.New(config.OutCfgs, outputCh)
	go pipeline.Run()
	go input.Run()
	go outputRunner.Start()

	<-doneCh
}
