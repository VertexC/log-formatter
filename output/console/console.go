package console

import (
	"fmt"
	"github.com/VertexC/log-formatter/output"
)

func init() {
	output.Register("console", NewConsole)
}

type Console struct {
	docCh chan map[string]interface{}
}

func NewConsole(content interface{}, docCh chan map[string]interface{}) (output.Output, error) {
	console := &Console{
		docCh: docCh,
	}
	return console, nil
}

func (console *Console) Run() {
	for doc := range console.docCh {
		fmt.Printf("%+v\n", doc)
	}
}
