package completion

import (
	"regexp"
	"sync/atomic"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/ui"
)

type Engine struct {
	config        *inputrc.Config
	cached        Completer
	autoCompleter Completer
	hint          *ui.Hint

	keys       *core.Keys
	line       *core.Line
	cursor     *core.Cursor
	selection  *core.Selection
	compLine   *core.Line
	compCursor *core.Cursor
	keymap     *keymap.Engine

	groups      []*group
	sm          SuffixMatcher
	selected    Candidate
	prefix      string
	suffix      string
	inserted    []rune
	usedY       int
	auto        bool
	autoForce   bool
	skipDisplay bool
	regenReq    int32

	IsearchRegex       *regexp.Regexp
	isearchBuf         *core.Line
	isearchCur         *core.Cursor
	isearchName        string
	isearchInsert      bool
	isearchForward     bool
	isearchSubstring   bool
	isearchReplaceLine bool
	isearchStartBuf    string
	isearchStartCursor int
	isearchLast        string
	isearchModeExit    keymap.Mode
}

func Init(eng *Engine, k *core.Keys, l *core.Line, cur *core.Cursor, s *core.Selection, comp Completer) {
	eng.keys = k
	eng.line = l
	eng.cursor = cur
	eng.selection = s
	eng.compLine = l
	eng.compCursor = cur
	eng.autoCompleter = comp
}

func NewEngine(h *ui.Hint, km *keymap.Engine, o *inputrc.Config) *Engine {
	return &Engine{
		config: o,
		hint:   h,
		keymap: km,
	}
}

func (e *Engine) ApplyRegen() {
	if atomic.SwapInt32(&e.regenReq, 0) == 0 {
		return
	}

	if !e.IsActive() {
		return
	}

	e.regenPreservingSelection()
}

func (e *Engine) AutoCompleting() bool {
	return e.keymap.Local() == keymap.Isearch || e.autoForce
}

func (e *Engine) Autocomplete() {
	e.auto = e.needsAutoComplete()

	if e.auto || (!e.IsActive() && e.cursor.Pos() == 0) {
		e.resetValues(true, false)
	}

	if !e.auto {
		return
	}

	if e.cached != nil {
		e.prepare(e.cached())
	} else if e.autoCompleter != nil {
		e.prepare(e.autoCompleter())
	}
}

func (e *Engine) AutocompleteForce() {
	e.autoForce = true
}

func (e *Engine) Cancel(inserted, cached bool) {
	if cached {
		e.cached = nil
		e.hint.Reset()
	}

	if len(e.selected.Value) == 0 && !inserted {
		return
	}

	defer e.cancelCompletedLine()

	if inserted {
		e.compLine.Set(*e.line...)
		e.compCursor.Set(e.cursor.Pos())
	} else {
		e.line.Set(*e.compLine...)
		e.cursor.Set(e.compCursor.Pos())
	}
}

func (e *Engine) ClearMenu(completions bool) {
	e.skipDisplay = false

	e.resetValues(completions, false)

	if e.keymap.Local() == keymap.MenuSelect {
		e.keymap.SetLocal("")
	}
}

func (e *Engine) DisplaySkipped() bool {
	return e.skipDisplay
}

func (e *Engine) Generate(completions Values) {
	e.prepare(completions)

	if e.noCompletions() {
		e.ClearMenu(true)
	}

	if e.hasUniqueCandidate() && e.keymap.Local() != keymap.Isearch {
		e.acceptCandidate()
		e.ClearMenu(true)
	}
}

func (e *Engine) GenerateCached() {
	e.GenerateWith(e.cached)
}

func (e *Engine) GenerateWith(completer Completer) {
	e.cached = completer
	if e.cached == nil {
		return
	}

	e.Generate(e.cached())
}

func (e *Engine) IsActive() bool {
	completing := e.keymap.Local() == keymap.MenuSelect

	isearching := e.keymap.Local() == keymap.Isearch ||
		e.auto || e.autoForce

	nonIsearching, _, _ := e.NonIncrementallySearching()

	return (completing || isearching) && !nonIsearching
}

func (e *Engine) IsInserting() bool {
	return e.selected.Value != ""
}

func (e *Engine) Line() (*core.Line, *core.Cursor) {
	if len(e.selected.Value) > 0 {
		return e.compLine, e.compCursor
	}

	return e.line, e.cursor
}

func (e *Engine) Matches() int {
	comps, _ := e.completionCount()
	return comps
}

func (e *Engine) RequestRegen() {
	atomic.StoreInt32(&e.regenReq, 1)
}

func (e *Engine) Reset() {
	e.autoForce = false
	if !e.IsActive() {
		e.ClearMenu(true)
		return
	}

	e.Cancel(false, true)
	e.ClearMenu(true)
	e.IsearchStop(false)
}

func (e *Engine) ResetForce() {
	e.Cancel(!e.autoForce, true)
	e.ClearMenu(true)

	revertLine := e.keymap.Local() == keymap.Isearch ||
		e.keymap.Local() == keymap.MenuSelect

	e.IsearchStop(revertLine)
}

func (e *Engine) ResetUsedRows() {
	e.usedY = 0
}

func (e *Engine) Select(row, column int) {
	grp := e.currentGroup()

	if grp == nil || len(grp.rows) == 0 {
		return
	}

	e.adjustSelectKeymap()

	row, column = e.adjustCycleKeys(row, column)

	if len(e.selected.Value) > 0 {
		e.cancelCompletedLine()
	}

	defer e.refreshLine()

	done, next := grp.moveSelector(row, column)
	if !done {
		return
	}

	var newGrp *group

	if next {
		e.cycleNextGroup()
		newGrp = e.currentGroup()
		newGrp.firstCell()
	} else {
		e.cyclePreviousGroup()
		newGrp = e.currentGroup()
		newGrp.lastCell()
	}
}

func (e *Engine) SelectTag(next bool) {

	e.adjustSelectKeymap()

	if len(e.groups) <= 1 {
		return
	}

	if len(e.selected.Value) > 0 {
		e.cancelCompletedLine()
	}

	defer e.refreshLine()

	if next {
		e.cycleNextGroup()
		newGrp := e.currentGroup()
		newGrp.firstCell()
	} else {
		e.cyclePreviousGroup()
		newGrp := e.currentGroup()
		newGrp.firstCell()
	}
}

func (e *Engine) SkipDisplay() {
	e.skipDisplay = true
}

func (e *Engine) regenPreservingSelection() {

	completer := e.cached
	if completer == nil {
		completer = e.autoCompleter
	}

	if completer == nil {
		return
	}

	if e.keymap.Local() == keymap.Isearch {
		e.GenerateCached()
		return
	}

	selTag, selValue := e.selected.Tag, e.selected.Value
	hadSelection := selValue != ""

	if hadSelection {
		e.cancelCompletedLine()
	}

	e.prepare(completer())

	if e.noCompletions() {
		e.ClearMenu(true)
		return
	}

	if hadSelection && e.selectCandidate(selTag, selValue) {
		e.insertCandidate()
	}
}

func (e *Engine) selectCandidate(tag, value string) bool {
	want := color.Strip(value)

	for _, grp := range e.groups {
		if tag != "" && grp.tag != tag {
			continue
		}

		for y, row := range grp.rows {
			for x, cand := range row {
				if color.Strip(cand.Value) != want {
					continue
				}

				for _, g := range e.groups {
					g.isCurrent = false
				}

				grp.isCurrent = true
				grp.posX = x
				grp.posY = y

				return true
			}
		}
	}

	return false
}
