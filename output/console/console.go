package console

import (
	"log"
)

func Execute(recordCh chan interface{}) {
	for {
		record := <-recordCh
		log.Printf("[Get Message] +v%\n", record)
	}
}
