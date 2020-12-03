package pipeline

import (
	"log"

	"github.com/VertexC/log-formatter/pipeline/filter"
	"github.com/VertexC/log-formatter/pipeline/parser"
	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Formatter interface {
	Format(util.Doc) (util.Doc, error)
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
	Worker        int               `yaml:"worker"`
}

type worker struct {
	inputCh  chan util.Doc
	outputCh chan util.Doc
	logger   *util.Logger
	// TODO: move labelling to proper component of log-formatter
	labels     map[string]string
	formatters []Formatter
}

type Pipeline struct {
	inputCh  chan util.Doc
	outputCh chan util.Doc
	logger   *util.Logger
	workers  []*worker
}

func NewPipeline(config PipelineConfig, inputCh chan util.Doc, outputCh chan util.Doc) *Pipeline {
	logger := util.NewLogger("pipeline")
	pipeline := new(Pipeline)
	pipeline.logger = logger
	if config.Worker == 0 {
		config.Worker = 1
	}
	for i := 0; i < config.Worker; i++ {
		fmts := []Formatter{}
		for _, fmtCfg := range config.FormatterCfgs {
			fmt := NewFormatter(fmtCfg)
			fmts = append(fmts, fmt)
		}
		labels := map[string]string{}
		for _, label := range config.Labels {
			labels[label.Key] = label.Val
		}
		w := &worker{
			inputCh:    inputCh,
			outputCh:   outputCh,
			logger:     logger,
			formatters: fmts,
		}
		pipeline.workers = append(pipeline.workers, w)
	}
	return pipeline
}

func (pipeline *Pipeline) Run() {
	for _, worker := range pipeline.workers {
		go worker.Run()
	}
}

func (w *worker) Run() {
	for doc := range w.inputCh {
		discard := false
		for _, fmt := range w.formatters {
			var err error
			doc, err = fmt.Format(doc)
			if err != nil {
				discard = true
				w.logger.Warning.Printf("Discard doc:%s **with err** %s", doc, err)
			}
		}
		if !discard {
			for k, v := range w.labels {
				doc[k] = v
			}
			w.outputCh <- doc
		}
	}
}
