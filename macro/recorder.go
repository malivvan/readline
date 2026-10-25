package macro

import (
	"sort"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/term"
	"github.com/malivvan/readline/ui"
)

type Engine struct {
	recording  bool
	current    []rune
	currentKey rune
	macros     map[rune]string
	started    bool

	keys   *core.Keys
	hint   *ui.Hint
	status string
}

var validMacroKeys = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\""

func NewEngine(keys *core.Keys, hint *ui.Hint) *Engine {
	return &Engine{
		current: make([]rune, 0),
		macros:  make(map[rune]string),
		keys:    keys,
		hint:    hint,
	}
}

func RecordKeys(eng *Engine) {
	if !eng.recording {
		return
	}

	keys := core.MacroKeys(eng.keys)
	if len(keys) == 0 {
		return
	}

	if !eng.started {
		eng.current = append(eng.current, keys...)
	}

	eng.started = false

	eng.hint.Persist(eng.status + inputrc.EscapeMacro(string(eng.current)) + color.Reset)
}

func validMacroID(key rune) bool {
	for _, char := range validMacroKeys {
		if char == key {
			return true
		}
	}

	return false
}

func (e *Engine) PrintAllMacros() {
	macroIDs := make([]rune, 0, len(e.macros))

	for key := range e.macros {
		macroIDs = append(macroIDs, key)
	}

	sort.Slice(macroIDs, func(i, j int) bool {
		return macroIDs[i] < macroIDs[j]
	})

	for _, macro := range macroIDs {
		sequence := e.macros[macro]
		if sequence == "" {
			continue
		}

		if macro == 0 {
			macro = '"'
		}

		term.Printf("\"%s\": %s\n", string(macro), sequence)
	}
}

func (e *Engine) PrintLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	term.Printf("\n%s\n", e.macros[e.currentKey])
}

func (e *Engine) Recording() bool {
	return e.recording
}

func (e *Engine) RunLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	macro := inputrc.Unescape(e.macros[rune(0)])

	if len(macro) == 0 {
		return
	}

	e.keys.Feed(false, []rune(macro)...)
}

func (e *Engine) RunMacro(key rune) {
	if !validMacroID(key) && key != 0 {
		return
	}

	macro := inputrc.Unescape(e.macros[key])
	if len(macro) == 0 {
		return
	}

	e.keys.Feed(false, []rune(macro)...)
}

func (e *Engine) StartRecord(key rune) {
	switch {
	case validMacroID(key), key == 0:
		e.currentKey = key
	default:
		return
	}

	e.started = true
	e.recording = true
	e.status = color.Dim + "Recording macro: " + color.Bold
	e.hint.Persist(e.status)
}

func (e *Engine) StopRecord(keys ...rune) {
	e.recording = false

	e.hint.ResetPersist()

	if len(e.current) == 0 {
		return
	}

	e.current = append(e.current, keys...)
	macro := inputrc.EscapeMacro(string(e.current))

	e.macros[e.currentKey] = macro
	e.macros[rune(0)] = macro

	e.current = make([]rune, 0)
}
