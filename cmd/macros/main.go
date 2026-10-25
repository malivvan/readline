// Command macros shows keyboard macros.
//
// Readline can replay recorded key sequences, and binds can be macros: instead
// of naming a command, a bind can feed raw input to the shell.
//
// Builtin recording keys (Emacs mode):
//
//	Ctrl-X (   start recording
//	Ctrl-X )   stop recording
//	Ctrl-X e   replay the last recorded macro
//
// In Vim mode, "q<register>" records and "@<register>" replays.
//
// This example also binds a static macro (Ctrl-X Ctrl-D) that types a date
// command, and adds a command that replays the last recorded macro from the
// application.
package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
)

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "macros> "
	})

	// A macro bind stores the keys to be fed to the shell as-is. Escape them
	// with inputrc.EscapeMacro when they contain control characters.
	if err := shell.Config.Bind(
		keymap.Emacs,
		inputrc.Unescape(`\C-x\C-d`),
		inputrc.EscapeMacro("date +%F"),
		true,
	); err != nil {
		fmt.Printf("bind: %v\n", err)
	}

	// The macro engine is reachable from the application: this command
	// replays whatever was last recorded with Ctrl-X ( ... Ctrl-X ).
	shell.Keymap.Register(map[string]func(){
		"replay-last-macro": func() {
			shell.Macros.RunLastMacro()
		},
	})

	if err := shell.Config.Bind(keymap.Emacs, inputrc.Unescape(`\C-x\C-l`), "replay-last-macro", false); err != nil {
		fmt.Printf("bind: %v\n", err)
	}

	fmt.Println("Ctrl-X ( record | Ctrl-X ) stop | Ctrl-X e replay | Ctrl-X Ctrl-D date macro")

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
