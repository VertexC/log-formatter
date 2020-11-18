package pipeline

import (
	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type PipelineConfig struct {
	FormatterCfgs []FormatterConfig `yaml:"formatters"`
	Labels []Label `yaml:"labels"`
}

type Pipeline struct {
	formatters []Formatter
	inputCh chan map[string]interface{}
	outputCh chan map[string]interface{}
	logger *util.Logger
	// TODO: move labelling to proper component of log-formatter
	labels map[string]string
}

func NewPipeline(config PipelineConfig, inputCh chan map[string]interface{}, outputCh chan map[string]interface{}) *Pipeline {
	fmts := []Formatter {}
	for _, fmtCfg := range config.FormatterCfgs {
		fmt := NewFormatter(fmtCfg)
		fmts = append(fmts, fmt)
	}
	pipeline := new(Pipeline)
	pipeline.formatters = fmts
	pipeline.labels = map[string]string {}
	for _, label := range config.Labels {
		pipeline.labels[label.Key] = label.Val
	}
	pipeline.inputCh = inputCh
	pipeline.outputCh = outputCh
	pipeline.logger = util.NewLogger("pipeline")
	return pipeline
}

func (pipeline *Pipeline) Run () {
	for doc := range pipeline.inputCh {
		discard := false
		for _, fmt := range pipeline.formatters {
			doc , err := fmt.Format(doc)
			if err != nil {
				discard = true
				pipeline.logger.Warning.Printf("Discard doc:%s **with err** %s", doc, err)
			}
		}
		if !discard {
			for k,v := range pipeline.labels {
				doc[k] = v
			}
			pipeline.outputCh <- doc
		}
	}
}
