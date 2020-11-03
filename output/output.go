package output

import (
	"github.com/VertexC/log-formatter/output/console"
	"github.com/VertexC/log-formatter/output/elasticsearch"
)

type Config struct {
	Target   string                 `yaml:"target"`
	EsCfg    elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg KafkaConfig            `yaml:"kafka,omitempty"`
}

func Execute(config Config, outputCh chan interface{}) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, outputCh)
	case "console":
		console.Execute(outputCh)
	}
}
