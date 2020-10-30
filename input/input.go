package input

import (
	"github.com/VertexC/log-formatter/input/elasticsearch"
	"github.com/VertexC/log-formatter/input/kafka"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg kafka.KafkaConfig      `yaml:"kafka,omitempty"`
}

func Execute(config Config, records chan []interface{}, done chan struct{}) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, records, done)
	case "kafka":
		kafka.Execute(config.KafkaCfg, records, done)
	}
}
