package output

import (
	"encoding/json"
	"github.com/VertexC/log-formatter/output/console"
	"github.com/VertexC/log-formatter/output/elasticsearch"
	"github.com/VertexC/log-formatter/output/kafka"
	"log"
	"os"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg kafka.KafkaConfig      `yaml:"kafka,omitempty"`
	File     string                 `yaml:"file"`
}

func writeFile(file string, outputCh chan interface{}) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0667)
	if err != nil {
		log.Fatalf("Failed to open file %s with error:%s \n", file, err)
	}
	defer f.Close()

	for doc := range outputCh {
		data, err := json.Marshal(doc)
		if err != nil {
			log.Fatalf("Failed to marshal doc: %+v into json. %s", doc, err)
		}
		f.WriteString(string(data) + "\n")
	}
}

func Execute(config Config, outputCh chan interface{}, logFile string, verbose bool) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, outputCh, logFile, verbose)
	case "kafka":
		kafka.Execute(config.KafkaCfg, outputCh, logFile, verbose)
	case "console":
		console.Execute(outputCh)
	case "file":
		writeFile(config.File, outputCh)
	default:
		panic("Invalid output target:" + config.Target)
	}
}
