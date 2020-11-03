package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"

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

var configFilePath = flag.String("c", "", "config file path")

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

func main() {

	flag.Parse()
	logger.Init("main")

	configFile := *configFilePath

	config := loadConfig(configFile)
	logger.Info.Printf("%+v\n", *config)

	// create log Dir
	_ = os.Mkdir(config.LogDir, os.ModePerm)

	inputCh := make(chan interface{}, 1000)
	outputCh := make(chan interface{}, 1000)
	doneCh := make(chan struct{})

	go input.Execute(config.InCfg, inputCh, doneCh)
	go formatter.Execute(config.FmtCfg, inputCh, outputCh)
	go output.Execute(config.OutCfg, outputCh)

	<-doneCh
}
