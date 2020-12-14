package pipeline

import (
	"fmt"

	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Formatter interface {
	Format(map[string]interface{}) (map[string]interface{}, error)
}

type Factory = func(interface{}) (Formatter, error)

var registry = make(map[string]Factory)
var logger = util.NewLogger("PIPLINE")

func Register(name string, factory Factory) error {
	logger.Info.Printf("Registering formatter <%s>\n", name)
	if name == "" {
		return fmt.Errorf("Error registering formatter: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering formatter '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering formatter '%v': already registered", name)
	}

	registry[name] = factory
	logger.Info.Printf("Successfully registered formatter <%s>\n", name)

	return nil
}

func NewFormatter(content interface{}) (Formatter, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot convert given formatter config to mapStr")
	}
	for name, val := range contentMapStr {
		if factory, ok := registry[name]; ok {
			output, err := factory(val)
			return output, err
		}
	}
	return nil, fmt.Errorf("Failed to creat any output target")
}

type worker struct {
	inputCh  chan map[string]interface{}
	outputCh chan map[string]interface{}
	logger   *util.Logger
	// TODO: move labelling to proper component of log-formatter
	labels     map[string]string
	formatters []Formatter
}

type PipelineConfig struct {
	Base   config.ConfigBase
	Worker int `yaml:"worker"`
}

type Pipeline struct {
	logger  *util.Logger
	workers []*worker
}

func NewPipeline(content interface{}, inputCh chan map[string]interface{}, outputCh chan map[string]interface{}) (*Pipeline, error) {
	logger := util.NewLogger("pipeline")
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to convert pipeline config to MapStr")
	}

	config := PipelineConfig{
		Base: config.ConfigBase{
			Content:          contentMapStr,
			MandantoryFields: []string{"formatters"},
		},
		Worker: 1,
	}

	if err := config.Base.Validate(); err != nil {
		return nil, err
	}

	util.YamlConvert(contentMapStr, &config)

	formatterCfgs, ok := contentMapStr["formatters"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to convert config to []MapStr")
	}

	pipeline := new(Pipeline)
	pipeline.logger = logger
	for i := 0; i < config.Worker; i++ {
		fmts := []Formatter{}
		for _, c := range formatterCfgs {
			fmt, err := NewFormatter(c)
			if err != nil {
				return nil, err
			}
			fmts = append(fmts, fmt)
		}

		w := &worker{
			inputCh:    inputCh,
			outputCh:   outputCh,
			logger:     logger,
			formatters: fmts,
		}
		pipeline.workers = append(pipeline.workers, w)
	}
	return pipeline, nil
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
