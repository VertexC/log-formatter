package forward

import (
	"github.com/VertexC/log-formatter/pipeline"
)

func init() {
	pipeline.Register("forwarder", NewForwarder)
}

type Forwarder struct{}

func NewForwarder(content interface{}) (pipeline.Formatter, error) {
	f := &Forwarder{}
	return f, nil
}

func (f *Forwarder) Format(doc map[string]interface{}) (map[string]interface{}, error) {
	return doc, nil
}
