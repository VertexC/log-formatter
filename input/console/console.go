package console

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/VertexC/log-formatter/input"
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
		time.Sleep(time.Duration(1) * time.Second)
		fmt.Printf(">")
		text, _ := reader.ReadString('\n')
		console.docCh <- map[string]interface{}{"message": text}
	}
}
