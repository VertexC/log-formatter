package general_test

import (
	"github.com/VertexC/log-formatter/formatter/general"
	"reflect"
	"testing"
)

var regex = `^(?P<id>[0-9]+)\s+(?P<msg>.*?)$`
var invalidRegex = `^(?P<id>[a-z]+)\s+(?P<msg>.*?)$`
var labels = []general.Label{{Component: "msg", Regexprs: []string{`(?P<foo>foo)`}}}
var invalidLabels = []general.Label{{Component: "msg", Regexprs: []string{`(?P<foo>???)`}}}

var msg = "123 hello foo"

var plainConfig = general.Config{
	Regex:  regex,
	Labels: []general.Label{},
}

// TODO: tear up test log file in the end
func TestComponents(t *testing.T) {
	expected := map[string]interface{}{"id": "123", "msg": "hello foo"}
	config := general.Config{Regex: regex, Labels: []general.Label{}}
	formatter := new(general.Formatter)
	formatter.SetConfig(config)
	formatter.Init("", false)
	formatter.DiscardLog()
	result := formatter.Format(msg)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Got %+v but expected %+v", result, expected)
	} else {
		t.Fatalf("Got %+v but expected %+v", result, expected)
	}
}

func TestLabels(t *testing.T) {

}

func TestInvalidComponents(t *testing.T) {

}

func TestInvalidLabels(t *testing.T) {

}
