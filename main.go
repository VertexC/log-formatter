package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/pkg/profile"

	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
	"github.com/VertexC/log-formatter/pipeline"

	"github.com/VertexC/log-formatter/config"
	_ "github.com/VertexC/log-formatter/include"
	"github.com/VertexC/log-formatter/util"
)

// Options contains the top level options of log-formatter
var options = &struct {
	configFile  string
	verboseFlag bool
	cpuProfile  bool
	memProfile  bool
}{}

type Config struct {
	Base   *config.ConfigBase
	LogDir string `yaml:"log" default:"logs"`
	// OutCfg      map[string]interface{}  `yaml:"output"`
	// InCfg       map[string]interface{} `yaml:"input"`
	// PipelineCfg map[string]interface{} `yaml:"pipeline"`
}

func init() {
	flag.StringVar(&options.configFile, "c", "config.yml", "config file path")
	flag.BoolVar(&options.verboseFlag, "v", false, "add TRACE/WARNING logging if enabled")
	flag.BoolVar(&options.cpuProfile, "cpuprof", false, "enable cpu profile")
	flag.BoolVar(&options.memProfile, "memprof", false, "enable mem profile")
}

// Validate: validate and set with default field in content
func (c *Config) Validate() error {
	// check mandantory field
	if err := c.Base.Validate(); err != nil {
		return err
	}
	return nil
}

func main() {
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

	// load config content
	content, err := config.LoadMapStrFromYamlFile(options.configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config From File: %s", err))
	}

	// default config
	config := &Config{
		Base: &config.ConfigBase{
			Content:          content,
			MandantoryFields: []string{"input", "output", "pipeline"},
		},
		LogDir: "logs",
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Failed to parse config: %s", err))
	}

	if _, err := os.Stat(config.LogDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(config.LogDir, os.ModePerm); err != nil {
				panic(fmt.Sprintf("Failed to create dir <%s>: %s", config.LogDir, err))
			}
		} else {
			panic(fmt.Sprintf("Failed to get stats of dir <%s>: %s", config.LogDir, err))
		}
	}

	util.Verbose = options.verboseFlag
	util.LogFile = path.Join(config.LogDir, "runtime.log")

	logger := util.NewLogger("Main")
	configPretty, _ := json.MarshalIndent(config, "", "  ")
	logger.Info.Printf("Get config\n %s\n", configPretty)

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

	// // TODO: make it configurable
	inputCh := make(chan map[string]interface{}, 1000)
	outputCh := make(chan map[string]interface{}, 1000)
	// construct components
	pipeline, err := pipeline.NewPipeline(config.Base.Content["pipeline"], inputCh, outputCh)
	if err != nil {
		panic(fmt.Sprintf("Failed to create pipeline: %s", err))
	}

	output, err := output.NewOutput(config.Base.Content["output"], outputCh)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Output: %s", err))
	}

	input, err := input.NewInput(config.Base.Content["input"], inputCh)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Input: %s", err))
	}

	// logger.Debug.Println("output: %+v", output)
	// logger.Debug.Println("input: %+v", input)
	// logger.Debug.Println("pipeline: %+v", pipeline)
	go pipeline.Run()
	go input.Run()
	go output.Run()

	<-doneCh
}
