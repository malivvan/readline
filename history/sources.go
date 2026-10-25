package history

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/ui"
)

type Sources struct {
	line   *core.Line
	cursor *core.Cursor
	hint   *ui.Hint
	config *inputrc.Config

	list       map[string]Source
	names      []string
	maxEntries int
	sourcePos  int
	hpos       int
	cpos       int

	skip    bool
	undoing bool
	last    inputrc.Bind
	lines   map[string]map[int]*lineHistory

	infer      bool
	accepted   bool
	acceptHold bool
	acceptLine core.Line
	acceptErr  error
}

func Complete(h *Sources, forward, filter bool, maxLines int, regex *regexp.Regexp) completion.Values {
	if len(h.list) == 0 {
		return completion.Values{}
	}

	history := h.Current()
	if history == nil {
		return completion.Values{}
	}

	h.hint.Set(color.Bold + color.FgCyanBright + h.names[h.sourcePos] + color.Reset)

	capHint := history.Len()
	if maxLines >= 0 && maxLines < capHint {
		capHint = maxLines + 1
	}

	compLines := make([]completion.Candidate, 0, capHint)
	seen := make(map[string]bool, capHint)

	var (
		histPos int
		done    func(i int) bool
		move    func(inc int) int
	)

	if forward {
		histPos = -1
		done = func(i int) bool { return i < history.Len()-1 && maxLines >= 0 }
		move = func(pos int) int { return pos + 1 }
	} else {
		histPos = history.Len()
		done = func(i int) bool { return i > 0 && maxLines >= 0 }
		move = func(pos int) int { return pos - 1 }
	}

	for done(histPos) {
		histPos = move(histPos)

		line, err := history.GetLine(histPos)
		if err != nil {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		if filter && !strings.HasPrefix(line, string(*h.line)) {
			continue
		} else if regex != nil && !regex.MatchString(line) {
			continue
		}

		if seen[line] {
			continue
		}

		seen[line] = true

		display := strings.ReplaceAll(line, "\n", ` `)

		indexStr := strconv.Itoa(histPos)
		pad := strings.Repeat(" ", len(strconv.Itoa(history.Len()))-len(indexStr))
		display = fmt.Sprintf("%s%s %s%s", color.Dim, indexStr+pad, color.DimReset, display)

		value := completion.Candidate{
			Display: display,
			Value:   line,
		}

		compLines = append(compLines, value)

		maxLines--
	}

	comps := completion.AddRaw(compLines)
	comps.NoSort["*"] = true
	comps.ListLong["*"] = true
	comps.PREFIX = string(*h.line)

	return comps
}

func Init(hist *Sources) {
	defer func() {
		hist.accepted = false
		hist.acceptLine = nil
		hist.acceptErr = nil
		hist.cpos = -1
	}()

	if hist.acceptHold {
		hist.hpos = -1
		hist.line.Set(hist.acceptLine...)
		hist.cursor.Set(hist.line.Len())

		return
	}

	if !hist.infer {
		hist.hpos = -1
		undoHist := hist.getHistoryLineChanges()
		undoHist[hist.hpos] = &lineHistory{}

		return
	}

	switch hist.hpos {
	case -1:
	case 0:
		hist.InferNext()
	default:
		hist.Walk(-1)
	}

	hist.infer = false
}

func NewSources(line *core.Line, cur *core.Cursor, hint *ui.Hint, opts *inputrc.Config) *Sources {
	sources := &Sources{

		list: make(map[string]Source),

		lines: make(map[string]map[int]*lineHistory),

		line:   line,
		cursor: cur,
		cpos:   -1,
		hpos:   -1,
		hint:   hint,
		config: opts,
	}

	sources.names = append(sources.names, standardSourceName)
	sources.list[standardSourceName] = new(memory)

	sources.maxEntries = opts.GetInt("history-size")
	sizeSet := opts.GetString("history-size") != ""

	if sources.maxEntries == 0 && !sizeSet {
		sources.maxEntries = -1
	} else if sources.maxEntries == 0 && sizeSet {
		sources.maxEntries = 500
	}

	return sources
}

func (h *Sources) Accept(hold, infer bool, err error) {
	h.accepted = true
	h.acceptHold = hold
	h.acceptLine = *h.line
	h.acceptErr = err

	if err == nil {
		h.Write(infer)
	}
}

func (h *Sources) Add(name string, hist Source) {
	if len(h.list) == 1 && h.names[0] == standardSourceName {
		delete(h.list, standardSourceName)
		h.names = make([]string, 0)
	}

	h.names = append(h.names, name)
	h.list[name] = hist
}

func (h *Sources) AddFromFile(name, file string) {
	hist := new(fileHistory)
	hist.file = file
	hist.lines, _ = beginHist(file)

	h.Add(name, hist)
}

func (h *Sources) Current() Source {
	if len(h.list) == 0 {
		return nil
	}

	return h.list[h.names[h.sourcePos]]
}

func (h *Sources) Cycle(next bool) {
	switch next {
	case true:
		h.sourcePos++

		if h.sourcePos == len(h.names) {
			h.sourcePos = 0
		}
	case false:
		h.sourcePos--

		if h.sourcePos < 0 {
			h.sourcePos = len(h.names) - 1
		}
	}
}

func (h *Sources) Delete(sources ...string) {
	if len(sources) == 0 {
		h.list = make(map[string]Source)
		h.names = make([]string, 0)

		return
	}

	for _, name := range sources {
		delete(h.list, name)

		for i, hname := range h.names {
			if hname == name {
				h.names = append(h.names[:i], h.names[i+1:]...)
				break
			}
		}
	}

	h.sourcePos = 0
	if !h.infer {
		h.hpos = -1
	}
}

func (h *Sources) Fetch(pos int) {
	history := h.Current()

	if history == nil || history.Len() == 0 {
		return
	}

	if pos < 0 || pos >= history.Len() {
		return
	}

	line, err := history.GetLine(pos)
	if err != nil {
		h.hint.Set(color.FgRed + "history error: " + err.Error())
		return
	}

	h.setLineCursorMatch(line)
}

func (h *Sources) GetLast() string {
	history := h.Current()

	if history == nil || history.Len() == 0 {
		return ""
	}

	last, err := history.GetLine(history.Len() - 1)
	if err != nil {
		return ""
	}

	return last
}

func (h *Sources) InferNext() {
	if len(h.list) == 0 {
		return
	}

	history := h.Current()
	if history == nil {
		return
	}

	_, pos, found := h.match(h.line, nil, false, false, false)
	if !found {
		return
	}

	if pos+1 >= history.Len() {
		return
	}

	line, err := history.GetLine(pos + 1)
	if err != nil {
		return
	}

	h.line.Set([]rune(line)...)
	h.cursor.Set(h.line.Len())
}

func (h *Sources) InsertMatch(line *core.Line, cur *core.Cursor, usePos, fwd, regexp bool) {
	if len(h.list) == 0 {
		return
	}

	if h.Current() == nil {
		return
	}

	line, cur = h.getLine(line, cur)
	preservePoint := cur.Pos() != 0

	if fwd && h.hpos <= -1 {
		h.hpos = -1
		return
	}

	match, pos, found := h.match(line, cur, usePos, fwd, regexp)

	if !found {
		if fwd {
			h.hpos = -1
			h.Undo()
		}

		return
	}

	h.hpos = h.Current().Len() - pos
	h.line.Set([]rune(match)...)

	if preservePoint {
		h.cursor.Set(cur.Pos())
	} else {
		h.cursor.Set(h.line.Len())
	}
}

func (h *Sources) LineAccepted() (bool, string, error) {
	if !h.accepted {
		return false, "", nil
	}

	line := string(h.acceptLine)

	if h.config.GetBool("revert-all-at-newline") {
		for source := range h.lines {
			h.lines[source] = make(map[int]*lineHistory)
		}
	}

	return true, line, h.acceptErr
}

func (h *Sources) Name() string {
	return h.names[h.sourcePos]
}

func (h *Sources) OnLastSource() bool {
	return h.sourcePos == len(h.names)-1
}

func (h *Sources) Suggest(line *core.Line) core.Line {
	if len(h.list) == 0 || len(*line) == 0 {
		return *line
	}

	if h.Current() == nil {
		return *line
	}

	suggested, _, found := h.match(line, nil, false, false, false)
	if !found {
		return *line
	}

	return core.Line([]rune(suggested))
}

func (h *Sources) Walk(pos int) {
	history := h.Current()

	if history == nil || history.Len() == 0 {
		return
	}

	if h.hpos == history.Len() && pos == 1 {
		return
	}

	if h.hpos == -1 && pos > 0 {
		h.skip = false
		h.Save()
		h.cpos = -1
		h.hpos = 0
	}

	h.hpos += pos

	switch {
	case h.hpos < -1:
		h.hpos = -1
		return
	case h.hpos == 0:
		h.restoreLineBuffer()
		return
	case h.hpos > history.Len():
		h.hpos = history.Len()
	}

	var line string
	var err error

	if hist := h.getLineHistory(); hist != nil && len(hist.items) > 0 {
		line = hist.items[len(hist.items)-1].line
	} else if line, err = history.GetLine(history.Len() - h.hpos); err != nil {
		h.hint.Set(color.FgRed + "history error: " + err.Error())
		return
	}

	h.setLineCursorMatch(line)
}

func (h *Sources) Write(infer bool) {
	if infer {
		h.infer = true
		return
	}

	line := string(*h.line)

	if len(strings.TrimSpace(line)) == 0 {
		return
	}

	for _, history := range h.list {
		if history == nil {
			continue
		}

		if h.maxEntries > 0 && history.Len() >= h.maxEntries {
			continue
		}

		var err error

		last, err := history.GetLine(history.Len() - 1)
		if err == nil && last != "" && strings.TrimSpace(last) == strings.TrimSpace(line) {
			return
		}

		_, err = history.Write(line)
		if err != nil {
			h.hint.Set(color.FgRed + err.Error())
		}
	}
}

func (h *Sources) getLine(line *core.Line, cur *core.Cursor) (*core.Line, *core.Cursor) {
	if h.hpos == -1 {
		skip := h.skip
		h.skip = false
		h.Save()
		h.skip = skip
	}

	if line == nil {
		line = new(core.Line)
		cur = core.NewCursor(line)

		hist := h.getHistoryLineChanges()
		if hist == nil {
			return line, cur
		}

		lh := hist[0]
		if lh == nil || len(lh.items) == 0 {
			return line, cur
		}

		undo := lh.items[len(lh.items)-1]
		line.Set([]rune(undo.line)...)
		cur.Set(undo.pos)
	}

	if cur == nil {
		cur = core.NewCursor(line)
	}

	return line, cur
}

func (h *Sources) match(match *core.Line, cur *core.Cursor, usePos, fwd, regex bool) (line string, pos int, found bool) {
	if len(h.list) == 0 {
		return line, pos, found
	}

	history := h.Current()
	if history == nil {
		return line, pos, found
	}

	var histPos int
	var done func(i int) bool
	var move func(inc int) int

	if fwd {
		histPos = -1
		done = func(i int) bool { return i < history.Len() }
		move = func(pos int) int { return pos + 1 }
	} else {
		histPos = history.Len()
		done = func(i int) bool { return i > 0 }
		move = func(pos int) int { return pos - 1 }
	}

	if usePos && h.hpos > -1 {
		histPos = history.Len() - h.hpos
	}

	cline := string(*match)
	if cur != nil && cur.Pos() < match.Len() {
		cline = cline[:cur.Pos()]
	}

	for done(histPos) {

		histPos = move(histPos)

		histline, err := history.GetLine(histPos)
		if err != nil {
			return line, pos, found
		}

		switch {
		case regex:

			if !strings.Contains(histline, cline) {
				continue
			}

		default:

			if len(histline) < len(cline) || (len(cline) > 0 && !strings.HasPrefix(histline, cline)) {
				continue
			}
		}

		return histline, histPos, true
	}

	return "", 0, false
}

func (h *Sources) setLineCursorMatch(next string) {

	if h.cpos == -1 && h.line.Len() > 0 && h.cursor.Pos() <= h.line.Len() {
		h.cpos = h.cursor.Pos()
	}

	h.line.Set([]rune(next)...)

	if h.config.GetBool("history-preserve-point") && h.line.Len() > h.cpos && h.cpos != -1 {
		h.cursor.Set(h.cpos)
	} else {
		h.cursor.Set(h.line.Len())
	}
}
