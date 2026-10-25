package readline

import (
	"unicode"

	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/strutil"
)

func (rl *Shell) viArgDigit() {
	rl.History.SkipSave()

	keys := rl.Keys.Caller()
	rl.Iterations.Add(string(keys))
}

func (rl *Shell) viBackToIndent() {
	rl.cursor.BeginningOfLine()
	rl.cursor.ToFirstNonSpace(true)
}

func (rl *Shell) viBackwardBlankWord() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) viBackwardBlankWordEnd() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Inc()

		rl.cursor.Move(rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos()))
		rl.cursor.Move(rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos()))

		rl.cursor.Move(rl.line.ForwardEnd(rl.line.TokenizeSpace, rl.cursor.Pos()))
	}
}

func (rl *Shell) viBackwardChar() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()

	if rl.cursor.Pos() == 0 {
		return
	}

	for i := 1; i <= vii; i++ {
		if (*rl.line)[rl.cursor.Pos()-1] == '\n' {
			break
		}

		rl.cursor.Dec()
	}
}

func (rl *Shell) viBackwardWord() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) viBackwardWordEnd() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()

	if rl.line.Len() == 0 {
		return
	}

	for i := 1; i <= vii; i++ {
		rl.cursor.Inc()

		rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))
		rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))

		if unicode.IsPunct(rl.cursor.Char()) {
			rl.cursor.Dec()
		}

		rl.cursor.Move(rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos()))
	}
}

func (rl *Shell) viCharSearch() {
	var forward, skip bool

	keys := rl.Keys.Caller()

	switch keys[0] {
	case 'F':
		forward = false
		skip = false
	case 't':
		forward = true
		skip = true
	case 'T':
		forward = false
		skip = true
	case 'f':
		fallthrough
	default:
		forward = true
		skip = false
	}

	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.viFindChar(forward, skip)
	}
}

func (rl *Shell) viEOFMaybe() {
	rl.endOfFile()
}

func (rl *Shell) viEndOfLine() {
	rl.History.SkipSave()

	rl.cursor.EndOfLineAppend()
}

func (rl *Shell) viFindChar(forward, skip bool) {
	rl.History.SkipSave()

	done := rl.Keymap.PendingCursor()
	defer done()

	char, esc := rl.Keys.ReadKey()
	if esc {
		return
	}

	times := rl.Iterations.Get()

	for i := 1; i <= times; i++ {
		pos := rl.line.Find(char, rl.cursor.Pos(), forward)

		if pos == rl.cursor.Pos() || pos == -1 {
			break
		}

		if forward && skip {
			pos--
		} else if !forward && skip {
			pos++
		}

		rl.cursor.Set(pos)
	}
}

func (rl *Shell) viFindNextChar() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.viFindChar(true, false)
	}
}

func (rl *Shell) viFindNextCharSkip() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.viFindChar(true, true)
	}
}

func (rl *Shell) viFindPrevChar() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.viFindChar(false, false)
	}
}

func (rl *Shell) viFindPrevCharSkip() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.viFindChar(false, true)
	}
}

func (rl *Shell) viFirstPrint() {
	rl.cursor.BeginningOfLine()
	rl.cursor.ToFirstNonSpace(true)
}

func (rl *Shell) viForwardBlankWord() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viForwardBlankWordEnd() {
	rl.History.SkipSave()
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Move(rl.line.ForwardEnd(rl.line.TokenizeSpace, rl.cursor.Pos()))
	}
}

func (rl *Shell) viForwardChar() {

	if rl.Config.GetBool("history-autosuggest") && rl.cursor.Pos() == rl.line.Len()-1 {
		rl.autosuggestAccept()
		return
	}

	if rl.acceptInlineSuggestion() {
		return
	}

	rl.History.SkipSave()

	if rl.Keymap.Main() != keymap.ViInsert && rl.cursor.Pos() < rl.line.Len()-1 {
		vii := rl.Iterations.Get()

		for i := 1; i <= vii; i++ {
			if (*rl.line)[rl.cursor.Pos()+1] == '\n' {
				break
			}

			rl.cursor.Inc()
		}
	}
}

func (rl *Shell) viForwardWord() {
	rl.History.Save()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {

		rl.insertAutosuggestPartial(false)

		forward := rl.line.Forward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viForwardWordEnd() {
	rl.History.SkipSave()
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viGotoColumn() {
	rl.History.SkipSave()

	column := rl.Iterations.Get()

	if column < 0 {
		return
	}

	cpos := rl.cursor.Pos()

	rl.cursor.BeginningOfLine()
	bpos := rl.cursor.Pos()
	rl.cursor.EndOfLine()
	epos := rl.cursor.Pos()

	rl.cursor.Set(cpos)

	switch {
	case column > epos-cpos:
		rl.cursor.Set(epos)
	default:
		rl.cursor.Set(bpos + column - 1)
	}
}

func (rl *Shell) viGotoMark() {
	switch {
	case rl.selection.Active():

		bpos, epos := rl.selection.Pos()
		if bpos != rl.cursor.Pos() {
			rl.cursor.Set(bpos)
		} else {
			rl.cursor.Set(epos)
		}

	case rl.cursor.Mark() != -1:

		rl.cursor.Set(rl.cursor.Mark())
	}
}

func (rl *Shell) viMatchBracket() {
	rl.History.SkipSave()

	nextPos := rl.cursor.Pos()
	found := false

	if !strutil.IsBracket(rl.cursor.Char()) {
		for i := rl.cursor.Pos() + 1; i < rl.line.Len(); i++ {
			char := (*rl.line)[i]
			if char == '}' || char == ')' || char == ']' {
				nextPos = i - rl.cursor.Pos()
				found = true

				break
			}
		}

		if !found {
			return
		}

		rl.cursor.Move(nextPos)
	}

	var adjust int

	split, index, pos := rl.line.TokenizeBlock(rl.cursor.Pos())

	switch {
	case len(split) == 0:
		return
	case pos == 0:
		adjust = len(split[index])
	default:
		adjust = pos * -1
	}

	rl.cursor.Move(adjust)
}

func (rl *Shell) viSearch() {
	var forward bool

	keys := rl.Keys.Caller()

	switch keys[0] {
	case '/':
		forward = true
	case '?':
		forward = false
	}

	rl.completer.NonIsearchStart(rl.History.Name()+" "+string(keys[0]), false, forward, true)
}

func (rl *Shell) viSearchAgain() {
	var forward bool
	var hint string

	keys := rl.Keys.Caller()

	switch keys[0] {
	case 'n':
		forward = true
		hint = " /"
	case 'N':
		forward = false
		hint = " ?"
	}

	rl.completer.NonIsearchStart(rl.History.Name()+hint, true, forward, true)

	line, cursor, _ := rl.completer.GetBuffer()

	rl.History.InsertMatch(line, cursor, true, forward, true)
	rl.completer.NonIsearchStop()
}

func (rl *Shell) viSearchAgainBackward() {
	rl.completer.NonIsearchStart(rl.History.Name()+" ?", true, false, true)

	line, cursor, _ := rl.completer.GetBuffer()

	rl.History.InsertMatch(line, cursor, true, false, true)
	rl.completer.NonIsearchStop()
}

func (rl *Shell) viSearchAgainForward() {
	rl.completer.NonIsearchStart(rl.History.Name()+" /", true, true, true)

	line, cursor, _ := rl.completer.GetBuffer()

	rl.History.InsertMatch(line, cursor, true, true, true)
	rl.completer.NonIsearchStop()
}

func (rl *Shell) viSearchBackward() {
	rl.completer.NonIsearchStart(rl.History.Name()+" ?", false, false, true)
}

func (rl *Shell) viSearchForward() {
	rl.completer.NonIsearchStart(rl.History.Name()+" /", false, true, true)
}

func (rl *Shell) viSetMark() {
	rl.History.SkipSave()
	rl.selection.Mark(rl.cursor.Pos())
}
