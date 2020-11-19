package console

import (
	"encoding/json"
	"log"
)

type Console struct {
	docCh chan map[string]interface{}
}

func NewConsoleOutput(docCh chan map[string]interface{}) *Console {
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
