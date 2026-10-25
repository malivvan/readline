package completion

import (
	"unicode"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
)

func UpdateInserted(eng *Engine) {

	choices := len(eng.selected.Value) != 0
	if !eng.auto {
		defer eng.ClearMenu(choices)
	}

	inserted := eng.mustRemoveInserted()
	cached := eng.keymap.Local() != keymap.Isearch && !eng.autoForce

	eng.Cancel(inserted, cached)

	if choices && eng.autoForce && len(eng.selected.Value) == 0 {
		eng.Reset()
	}
}

func commonStringPrefix(a, b string) string {
	ra, rb := []rune(a), []rune(b)
	bound := min(len(ra), len(rb))

	i := 0
	for i < bound && ra[i] == rb[i] {
		i++
	}

	return string(ra[:i])
}

func notMatcher(key rune, matchers string) bool {
	for _, r := range matchers {
		if r == key {
			return false
		}
	}

	return true
}

func (e *Engine) InsertCommonPrefix() {
	common := e.commonPrefix()

	if len([]rune(common)) <= len([]rune(e.prefix)) {
		return
	}

	e.cursor.Move(-1 * e.prefixLength())
	e.line.Cut(e.cursor.Pos(), e.cursor.Pos()+e.prefixLength())
	e.cursor.InsertAt([]rune(common)...)

	e.prefix = common
}

func (e *Engine) TrimSuffix() {
	if e.line.Len() == 0 || e.cursor.Pos() == 0 || len(e.selected.Value) > 0 {
		return
	}

	if e.sm.pos != e.cursor.Pos()-1 || e.sm.string == "" {
		e.sm = SuffixMatcher{}
		return
	}

	suf := (*e.line)[e.cursor.Pos()-1]
	keys := e.keys.Caller()
	key := keys[0]

	if suf == '/' && key != inputrc.Space && notMatcher(key, e.sm.string) {
		return
	}

	if e.sm.Matches(string(key)) || unicode.IsSpace(key) {
		e.cursor.Dec()
		e.line.CutRune(e.cursor.Pos())
	}

	if unicode.IsSpace(key) {
		e.sm = SuffixMatcher{}
	}
}

func (e *Engine) acceptCandidate() {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	e.selected = cur.selected()

	completion := e.prepareSuffix()
	e.inserted = []rune(completion)

	e.cursor.Move(-1 * e.prefixLength())
	e.line.Cut(e.cursor.Pos(), e.cursor.Pos()+e.prefixLength())
	e.cursor.InsertAt(e.inserted...)

	e.inserted = make([]rune, 0)
	e.prefix = ""
	e.suffix = ""
}

func (e *Engine) cancelCompletedLine() {

	e.compLine.Set(*e.line...)
	e.compCursor.Set(e.cursor.Pos())

	e.selected = Candidate{}
}

func (e *Engine) commonPrefix() string {
	var (
		common string
		seen   bool
	)

	for _, grp := range e.groups {
		for _, row := range grp.rows {
			for _, cand := range row {
				if !seen {
					common = cand.Value
					seen = true

					continue
				}

				common = commonStringPrefix(common, cand.Value)
				if common == "" {
					return ""
				}
			}
		}
	}

	return common
}

func (e *Engine) insertCandidate() {
	grp := e.currentGroup()
	if grp == nil {
		return
	}

	e.selected = grp.selected()

	if len([]rune(e.selected.Value)) < e.prefixLength() {
		return
	}

	completion := e.prepareSuffix()
	e.inserted = []rune(completion)

	completed := core.Line(string(*e.line))
	e.compLine = &completed

	e.compCursor = core.NewCursor(e.compLine)
	e.compCursor.Set(e.cursor.Pos())

	e.compCursor.Move(-1 * e.prefixLength())
	e.compLine.Cut(e.compCursor.Pos(), e.compCursor.Pos()+e.prefixLength())
	e.compCursor.InsertAt(e.inserted...)
}

func (e *Engine) mustRemoveInserted() bool {

	if e.keymap.Local() != keymap.Isearch {
		return false
	}

	key, empty := core.PeekKey(e.keys)
	if empty {
		return false
	}

	switch rune(key) {
	case inputrc.Esc, inputrc.Return:
		return false
	default:
		return true
	}
}

func (e *Engine) prepareSuffix() (comp string) {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	comp = e.selected.Value

	inserted := []rune(comp)
	prefix := e.prefixLength()

	if prefix > len(inserted) {
		prefix = len(inserted)
	}

	if len(inserted) > 0 && len(inserted[prefix:]) <= 1 {
		return
	}

	e.sm = cur.noSpace
	e.sm.pos = e.cursor.Pos() + len(inserted) - prefix - 1

	return comp
}

// prefixLength returns the length of the completion prefix in runes. The line,
// cursor and selection APIs all count positions in runes, so the byte length of
// the prefix must never be used to move or cut them.
func (e *Engine) prefixLength() int {
	return len([]rune(e.prefix))
}

func (e *Engine) refreshLine() {
	if e.noCompletions() {
		e.Cancel(true, true)
		return
	}

	if e.currentGroup() == nil {
		return
	}

	if e.hasUniqueCandidate() && e.keymap.Local() != keymap.Isearch {
		e.acceptCandidate()
		e.ResetForce()
	} else {
		e.insertCandidate()
	}
}
