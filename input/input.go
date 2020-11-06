package input

import (
	"github.com/VertexC/log-formatter/input/elasticsearch"
	"github.com/VertexC/log-formatter/input/kafka"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg kafka.Config           `yaml:"kafka,omitempty"`
}

func Execute(config Config, inputCh chan interface{}, logFile string, verbose bool) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, inputCh, logFile, verbose)
	case "kafka":
		kafka.ExecuteGroup(config.KafkaCfg, inputCh, logFile, verbose)
	}
}
