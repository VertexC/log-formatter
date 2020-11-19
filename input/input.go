package input

import (
	"github.com/VertexC/log-formatter/input/elasticsearch"
	"github.com/VertexC/log-formatter/input/kafka"
	"github.com/VertexC/log-formatter/input/file"
)

type InputConfig struct {
	Target   string                 `yaml:"target"`
	EsCfg    *elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg *kafka.KafkaConfig           `yaml:"kafka,omitempty"`
	FilePath string                 `yaml:"file"`
}

type Input interface {
	// TODO: wrap inputCh and outputCh into contextChannl
	Run()
}

func NewInput(config InputConfig, inputCh chan map[string]interface{}) (input Input) {
	switch config.Target {
	case "elasticsearch":
		input = elasticsearch.NewEsInput(*config.EsCfg, inputCh)
	case "kafka":
		input = kafka.NewKafkaInput(*config.KafkaCfg, inputCh)
	case "file":
		input = file.NewFileInput(config.FilePath, inputCh)
	default:
		panic("Invalid input Target:" + config.Target)
	}
	return
}
