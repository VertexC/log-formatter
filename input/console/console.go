package console

import (
	"bufio"
	"github.com/VertexC/log-formatter/input"
	"os"
)

func init() {
	input.Register("console", NewConsole)
}

type Console struct {
	docCh chan map[string]interface{}
}

func NewConsole(content interface{}, docCh chan map[string]interface{}) (input.Input, error) {
	console := &Console{
		docCh: docCh,
	}
	return console, nil
}

func (console *Console) Run() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		console.docCh <- map[string]interface{}{"message": text}
	}
}
