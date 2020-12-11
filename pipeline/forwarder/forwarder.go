package forward

import (
	"github.com/VertexC/log-formatter/pipeline"
	"github.com/VertexC/log-formatter/util"
)

func init() {
	pipeline.Register("forwarder", NewForwarder)
}

type Forwarder struct{}

func NewForwarder(content interface{}) (pipeline.Formatter, error) {
	f := &Forwarder{}
	return f, nil
}

func (f *Forwarder) Format(doc util.Doc) (util.Doc, error) {
	return doc, nil
}
