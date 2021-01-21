package util

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SigControl provides a block over SIGTERM(Ctrl-C) and SIGINT
// which provides a graceful shutdown control when pod is killed
func SigControl(handler func()) struct{} {
	doneCh := make(chan struct{})

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// go routine to catch signal interrupt
	go func() {
		select {
		case <-sigterm:
			fmt.Println("terminating: via signal")
			handler()
			doneCh <- struct{}{}
		}
	}()

	return <-doneCh
}
