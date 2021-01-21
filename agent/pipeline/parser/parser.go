package parser

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/config"
	"github.com/VertexC/log-formatter/agent/pipeline"
	"github.com/VertexC/log-formatter/agent/pipeline/protocol"
	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Component string   `yaml:"component"`
	Regexprs  []string `yaml:"regexprs"`
}

type ParserConfig struct {
	Base config.ConfigBase
	// ComponentsRegex is regexpr with components as groupname. To discard a component, add _ behind groupname like (?P<foo_>)
	ComponentsRegex string `yaml:"components_regex"`
	// Labels further extract label form certain component based on regexpr
	Labels []Label `yaml:"labels"`
	// TargetField specify which field to parse
	TargetField string `yaml:"target_field"`
	// ErrTolerant if set, parser will return err if fails to parse components
	ErrTolerant bool `yaml:"error_tolerant"`
}

type Parser struct {
	config ParserConfig
	logger *util.Logger
}

func init() {
	pipeline.Register("parser", New)
}

// TODO: finish regexpr compile jobs while new parser
func New(content interface{}) (protocol.Formatter, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to convert config to MapStr")
	}
	config := ParserConfig{
		Base: config.ConfigBase{
			Content:          contentMapStr,
			MandantoryFields: []string{"components_regex", "target_field"},
		},
		ErrTolerant: false,
	}
	if err := util.YamlConvert(contentMapStr, &config); err != nil {
		return nil, err
	}
	parser := &Parser{
		config: config,
		logger: util.NewLogger("pipeline-parser"),
	}
	return parser, nil
}

func (parser *Parser) Format(doc map[string]interface{}) (map[string]interface{}, error) {
	target, exist := doc[parser.config.TargetField]
	if !exist {
		if parser.config.ErrTolerant {
			return doc, nil
		} else {
			return doc, fmt.Errorf("Target field %s not found", parser.config.TargetField)
		}
	}
	componentMap, err := util.SubMatchMapRegex(parser.config.ComponentsRegex, target.(string))
	if err != nil {
		parser.logger.Error.Printf("Error occurs while get compponents: %s\n", err)
		if !parser.config.ErrTolerant {
			return doc, fmt.Errorf("Failed to parse %+v", doc)
		}
	}

	for _, label := range parser.config.Labels {
		if component, ok := componentMap[label.Component]; !ok {
			parser.logger.Error.Printf("Componet %s not found in matchMap %+v: %s\n", label.Component, componentMap, err)
		} else {
			for _, regex := range label.Regexprs {
				labelMap, err := util.SubMatchMapRegex(regex, component)
				if err != nil {
					parser.logger.Warning.Printf("Error occurs while get labels: %s\n", err)
					continue
				}
				for key, val := range labelMap {
					if key[len(key)-1:] == "_" {
						continue
					}
					doc[key] = val
				}
			}
		}
	}

	for key, val := range componentMap {
		if key[len(key)-1:] == "_" {
			continue
		}
		doc[key] = val
	}

	return doc, nil
}
