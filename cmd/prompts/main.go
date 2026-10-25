// Command prompts shows every prompt handler and the two ways of logging
// output into an interactive shell.
//
// Besides the primary prompt, readline supports:
//   - a right-aligned prompt (RPROMPT),
//   - a secondary prompt, used for continuation lines,
//   - a transient prompt, printed in place of the primary prompt once the line
//     is accepted (enabled with the "transient-prompt" option),
//   - a tooltip, recomputed for the word under the cursor on every keypress.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

func main() {
	shell := readline.NewShell()

	// The transient prompt only replaces the primary prompt when enabled.
	_ = shell.Config.Set("transient-prompt", true)

	shell.Prompt.Primary(func() string {
		return color.Bold + color.FgCyan + "prompts" + color.Reset + "> "
	})

	shell.Prompt.Secondary(func() string {
		return color.Dim + "... " + color.Reset
	})

	shell.Prompt.Right(func() string {
		return color.Dim + time.Now().Format("15:04:05") + color.Reset
	})

	shell.Prompt.Transient(func() string {
		return color.Dim + "prompts (done)" + color.Reset
	})

	// The tooltip receives the word under the cursor.
	shell.Prompt.Tooltip(func(word string) string {
		word = strings.TrimSpace(word)
		if word == "" {
			return ""
		}

		return color.Dim + "word: " + color.Reset + color.FgYellow + word + color.Reset
	})

	fmt.Println("commands: log <msg> | transient <msg> | ticker | quit")

	var ticker atomic.Bool

	for {
		line, err := shell.Readline()

		switch {
		case errors.Is(err, readline.ErrInterrupt):
			continue

		case errors.Is(err, io.EOF):
			return

		case err != nil:
			fmt.Printf("error: %v\n", err)
			return
		}

		switch fields := strings.Fields(line); {
		case len(fields) == 0:
		case fields[0] == "quit":
			ticker.Store(false)
			return

		case fields[0] == "log":
			// Printed below the current line; a fresh prompt is redrawn.
			shell.Printf("%s[log]%s %s", color.FgBlue, color.Reset, strings.Join(fields[1:], " "))

		case fields[0] == "transient":
			// The current line is cleared, the message is printed in its
			// place, and the prompt/line are redrawn below it.
			shell.PrintTransientf("%s[transient]%s %s", color.FgMagenta, color.Reset, strings.Join(fields[1:], " "))

		case fields[0] == "ticker":
			if ticker.CompareAndSwap(false, true) {
				go func() {
					for ticks := 1; ticker.Load(); ticks++ {
						time.Sleep(2 * time.Second)
						shell.PrintTransientf("%s[ticker]%s notification %d", color.FgGreen, color.Reset, ticks)
					}
				}()
			}

		default:
			fmt.Printf("accepted: %q\n", line)
		}
	}
}
