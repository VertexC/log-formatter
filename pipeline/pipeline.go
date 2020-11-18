package pipeline

import (
	"github.com/VertexC/log-formatter/pipeline/formatter"
	"log"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type PipelineConfig {
	FormatterCfgs []formatter.FormatterConfig `yaml:"formatters"`
}

type Pipeline struct {
	formatters []formatter.Formatter
}

func NewPipeline(config PipelineConfig) *Pipeline {
	fmts := []formatter.Formatter {}
	for _, fmtCfg := range config.FormatterCfgs {
		fmt := formatter.NewFormatter(fmtCfg)
		fmts = append(fmts, fmt)
	}
	pipeline := new(Pipeline)
	pipeline.formatters := fmts
	return pipeline
}

// TODO: maybe move to util
func Merge(a map[string]interface{}, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

// TODO: Cache
func Filter(includeFields []string, record map[string]interface{}) map[string]interface{} {
	regList := []*regexp.Regexp{}
	for _, s := range includeFields {
		regList = append(regList, regexp.MustCompile(s))
	}

	result := map[string]interface{}{}
	for k, v := range record {
		for _, r := range regList {
			if r.MatchString(k) {
				result[k] = v
				break
			}
		}
	}
	return result
}

func (pipeline *Pipeline) Run (inputCh chan map[string]interface{}, outputCh chan interface{}) {
	for msg := range inputCh {
		for _, fmt := range pipeline.Formatters {
			msg, err = fmt.Format(msg)
			if fmt.BreakOnErr && err != nil {
				break
			}
		}
		ouputCh <- msg
	}
}
