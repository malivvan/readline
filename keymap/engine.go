package keymap

import (
	"sort"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
)

type Engine struct {
	local        Mode
	main         Mode
	prefixed     inputrc.Bind
	active       inputrc.Bind
	pending      []inputrc.Bind
	skip         bool
	isCaller     bool
	nonIncSearch bool

	keys       *core.Keys
	iterations *core.Iterations
	config     *inputrc.Config
	commands   map[string]func()
}

type Mode string

const (
	Emacs         = "emacs"
	EmacsMeta     = "emacs-meta"
	EmacsCtrlX    = "emacs-ctlx"
	EmacsStandard = "emacs-standard"

	ViInsert  = "vi-insert"
	Vi        = "vi"
	ViCommand = "vi-command"
	ViMove    = "vi-move"
	Visual    = "vi-visual"
	ViOpp     = "vi-opp"

	Isearch    = "isearch"
	MenuSelect = "menu-select"
)

func NewEngine(keys *core.Keys, i *core.Iterations, opts ...inputrc.Option) (*Engine, *inputrc.Config) {
	modes := &Engine{
		main:       Emacs,
		keys:       keys,
		iterations: i,
		config:     inputrc.NewDefaultConfig(),
		commands:   make(map[string]func()),
	}

	_ = modes.ReloadConfig(opts...)

	return modes, modes.config
}

func (m *Engine) ActiveCommand() inputrc.Bind {
	return m.active
}

func (m *Engine) Commands() map[string]func() {
	return m.commands
}

func (m *Engine) InputIsTerminator() bool {
	terminators := []string{
		inputrc.Unescape(`\C-G`),
		inputrc.Unescape(`\C-]`),
	}

	binds := make(map[string]inputrc.Bind)

	for _, sequence := range terminators {
		binds[sequence] = inputrc.Bind{Action: "abort", Macro: false}
	}

	bind, _, _, _ := m.dispatchKeys(binds)

	return bind.Action == "abort"
}

func (m *Engine) IsEmacs() bool {
	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		return true
	default:
		return false
	}
}

func (m *Engine) Local() Mode {
	return m.local
}

func (m *Engine) Main() Mode {
	return m.main
}

func (m *Engine) NonIncrementalSearchStart() {
	m.nonIncSearch = true
}

func (m *Engine) NonIncrementalSearchStop() {
	m.nonIncSearch = false
}

func (m *Engine) PrintBinds(keymap string, inputrcFormat bool) {
	commands := make([]string, 0, len(m.commands))

	for command := range m.commands {
		commands = append(commands, command)
	}

	sort.Strings(commands)

	binds := m.config.Binds[keymap]
	if binds == nil {
		return
	}

	allBinds := make(map[string][]string)

	for _, command := range commands {
		for key, bind := range binds {
			if bind.Action != command {
				continue
			}

			commandBinds := allBinds[command]
			commandBinds = append(commandBinds, inputrc.Escape(key))
			allBinds[command] = commandBinds
		}
	}

	if inputrcFormat {
		revealBindsInputrc(commands, allBinds)
	} else {
		revealBindsReadable(commands, allBinds)
	}
}

func (m *Engine) Register(commands map[string]func()) {
	for name, command := range commands {
		m.commands[name] = command
	}
}

func (m *Engine) ResetLocal() {
	m.local = ""
	m.UpdateCursor()
}

func (m *Engine) SetLocal(keymap string) {
	m.local = Mode(keymap)
	m.UpdateCursor()
}

func (m *Engine) SetMain(keymap string) {
	m.main = Mode(keymap)
	m.UpdateCursor()
}

func (m *Engine) UpdateCursor() {
	switch m.local {
	case ViOpp:
		m.PrintCursor(ViOpp)
		return
	case Visual:
		m.PrintCursor(Visual)
		return
	}

	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		m.PrintCursor(Emacs)
	case ViInsert:
		m.PrintCursor(ViInsert)
	case ViCommand, ViMove, Vi:
		m.PrintCursor(ViCommand)
	}
}
