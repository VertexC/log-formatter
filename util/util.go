package util

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// ExitControl provides a block over SIGTERM (Ctrl-C)
func ExitControl() struct{} {
	doneCh := make(chan struct{})

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// go routine to catch signal interrupt
	go func() {
		select {
		case <-sigterm:
			fmt.Println("terminating: via signal")
			doneCh <- struct{}{}
		}
	}()

	return <-doneCh
}
