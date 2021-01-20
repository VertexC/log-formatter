package console

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/output"
	"github.com/VertexC/log-formatter/agent/output/protocol"
)

func init() {
	output.Register("console", NewConsole)
}

type Console struct{}

func NewConsole(content interface{}) (protocol.Output, error) {
	console := &Console{}
	return console, nil
}

func (console *Console) Run() {}

func (console *Console) Stop() {}

func (console *Console) Send(doc map[string]interface{}) {
	fmt.Printf("%+v\n", doc)
}
