//go:build unix
// +build unix

package display

import (
	"os"
	"os/signal"
	"syscall"
)

func WatchResize(eng *Engine) chan<- bool {
	done := make(chan bool, 1)

	resizeChannel := make(chan os.Signal, 1)
	signal.Notify(resizeChannel, syscall.SIGWINCH)

	go func() {
		defer signal.Stop(resizeChannel)

		for {
			select {
			case <-resizeChannel:

				eng.completer.RequestRegen()
				eng.keys.RequestRefresh()
			case <-done:
				return
			}
		}
	}()

	return done
}
