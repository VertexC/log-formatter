package parser

import (
	"fmt"
	"github.com/VertexC/log-formatter/util"
	"regexp"
)

type Label struct {
	Component string   `yaml:"component"`
	Regexprs  []string `yaml:"regexprs"`
}

type ParserConfig struct {
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

// TODO: finish regexpr compile jobs while new parser
func NewParser(parserCfg ParserConfig) (parser *Parser) {
	parser = new(Parser)
	parser.config = parserCfg
	parser.logger = util.NewLogger("pipeline-parser")
	return
}

func (parser *Parser) Format(doc util.Doc) (util.Doc, error) {
	target, exist := doc[parser.config.TargetField]
	if !exist {
		if parser.config.ErrTolerant {
			return doc, nil
		} else {
			return doc, fmt.Errorf("Target field %s not found", parser.config.TargetField)
		}
	}
	componentMap, err := reSubMatchMap(parser.config.ComponentsRegex, target.(string))
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
				labelMap, err := reSubMatchMap(regex, component)
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

func reSubMatchMap(reg string, str string) (map[string]string, error) {
	r := regexp.MustCompile(reg)
	match := r.FindStringSubmatch(str)
	groupNames := r.SubexpNames()
	if len(match) != len(groupNames) {
		return nil, fmt.Errorf("Failed to extract groups %s from %s with %s, match:%s", r.SubexpNames(), str, reg, match)
	}
	subMatchMap := map[string]string{}
	for i, name := range groupNames {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap, nil
}
