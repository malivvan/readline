package readline

import (
	"unicode"

	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/strutil"
)

func (rl *Shell) adjustSelectionPending() {
	if !rl.selection.Active() {
		return
	}

	switch rl.Keymap.ActiveCommand().Action {

	case "vi-end-word", "vi-end-bigword",
		"vi-find-next-char", "vi-find-next-char-skip",
		"vi-find-prev-char", "vi-find-prev-char-skip",
		"vi-match":
		rl.selection.Visual(false)

	case "select-in-word", "select-a-word",
		"select-in-blank-word", "select-a-blank-word",
		"select-in-shell-word", "select-a-shell-word",
		"vi-select-inside":
		rl.selection.Visual(false)

	case "vi-change-to":
		rl.selection.Visual(false)
	}
}

func (rl *Shell) viAddEol() {
	rl.Iterations.Reset()

	if rl.Keymap.Local() == keymap.Visual {
		rl.cursor.Inc()
		rl.viInsertMode()

		return
	}

	rl.endOfLine()
	rl.viInsertMode()
}

func (rl *Shell) viAddNext() {
	if rl.line.Len() > 0 {
		rl.cursor.Inc()
	}

	rl.viInsertMode()
}

func (rl *Shell) viAddSurround() {

	done := rl.Keymap.PendingCursor()
	defer done()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		rl.History.SkipSave()
		return
	}

	bchar, echar := strutil.MatchSurround(key)

	rl.History.Save()

	rl.selection.Surround(bchar, echar)
}

func (rl *Shell) viBackwardDeleteChar() {
	if !rl.cursor.AtBeginningOfLine() {
		rl.backwardDeleteChar()
	}
}

func (rl *Shell) viChangeCase() {
	switch {
	case rl.selection.Active() && rl.selection.IsVisual():
		rl.selection.ReplaceWith(func(char rune) rune {
			if unicode.IsLower(char) {
				return unicode.ToUpper(char)
			}

			return unicode.ToLower(char)
		})

	default:
		if rl.line.Len() == 0 || rl.cursor.Pos() == rl.line.Len() {
			return
		}

		char := rl.cursor.Char()
		if unicode.IsLower(char) {
			char = unicode.ToUpper(char)
		} else {
			char = unicode.ToLower(char)
		}

		rl.cursor.ReplaceWith(char)
	}
}

func (rl *Shell) viChangeChar() {
	rl.History.Save()

	done := rl.Keymap.PendingCursor()
	defer done()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		rl.History.SkipSave()
		return
	}

	switch {
	case rl.selection.Active() && rl.selection.IsVisual():

		rl.selection.ReplaceWith(func(_ rune) rune {
			return key
		})
	default:

		rl.cursor.ReplaceWith(key)
	}
}

func (rl *Shell) viChangeEol() {
	rl.History.Save()
	rl.History.SkipSave()

	pos := rl.cursor.Pos()
	rl.selection.Mark(pos)
	rl.cursor.EndOfLineAppend()
	rl.selection.Cut()
	rl.cursor.Set(pos)

	rl.Iterations.Reset()
	rl.Display.ResetHelpers()
	rl.viInsertMode()
}

func (rl *Shell) viChangeTo() {
	switch {
	case rl.Keymap.IsPending():

		rl.Keymap.CancelPending()
		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.Cut()
		rl.viInsertMode()

	case len(rl.selection.Surrounds()) == 2:

		rl.Display.Refresh()
		defer rl.selection.Reset()

		done := rl.Keymap.PendingCursor()
		defer done()

		rchar, isAbort := rl.Keys.ReadKey()
		if isAbort {
			return
		}

		rl.History.Save()

		bchar, echar := strutil.MatchSurround(rchar)

		surrounds := rl.selection.Surrounds()

		bpos, _ := surrounds[0].Pos()
		epos, _ := surrounds[1].Pos()

		(*rl.line)[bpos] = bchar
		(*rl.line)[epos] = echar

	case rl.selection.Active():

		rl.History.Save()

		rl.adjustSelectionPending()
		cpos := rl.selection.Cursor()
		cut := rl.selection.Cut()
		rl.Buffers.Write([]rune(cut)...)
		rl.cursor.Set(cpos)

		rl.viInsertMode()

	default:

		keys := rl.Keys.Caller()

		switch keys[0] {
		case 'c':
			rl.Keymap.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'C':
			rl.viChangeEol()
		}
	}
}

func (rl *Shell) viCommandMode() {

	rl.selection.Reset()
	rl.Iterations.Reset()
	rl.Buffers.Reset()

	rl.Hint.Reset()
	rl.completer.Reset()
	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	if rl.Keymap.Main() == keymap.ViInsert && !rl.cursor.AtBeginningOfLine() {
		rl.cursor.Dec()
	}

	rl.cursor.CheckCommand()
	rl.Keymap.SetLocal("")
	rl.Keymap.SetMain(keymap.ViCommand)
}

func (rl *Shell) viDeleteChar() {
	if rl.line.Len() == 0 || rl.cursor.Pos() == rl.line.Len() {
		return
	}

	rl.History.Save()

	cutBuf := make([]rune, 0)

	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		cutBuf = append(cutBuf, rl.cursor.Char())
		rl.line.CutRune(rl.cursor.Pos())
	}

	rl.Buffers.Write(cutBuf...)
}

func (rl *Shell) viDeleteTo() {
	switch {
	case rl.Keymap.IsPending():

		rl.Keymap.CancelPending()
		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		cpos := rl.selection.Cursor()

		text := rl.selection.Cut()

		if len(text) > 0 && rune(text[len(text)-1]) != inputrc.Newline {
			text += string(inputrc.Newline)
		}

		rl.Buffers.Write([]rune(text)...)
		rl.cursor.Set(cpos)

	case rl.selection.Active():

		rl.History.Save()

		rl.adjustSelectionPending()
		cpos := rl.selection.Cursor()
		cut := rl.selection.Cut()
		rl.Buffers.Write([]rune(cut)...)
		rl.cursor.Set(cpos)

		rl.viCommandMode()

	default:

		keys := rl.Keys.Caller()

		switch keys[0] {
		case 'd':
			rl.Keymap.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'D':
			rl.viKillEol()
		}
	}
}

func (rl *Shell) viDownCase() {
	switch {
	case rl.Keymap.IsPending():

		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.ReplaceWith(unicode.ToLower)
		rl.viCommandMode()

	case rl.selection.Active():
		rl.History.Save()
		rl.selection.ReplaceWith(unicode.ToLower)
		rl.viCommandMode()

	default:

		rl.History.SkipSave()
		rl.Keymap.Pending()
		rl.selection.Mark(rl.cursor.Pos())
	}
}

func (rl *Shell) viEditAndExecuteCommand() {
	rl.editAndExecuteCommand()
}

func (rl *Shell) viEditCommandLine() {
	keymapCur := rl.Keymap.Main()

	rl.editCommandLine()

	switch keymapCur {
	case keymap.ViCommand, keymap.Vi:
		rl.viCommandMode()
	default:
		rl.viInsertMode()
	}
}

func (rl *Shell) viInsertBol() {
	rl.Iterations.Reset()
	rl.beginningOfLine()
	rl.viInsertMode()
}

func (rl *Shell) viInsertMode() {
	rl.History.Save()

	rl.selection.Reset()
	rl.Iterations.Reset()
	rl.Buffers.Reset()

	rl.Keymap.SetLocal("")
	rl.Keymap.SetMain(keymap.ViInsert)
	rl.cursor.SetMark()
}

func (rl *Shell) viKillEol() {
	rl.History.Save()

	pos := rl.cursor.Pos()
	rl.selection.Mark(rl.cursor.Pos())
	rl.cursor.EndOfLineAppend()

	cut := rl.selection.Cut()
	rl.Buffers.Write([]rune(cut)...)
	rl.cursor.Set(pos)

	if !rl.cursor.AtBeginningOfLine() {
		rl.cursor.Dec()
	}

	rl.Iterations.Reset()
	rl.Display.ResetHelpers()
}

func (rl *Shell) viKillLine() {
	if rl.cursor.Pos() <= rl.cursor.Mark() || rl.cursor.Pos() == 0 {
		return
	}

	rl.History.Save()

	rl.selection.MarkRange(rl.cursor.Mark(), rl.line.Len())
	rl.cursor.Dec()
	cut := rl.selection.Cut()
	rl.Buffers.Write([]rune(cut)...)
}

func (rl *Shell) viOpenLineAbove() {
	rl.History.Save()

	if !rl.cursor.OnEmptyLine() {
		rl.beginningOfLine()
	}

	rl.cursor.InsertAt('\n')
	rl.cursor.Dec()
	rl.viInsertMode()
}

func (rl *Shell) viOpenLineBelow() {
	rl.History.Save()

	if !rl.cursor.OnEmptyLine() {
		rl.endOfLine()
	}

	rl.cursor.InsertAt('\n')
	rl.viInsertMode()
}

func (rl *Shell) viPut() {
	keys := rl.Keys.Caller()

	switch keys[0] {
	case 'P':
		rl.viPutBefore()
	case 'p':
		fallthrough
	default:
		rl.viPutAfter()
	}
}

func (rl *Shell) viPutAfter() {
	rl.History.Save()

	buffer := rl.Buffers.Active()

	if len(buffer) == 0 {
		return
	}

	if buffer[len(buffer)-1] == '\n' {
		if !rl.cursor.OnEmptyLine() {
			rl.cursor.EndOfLineAppend()
		}

		if rl.cursor.Pos() == rl.line.Len() {
			buffer = append([]rune{'\n'}, buffer[:len(buffer)-1]...)
		}
	}

	rl.cursor.Inc()
	pos := rl.cursor.Pos()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		rl.line.Insert(pos, buffer...)
	}
}

func (rl *Shell) viPutBefore() {
	rl.History.Save()

	buffer := rl.Buffers.Active()

	if len(buffer) == 0 {
		return
	}

	if buffer[len(buffer)-1] == '\n' {
		rl.cursor.BeginningOfLine()

		if rl.cursor.OnEmptyLine() {
			buffer = append(buffer, '\n')

			rl.cursor.Dec()
		}
	}

	pos := rl.cursor.Pos()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		rl.line.Insert(pos, buffer...)
	}

	rl.cursor.Set(pos)
}

func (rl *Shell) viRedo() {
	if rl.History.Pos() > 0 {
		rl.History.Redo()
		return
	}

	rl.viInsertMode()
}

func (rl *Shell) viReplace() {

	rl.overwriteMode()

	rl.cursor.Dec()
}

func (rl *Shell) viRubout() {
	if rl.Keymap.Main() != keymap.ViInsert {
		rl.History.Save()
	}

	vii := rl.Iterations.Get()

	cut := make([]rune, 0)

	for i := 1; i <= vii; i++ {
		if rl.cursor.Pos() == 0 {
			break
		}

		rl.cursor.Dec()
		cut = append(cut, rl.cursor.Char())
		rl.line.CutRune(rl.cursor.Pos())
	}

	rl.Buffers.Write(cut...)
}

func (rl *Shell) viSelectABlankWord() {
	rl.History.SkipSave()
	rl.cursor.CheckCommand()

	rl.selection.SelectABlankWord()
}

func (rl *Shell) viSelectAShellWord() {
	rl.History.SkipSave()
	rl.cursor.CheckCommand()

	bpos, _ := rl.line.SelectBlankWord(rl.cursor.Pos())
	rl.cursor.Set(bpos)

	rl.selection.SelectAShellWord()
}

func (rl *Shell) viSelectAWord() {
	rl.History.SkipSave()
	rl.selection.SelectAWord()
}

func (rl *Shell) viSelectInBlankWord() {
	rl.History.SkipSave()

	bpos, epos := rl.line.SelectBlankWord(rl.cursor.Pos())
	rl.cursor.Set(epos)
	rl.selection.Mark(bpos)
}

func (rl *Shell) viSelectInShellWord() {
	rl.History.SkipSave()

	bpos, _ := rl.line.SelectBlankWord(rl.cursor.Pos())
	rl.cursor.Set(bpos)

	sBpos, sEpos := rl.line.SurroundQuotes(true, rl.cursor.Pos())
	dBpos, dEpos := rl.line.SurroundQuotes(false, rl.cursor.Pos())
	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	if mark == -1 && cpos == -1 {
		rl.viSelectInBlankWord()

		return
	}

	rl.cursor.Set(cpos - 1)

	rl.selection.Mark(mark + 1)
}

func (rl *Shell) viSelectInWord() {
	rl.History.SkipSave()

	bpos, epos := rl.line.SelectWord(rl.cursor.Pos())
	rl.cursor.Set(epos)
	rl.selection.Mark(bpos)
}

func (rl *Shell) viSelectInside() {
	rl.History.SkipSave()

	var inside bool

	keys := rl.Keys.Caller()
	if keys[0] == 'i' {
		inside = true
	}

	char, empty := rl.Keys.Pop()
	if empty {
		return
	}

	bpos, epos, _, _ := rl.line.FindSurround(rune(char), rl.cursor.Pos())
	if bpos == -1 && epos == -1 {
		return
	}

	if inside {
		bpos++
		epos--
	}

	rl.selection.Mark(bpos)
	rl.cursor.Set(epos)
}

func (rl *Shell) viSelectSurround() {
	rl.History.SkipSave()

	done := rl.Keymap.PendingCursor()
	defer done()

	char, isAbort := rl.Keys.ReadKey()
	if isAbort {
		return
	}

	bpos, epos, _, _ := rl.line.FindSurround(char, rl.cursor.Pos())
	if bpos == -1 || epos == -1 {
		return
	}

	rl.selection.MarkSurround(bpos, epos)
}

func (rl *Shell) viSetBuffer() {
	rl.History.SkipSave()

	rl.Buffers.Reset()

	done := rl.Keymap.PendingCursor()
	defer done()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		return
	}

	rl.Buffers.SetActive(key)
}

func (rl *Shell) viSubstitute() {
	rl.History.Save()

	defer rl.viInsertMode()

	switch {
	case rl.selection.Active():

		cpos := rl.selection.Cursor()
		rl.selection.Cut()
		rl.cursor.Set(cpos)

	default:

		keys := rl.Keys.Caller()

		switch keys[0] {
		case 's':

			vii := rl.Iterations.Get()
			for i := 1; i <= vii; i++ {
				rl.line.CutRune(rl.cursor.Pos())
			}
		case 'S':
			if rl.cursor.OnEmptyLine() {
				return
			}

			rl.selection.Mark(rl.cursor.Pos())
			rl.selection.Visual(true)

			bpos, epos := rl.selection.Pos()
			rl.Buffers.Write((*rl.line)[bpos:epos]...)

			if (*rl.line)[epos-1] == '\n' {
				epos--
			}

			rl.line.Cut(bpos, epos)
			rl.cursor.Set(bpos)
		}
	}
}

func (rl *Shell) viUpCase() {
	switch {
	case rl.Keymap.IsPending():

		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.ReplaceWith(unicode.ToUpper)
		rl.viCommandMode()

	case rl.selection.Active():
		rl.History.Save()
		rl.selection.ReplaceWith(unicode.ToUpper)
		rl.viCommandMode()

	default:

		rl.History.SkipSave()
		rl.Keymap.Pending()
		rl.selection.Mark(rl.cursor.Pos())
	}
}

func (rl *Shell) viVisualLineMode() {
	rl.History.SkipSave()
	rl.Iterations.Reset()
	rl.Buffers.Reset()

	rl.Hint.Reset()
	rl.completer.Reset()

	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(true)
	rl.Keymap.SetLocal(keymap.Visual)

	rl.Keymap.PrintCursor(keymap.Visual)
}

func (rl *Shell) viVisualMode() {
	rl.History.SkipSave()
	rl.Iterations.Reset()
	rl.Buffers.Reset()

	rl.Hint.Reset()
	rl.completer.Reset()

	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(false)
	rl.Keymap.SetLocal(keymap.Visual)
}

func (rl *Shell) viYankTo() {
	switch {
	case rl.Keymap.IsPending():

		rl.Keymap.CancelPending()
		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)

		text, _, _, _ := rl.selection.Pop()
		if len(text) > 0 && rune(text[len(text)-1]) != inputrc.Newline {
			text += string(inputrc.Newline)
		}

		rl.Buffers.Write([]rune(text)...)

	case rl.selection.Active():

		rl.History.Save()
		rl.adjustSelectionPending()
		text, _, _, cpos := rl.selection.Pop()

		rl.Buffers.Write([]rune(text)...)
		rl.cursor.Set(cpos)

		rl.viCommandMode()

	default:

		keys := rl.Keys.Caller()

		switch keys[0] {
		case 'y':
			rl.Keymap.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'Y':
			rl.viYankWholeLine()
		}
	}
}

func (rl *Shell) viYankWholeLine() {
	rl.History.SkipSave()

	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(true)

	bpos, epos := rl.selection.Pos()

	if (*rl.line)[epos-1] == '\n' {
		epos--
	}

	buffer := (*rl.line)[bpos:epos]
	rl.Buffers.Write(buffer...)

	rl.selection.Reset()
}
