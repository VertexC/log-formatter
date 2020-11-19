package input

import (
	"bufio"
	"log"
	"os"

	"github.com/VertexC/log-formatter/input/elasticsearch"
	"github.com/VertexC/log-formatter/input/kafka"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    *elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg *kafka.Config           `yaml:"kafka,omitempty"`
	FilePath string                 `yaml:"file"`
}

func readFile(filePath string, inputCh chan map[string]interface{}) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open file %s: %s", filePath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		inputCh <- map[string]interface{}{"message": scanner.Text()}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read from file %s", err)
	}
}

func Execute(config Config, inputCh chan map[string]interface{}) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(*config.EsCfg, inputCh)
	case "kafka":
		kafka.ExecuteGroup(*config.KafkaCfg, inputCh)
	case "file":
		readFile(config.FilePath, inputCh)
	default:
		panic("Invalid input Target:" + config.Target)
	}
}
