package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/VertexC/log-formatter/agent/input/protocol"
)

type Console struct {
	reader *bufio.Reader
}

func New(content interface{}) (protocol.Input, error) {
	console := &Console{
		reader: bufio.NewReader(os.Stdin),
	}
	return console, nil
}

func (console *Console) Run() {}

func (console *Console) Emit() map[string]interface{} {
	time.Sleep(time.Duration(1) * time.Second)
	fmt.Printf(">")
	text, _ := console.reader.ReadString('\n')
	return map[string]interface{}{"message": text}
}
