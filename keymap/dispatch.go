package keymap

import (
	"slices"
	"sort"
	"strings"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
)

func MatchLocal(eng *Engine) (bind inputrc.Bind, command func(), prefix bool) {
	if eng.local == "" {
		return bind, command, prefix
	}

	binds := eng.getContextBinds(false)
	if len(binds) == 0 {
		return bind, command, prefix
	}

	bind, prefix, read, matched := eng.dispatchKeys(binds)

	if !bind.Macro {
		command = eng.commands[bind.Action]
	}

	if prefix {
		core.MatchedPrefix(eng.keys, read...)
	} else {
		core.MatchedKeys(eng.keys, matched, read[len(matched):]...)
	}

	if eng.isEscapeKey() && (prefix || command == nil) {
		bind, command, prefix = eng.handleEscape(false)
	}

	return bind, command, prefix
}

func MatchMain(eng *Engine) (bind inputrc.Bind, command func(), prefix bool) {
	if eng.main == "" {
		return bind, command, prefix
	}

	binds := eng.getContextBinds(true)
	if len(binds) == 0 {
		return bind, command, prefix
	}

	bind, prefix, read, _ := eng.dispatchKeys(binds)

	if !bind.Macro {
		command = eng.commands[bind.Action]
	}

	if prefix {
		core.MatchedPrefix(eng.keys, read...)
	} else {
		core.MatchedKeys(eng.keys, read)
	}

	if eng.nonIncSearch && (command == nil || prefix) {
		bind = inputrc.Bind{Action: "self-insert"}
		eng.active = bind
		command = eng.resolve(bind)
		prefix = false
	}

	if eng.isEscapeKey() && !eng.IsEmacs() && prefix {
		bind, command, prefix = eng.handleEscape(true)
	}

	return bind, command, prefix
}

func validCommand(widget string, commands []string) bool {
	return slices.Contains(commands, widget)
}

func (m *Engine) CancelPending() {
	if len(m.pending) == 0 {
		return
	}

	m.pending = m.pending[:len(m.pending)-1]

	if len(m.pending) == 0 && m.Local() == ViOpp {
		m.SetLocal("")
	}
}

func (m *Engine) IsPending() bool {
	if len(m.pending) == 0 {
		return false
	}

	return m.active.Action == m.pending[0].Action
}

func (m *Engine) Pending() {
	m.SetLocal(ViOpp)
	m.skip = true

	m.pending = append(m.pending, m.active)
}

func (m *Engine) PendingCursor() (restore func()) {
	m.PrintCursor(ViOpp)

	return func() {
		m.UpdateCursor()
	}
}

func (m *Engine) RunPending() {
	if len(m.pending) == 0 {
		return
	}

	if m.skip {
		m.skip = false
		return
	}

	defer m.UpdateCursor()

	pending := m.pending[len(m.pending)-1]
	m.pending = m.pending[:len(m.pending)-1]

	if pending.Action == m.active.Action {
		m.isCaller = true
		defer func() { m.isCaller = false }()
	}

	if pending.Action == "" {
		return
	}

	command := m.resolve(pending)

	command()

	if len(m.pending) == 0 && m.Local() == ViOpp {
		m.SetLocal("")
	}
}

func (m *Engine) dispatchKeys(binds map[string]inputrc.Bind) (bind inputrc.Bind, prefix bool, read, matched []byte) {

	char, empty := core.PeekChar(m.keys)
	if empty {
		return m.active, prefix, read, matched
	}

	if len(char) > 1 {
		read = append(read, char...)
		match := inputrc.Bind{
			Action: "self-insert",
			Macro:  false,
		}

		matched = append(matched, char...)
		prefix = m.makeMatch(match, inputrc.Bind{})
		core.PopChar(m.keys)

		return m.active, prefix, read, matched
	}

	for {

		key, empty := core.PeekKey(m.keys)
		if empty {
			break
		}

		read = append(read, key)

		match, prefixed := m.matchBind(read, binds)

		if match.Action == "" && len(prefixed) == 0 {
			prefix = m.makeMatch(m.prefixed, inputrc.Bind{})

			core.PopKey(m.keys)

			break
		}

		core.PopKey(m.keys)

		matched = append(matched, key)

		if len(prefixed) > 0 {
			prefix = true

			if match.Action != "" {
				m.prefixed = match
			}

			continue
		}

		prefix = m.makeMatch(match, inputrc.Bind{})

		break
	}

	return m.active, prefix, read, matched
}

func (m *Engine) getContextBinds(main bool) (binds map[string]inputrc.Bind) {

	if main {
		binds = m.config.Binds[string(m.main)]
	} else {
		binds = m.config.Binds[string(m.local)]
	}

	if !main || len(binds) == 0 {
		return
	}

	switch {
	case m.Local() == Isearch:
		binds = m.restrictCommands(m.main, isearchCommands)
	case m.nonIncSearch:
		binds = m.restrictCommands(m.main, nonIsearchCommands)
	}

	return
}

func (m *Engine) handleEscape(main bool) (bind inputrc.Bind, cmd func(), pref bool) {
	switch {
	case m.prefixed.Action == "vi-movement-mode":

		bind = m.prefixed

	case !main && m.IsEmacs() && m.Local() == Isearch:

		bind = inputrc.Bind{Action: "emacs-editing-mode"}

		core.PopForce(m.keys)

	case !main:

		bind = inputrc.Bind{}
	}

	m.prefixed = inputrc.Bind{}

	if bind.Action != "" && !bind.Macro {
		cmd = m.resolve(bind)
	}

	if main {
		core.PopForce(m.keys)
	}

	return bind, cmd, pref
}

func (m *Engine) isEscapeKey() bool {
	keys := m.keys.Caller()
	if len(keys) == 0 {
		return false
	}

	if len(keys) > 1 {
		return false
	}

	if keys[0] != inputrc.Esc {
		return false
	}

	return true
}

func (m *Engine) makeMatch(active, prefixed inputrc.Bind) (prefix bool) {
	m.active = active
	m.prefixed = prefixed

	return m.prefixed.Action != ""
}

func (m *Engine) matchBind(keys []byte, binds map[string]inputrc.Bind) (inputrc.Bind, []inputrc.Bind) {
	var match inputrc.Bind
	var prefixed []inputrc.Bind

	sequences := make([]string, 0, len(binds))
	for sequence := range binds {
		sequences = append(sequences, sequence)
	}

	sort.Slice(sequences, func(i, j int) bool {
		if len(sequences[i]) == len(sequences[j]) {
			return sequences[i] < sequences[j]
		}

		return len(sequences[i]) < len(sequences[j])
	})

	for _, sequence := range sequences {
		seq := strutil.ConvertMeta([]rune(sequence))

		if len(string(keys)) < len(seq) && strings.HasPrefix(seq, string(keys)) {
			prefixed = append(prefixed, binds[sequence])
		}

		if string(keys) == seq {
			match = binds[sequence]
		}
	}

	return match, prefixed
}

func (m *Engine) resolve(bind inputrc.Bind) func() {
	if bind.Macro {
		return nil
	}

	if bind.Action == "" {
		return nil
	}

	return m.commands[bind.Action]
}

func (m *Engine) restrictCommands(mode Mode, commands []string) map[string]inputrc.Bind {
	if len(commands) == 0 {
		return m.config.Binds[string(mode)]
	}

	isearch := make(map[string]inputrc.Bind)

	for seq, command := range m.config.Binds[string(mode)] {

		if !validCommand(command.Action, commands) {
			continue
		}

		isearch[seq] = command
	}

	return isearch
}
