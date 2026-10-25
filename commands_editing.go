package readline

import (
	"strings"
	"unicode"

	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/strutil"
	"github.com/rivo/uniseg"
)

func (rl *Shell) backwardChar() {
	rl.History.SkipSave()
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Dec()
	}
}

func (rl *Shell) backwardDeleteChar() {
	if rl.Keymap.Main() == keymap.ViInsert {
		rl.History.SkipSave()
	} else {
		rl.History.Save()
	}

	completion.UpdateInserted(rl.completer)

	if rl.cursor.Pos() == 0 {
		return
	}

	vii := rl.Iterations.Get()

	switch vii {
	case 1:

		if rl.Config.GetBool("autopairs") {
			completion.AutopairDelete(rl.line, rl.cursor)
		}

		rl.cursor.Dec()
		rl.line.CutRune(rl.cursor.Pos())

	default:
		for i := 1; i <= vii; i++ {
			rl.cursor.Dec()
			rl.line.CutRune(rl.cursor.Pos())
		}
	}
}

func (rl *Shell) backwardKillLine() {
	rl.Iterations.Reset()
	rl.History.Save()

	if rl.line.Len() == 0 {
		return
	}

	cpos := rl.cursor.Pos()
	rl.cursor.BeginningOfLine()

	rl.selection.MarkRange(rl.cursor.Pos(), cpos)
	text := rl.selection.Cut()

	rl.Buffers.Write([]rune(text)...)
}

func (rl *Shell) backwardKillWord() {
	rl.History.Save()
	rl.History.SkipSave()

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust)

	rl.Buffers.Write([]rune(rl.selection.Cut())...)
}

func (rl *Shell) backwardShellWord() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {

		startPos := rl.cursor.Pos()
		backward := rl.line.Backward(rl.line.TokenizeSpace, startPos)
		rl.cursor.Move(backward)

		bpos, _ := rl.selection.SelectAShellWord()
		rl.cursor.Set(bpos)
	}
}

func (rl *Shell) backwardWord() {
	rl.History.SkipSave()

	vii := rl.Iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) beginningOfLine() {
	rl.History.SkipSave()

	if !rl.Keymap.IsEmacs() && rl.Iterations.IsSet() {
		rl.Iterations.Add("0")
		return
	}

	rl.cursor.BeginningOfLine()
}

func (rl *Shell) capitalizeWord() {
	if rl.line.Len() == 0 {
		return
	}

	rl.History.Save()

	startPos := rl.cursor.Pos()

	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	letter := rl.cursor.Char()
	rl.cursor.ReplaceWith(unicode.ToUpper(letter))
	rl.cursor.Set(startPos)
}

func (rl *Shell) characterSearch() {
	if rl.Iterations.Get() < 0 {
		rl.viFindChar(false, false)
	} else {
		rl.viFindChar(true, false)
	}
}

func (rl *Shell) characterSearchBackward() {
	if rl.Iterations.Get() < 0 {
		rl.viFindChar(true, false)
	} else {
		rl.viFindChar(false, false)
	}
}

func (rl *Shell) copyBackwardWord() {
	rl.History.Save()

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust)

	rl.Buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) copyForwardWord() {
	rl.History.Save()

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Forward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust + 1)

	rl.Buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) copyPrevShellWord() {
	rl.History.Save()

	posInit := rl.cursor.Pos()

	rl.backwardShellWord()
	rl.backwardShellWord()

	rl.viSelectAShellWord()

	word := rl.selection.Text()

	rl.cursor.Set(posInit)
	rl.selection.InsertAt(rl.cursor.Pos(), -1)
	rl.cursor.Move(len(word))
}

func (rl *Shell) copyRegionAsKill() {
	rl.History.SkipSave()

	if !rl.selection.Active() {
		return
	}

	rl.Buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) deleteChar() {
	rl.History.Save()

	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.line.CutRune(rl.cursor.Pos())
	}
}

func (rl *Shell) deleteHorizontalWhitespace() {
	rl.History.Save()

	startPos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(false)

	if rl.cursor.Pos() != startPos {
		rl.cursor.Inc()
	}

	bpos := rl.cursor.Pos()
	rl.cursor.ToFirstNonSpace(true)

	if rl.cursor.Pos() != startPos {
		rl.cursor.Dec()
	}

	epos := rl.cursor.Pos()
	rl.line.Cut(bpos, epos)
	rl.cursor.Set(bpos)
}

func (rl *Shell) deleteWord() {
	rl.History.Save()

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.Cut()
}

func (rl *Shell) digitArgument() {
	rl.History.SkipSave()
	keys := rl.Keys.Caller()

	if len(keys) > 1 && keys[0] == inputrc.Esc {
		keys = keys[1:]
	}

	rl.Iterations.Add(string(keys))
}

func (rl *Shell) downCaseWord() {
	rl.History.Save()

	startPos := rl.cursor.Pos()

	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.ReplaceWith(unicode.ToLower)
	rl.cursor.Set(startPos)
}

func (rl *Shell) downLine() {
	lines := rl.Iterations.Get()
	rl.cursor.LineMove(lines)
}

func (rl *Shell) endOfLine() {
	rl.History.SkipSave()

	rl.cursor.EndOfLineAppend()
}

func (rl *Shell) exchangePointAndMark() {

	if rl.cursor.Mark() > rl.line.Len() {
		rl.cursor.ResetMark()
	}

	if rl.cursor.Mark() < 0 {
		cpos := rl.cursor.Pos()
		rl.cursor.Set(0)
		rl.cursor.SetMark()
		rl.cursor.Set(cpos)
	} else {
		mark := rl.cursor.Mark()

		rl.cursor.SetMark()
		rl.cursor.Set(mark)

		rl.selection.MarkRange(rl.cursor.Mark(), rl.cursor.Pos())
		rl.selection.Visual(false)
	}
}

func (rl *Shell) forwardBackwardDeleteChar() {
	switch rl.cursor.Pos() {
	case rl.line.Len():
		rl.backwardDeleteChar()
	default:
		rl.deleteChar()
	}
}

func (rl *Shell) forwardChar() {
	startPos := rl.cursor.Pos()

	if rl.Config.GetBool("history-autosuggest") && rl.cursor.Pos() >= rl.line.Len()-1 {
		rl.autosuggestAccept()
	}

	if rl.cursor.Pos() == startPos {
		rl.acceptInlineSuggestion()
	}

	if rl.cursor.Pos() > startPos {
		return
	}

	rl.History.SkipSave()
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Inc()
	}
}

func (rl *Shell) forwardShellWord() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.selection.SelectAShellWord()
		_, _, tepos, _ := rl.selection.Pop()
		rl.cursor.Set(tepos)
	}
}

func (rl *Shell) forwardWord() {
	rl.History.SkipSave()
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {

		rl.insertAutosuggestPartial(true)

		forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward + 1)
	}
}

func (rl *Shell) insertComment() {
	comment := strings.Trim(rl.Config.GetString("comment-begin"), "\"")

	switch {
	case !rl.Iterations.IsSet():

		cpos := rl.cursor.Pos()
		rl.cursor.BeginningOfLine()
		rl.cursor.InsertAt([]rune(comment)...)
		rl.cursor.Set(cpos)

	default:

		cpos := rl.cursor.Pos()
		rl.cursor.BeginningOfLine()

		bpos := rl.cursor.Pos()
		epos := bpos + len(comment)

		rl.cursor.Set(cpos)

		commentFits := epos < rl.line.Len()

		if commentFits && string((*rl.line)[bpos:epos]) == comment {
			rl.line.Cut(bpos, epos)
			rl.cursor.Move(-1 * len(comment))
		} else {
			rl.line.Insert(bpos, []rune(comment)...)
			rl.cursor.Move(1 * len(comment))
		}
	}

	rl.acceptLineWith(false, false)
}

func (rl *Shell) keywordDecrease() {
	rl.History.Save()
	rl.keywordSwitch(false)
}

func (rl *Shell) keywordIncrease() {
	rl.History.Save()
	rl.keywordSwitch(true)
}

func (rl *Shell) keywordSwitch(increase bool) {
	cpos := strutil.AdjustNumberOperatorPos(rl.cursor.Pos(), *rl.line)

	bpos, epos := rl.line.SelectWord(cpos)
	epos++

	if bpos != 0 && ((*rl.line)[bpos-1] == '+' || (*rl.line)[bpos-1] == '-') {
		bpos--
	}

	selection := string((*rl.line)[bpos:epos])

	for _, switcher := range strutil.KeywordSwitchers() {
		vii := rl.Iterations.Get()

		changed, word, obpos, oepos := switcher(selection, increase, vii)
		if !changed {
			continue
		}

		epos = bpos + oepos
		bpos += obpos

		if cpos < bpos || cpos >= epos {
			continue
		}

		begin := string((*rl.line)[:bpos])
		end := string((*rl.line)[epos:])

		newLine := append([]rune(begin), []rune(word)...)
		newLine = append(newLine, []rune(end)...)
		rl.line.Set(newLine...)
		rl.cursor.Set(bpos + len(word) - 1)

		return
	}
}

func (rl *Shell) killBuffer() {
	rl.History.Save()

	if rl.line.Len() == 0 {
		return
	}

	rl.Buffers.Write(*rl.line...)
	rl.line.Cut(0, rl.line.Len())
}

func (rl *Shell) killLine() {
	rl.Iterations.Reset()
	rl.History.Save()

	if rl.line.Len() == 0 {
		return
	}

	cpos := rl.cursor.Pos()
	rl.cursor.EndOfLineAppend()

	rl.selection.MarkRange(cpos, rl.cursor.Pos())
	text := rl.selection.Cut()

	rl.Buffers.Write([]rune(text)...)
	rl.cursor.Set(cpos)
}

func (rl *Shell) killRegion() {
	rl.History.Save()

	if !rl.selection.Active() {
		return
	}

	rl.Buffers.Write([]rune(rl.selection.Cut())...)
}

func (rl *Shell) killWholeLine() {
	rl.History.Save()

	if rl.line.Len() == 0 {
		return
	}

	rl.Buffers.Write(*rl.line...)
	rl.line.Cut(0, rl.line.Len())
}

func (rl *Shell) killWord() {
	rl.History.Save()

	bpos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(true)
	forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(forward - 1)
	epos := rl.cursor.Pos()

	rl.selection.MarkRange(bpos, epos)
	rl.Buffers.Write([]rune(rl.selection.Cut())...)
	rl.cursor.Set(bpos)
}

func (rl *Shell) overwriteMode() {

	rl.History.Save()

	done := rl.Keymap.PendingCursor()
	defer done()

	cache := make([]rune, 0)

	lineStart := rl.line.Len()

	for {

		key, isAbort := rl.Keys.ReadKey()
		if isAbort {
			break
		}

		if string(key) == inputrc.Unescape(string(`\C-?`)) {
			if rl.cursor.Pos() > lineStart {
				rl.backwardDeleteChar()
			} else if rl.cursor.Pos() > 0 {
				rl.cursor.Dec()
			}

			if len(cache) > 0 && rl.cursor.Pos() < lineStart {
				key = cache[len(cache)-1]
				cache = cache[:len(cache)-1]

				rl.cursor.ReplaceWith(key)
			}
		} else {

			if rl.line.Len() == rl.cursor.Pos() {
				rl.cursor.InsertAt(key)
			} else {
				cache = append(cache, rl.cursor.Char())
				rl.cursor.ReplaceWith(key)
				rl.cursor.Inc()
			}
		}

		rl.Display.Refresh()
	}
}

func (rl *Shell) quoteLine() {
	if rl.line.Len() == 0 {
		return
	}

	for pos, char := range *rl.line {
		if char == '\n' {
			break
		}

		if char == '\'' {
			(*rl.line)[pos] = '"'
		}
	}

	rl.line.Insert(0, '\'')
	rl.line.Insert(rl.line.Len(), '\'')
}

func (rl *Shell) quoteRegion() {
	rl.History.Save()

	rl.selection.Surround('\'', '\'')
	rl.cursor.Inc()
}

func (rl *Shell) quotedInsert() {
	rl.History.SkipSave()
	rl.completer.TrimSuffix()

	done := rl.Keymap.PendingCursor()
	defer done()

	key, _ := rl.Keys.ReadKey()

	quoted, _ := strutil.Quote(key)

	rl.cursor.InsertAt(quoted...)
}

func (rl *Shell) redo() {
	rl.History.Redo()
}

func (rl *Shell) revertLine() {
	rl.History.Revert()
}

func (rl *Shell) selectKeywordNext() {
	rl.History.SkipSave()

	bpos, epos := rl.line.SelectBlankWord(rl.cursor.Pos())

	_, epos, match := rl.selection.SelectKeyword(bpos, epos, true)
	if !match {
		return
	}

	rl.cursor.Set(epos)
	rl.selection.Visual(false)
}

func (rl *Shell) selectKeywordPrev() {
	rl.History.SkipSave()

	bpos, epos := rl.line.SelectBlankWord(rl.cursor.Pos())

	_, epos, match := rl.selection.SelectKeyword(bpos, epos, false)
	if !match {
		return
	}

	rl.cursor.Set(epos)
	rl.selection.Visual(false)
}

func (rl *Shell) selfInsert() {
	rl.History.SkipSave()

	rl.completer.TrimSuffix()

	key := rl.Keys.Caller()

	searching, _, _ := rl.completer.NonIncrementallySearching()
	isearch := rl.Keymap.Local() == keymap.Isearch

	if !searching && !isearch && rl.Config.GetBool("autopairs") {
		if jump := completion.AutopairInsertOrJump(key[0], rl.line, rl.cursor); jump {
			return
		}
	}

	var quoted []rune
	var length int

	if rl.Config.GetBool("output-meta") && key[0] != inputrc.Esc {
		quoted = append(quoted, key[0])
		length = uniseg.StringWidth(string(quoted))
	} else {
		quoted, length = strutil.Quote(key[0])
	}

	rl.cursor.InsertAt(quoted...)
	rl.cursor.Move(-1 * len(quoted))
	rl.cursor.Move(length)
}

func (rl *Shell) setMark() {
	switch {
	case rl.Iterations.IsSet():
		rl.cursor.SetMark()
	default:
		cpos := rl.cursor.Pos()
		mark := rl.Iterations.Get()

		if mark > rl.line.Len()-1 {
			return
		}

		rl.cursor.Set(mark)
		rl.cursor.SetMark()
		rl.cursor.Set(cpos)
	}
}

func (rl *Shell) shellBackwardKillWord() {
	startPos := rl.cursor.Pos()
	if rl.line.Len() == 0 || startPos == 0 {
		return
	}

	rl.History.Save()

	rl.cursor.Dec()
	rl.cursor.ToFirstNonSpace(false)

	unclosed, bpos := strutil.GetQuotedWordStart((*rl.line)[:startPos])

	if !unclosed {
		rl.viSelectAShellWord()
		bpos, _ = rl.selection.Pos()
	}

	rl.cursor.Set(bpos)
	rl.cursor.ToFirstNonSpace(true)
	bpos = rl.cursor.Pos()

	rl.Buffers.Write([]rune((*rl.line)[bpos:startPos])...)
	rl.line.Cut(bpos, startPos)
	rl.selection.Reset()
}

func (rl *Shell) shellKillWord() {
	startPos := rl.cursor.Pos()

	rl.viSelectAShellWord()

	_, epos := rl.selection.Pos()

	rl.Buffers.Write([]rune((*rl.line)[startPos:epos])...)
	rl.line.Cut(startPos, epos)
	rl.cursor.Set(startPos)

	rl.selection.Reset()
}

func (rl *Shell) shellTransposeWords() {
	rl.History.Save()

	startPos := rl.cursor.Pos()

	rl.viSelectAShellWord()
	toTranspose, tbpos, tepos, _ := rl.selection.Pop()

	rl.cursor.Set(tbpos)
	rl.backwardShellWord()

	rl.viSelectAShellWord()
	transposeWith, wbpos, wepos, _ := rl.selection.Pop()

	if wepos > tbpos {
		rl.cursor.Set(startPos)
		return
	}

	begin := string((*rl.line)[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, (*rl.line)[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, (*rl.line)[tepos:]...)
	rl.line.Set(newLine...)

	rl.cursor.Set(tepos)
}

func (rl *Shell) tabInsert() {
	rl.History.SkipSave()

	rl.cursor.InsertAt('\t')
}

func (rl *Shell) transposeChars() {
	if rl.cursor.Pos() < 2 || rl.line.Len() < 2 {
		rl.History.SkipSave()
		return
	}

	rl.History.Save()

	switch {
	case rl.cursor.Pos() == rl.line.Len():
		last := (*rl.line)[rl.cursor.Pos()-1]
		blast := (*rl.line)[rl.cursor.Pos()-2]
		(*rl.line)[rl.cursor.Pos()-2] = last
		(*rl.line)[rl.cursor.Pos()-1] = blast
	default:
		last := rl.cursor.Char()
		blast := (*rl.line)[rl.cursor.Pos()-1]
		(*rl.line)[rl.cursor.Pos()-1] = last
		rl.cursor.ReplaceWith(blast)
	}
}

func (rl *Shell) transposeWords() {
	rl.History.Save()

	startPos := rl.cursor.Pos()
	rl.cursor.ToFirstNonSpace(true)
	rl.cursor.CheckCommand()

	rl.viSelectInWord()
	rl.selection.Visual(false)
	toTranspose, tbpos, tepos, _ := rl.selection.Pop()

	rl.cursor.Set(tbpos)

	if tepos >= rl.line.Len()-1 || rl.Iterations.IsSet() {
		rl.backwardWord()
	} else {
		rl.viForwardWord()
	}

	rl.viSelectInWord()
	rl.selection.Visual(false)
	transposeWith, wbpos, wepos, _ := rl.selection.Pop()

	if tbpos == 0 {
		rl.cursor.Set(startPos)
		return
	}

	if wbpos > tbpos {
		wbpos, tbpos = tbpos, wbpos
		wepos, tepos = tepos, wepos
		transposeWith, toTranspose = toTranspose, transposeWith
	}

	begin := string((*rl.line)[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, (*rl.line)[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, (*rl.line)[tepos:]...)
	rl.line.Set(newLine...)

	rl.cursor.Set(tepos)
}

func (rl *Shell) undoLast() {
	rl.History.Undo()
}

func (rl *Shell) upCaseWord() {
	rl.History.Save()

	startPos := rl.cursor.Pos()

	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.ReplaceWith(unicode.ToUpper)
	rl.cursor.Set(startPos)
}

func (rl *Shell) upLine() {
	lines := rl.Iterations.Get()
	rl.cursor.LineMove(lines * -1)
}

func (rl *Shell) yank() {
	buf := rl.Buffers.Active()

	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.InsertAt(buf...)
	}
}

func (rl *Shell) yankPop() {
	vii := rl.Iterations.Get()

	for i := 1; i <= vii; i++ {
		buf := rl.Buffers.Pop()
		rl.cursor.InsertAt(buf...)
	}
}
