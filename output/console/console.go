package console

import (
	"log"
)

func Execute(recordCh chan []interface{}) {
	for {
		records := <- recordCh
		for _, record := range records {
			log.Printf("[Get Message] +v%\n", record)
		}
	}
}