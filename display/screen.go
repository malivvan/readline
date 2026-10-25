package display

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/history"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/term"
	"github.com/malivvan/readline/ui"
)

type Engine struct {
	highlighter    func(line []rune) string
	startCols      int
	startRows      int
	lineCol        int
	lineRows       int
	cursorRow      int
	cursorCol      int
	hintRows       int
	compRows       int
	primaryPrinted bool

	commentToken string
	commentRegex *regexp.Regexp

	keys      *core.Keys
	line      *core.Line
	suggested core.Line
	cursor    *core.Cursor
	inline    string
	selection *core.Selection
	histories *history.Sources
	prompt    *ui.Prompt
	hint      *ui.Hint
	completer *completion.Engine
	opts      *inputrc.Config
}

var (
	oneThirdTerminalHeight = 3
	halfTerminalHeight     = 2
)

func Init(e *Engine, highlighter func([]rune) string) {
	e.highlighter = highlighter
}

func NewEngine(k *core.Keys, s *core.Selection, h *history.Sources, p *ui.Prompt, i *ui.Hint, c *completion.Engine, opts *inputrc.Config) *Engine {
	return &Engine{
		keys:      k,
		selection: s,
		histories: h,
		prompt:    p,
		hint:      i,
		completer: c,
		opts:      opts,
	}
}

func (e *Engine) AcceptInline(line *core.Line, cursor *core.Cursor) bool {
	if !e.inlineAcceptable(line, cursor) {
		return false
	}

	line.Set([]rune(e.inline)...)
	cursor.Set(line.Len())
	e.inline = ""

	return true
}

func (e *Engine) AcceptInlineWord(line *core.Line, cursor *core.Cursor) {
	if !e.inlineAcceptable(line, cursor) {
		return
	}

	sug := []rune(e.inline)
	from := line.Len()
	remainder := sug[from:]

	pos := 0
	for pos < len(remainder) && unicode.IsSpace(remainder[pos]) {
		pos++
	}

	for pos < len(remainder) && !unicode.IsSpace(remainder[pos]) {
		pos++
	}

	line.Set(sug[:from+pos]...)
	cursor.Set(line.Len())

	if from+pos >= len(sug) {
		e.inline = ""
	}
}

func (e *Engine) AcceptLine() {
	term.BeginBuffer()
	defer term.EndBuffer()

	e.CursorToLineStart()

	e.computeCoordinates(false)

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(e.lineRows)
	term.MoveCursorForwards(e.lineCol)
	term.WriteString(term.ClearScreenBelow)

	e.prompt.RightPrint(e.lineCol, false)

	term.MoveCursorBackwards(term.GetWidth())
	term.WriteString(term.NewlineReturn)
}

func (e *Engine) AvailableHelperLines() int {
	termHeight := term.GetLength()
	compLines := termHeight - e.startRows - e.lineRows - e.hintRows

	if compLines < (termHeight / oneThirdTerminalHeight) {
		compLines = (termHeight / halfTerminalHeight)
	}

	return compLines
}

func (e *Engine) ClearHelpers() {
	term.BeginBuffer()
	defer term.EndBuffer()

	e.CursorBelowLine()
	term.WriteString(term.ClearScreenBelow)

	term.MoveCursorUp(1)
	term.MoveCursorUp(e.lineRows)
	term.MoveCursorDown(e.cursorRow)
	term.MoveCursorForwards(e.cursorCol)
}

func (e *Engine) ClearInlineSuggestion() {
	e.inline = ""
}

func (e *Engine) CursorBelowLine() {
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorDown(e.lineRows)
	term.WriteString(term.NewlineReturn)
}

func (e *Engine) CursorToLineStart() {
	term.MoveCursorBackwards(e.cursorCol)
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorForwards(e.startCols)
}

func (e *Engine) GetInlineSuggestion() string {
	return e.inline
}

func (e *Engine) PrintPrimaryPrompt() {
	e.prompt.PrimaryPrint()
	e.primaryPrinted = true
}

func (e *Engine) RefreshTransient() {
	if !e.opts.GetBool("prompt-transient") {
		return
	}

	term.BeginBuffer()
	defer term.EndBuffer()

	e.CursorToLineStart()
	term.MoveCursorUp(e.prompt.PrimaryUsed())

	e.prompt.TransientPrint()
	e.displayLine()
	term.WriteString(term.NewlineReturn)
}

func (e *Engine) ResetHelpers() {
	e.hint.Reset()
	e.completer.ClearMenu(true)
}

func (e *Engine) SetInlineSuggestion(suggestion string) {
	e.inline = suggestion
}

func (e *Engine) computeCoordinates(suggested bool) {

	e.line, e.cursor = e.completer.Line()
	if e.completer.IsInserting() {
		e.suggested = *e.line
	} else {
		e.suggested = e.histories.Suggest(e.line)
	}

	e.hint.UpdateProvided([]rune(*e.line), e.cursor.Pos())

	if e.opts.GetBool("cursor-position-probe") {
		e.startCols, e.startRows = e.keys.GetCursorPos()
	} else {
		e.startCols, e.startRows = -1, -1
	}

	if e.startCols > 0 {
		e.startCols--
	}

	if e.startCols == -1 {
		e.startCols = e.prompt.LastUsed()
	}

	e.cursorCol, e.cursorRow = core.CoordinatesCursor(e.cursor, e.startCols)

	e.lineCol, e.lineRows = core.CoordinatesLine(e.coordinatesLine(suggested), e.startCols)

	e.primaryPrinted = false
}

func (e *Engine) coordinatesLine(suggested bool) *core.Line {
	currentLine := string(*e.line)
	if e.opts.GetBool("history-autosuggest") && suggested && e.suggested.Len() > e.line.Len() {
		return &e.suggested
	}
	if e.inlineSuggestionApplies(currentLine) {
		inline := core.Line{}
		inline.Set([]rune(e.inline)...)
		return &inline
	}

	return e.line
}

func (e *Engine) inlineSuggestionApplies(currentLine string) bool {
	if e.inline == "" {
		return false
	}
	if e.cursor.Pos() != e.line.Len() {
		return false
	}

	return strings.HasPrefix(e.inline, currentLine) && len(e.inline) > len(currentLine)
}

func (e *Engine) startRowKnown() bool {
	return e.startRows > 0
}
