package console

import (
	"log"
	"encoding/json"
)

func Execute(recordCh chan map[string]interface{}) {
	for {
		doc := <-recordCh
		docPretty, _ := json.MarshalIndent(doc, "", "\t")
		log.Printf("%s\n", docPretty)
	}
}
