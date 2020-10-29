package input

import (
	"github.com/VertexC/log-formatter/input/elasticsearch"
)

type Config struct {
	Target string      `yaml:"target"`
	EsCfg     elasticsearch.EsConfig    `yaml:"elasticsearch,omitempty"`
	KafkaCfg  KafkaConfig `yaml:"kafka,omitempty"`
}


func Execute(config Config, records chan[] interface{}, inLastJobCh chan int) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, records, inLastJobCh)
	}
}
