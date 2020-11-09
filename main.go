package main

import (
	"errors"
	"flag"
	"github.com/VertexC/log-formatter/formatter"
	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
	"github.com/VertexC/log-formatter/util"
	"github.com/pkg/profile"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

// used for loading included files
type Fragment struct {
	content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(value *yaml.Node) error {
	var err error
	// process includes in fragments
	f.content, err = resolveIncludes(value)
	return err
}

type IncludeProcessor struct {
	target interface{}
}

func (i *IncludeProcessor) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := resolveIncludes(value)
	if err != nil {
		return err
	}
	return resolved.Decode(i.target)
}

func resolveIncludes(node *yaml.Node) (*yaml.Node, error) {
	if node.Tag == "!include" {
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!include on a non-scalar node")
		}
		file, err := ioutil.ReadFile(node.Value)
		if err != nil {
			return nil, err
		}
		var f Fragment
		err = yaml.Unmarshal(file, &f)
		return f.content, err
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = resolveIncludes(node.Content[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}

type Config struct {
	LogDir string           `yaml:"log" default:"logs"`
	OutCfg output.Config    `yaml:"output"`
	InCfg  input.Config     `yaml:"input"`
	FmtCfg formatter.Config `yaml:"formatter"`
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
	if err := yaml.Unmarshal(yamlFile, &IncludeProcessor{&config}); err != nil {
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

	logFile := path.Join(config.LogDir, "runtime.log")
	logger.Init(logFile, "Main", verbose)

	logger.Info.Printf("Get config %+v\n", *config)

	// TODO: make it configurable
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
