package output

import (
	"github.com/VertexC/log-formatter/output/elasticsearch"
)

type Config struct {
	Target string      `yaml:"target"`
	EsCfg     elasticsearch.EsConfig    `yaml:"elasticsearch,omitempty"`
	KafkaCfg  KafkaConfig `yaml:"kafka,omitempty"`
}



func Execute(config Config, records chan[] interface{}, outJobCh chan int) {
	switch config.Target {
	case "elasticsearch":
		elasticsearch.Execute(config.EsCfg, records, outJobCh)
	}
}
