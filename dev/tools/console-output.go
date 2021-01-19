package main

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/output/protocol"
)

type Console struct{}

func New(content interface{}) (protocol.Output, error) {
	console := &Console{}
	return console, nil
}

func (console *Console) Run() {}

func (console *Console) Send(doc map[string]interface{}) {
	fmt.Printf("%+v\n", doc)
}
