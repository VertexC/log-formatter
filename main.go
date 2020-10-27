package main

import (
	"fmt"

	"github.com/VertexC/log-formatter/formatter"
	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/output"
)


func main() {
	fmt.Println(formatter.Version)
	records := input.EsSearch()
	output.EsUpdate(records)
}