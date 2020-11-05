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

func Execute(config Config, inputCh chan interface{}, doneCh chan struct{}) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, inputCh, doneCh)
	case "kafka":
		// kafka.Execute(config.KafkaCfg, inputCh, doneCh)
		kafka.ExecuteGroup(config.KafkaCfg, inputCh, doneCh)
	}
}
