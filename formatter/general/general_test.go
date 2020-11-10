package general_test

import (
	"github.com/VertexC/log-formatter/formatter/general"
	"testing"

	"github.com/stretchr/testify/assert"
)

var regex = `^(?P<id>[0-9]+)\s+(?P<msg>.*?)$`
var invalidRegex = `^(?P<id>[a-z]+)\s+(?P<msg>.*?)$`
var labels = []general.Label{{Component: "msg", Regexprs: []string{`(?P<action>.*?)\s+foo`}}}
var invalidLabels = []general.Label{{Component: "msg", Regexprs: []string{`(?P<action>)\s+bar`}}}

var msg = "123 hello foo"

var plainConfig = general.Config{
	Regex:  regex,
	Labels: []general.Label{},
}

func doCheck(config general.Config, expected map[string]interface{}, t *testing.T) {
	formatter := new(general.Formatter)
	formatter.SetConfig(config)
	formatter.Init("", false)
	formatter.DiscardLog()
	result := formatter.Format(msg)
	assert.Equal(t, result, expected)
}

// TODO: tear up test log file in the end
func TestComponents(t *testing.T) {
	expected := map[string]interface{}{"id": "123", "msg": "hello foo"}
	config := general.Config{Regex: regex, Labels: []general.Label{}}
	doCheck(config, expected, t)
}

func TestLabels(t *testing.T) {
	expected := map[string]interface{}{"id": "123", "msg": "hello foo", "action": "hello"}
	config := general.Config{Regex: regex, Labels: labels}
	doCheck(config, expected, t)
}

func TestInvalidComponents(t *testing.T) {
	config := general.Config{Regex: invalidRegex, Labels: []general.Label{}}
	doCheck(config, nil, t)
}

func TestInvalidLabels(t *testing.T) {
	expected := map[string]interface{}{"id": "123", "msg": "hello foo"}
	config := general.Config{Regex: regex, Labels: invalidLabels}
	doCheck(config, expected, t)
}
