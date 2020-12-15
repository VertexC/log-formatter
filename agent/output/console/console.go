package console

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/output"
)

func init() {
	output.Register("console", NewConsole)
}

type Console struct{}

func NewConsole(content interface{}) (output.Output, error) {
	console := &Console{}
	return console, nil
}

func (console *Console) Run() {}

func (console *Console) Send(doc map[string]interface{}) {
	fmt.Printf("%+v\n", doc)
}
