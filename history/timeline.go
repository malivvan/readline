package history

import (
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
)

type lineHistory struct {
	pos   int
	items []undoItem
}

type undoItem struct {
	line string
	pos  int
}

func (h *Sources) Last() inputrc.Bind {
	return h.last
}

func (h *Sources) Pos() int {
	lh := h.getLineHistory()
	if lh == nil {
		return 0
	}

	return lh.pos
}

func (h *Sources) Redo() {
	h.skip = true
	h.undoing = true

	line := h.getLineHistory()
	if line == nil || len(line.items) == 0 {
		return
	}

	line.pos--

	if line.pos < 1 {
		return
	}

	undo := line.items[len(line.items)-line.pos]
	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}

func (h *Sources) Reset() {
	h.skip = false

	line := h.getLineHistory()
	if line == nil {
		return
	}

	if !h.undoing {
		line.pos = 0
	}

	h.undoing = false
}

func (h *Sources) Revert() {
	line := h.getLineHistory()
	if line == nil || len(line.items) == 0 {
		return
	}

	undo := line.items[0]

	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)

	line.items = make([]undoItem, 0)

	h.Reset()
}

func (h *Sources) Save() {
	defer h.Reset()

	if h.skip {
		return
	}

	line := h.getLineHistory()
	if line == nil {
		return
	}

	if len(line.items) > 0 && line.items[len(line.items)-1].line == string(*h.line) {
		line.items[len(line.items)-1].pos = h.cursor.Pos()
		return
	}

	if line.pos > len(line.items) {
		line.pos = len(line.items)
	}

	line.items = line.items[:len(line.items)-line.pos]

	cur := core.NewCursor(h.line)
	cur.Set(h.cursor.Pos())
	cur.CheckCommand()

	line.items = append(line.items, undoItem{
		line: string(*h.line),
		pos:  cur.Pos(),
	})
}

func (h *Sources) SaveWithCommand(bind inputrc.Bind) {
	h.last = bind
	h.Save()
}

func (h *Sources) SkipSave() {
	h.skip = true
}

func (h *Sources) Undo() {
	h.skip = true
	h.undoing = true

	line := h.getLineHistory()
	if line == nil || len(line.items) == 0 {
		return
	}

	var undo undoItem

	for {
		line.pos++

		if line.pos > len(line.items) {
			line.pos = len(line.items)
			return
		}

		undo = line.items[len(line.items)-line.pos]
		if undo.line != string(*h.line) {
			break
		}
	}

	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}

func (h *Sources) getHistoryLineChanges() map[int]*lineHistory {
	history := h.Current()
	if history == nil {
		return map[int]*lineHistory{}
	}

	source := h.names[h.sourcePos]

	hist := h.lines[source]
	if hist == nil {
		h.lines[source] = make(map[int]*lineHistory)
		hist = h.lines[source]
	}

	return hist
}

func (h *Sources) getLineHistory() *lineHistory {
	hist := h.getHistoryLineChanges()
	if hist == nil {
		return &lineHistory{}
	}

	linePos := -1

	history := h.Current()
	if h.hpos > -1 && history != nil {
		linePos = history.Len() - h.hpos
	}

	if hist[linePos] == nil {
		hist[linePos] = &lineHistory{}
	}

	return hist[linePos]
}

func (h *Sources) restoreLineBuffer() {
	h.hpos = -1

	hist := h.getHistoryLineChanges()
	if hist == nil {
		return
	}

	lh := hist[h.hpos]
	if lh == nil || len(lh.items) == 0 {
		return
	}

	undo := lh.items[len(lh.items)-1]

	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}
