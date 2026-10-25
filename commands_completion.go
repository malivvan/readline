package readline

import (
	"fmt"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/history"
	"github.com/malivvan/readline/keymap"
)

func (rl *Shell) RefreshCompletions() {
	rl.completer.RequestRegen()
	rl.Keys.RequestRefresh()
}

func (rl *Shell) acceptAndMenuComplete() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	if !rl.completer.IsInserting() {
		return
	}

	rl.completer.Cancel(false, false)

	rl.completer.Select(1, 0)
}

func (rl *Shell) commandCompletion() (values completion.Values) {
	if rl.Completer == nil {
		return completion.Values{}
	}

	defer func() {
		if r := recover(); r != nil {
			msg := CompleteMessage("completion error: %v", r)
			values = msg.convert()
		}
	}()

	line, cursor := rl.completer.Line()
	comps := rl.Completer(*line, cursor.Pos())

	return comps.convert()
}

func (rl *Shell) completeWord() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)

		if rl.Config.GetBool("menu-complete-display-prefix") {

			rl.completer.InsertCommonPrefix()
			return
		}
	}

	rl.completer.Select(1, 0)
	rl.completer.SkipDisplay()
}

func (rl *Shell) deleteCharOrList() {
	switch {
	case rl.cursor.Pos() < rl.line.Len():
		rl.line.CutRune(rl.cursor.Pos())
	default:
		rl.possibleCompletions()
	}
}

func (rl *Shell) historyCompletion(forward, filterLine, substring bool) {
	switch {
	case rl.Keymap.Local() == keymap.MenuSelect || rl.Keymap.Local() == keymap.Isearch || rl.completer.AutoCompleting():

		if rl.History.OnLastSource() {
			rl.History.Cycle(true)
			rl.completer.ResetForce()
			rl.Hint.Reset()

			return
		}

		rl.History.Cycle(true)

		fallthrough

	default:

		if rl.History.Current() == nil {
			rl.Hint.SetTemporary(fmt.Sprintf("%s%s%s %s", color.Dim, color.FgRed, "No command history source", color.Reset))
			return
		}

		completer := func() completion.Values {
			maxLines := rl.Display.AvailableHelperLines()
			return history.Complete(rl.History, forward, filterLine, maxLines, rl.completer.IsearchRegex)
		}

		if substring {
			rl.completer.GenerateWith(completer)
			rl.completer.IsearchStart(rl.History.Name(), true, true)
		} else {
			rl.startMenuComplete(completer)
			rl.completer.AutocompleteForce()
		}
	}
}

func (rl *Shell) insertCompletions() {
	rl.History.Save()

	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)
	}

	for range rl.completer.Matches() {
		rl.completer.Select(1, 0)
		rl.completer.Cancel(false, false)
	}

	rl.completer.Cancel(false, false)
	rl.completer.ClearMenu(true)
}

func (rl *Shell) menuComplete() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)

		if rl.Config.GetBool("menu-complete-display-prefix") {

			rl.completer.InsertCommonPrefix()
			return
		}
	}

	rl.completer.Select(1, 0)
}

func (rl *Shell) menuCompleteBackward() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)
	}

	rl.completer.Select(-1, 0)
}

func (rl *Shell) menuCompleteNextTag() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(true)
}

func (rl *Shell) menuCompletePrevTag() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(false)
}

func (rl *Shell) menuIncrementalSearch() {
	rl.History.SkipSave()

	rl.completer.GenerateWith(rl.commandCompletion)
	rl.completer.IsearchStart("completions", false, false)
}

func (rl *Shell) possibleCompletions() {
	rl.History.SkipSave()

	rl.startMenuComplete(rl.commandCompletion)
}

func (rl *Shell) startMenuComplete(completer completion.Completer) {
	rl.History.SkipSave()

	rl.Keymap.SetLocal(keymap.MenuSelect)
	rl.completer.GenerateWith(completer)
}

func (rl *Shell) viRegistersComplete() {
	rl.History.SkipSave()
	rl.startMenuComplete(rl.Buffers.Complete)
}
