// Command completions shows the completion API: plain values, described and
// styled values, tags, usage hints, error messages, and live (asynchronous)
// updates while a completion menu is open.
//
// Type "show " and press Tab to complete a host name; the list grows in the
// background, and an open menu refreshes itself when new hosts are discovered.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

var commands = []string{"help", "hosts", "show", "quit"}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "completions> "
	})

	var (
		mu    sync.Mutex
		hosts = []string{"alpha.local", "beta.local"}
	)

	snapshot := func() []string {
		mu.Lock()
		defer mu.Unlock()

		return append([]string(nil), hosts...)
	}

	// A background producer grows the cache and asks the shell to rebuild the
	// open completion menu. Any candidate the user selected is preserved.
	go func() {
		for _, host := range []string{"gamma.local", "delta.local", "epsilon.local"} {
			time.Sleep(4 * time.Second)

			mu.Lock()
			hosts = append(hosts, host)
			mu.Unlock()

			shell.RefreshCompletions()
		}
	}()

	shell.Completer = func(line []rune, cursor int) readline.Completions {
		fields := strings.Fields(string(line[:cursor]))

		switch {
		case len(fields) == 0:
			// Several tags and styles in a single container. Only the
			// "commands" tag is displayed as a long list.
			return readline.CompleteValues(commands...).
				Tag("commands").
				Style(color.FgBlue).
				DisplayList().
				Usage("show <host> | hosts | help | quit")

		case fields[0] != "show":
			// Messages are printed below the input line, in the hint area.
			return readline.CompleteMessage("%q does not take arguments", fields[0])
		}

		found := snapshot()
		if len(found) == 0 {
			return readline.CompleteMessage("no hosts discovered yet")
		}

		// Described values are aligned and separated by the configured
		// "completion-list-separator" when the list is displayed.
		values := make([]string, 0, len(found)*2)
		for _, host := range found {
			values = append(values, host, "discovered host")
		}

		return readline.CompleteValuesDescribed(values...).
			Tag("hosts").
			JustifyDescriptions().
			DisplayList()
	}

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
			return
		case fields[0] == "hosts":
			fmt.Println(strings.Join(snapshot(), " "))
		case fields[0] == "help":
			fmt.Println("commands:", strings.Join(commands, ", "))
		default:
			fmt.Printf("running %q\n", line)
		}
	}
}
