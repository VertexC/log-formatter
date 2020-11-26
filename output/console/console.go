package console

import (
	"encoding/json"
	"github.com/VertexC/log-formatter/util"
	"log"
)

type Console struct {
	docCh chan util.Doc
}

func NewConsoleOutput() *Console {
	return &Console{
		docCh: make(chan util.Doc, 1000),
	}
}

func (console *Console) Append(doc util.Doc) {
	console.docCh <- doc
}

func (console *Console) Run() {
	for doc := range console.docCh {
		docPretty, _ := json.MarshalIndent(doc, "", "\t")
		log.Printf("%s\n", docPretty)
	}
}
