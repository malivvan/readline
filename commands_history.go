package readline

import (
	"strings"

	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
)

func (rl *Shell) acceptAndHold() {
	rl.acceptLineWith(false, true)
}

func (rl *Shell) acceptAndInferNextHistory() {
	rl.acceptLineWith(true, false)
}

func (rl *Shell) acceptLine() {
	rl.acceptLineWith(false, false)
}

func (rl *Shell) acceptLineAndDownHistory() {
	rl.acceptLineWith(true, false)
}

func (rl *Shell) acceptLineWith(infer, hold bool) {

	rl.completer.Reset()

	searching, forward, substring := rl.completer.NonIncrementallySearching()
	if searching {
		defer rl.completer.NonIsearchStop()

		line, cursor, _ := rl.completer.GetBuffer()
		rl.History.InsertMatch(line, cursor, true, forward, substring)

		return
	}

	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	if rl.AcceptMultiline == nil {
		rl.Macros.StopRecord(rl.Keys.Caller()...)

		rl.Display.AcceptLine()
		rl.History.Accept(hold, infer, nil)

		return
	}

	if rl.AcceptMultiline(*rl.line) {
		rl.Macros.StopRecord(rl.Keys.Caller()...)

		rl.Display.AcceptLine()
		rl.History.Accept(hold, infer, nil)

		return
	}

	rl.line.Insert(rl.cursor.Pos(), '\n')
	rl.cursor.Inc()
}

func (rl *Shell) autosuggestAccept() {
	suggested := rl.History.Suggest(rl.line)

	if suggested.Len() <= rl.line.Len() {
		return
	}

	rl.line.Set(suggested...)
	rl.cursor.Set(len(suggested))
}

func (rl *Shell) autosuggestDisable() {
	rl.History.SkipSave()
	rl.Config.Vars["history-autosuggest"] = false
}

func (rl *Shell) autosuggestEnable() {
	rl.History.SkipSave()
	rl.Config.Vars["history-autosuggest"] = true
}

func (rl *Shell) autosuggestExecute() {
	suggested := rl.History.Suggest(rl.line)

	if suggested.Len() <= rl.line.Len() {
		return
	}

	rl.line.Set(suggested...)
	rl.cursor.Set(len(suggested))

	rl.acceptLine()
}

func (rl *Shell) autosuggestToggle() {
	if rl.Config.GetBool("history-autosuggest") {
		rl.autosuggestDisable()
	} else {
		rl.autosuggestEnable()
	}
}

func (rl *Shell) beginningOfBufferOrHistory() {
	if rl.cursor.Pos() > 0 {
		rl.History.SkipSave()
		rl.cursor.Set(0)

		return
	}

	rl.History.Save()
	rl.beginningOfHistory()
}

func (rl *Shell) beginningOfHistory() {
	rl.History.SkipSave()

	history := rl.History.Current()
	if history == nil {
		return
	}

	rl.History.Walk(history.Len())
}

func (rl *Shell) beginningOfLineHist() {
	switch {
	case rl.cursor.Pos() > 0:
		rl.History.SkipSave()

		if rl.cursor.AtBeginningOfLine() {
			rl.cursor.Dec()
		}

		rl.beginningOfLine()
	default:
		rl.History.Save()
		rl.History.Walk(1)
	}
}

func (rl *Shell) downHistory() {
	rl.History.Save()
	rl.History.Walk(-1)
}

func (rl *Shell) downLineOrHistory() {
	times := rl.Iterations.Get()
	linesDown := rl.line.Lines() - rl.cursor.LinePos()

	if linesDown > 0 {
		rl.cursor.LineMove(times)
		times -= linesDown
	}

	if times > 0 {
		rl.History.Walk(times * -1)
	}
}

func (rl *Shell) downLineOrSearch() {
	rl.History.SkipSave()

	switch {
	case rl.cursor.LinePos() < rl.line.Lines():
		rl.cursor.LineMove(1)
	default:
		rl.historySearchForward()
	}
}

func (rl *Shell) downLineOrSelect() {
	rl.History.SkipSave()

	switch {
	case rl.cursor.LinePos() < rl.line.Lines():
		rl.cursor.LineMove(1)
	default:
		rl.menuComplete()
	}
}

func (rl *Shell) endOfBufferOrHistory() {
	if rl.cursor.Pos() < rl.line.Len()-1 {
		rl.History.SkipSave()
		rl.cursor.Set(rl.line.Len())

		return
	}

	rl.History.Save()
	rl.endOfHistory()
}

func (rl *Shell) endOfHistory() {
	history := rl.History.Current()

	if history == nil {
		return
	}

	rl.History.Walk(-history.Len() + 1)
}

func (rl *Shell) endOfLineHist() {
	rl.History.SkipSave()

	switch {
	case rl.cursor.Pos() < rl.line.Len()-1:
		rl.History.SkipSave()

		if rl.cursor.AtEndOfLine() {
			rl.cursor.Inc()
		}

		rl.endOfLine()

	default:
		rl.History.Save()
		rl.History.Walk(-1)
	}
}

func (rl *Shell) fetchHistory() {
	if rl.Iterations.IsSet() {
		rl.History.Fetch(rl.Iterations.Get())
	} else {
		rl.History.Fetch(0)
	}
}

func (rl *Shell) forwardSearchHistory() {
	rl.History.SkipSave()

	forward := true
	filterLine := false
	regexp := true

	rl.historyCompletion(forward, filterLine, regexp)
}

func (rl *Shell) historySearchBackward() {
	rl.History.Save()

	usePos := true
	forward := false
	regexp := false

	rl.History.InsertMatch(nil, nil, usePos, forward, regexp)
}

func (rl *Shell) historySearchForward() {
	rl.History.Save()

	usePos := true
	forward := true
	regexp := false

	rl.History.InsertMatch(nil, nil, usePos, forward, regexp)
}

func (rl *Shell) historySourceNext() {
	rl.History.Cycle(true)
}

func (rl *Shell) historySourcePrev() {
	rl.History.Cycle(false)
}

func (rl *Shell) historySubstringSearchBackward() {
	usePos := true
	forward := false
	regexp := true

	rl.History.InsertMatch(rl.line, rl.cursor, usePos, forward, regexp)
}

func (rl *Shell) historySubstringSearchForward() {
	usePos := true
	forward := true
	regexp := true

	rl.History.InsertMatch(rl.line, rl.cursor, usePos, forward, regexp)
}

func (rl *Shell) incrementalForwardSearchHistory() {
	rl.History.SkipSave()

	forward := true
	filter := true
	regexp := false

	rl.historyCompletion(forward, filter, regexp)
}

func (rl *Shell) incrementalReverseSearchHistory() {
	rl.History.SkipSave()

	forward := false
	filter := true
	regexp := false

	rl.historyCompletion(forward, filter, regexp)
}

func (rl *Shell) inferNextHistory() {
	rl.History.SkipSave()
	rl.History.InferNext()
}

func (rl *Shell) insertAutosuggestPartial(emacs bool) {
	cpos := rl.cursor.Pos()
	if cpos < rl.line.Len()-1 {
		return
	}

	if !rl.Config.GetBool("history-autosuggest") {
		return
	}

	suggested := rl.History.Suggest(rl.line)

	if suggested.Len() > rl.line.Len() {
		var forward int

		if emacs {
			forward = suggested.ForwardEnd(suggested.Tokenize, cpos)
		} else {
			forward = suggested.Forward(suggested.Tokenize, cpos)
		}

		if cpos+1+forward > suggested.Len() {
			forward = suggested.Len() - cpos - 1
		}

		rl.line.Insert(cpos+1, suggested[cpos+1:cpos+forward+1]...)
	}
}

func (rl *Shell) magicSpace() {
	cpos := rl.cursor.Pos()
	lineLen := rl.line.Len()

	if lineLen == 0 || (cpos == lineLen && (*rl.line)[cpos-1] == inputrc.Space) {
		rl.selfInsert()
		return
	}

	rl.viSelectInBlankWord()
	word, bpos, _, _ := rl.selection.Pop()
	rl.cursor.Set(cpos)

	if len(strings.TrimSpace(word)) == 0 {
		rl.selfInsert()
		return
	}

	if !strings.HasPrefix(word, "!") || word == "!" {
		rl.selfInsert()
		return
	}

	pattern := (*rl.line)[bpos+1:]
	suggested := rl.History.Suggest(&pattern)

	if string(suggested) == string(pattern) {
		rl.selfInsert()
		return
	}

	rl.History.Save()
	rl.line.Cut(bpos, lineLen)
	rl.line.Insert(bpos, suggested...)
	rl.cursor.Set(bpos + suggested.Len())
}

func (rl *Shell) nonIncrementalForwardSearchHistory() {
	repeat := false
	forward := true
	regexp := true

	rl.completer.NonIsearchStart(rl.History.Name(), repeat, forward, regexp)
}

func (rl *Shell) nonIncrementalReverseSearchHistory() {
	repeat := false
	forward := false
	regexp := true

	rl.completer.NonIsearchStart(rl.History.Name(), repeat, forward, regexp)
}

func (rl *Shell) reverseSearchHistory() {
	rl.History.SkipSave()

	forward := false
	filterLine := false
	regexp := true

	rl.historyCompletion(forward, filterLine, regexp)
}

func (rl *Shell) saveLine() {
	rl.History.Write(false)
	rl.History.Revert()
}

func (rl *Shell) upHistory() {
	rl.History.Save()
	rl.History.Walk(1)
}

func (rl *Shell) upLineOrHistory() {
	times := rl.Iterations.Get()
	linesUp := rl.cursor.LinePos()

	if linesUp > 0 {
		rl.cursor.LineMove(times * -1)
		times -= linesUp
	}

	if times > 0 {
		rl.History.Walk(times)
	}
}

func (rl *Shell) upLineOrSearch() {
	rl.History.SkipSave()

	switch {
	case rl.cursor.LinePos() > 0:
		rl.cursor.LineMove(-1)
	default:
		rl.historySearchBackward()
	}
}

func (rl *Shell) viDownLineOrHistory() {
	rl.downLineOrHistory()
	rl.viFirstPrint()
}

func (rl *Shell) yankLastArg() {

	last := rl.History.GetLast()
	if last == "" {
		return
	}

	words, err := strutil.Split(last)
	if err != nil || len(words) == 0 {
		return
	}

	lastArg := words[len(words)-1]
	if strings.ContainsAny(lastArg, " \t") {
		if strings.Contains(lastArg, "\"") {
			lastArg = "'" + lastArg + "'"
		} else {
			lastArg = "\"" + lastArg + "\""
		}
	}

	rl.line.Insert(rl.cursor.Pos(), []rune(lastArg)...)
	rl.cursor.Move(len(lastArg))
}

func (rl *Shell) yankNthArg() {

	last := rl.History.GetLast()
	if last == "" {
		return
	}

	words, err := strutil.Split(last)
	if err != nil || len(words) == 0 {
		return
	}

	var lastArg string

	argNth := rl.Iterations.Get()

	if argNth < 0 {
		argNth = len(words) + argNth + 1
	}

	if argNth < 1 || argNth > len(words) {
		return
	}

	lastArg = words[argNth-1]

	if strings.ContainsAny(lastArg, " \t") {
		if strings.Contains(lastArg, "\"") {
			lastArg = "'" + lastArg + "'"
		} else {
			lastArg = "\"" + lastArg + "\""
		}
	}

	rl.line.Insert(rl.line.Len(), []rune(lastArg)...)
	rl.cursor.Move(len(lastArg))
}
