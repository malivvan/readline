// Command options shows how to configure the shell programmatically.
//
// Options are the same as those of an .inputrc file: they can be set with
// Config.Set, read back with Config.Get/GetBool/GetString/GetInt, and loaded
// from a file with inputrc.ParseFile. A user's ~/.inputrc is loaded
// automatically by every new shell.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
)

// options are the library-specific options, on top of every standard readline
// option supported through .inputrc files.
var options = []struct {
	name  string
	value interface{}
	help  string
}{
	{"autopairs", true, "insert the matching closing character"},
	{"autocomplete", true, "complete automatically while typing"},
	{"completion-ignore-case", true, "ignore case when completing"},
	{"usage-hint-always", true, "always show completion usage hints"},
	{"history-autosuggest", true, "show history suggestions after the cursor"},
	{"transient-prompt", false, "replace the primary prompt once accepted"},
	{"completion-list-separator", "--", "separator used in completion lists"},
}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "options> "
	})

	// Programmatic configuration.
	for _, option := range options {
		if err := shell.Config.Set(option.name, option.value); err != nil {
			fmt.Printf("set %s: %v\n", option.name, err)
		}
	}

	// Optional: load a per-application configuration file, if the user has
	// one. It is parsed into the very same configuration object, so its
	// options and binds take precedence over the ones set above.
	if file, err := os.UserHomeDir(); err == nil {
		_ = inputrc.ParseFile(
			file+"/.readline-cmd-inputrc",
			shell.Config,
			inputrc.WithApp("options"),
			inputrc.WithMode("emacs"),
		)
	}

	// A command reading the current configuration, to show live access.
	shell.Keymap.Register(map[string]func(){
		"show-options": func() {
			shell.Hint.SetTemporary(fmt.Sprintf(
				"%sautopairs=%v autocomplete=%v ignore-case=%v mode=%s%s",
				color.Dim,
				shell.Config.GetBool("autopairs"),
				shell.Config.GetBool("autocomplete"),
				shell.Config.GetBool("completion-ignore-case"),
				shell.Keymap.Main(),
				color.Reset,
			))
		},
	})

	if err := shell.Config.Bind(keymap.Emacs, inputrc.Unescape(`\C-x\C-o`), "show-options", false); err != nil {
		fmt.Printf("bind: %v\n", err)
	}

	fmt.Println("commands: options | toggle-autocomplete | reload | quit")
	fmt.Println("Ctrl-X Ctrl-O shows the current options; Ctrl-X Ctrl-R reloads the init file")

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
		case "quit":
			return

		case "options":
			names := make([]string, 0, len(options))
			for _, option := range options {
				names = append(names, fmt.Sprintf("  %-30s %v  # %s", option.name, option.value, option.help))
			}

			sort.Strings(names)
			fmt.Println(strings.Join(names, "\n"))

		case "toggle-autocomplete":
			value := !shell.Config.GetBool("autocomplete")
			_ = shell.Config.Set("autocomplete", value)
			fmt.Printf("autocomplete = %v\n", value)

		case "reload":
			if err := shell.Keymap.ReloadConfig(shell.Opts...); err != nil {
				fmt.Printf("reload: %v\n", err)
			} else {
				fmt.Println("configuration reloaded")
			}

		default:
			fmt.Printf("accepted: %q\n", line)
		}
	}
}
