package completion

import (
	"regexp"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/keymap"
)

func (e *Engine) GetBuffer() (*core.Line, *core.Cursor, *core.Selection) {

	searching, _, _ := e.NonIncrementallySearching()

	if e.keymap.Local() == keymap.Isearch || searching {
		selection := core.NewSelection(e.isearchBuf, e.isearchCur)
		return e.isearchBuf, e.isearchCur, selection
	}

	if len(e.selected.Value) > 0 {
		return e.compLine, e.compCursor, e.selection
	}

	return e.line, e.cursor, e.selection
}

func (e *Engine) IsearchStart(name string, autoinsert, replaceLine bool) {

	e.isearchInsert = autoinsert
	e.isearchReplaceLine = replaceLine

	e.isearchStartBuf = string(*e.line)
	e.isearchStartCursor = e.cursor.Pos()

	e.isearchBuf = new(core.Line)
	e.isearchCur = core.NewCursor(e.isearchBuf)

	e.auto = true
	e.keymap.SetLocal(keymap.Isearch)
	e.adaptIsearchInsertMode()

	e.isearchName = name
	e.hint.Set(color.Bold + color.FgCyan + e.isearchName + " (isearch): " + color.Reset + string(*e.isearchBuf))
}

func (e *Engine) IsearchStop(revertLine bool) {

	e.isearchBuf = nil
	e.IsearchRegex = nil
	e.isearchCur = nil

	if e.isearchReplaceLine && revertLine {
		e.line.Set([]rune(e.isearchStartBuf)...)
		e.cursor.Set(e.isearchStartCursor)
	}

	e.isearchStartBuf = ""
	e.isearchStartCursor = 0
	e.isearchReplaceLine = false

	e.auto = false
	e.autoForce = false
	e.keymap.SetLocal("")
	e.resetIsearchInsertMode()
}

func (e *Engine) NonIncrementallySearching() (searching, forward, substring bool) {
	searching = e.isearchCur != nil && e.keymap.Local() != keymap.Isearch
	forward = e.isearchForward
	substring = e.isearchSubstring

	return
}

func (e *Engine) NonIsearchStart(name string, repeat, forward, substring bool) {
	if repeat {
		e.isearchBuf = new(core.Line)
		e.isearchBuf.Set([]rune(e.isearchLast)...)
	} else {
		e.isearchBuf = new(core.Line)
	}

	e.isearchCur = core.NewCursor(e.isearchBuf)
	e.isearchCur.Set(e.isearchBuf.Len())

	e.isearchName = name
	e.isearchForward = forward
	e.isearchSubstring = substring

	e.keymap.NonIncrementalSearchStart()
	e.adaptIsearchInsertMode()
}

func (e *Engine) NonIsearchStop() {
	e.isearchLast = string(*e.isearchBuf)
	e.isearchBuf = nil
	e.IsearchRegex = nil
	e.isearchCur = nil
	e.isearchForward = false
	e.isearchSubstring = false

	e.keymap.NonIncrementalSearchStop()
	e.resetIsearchInsertMode()
	e.hint.Reset()
}

func (e *Engine) UpdateIsearch() {
	searching, _, _ := e.NonIncrementallySearching()

	if e.keymap.Local() != keymap.Isearch && !searching {
		return
	}

	if len(e.selected.Value) > 0 {
		return
	}

	if e.keymap.Local() == keymap.Isearch {
		e.updateIncrementalSearch()
	} else {
		e.updateNonIncrementalSearch()
	}
}

func (e *Engine) adaptIsearchInsertMode() {
	e.isearchModeExit = e.keymap.Main()

	if !e.keymap.IsEmacs() && e.keymap.Main() != keymap.ViInsert {
		e.keymap.SetMain(keymap.ViInsert)
	}
}

func (e *Engine) resetIsearchInsertMode() {
	if e.isearchModeExit == "" {
		return
	}

	if e.keymap.Main() != e.isearchModeExit {
		e.keymap.SetMain(string(e.isearchModeExit))
		e.isearchModeExit = ""
	}

	if e.keymap.Main() == keymap.ViCommand {
		e.cursor.CheckCommand()
	}
}

func (e *Engine) updateIncrementalSearch() {
	var regexStr string
	if containsUpper(*e.isearchBuf) {
		regexStr = string(*e.isearchBuf)
	} else {
		regexStr = "(?i)" + string(*e.isearchBuf)
	}

	var err error

	e.IsearchRegex, err = regexp.Compile(regexStr)
	if err != nil {
		e.IsearchRegex = nil
		e.hint.Set(color.FgRed + "Failed to compile i-search regexp")

		return
	}

	e.GenerateWith(e.cached)

	for _, g := range e.groups {
		g.updateIsearch(e)
	}

	isearchHint := color.Bold + color.FgCyan + e.isearchName + " (inc-search)"

	if e.Matches() == 0 {
		isearchHint += color.Reset + color.Bold + color.FgRed + " (no matches)"
	}

	isearchHint += ": " + color.Reset + color.Bold + string(*e.isearchBuf) + color.Reset + "_"

	e.hint.Set(isearchHint)

	if e.isearchInsert && e.Matches() > 0 && e.isearchBuf.Len() > 0 {

		if e.isearchReplaceLine {
			e.prefix = ""
			e.line.Set()
			e.cursor.Set(0)
		}

		e.Select(1, 0)
	} else if e.isearchReplaceLine {

		e.line.Set([]rune(e.isearchStartBuf)...)
		e.cursor.Set(e.isearchStartCursor)
	}
}

func (e *Engine) updateNonIncrementalSearch() {
	isearchHint := color.Bold + color.FgCyan + e.isearchName +
		" (non-inc-search): " + color.Reset + color.Bold + string(*e.isearchBuf) + color.Reset + "_"
	e.hint.Set(isearchHint)
}
