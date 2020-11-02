package output

import (
	"github.com/VertexC/log-formatter/output/elasticsearch"
	"github.com/VertexC/log-formatter/output/console"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg KafkaConfig            `yaml:"kafka,omitempty"`
}

func Execute(config Config, records chan []interface{}, doneCh chan struct{}) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, records)
	case "console":
		console.Execute(records)
	}
}
