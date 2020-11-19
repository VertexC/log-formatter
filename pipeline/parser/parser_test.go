package parser_test

import (
	"github.com/VertexC/log-formatter/pipeline/parser"
	"github.com/VertexC/log-formatter/util"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

var invalidRegex = `^(?P<id>[a-z]+)\s+(?P<msg>.*?)$`
var labels = []parser.Label{{Component: "msg", Regexprs: []string{}}}
var invalidLabels = []parser.Label{{Component: "msg", Regexprs: []string{`(?P<action>)\s+bar`}}}

// default config for test cases
var defaultConfigYml = `
components_regex: ^(?P<id>[0-9]+)\s+\[(?P<context>[A-Z]+)\]\s+(?P<msg>.*?)$
labels:
  - component: msg
    regexprs:
      - (?P<action>.*?)\s+foo
target_field: message
error_tolerant: true
`

var doc = map[string]interface{}{"message": "123 [HELLO] hello foo", "user": "vertexc"}

func init() {
	util.LogFile = ""
}

func createDefaultConfig() parser.ParserConfig {
	return parser.ParserConfig{
		ComponentsRegex: `^(?P<id>[0-9]+)\s+\[(?P<context>[A-Z]+)\]\s+(?P<msg>.*?)$`,
		Labels: []parser.Label{{
			Component: `msg`,
			Regexprs:  []string{`(?P<action>.*?)\s+foo`},
		},
		},
		TargetField: `message`,
		ErrTolerant: true,
	}
}

func TestParserConfigFromYaml(t *testing.T) {
	config := new(parser.ParserConfig)
	err := yaml.Unmarshal([]byte(defaultConfigYml), config)
	if err != nil {
		assert.Fail(t, "Cannot parse from yaml")
	}
	defaultCfg := createDefaultConfig()
	assert.Equal(t, defaultCfg, *config)
}

func TestParserFormatWithDefaultConfig(t *testing.T) {
	expectedDoc := map[string]interface{}{
		"id":      "123",
		"context": "HELLO",
		"msg":     "hello foo",
		"action":  "hello",
		"message": "123 [HELLO] hello foo",
		"user":    "vertexc",
	}
	defaultCfg := createDefaultConfig()
	p := parser.NewParser(defaultCfg)
	result, err := p.Format(doc)
	assert.Equal(t, err, nil)
	assert.Equal(t, expectedDoc, result)
}
