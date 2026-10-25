// Command keymaps shows how to extend the shell with your own commands and
// binds, and how to switch editing modes at runtime.
//
// Bindings added here:
//
//	Ctrl-T          insert the current date and time at the cursor
//	Ctrl-X Ctrl-U   uppercase the whole line
//	Ctrl-G          toggle between Emacs and Vi insert mode
//	Ctrl-X Ctrl-S   print the line, its length and the cursor position
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
)

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		mode := string(shell.Keymap.Main())
		if shell.Keymap.IsEmacs() {
			mode = string(keymap.Emacs)
		}

		return color.FgCyan + mode + color.Reset + "> "
	})

	// Commands are plain func() values. They read and mutate the editor
	// through the shell's accessors, which never expose invalid positions.
	shell.Keymap.Register(map[string]func(){
		"insert-date": func() {
			shell.Cursor().InsertAt([]rune(time.Now().Format(time.RFC3339))...)
		},

		"uppercase-line": func() {
			shell.Line().Set([]rune(strings.ToUpper(string(*shell.Line())))...)
			shell.Cursor().Set(shell.Line().Len())
		},

		"toggle-mode": func() {
			if shell.Keymap.IsEmacs() {
				shell.Keymap.SetMain(keymap.ViInsert)
			} else {
				shell.Keymap.SetMain(keymap.Emacs)
			}

			shell.Keymap.UpdateCursor()
		},

		"describe-line": func() {
			shell.Hint.SetTemporary(fmt.Sprintf(
				"%s%d characters, cursor at %d%s",
				color.Dim, shell.Line().Len(), shell.Cursor().Pos(), color.Reset,
			))
		},
	})

	// Binds are declared with the same syntax as an .inputrc file. Sequences
	// must be unescaped before being registered.
	binds := []struct {
		keymap string
		key    string
		action string
	}{
		{keymap.Emacs, `\C-t`, "insert-date"},
		{keymap.Emacs, `\C-Xu`, "uppercase-line"},
		{keymap.Emacs, `\C-g`, "toggle-mode"},
		{keymap.Emacs, `\C-x\C-s`, "describe-line"},
		{keymap.ViInsert, `\C-t`, "insert-date"},
		{keymap.ViInsert, `\C-g`, "toggle-mode"},
	}

	for _, bind := range binds {
		if err := shell.Config.Bind(bind.keymap, inputrc.Unescape(bind.key), bind.action, false); err != nil {
			fmt.Printf("bind %s: %v\n", bind.key, err)
		}
	}

	// The main keymap can also be selected from the start.
	shell.Keymap.SetMain(keymap.Emacs)

	fmt.Println("Ctrl-T date | Ctrl-X Ctrl-U uppercase | Ctrl-G toggle Emacs/Vi | Ctrl-X Ctrl-S describe")

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

		fmt.Printf("accepted: %q\n", line)
	}
}
