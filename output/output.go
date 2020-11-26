package output

import (
	"github.com/VertexC/log-formatter/output/console"
	"github.com/VertexC/log-formatter/output/elasticsearch"
	"github.com/VertexC/log-formatter/output/file"
	"github.com/VertexC/log-formatter/output/kafka"
	"github.com/VertexC/log-formatter/util"
)

type OutputConfig struct {
	Target   string                  `yaml:"target"`
	EsCfg    *elasticsearch.EsConfig `yaml:"elasticsearch,omitempty"`
	KafkaCfg *kafka.KafkaConfig      `yaml:"kafka,omitempty"`
	File     string                  `yaml:"file"`
}

type Output interface {
	Run()
	Append(util.Doc)
}

type Runner struct {
	outputs []Output
	docCh   chan util.Doc
}

func (r *Runner) Start() {
	// start each output
	for _, output := range r.outputs {
		go output.Run()
	}
	// distribute doc to each output
	for doc := range r.docCh {
		// TODO: no deepCoy now since doc is read only under current implemention
		for _, output := range r.outputs {
			output.Append(doc)
		}
	}
}

func New(configs []OutputConfig, docCh chan util.Doc) *Runner {
	r := &Runner{
		outputs: []Output{},
		docCh:   docCh,
	}
	for _, config := range configs {
		var output Output
		switch config.Target {
		case "elasticsearch":
			output = elasticsearch.NewEsOutput(*config.EsCfg)
		case "kafka":
			output = kafka.NewKafkaOutput(*config.KafkaCfg)
		case "console":
			output = console.NewConsoleOutput()
		case "file":
			output = file.NewFileOutput(config.File)
		default:
			panic("Invalid output target:" + config.Target)
		}
		r.outputs = append(r.outputs, output)
	}
	return r
}
