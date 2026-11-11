package hotkey

import (
	"fmt"
	"os"
	"os/signal"
)

// watchInterrupt listens for Ctrl+C and wakes the Windows message loop.
//
// PostThreadMessageW is needed because GetMessageW blocks while
// waiting for messages. The returned function stops the listener.
func watchInterrupt(threadID uintptr) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	done := make(chan struct{})
	exited := make(chan struct{})

	go func() {
		defer close(exited)

		select {
		case <-signals:
			result, _, err := postThreadMessageW.Call(
				threadID,
				wmQuit,
				0,
				0,
			)

			if result == 0 {
				fmt.Fprintf(
					os.Stderr,
					"PostThreadMessageW failed: %v\n",
					err,
				)
			}

		case <-done:
			return
		}
	}()

	return func() {
		signal.Stop(signals)
		close(done)
		<-exited
	}
}
