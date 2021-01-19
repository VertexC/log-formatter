package forward

import (
	"github.com/VertexC/log-formatter/agent/pipeline"
	"github.com/VertexC/log-formatter/agent/pipeline/protocol"
)

func init() {
	pipeline.Register("forwarder", New)
}

type Forwarder struct{}

func New(content interface{}) (protocol.Formatter, error) {
	f := &Forwarder{}
	return f, nil
}

func (f *Forwarder) Format(doc map[string]interface{}) (map[string]interface{}, error) {
	return doc, nil
}
