package formatter

import (
	"github.com/VertexC/log-formatter/pipeline/formatter/parser"
	"github.com/VertexC/log-formatter/pipeline/formatter/labeller"
	"github.com/VertexC/log-formatter/pipeline/formatter/filter"
)

type Formatter interface {
	Setup()
	Format(map[string]interface{}) map[string]interface{}, error
	BreakOnErr() bool
}

type FormatterConfig struct {
	Type          string         `yaml:"type"`
	ParserCfg    formatter.ParserConfig `yaml:"parser"`
	// FilterCfg formatter.FilterConfig `yaml:"filter"`
	// Labeller formatter.Labeller `yaml:"include_fields"`
	// stop forward the message in pipeline if set
}

func NewFormatter(fmtCfg FormatterConfig) Formatter {
	switch fmtCfg.Type {
	case "parser":
		return parser.NewParser(fmtCfg.ParserCfg)
	default:
		panic("Invalid Formatter Type:", fmtCfg.Type)
	}
}