package display

import (
	"fmt"
	"strconv"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/strutil"
	"github.com/malivvan/readline/term"
	"github.com/malivvan/readline/ui"
)

func (e *Engine) Refresh() {

	term.BeginBuffer()
	defer term.EndBuffer()

	term.WriteString(term.HideCursor)

	term.MoveCursorBackwards(term.GetWidth())

	if !e.primaryPrinted {
		term.MoveCursorUp(e.cursorRow)
	}

	e.prompt.LastPrint()

	e.computeCoordinates(true)

	e.ensureIndicatorSpace()

	if e.line.Lines() > 0 {
		e.cursorCol, e.cursorRow = core.CoordinatesCursor(e.cursor, e.startCols)
		e.lineCol, e.lineRows = core.CoordinatesLine(e.coordinatesLine(true), e.startCols)
	}

	e.ensureInputSpace()

	e.repaintPromptUpperLines()

	e.renderInputArea()

	termHeight := term.GetLength()
	atBottom := e.startRowKnown() && (e.startRows+e.lineRows) >= termHeight

	if !atBottom {
		term.MoveCursorDown(1)
		term.MoveCursorBackwards(term.GetWidth())
		term.WriteString(term.ClearScreenBelow)
		term.MoveCursorUp(1)
		term.MoveCursorForwards(e.lineCol)
	}

	e.renderHelpers()

	if e.lineRows > e.cursorRow {
		term.MoveCursorUp(e.lineRows - e.cursorRow)
	}

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.cursorCol)

	term.WriteString(term.ShowCursor)
}

func (e *Engine) displayLine() {
	var line string

	if e.highlighter != nil {
		line = e.highlighter(*e.line)
	} else {
		line = string(*e.line)
	}

	if e.opts.GetBool("blink-matching-paren") {
		core.HighlightMatchers(e.selection)
		defer core.ResetMatchers(e.selection)
	}

	line = e.highlightLine([]rune(line), *e.selection)

	suggestionAdded := false
	if len(e.suggested) > e.line.Len() && e.opts.GetBool("history-autosuggest") {
		line += color.Dim + color.Fmt(color.Fg+"242") + string(e.suggested[e.line.Len():]) + color.Reset
		suggestionAdded = true
	}

	currentLine := string(*e.line)
	if !suggestionAdded && e.inlineSuggestionApplies(currentLine) {
		line += color.Dim + color.Fmt(color.Fg+"242") + e.inline[len(currentLine):] + color.Reset
	}

	wrappedAtRightEdge := e.lineCol == 0 && e.line.Len() > 0
	line = strutil.FormatTabs(line)
	if !wrappedAtRightEdge {
		line += term.ClearLineAfter
	}

	e.suggested.Set([]rune(line)...)
	core.DisplayLine(&e.suggested, e.startCols)

	if wrappedAtRightEdge {
		term.WriteString(term.NewlineReturn)
	}
}

func (e *Engine) ensureIndicatorSpace() {

	var indicatorWidth int
	if e.opts.GetBool("multiline-column-numbered") {
		indicatorWidth = len(strconv.Itoa(1)) + 1
	} else {
		indicatorWidth = 2
	}

	if e.prompt.LastUsed() == 0 && e.line.Lines() > 0 {
		var indicator string
		if e.opts.GetBool("multiline-column-numbered") {
			indicator = fmt.Sprintf(color.FgBlackBright+"%d"+color.Reset+" ", 1)
		} else {
			indicator = ui.DefaultMultilineColumn
		}

		e.startCols += indicatorWidth

		term.WriteString(indicator)
	} else if e.line.Lines() > 0 && e.startCols < indicatorWidth {

		padding := indicatorWidth - e.startCols
		term.Printf("%*s", padding, "")

		e.startCols = indicatorWidth
	}
}

func (e *Engine) ensureInputSpace() {

	reserve := e.lineRows + 1

	deficit := 1
	if e.startRowKnown() {
		deficit = (e.startRows + reserve) - term.GetLength()
	}

	if deficit <= 0 {
		return
	}

	term.MoveCursorDown(e.lineRows)

	for range deficit {
		term.WriteString(term.NewlineReturn)
	}

	if e.startRowKnown() {
		e.startRows -= deficit
	}

	term.MoveCursorUp(reserve)
	term.MoveCursorForwards(e.startCols)
}

func (e *Engine) renderHelpers() {
	e.completer.Autocomplete()

	hintRows := ui.CoordinatesHint(e.hint)
	compMatches := e.completer.Matches()
	compSkip := e.completer.DisplaySkipped()

	if hintRows == 0 && (compMatches == 0 || compSkip) {
		e.hintRows = 0
		e.compRows = 0

		return
	}

	term.WriteString(term.NewlineReturn)

	ui.DisplayHint(e.hint)
	e.hintRows = ui.CoordinatesHint(e.hint)

	if compMatches > 0 && !compSkip {
		completion.Display(e.completer, e.AvailableHelperLines())
		e.compRows = completion.Coordinates(e.completer)
	} else {
		e.completer.ResetUsedRows()
		e.compRows = 0
	}

	term.MoveCursorUp(e.compRows)
	term.MoveCursorUp(e.hintRows)
	term.MoveCursorUp(1)

	term.MoveCursorForwards(e.lineCol)
}

func (e *Engine) renderInputArea() {
	e.displayLine()
	e.renderMultilineIndicators()
	e.renderRightPrompt()
}

func (e *Engine) renderMultilineIndicators() {

	if e.line.Lines() == 0 {
		return
	}

	columns := e.opts.GetBool("multiline-column") ||
		e.opts.GetBool("multiline-column-numbered") ||
		e.opts.GetString("multiline-column-custom") != ""
	promptEmpty := e.prompt.LastUsed() == 0

	if !columns && !promptEmpty {
		return
	}

	term.MoveCursorUp(e.lineRows)
	term.MoveCursorBackwards(term.GetWidth())

	printedLines := 0
	numbered := e.opts.GetBool("multiline-column-numbered")

	pipe := ui.DefaultMultilineColumn

	for i := 1; i <= e.line.Lines(); i++ {
		term.WriteString("\n")

		switch {
		case numbered:
			term.Printf(color.FgBlackBright+"%d"+color.Reset+" ", i+1)
		case i == e.line.Lines():
			e.prompt.SecondaryPrint()
		default:
			term.WriteString(pipe)
		}

		printedLines++
	}

	correction := e.lineRows - printedLines
	if correction > 0 {
		term.MoveCursorDown(correction)
	}

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.lineCol)
}

func (e *Engine) renderRightPrompt() {
	e.prompt.RightPrint(e.lineCol, true)

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.lineCol)
}

func (e *Engine) repaintPromptUpperLines() {
	rows := e.prompt.PrimaryUsed()
	if rows == 0 {
		return
	}

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(rows)
	e.prompt.UpperPrint()
	term.MoveCursorForwards(e.startCols)
}
