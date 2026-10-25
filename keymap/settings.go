package keymap

import (
	"os"
	"os/user"
	"sort"
	"strings"

	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/term"
)

var readlineOptions = map[string]interface{}{

	"autopairs": false,

	"autocomplete":               false,
	"completion-list-separator":  "--",
	"completion-selection-style": "\x1b[1;30m",

	"transient-prompt":          false,
	"usage-hint-always":         false,
	"history-autosuggest":       false,
	"multiline-column":          true,
	"multiline-column-numbered": false,

	"cursor-position-probe": true,
}

func revealBindsInputrc(commands []string, all map[string][]string) {
	for _, command := range commands {
		commandBinds := all[command]
		sort.Strings(commandBinds)

		if len(commandBinds) > 0 {
			for _, bind := range commandBinds {
				term.Printf("\"%s\": %s\n", bind, command)
			}
		}
	}
}

func revealBindsReadable(commands []string, all map[string][]string) {
	for _, command := range commands {
		commandBinds := all[command]
		sort.Strings(commandBinds)

		switch {
		case len(commandBinds) == 0:
		case len(commandBinds) > 5:
			firstBinds := make([]string, 0, 5)

			for i := range 5 {
				firstBinds = append(firstBinds, "\""+commandBinds[i]+"\"")
			}

			bindsStr := strings.Join(firstBinds, ", ")
			term.Printf("%s can be found on %s ...\n", command, bindsStr)

		default:
			var firstBinds []string

			for _, bind := range commandBinds {
				firstBinds = append(firstBinds, "\""+bind+"\"")
			}

			bindsStr := strings.Join(firstBinds, ", ")
			term.Printf("%s can be found on %s\n", command, bindsStr)
		}
	}
}

func (m *Engine) ReloadConfig(opts ...inputrc.Option) (err error) {

	m.loadBuiltinOptions()
	m.loadBuiltinBinds()

	user, _ := user.Current()

	_ = inputrc.UserDefault(user, m.config, inputrc.WithApp("go"))

	defaults := make([]inputrc.Option, 0, 2+len(opts))
	defaults = append(defaults,
		inputrc.WithMode("emacs"),
		inputrc.WithTerm(os.Getenv("TERM")),
	)

	opts = append(defaults, opts...)

	err = inputrc.UserDefault(user, m.config, opts...)
	if err != nil {
		return err
	}

	m.overrideBindsSpecial()

	switch m.config.GetString("editing-mode") {
	case "emacs":
		m.main = Emacs
	case "vi":
		m.main = ViInsert
	}

	return nil
}

func (m *Engine) loadBuiltinBinds() {

	for seq, bind := range emacsKeys {
		m.config.Binds[string(Emacs)][seq] = bind
	}

	for seq, bind := range vicmdKeys {
		m.config.Binds[string(ViCommand)][seq] = bind
		m.config.Binds[string(ViMove)][seq] = bind
		m.config.Binds[string(Vi)][seq] = bind
	}

	for seq, bind := range viinsKeys {
		m.config.Binds[string(ViInsert)][seq] = bind
	}

	m.config.Binds[string(Visual)] = visualKeys
	m.config.Binds[string(ViOpp)] = vioppKeys
	m.config.Binds[string(MenuSelect)] = menuselectKeys
	m.config.Binds[string(Isearch)] = menuselectKeys

	for _, keymap := range m.config.Binds {
		keymap[inputrc.Unescape(`\C-C`)] = inputrc.Bind{Action: "abort"}
	}
}

func (m *Engine) loadBuiltinOptions() {
	for name, value := range readlineOptions {
		if val := m.config.Get(name); val == nil {
			_ = m.config.Set(name, value)
		}
	}
}

func (m *Engine) overrideBindsSpecial() {

	if m.config.GetBool("disable-completion") {
		for _, keymap := range m.config.Binds {
			for seq, bind := range keymap {
				switch bind.Action {
				case "complete", "menu-complete", "possible-completions":
					keymap[seq] = inputrc.Bind{Action: "self-insert"}
				}
			}
		}
	}
}
