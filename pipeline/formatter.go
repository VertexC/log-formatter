package pipeline

import (
	"log"

	"github.com/VertexC/log-formatter/pipeline/parser"
	"github.com/VertexC/log-formatter/pipeline/filter"
)

type Formatter interface {
	Format(map[string]interface{}) (map[string]interface{}, error)
}

type FormatterConfig struct {
	Type          string         `yaml:"type"`
	ParserCfg    parser.ParserConfig `yaml:"parser"`
	FilterCfg filter.FilterConfig `yaml:"filter"`
	// Labeller Labeller `yaml:"include_fields"`
	// stop forward the message in pipeline if set
}

func NewFormatter(config FormatterConfig) Formatter {
	switch config.Type {
	case "parser":
		return parser.NewParser(config.ParserCfg)
	case "filter":
		formatter, err := filter.NewFilter(config.FilterCfg)
		if err != nil {
			log.Fatalf("Error when create filter: %s:", err)
		}
		return formatter
	default:
		panic("Invalid Formatter Type:" + config.Type)
	}
}