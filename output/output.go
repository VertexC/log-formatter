package output

import (
	"github.com/VertexC/log-formatter/output/console"
	"github.com/VertexC/log-formatter/output/elasticsearch"
	"github.com/VertexC/log-formatter/output/file"
	"github.com/VertexC/log-formatter/output/kafka"
)

type OutputConfig struct {
	Target   string                  `yaml:"target"`
	EsCfg    *elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg *kafka.KafkaConfig      `yaml:"kafka,omitempty"`
	File     string                  `yaml:"file"`
}

type Output interface {
	Run()
}

func NewOutput(config OutputConfig, docCh chan map[string]interface{}) (output Output) {
	switch config.Target {
	case "elasticsearch":
		output = elasticsearch.NewEsOutput(*config.EsCfg, docCh)
	case "kafka":
		output = kafka.NewKafkaOutput(*config.KafkaCfg, docCh)
	case "console":
		output = console.NewConsoleOutput(docCh)
	case "file":
		output = file.NewFileOutput(config.File, docCh)
	default:
		panic("Invalid output target:" + config.Target)
	}
	return
}
