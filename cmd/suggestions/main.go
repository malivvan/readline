// Command suggestions shows inline (ghost text) suggestions.
//
// The shell stores an application-owned suggestion and renders it as dim text
// after the cursor, as long as the current line is a prefix of it. The
// "forward-char" binding (Ctrl-F or the right arrow) accepts it.
//
// Here the suggestion is the first known command extending what has been typed.
// The shell recomputes it by wrapping the builtin insertion/deletion commands:
// a key event is just a command of your own.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

var knownCommands = []string{
	"describe", "deploy", "destroy", "help", "list", "logs", "quit", "restart", "status",
}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "suggestions> "
	})

	updateSuggestion := func() {
		line := string(*shell.Line())

		suggestion := ""
		if line != "" {
			for _, command := range knownCommands {
				if strings.HasPrefix(command, line) && command != line {
					suggestion = command

					break
				}
			}
		}

		shell.SetInlineSuggestion(suggestion)
	}

	// Grab the builtin commands before replacing them, so the wrappers can
	// still perform the actual insertion/deletion.
	commands := shell.Keymap.Commands()
	insert := commands["self-insert"]
	backspace := commands["backward-delete-char"]

	shell.Keymap.Register(map[string]func(){
		"self-insert": func() {
			insert()
			updateSuggestion()
		},
		"backward-delete-char": func() {
			backspace()
			updateSuggestion()
		},
		// Ctrl-G accepts the whole suggestion, wherever the cursor is.
		"accept-inline-suggestion": func() {
			suggestion := shell.GetInlineSuggestion()
			if suggestion == "" {
				return
			}

			shell.Line().Set([]rune(suggestion)...)
			shell.Cursor().Set(shell.Line().Len())
			shell.ClearInlineSuggestion()
		},
	})

	if err := shell.Config.Bind("emacs", "\x07", "accept-inline-suggestion", false); err != nil {
		fmt.Printf("bind: %v\n", err)
	}

	fmt.Println(color.Dim + "type \"d\" and press the right arrow or Ctrl-G to accept the suggestion" + color.Reset)

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

		if line == "quit" {
			return
		}

		fmt.Printf("accepted: %q\n", line)
	}
}
