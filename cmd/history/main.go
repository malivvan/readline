// Command history shows the history sources API.
//
// The shell always starts with an in-memory history source. Adding a file
// backed source replaces it, and any number of additional sources can be
// registered: the shell can then cycle through them with the
// "history-source-next"/"history-source-prev" commands (Ctrl-X n / Ctrl-X p
// in Emacs mode, or the bound keys listed by Ctrl-X Ctrl-P).
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
)

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "history[" + shell.History.Name() + "]> "
	})

	// A file-backed history shared by all sessions of this program.
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
		os.Exit(1)
	}

	historyFile := filepath.Join(home, ".readline-cmd-history")

	// A history source is opened from an existing file, so create it on the
	// first run before opening it.
	source, err := readline.NewHistoryFromFile(historyFile)
	if err != nil {
		if file, createErr := os.OpenFile(historyFile, os.O_CREATE|os.O_WRONLY, 0o600); createErr == nil {
			_ = file.Close()

			source, err = readline.NewHistoryFromFile(historyFile)
		}
	}

	// Adding the first source removes the default in-memory one. When the file
	// cannot be used, keep the in-memory history instead of failing.
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %v (using in-memory history)\n", err)
	} else {
		shell.History.Add("home", source)
	}

	// Additional sources live next to each other, and the shell provides
	// commands to cycle through them.
	shell.History.AddFromFile("project", ".readline-cmd-project-history")

	// Sources can also be removed at runtime.
	shell.History.Add("temporary", readline.NewInMemoryHistory())

	// Bind the source-cycling commands: Ctrl-X n / Ctrl-X p.
	for key, action := range map[string]string{
		`\C-xn`: "history-source-next",
		`\C-xp`: "history-source-prev",
	} {
		if err := shell.Config.Bind(keymap.Emacs, inputrc.Unescape(key), action, false); err != nil {
			fmt.Printf("bind %s: %v\n", key, err)
		}
	}

	fmt.Println("history sources: home (file), project (file), temporary (memory)")
	fmt.Println("Ctrl-X n / Ctrl-X p switch sources; type \"drop\" to remove the temporary one")

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

		switch line {
		case "drop":
			shell.History.Delete("temporary")
			fmt.Println("temporary history removed")
		case "sources":
			fmt.Println("current:", shell.History.Name(), "position:", shell.History.Pos())
		default:
			fmt.Printf("accepted: %q\n", line)
		}
	}
}
