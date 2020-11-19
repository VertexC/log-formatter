package pipeline

import (
	"github.com/VertexC/log-formatter/pipeline/filter"
	"github.com/VertexC/log-formatter/pipeline/parser"
	"github.com/VertexC/log-formatter/util"
	"log"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Formatter interface {
	Format(map[string]interface{}) (map[string]interface{}, error)
}

type FormatterConfig struct {
	Type      string               `yaml:"type"`
	ParserCfg *parser.ParserConfig `yaml:"parser"`
	FilterCfg *filter.FilterConfig `yaml:"filter"`
}

func NewFormatter(config FormatterConfig) Formatter {
	switch config.Type {
	case "parser":
		return parser.NewParser(*config.ParserCfg)
	case "filter":
		formatter, err := filter.NewFilter(*config.FilterCfg)
		if err != nil {
			log.Fatalf("Error when create filter: %s:", err)
		}
		return formatter
	default:
		panic("Invalid Formatter Type:" + config.Type)
	}
}

type PipelineConfig struct {
	FormatterCfgs []FormatterConfig `yaml:"formatters"`
	Labels        []Label           `yaml:"labels"`
}

type Pipeline struct {
	formatters []Formatter
	inputCh    chan map[string]interface{}
	outputCh   chan map[string]interface{}
	logger     *util.Logger
	// TODO: move labelling to proper component of log-formatter
	labels map[string]string
}

func NewPipeline(config PipelineConfig, inputCh chan map[string]interface{}, outputCh chan map[string]interface{}) *Pipeline {
	fmts := []Formatter{}
	for _, fmtCfg := range config.FormatterCfgs {
		fmt := NewFormatter(fmtCfg)
		fmts = append(fmts, fmt)
	}
	pipeline := new(Pipeline)
	pipeline.formatters = fmts
	pipeline.labels = map[string]string{}
	for _, label := range config.Labels {
		pipeline.labels[label.Key] = label.Val
	}
	pipeline.inputCh = inputCh
	pipeline.outputCh = outputCh
	pipeline.logger = util.NewLogger("pipeline")
	return pipeline
}

func (pipeline *Pipeline) Run() {
	for doc := range pipeline.inputCh {
		discard := false
		for _, fmt := range pipeline.formatters {
			var err error
			doc, err = fmt.Format(doc)
			if err != nil {
				discard = true
				pipeline.logger.Warning.Printf("Discard doc:%s **with err** %s", doc, err)
			}
		}
		if !discard {
			for k, v := range pipeline.labels {
				doc[k] = v
			}
			pipeline.outputCh <- doc
		}
	}
}
