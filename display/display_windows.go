//go:build windows
// +build windows

package display

import (
	"github.com/malivvan/readline/core"
)

func WatchResize(eng *Engine) chan<- bool {
	resizeChannel := core.GetTerminalResize(eng.keys)
	done := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-resizeChannel:

				eng.completer.GenerateCached()
				eng.Refresh()
			case <-done:
				return
			}
		}
	}()

	return done
}
