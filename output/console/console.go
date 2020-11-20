package console

import (
	"encoding/json"
	"github.com/VertexC/log-formatter/util"
	"log"
)

type Console struct {
	docCh chan util.Doc
}

func NewConsoleOutput(docCh chan util.Doc) *Console {
	return &Console{
		docCh: docCh,
	}
}

func (console *Console) Run() {
	for doc := range console.docCh {
		docPretty, _ := json.MarshalIndent(doc, "", "\t")
		log.Printf("%s\n", docPretty)
	}
}
