package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"flag"

	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
)

type Config struct {
	LogDir string        `yaml:"log" default:"logs"`
	OutCfg output.Config `yaml:"output"`
	InCfg  input.Config  `yaml:"input"`
}


var configFilePath = flag.String("c", "config.yml", "config file path")

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

	configFile := *configFilePath

	config := loadConfig(configFile)
	fmt.Printf("%+v\n", *config)

	// create log Dir
	_ = os.Mkdir(config.LogDir, os.ModePerm)

	records := make(chan []interface{})
	doneCh := make(chan struct{})

	go input.Execute(config.InCfg, records, doneCh)
	go output.Execute(config.OutCfg, records, doneCh)

	<-doneCh
}
