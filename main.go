package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"

	"github.com/VertexC/log-formatter/formatter"
	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
)

type Config struct {
	Output output.OutputConfig `yaml:"output"`
	Input  input.InputConfig   `yaml:"input"`
}

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
	configFile := "formatter-wish.yml"

	config := loadConfig(configFile)
	fmt.Printf("%+v\n", *config)
	fmt.Println(formatter.Version)

	records := make(chan []interface{})
	inLastJobCh := make(chan int)
	outJobCh := make(chan int)

	go input.EsSearch(config.Input, records, inLastJobCh)
	go output.EsUpdate(config.Output, records, outJobCh)

	// check if last input records has finihsed
	for {
		inLastJobId := <-inLastJobCh
		outJobId := <-outJobCh
		if outJobId == inLastJobId {
			break
		}
	}
}
