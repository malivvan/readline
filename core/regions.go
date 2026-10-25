package core

import (
	"regexp"
	"unicode"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
)

type Selection struct {
	Type       string
	active     bool
	visual     bool
	visualLine bool
	bpos       int
	epos       int
	kpos       int
	kmpos      int

	fg        string
	bg        string
	surrounds []Selection

	line   *Line
	cursor *Cursor
}

func HighlightMatchers(sel *Selection) {
	cpos := sel.cursor.Pos()

	if sel.line.Len() == 0 || cpos == sel.line.Len() {
		return
	}

	if strutil.IsBracket(sel.cursor.Char()) {
		var adjust, ppos int

		split, index, pos := sel.line.TokenizeBlock(cpos)

		switch {
		case len(split) == 0:
			return
		case pos == 0 && len(split) > index:
			adjust = len(split[index])
		default:
			adjust = pos * -1
		}

		ppos = cpos + adjust

		sel.surrounds = append(sel.surrounds, Selection{
			Type:   "matcher",
			active: true,
			visual: true,
			bpos:   ppos,
			epos:   ppos,
			bg:     color.Fmt("240"),
			line:   sel.line,
			cursor: sel.cursor,
		})
	}
}

func NewSelection(line *Line, cursor *Cursor) *Selection {
	return &Selection{
		bpos:   -1,
		epos:   -1,
		line:   line,
		cursor: cursor,
	}
}

func ResetMatchers(sel *Selection) {
	var surrounds []Selection

	for _, surround := range sel.surrounds {
		if surround.Type == "matcher" {
			continue
		}

		surrounds = append(surrounds, surround)
	}

	sel.surrounds = surrounds
}

func blankRune(char rune) bool {
	return unicode.IsSpace(char) && char != inputrc.Newline
}

func (s *Selection) Active() bool {
	return s.active
}

func (s *Selection) Cursor() int {
	bpos, epos := s.Pos()
	if bpos == -1 && epos == -1 {
		return s.cursor.Pos()
	}

	cpos := bpos

	if !s.visual || !s.visualLine {
		return cpos
	}

	var indent int
	pos := s.cursor.Pos()

	for cpos = pos - 1; cpos >= 0; cpos-- {
		if (*s.line)[cpos] == '\n' {
			break
		}
	}

	indent = pos - cpos - 1

	var hpos, rpos int

	if epos < s.line.Len() {
		hpos = epos + 1
		rpos = bpos
	} else {
		for hpos = bpos - 2; hpos >= 0; hpos-- {
			if (*s.line)[hpos] == '\n' {
				break
			}
		}

		if hpos < -1 {
			hpos = -1
		}

		hpos++
		rpos = hpos
	}

	for cpos = hpos; cpos < s.line.Len(); cpos++ {
		if (*s.line)[cpos] == '\n' {
			break
		}

		if hpos+indent <= cpos {
			break
		}
	}

	cpos = rpos + cpos - hpos

	return cpos
}

func (s *Selection) Cut() (buf string) {
	if s.line.Len() == 0 {
		return
	}

	defer s.Reset()

	switch {
	case len(s.surrounds) > 0:
		offset := 0

		for _, surround := range s.surrounds {
			s.line.CutRune(surround.bpos - offset)

			offset++
		}

	default:
		bpos, epos := s.Pos()
		if bpos == -1 || epos == -1 {
			return
		}

		buf = s.Text()

		s.line.Cut(bpos, epos)
	}

	return
}

func (s *Selection) Highlights() (fg, bg string) {
	return s.fg, s.bg
}

func (s *Selection) InsertAt(bpos, epos int) {
	bpos, epos, valid := s.checkRange(bpos, epos)
	if !valid {
		return
	}

	defer s.Reset()

	buf := s.Text()

	switch {
	case epos == -1, bpos == epos:
		s.line.Insert(bpos, []rune(buf)...)
	default:
		s.line.InsertBetween(bpos, epos, []rune(buf)...)
	}
}

func (s *Selection) IsVisual() bool {
	return s.visual
}

func (s *Selection) Len() int {
	if s.line.Len() == 0 || (s.bpos == s.epos) {
		return 0
	}

	bpos, epos := s.Pos()
	buf := (*s.line)[bpos:epos]

	return len(buf)
}

func (s *Selection) Mark(pos int) {
	if pos < 0 || pos > s.line.Len() {
		return
	}

	s.MarkRange(pos, -1)
}

func (s *Selection) MarkRange(bpos, epos int) {
	bpos, epos, valid := s.checkRange(bpos, epos)
	if !valid {
		return
	}

	s.Type = "visual"
	s.active = true
	s.bpos = bpos
	s.epos = epos
	s.bg = color.BgBlue
}

func (s *Selection) MarkSurround(bpos, epos int) {
	if bpos < 0 || epos > s.line.Len()-1 {
		return
	}

	s.active = true

	for _, pos := range []int{bpos, epos} {
		s.surrounds = append(s.surrounds, Selection{
			Type:   "surround",
			active: true,
			visual: true,
			bpos:   pos,
			epos:   pos,
			bg:     color.BgRed,
			line:   s.line,
			cursor: s.cursor,
		})
	}
}

func (s *Selection) Pop() (buf string, bpos, epos, cpos int) {
	if s.line.Len() == 0 {
		return "", -1, -1, 0
	}

	defer s.Reset()

	bpos, epos = s.Pos()
	if bpos == -1 || epos == -1 {
		return "", -1, -1, 0
	}

	cpos = s.Cursor()
	buf = string((*s.line)[bpos:epos])

	return buf, bpos, epos, cpos
}

func (s *Selection) Pos() (bpos, epos int) {
	if s.line.Len() == 0 || !s.active {
		return -1, -1
	}

	bpos, epos, valid := s.checkRange(s.bpos, s.epos)
	if !valid {
		return bpos, epos
	}

	s.bpos, s.epos = bpos, epos

	if epos == -1 {
		bpos, epos = s.selectToCursor(bpos)
	}

	if s.visual {
		epos++
	}

	bpos, epos, valid = s.checkRange(bpos, epos)
	if !valid {
		return -1, -1
	}

	return bpos, epos
}

func (s *Selection) ReplaceWith(replacer func(r rune) rune) {
	if s.line.Len() == 0 || s.Len() == 0 {
		return
	}

	defer s.Reset()

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return
	}

	for pos := bpos; pos < epos; pos++ {
		char := (*s.line)[pos]
		char = replacer(char)
		(*s.line)[pos] = char
	}
}

func (s *Selection) Reset() {
	s.Type = ""
	s.active = false
	s.visual = false
	s.visualLine = false
	s.bpos = -1
	s.epos = -1
	s.kpos = 0
	s.fg = ""
	s.bg = ""

	surrounds := make([]Selection, 0)

	for _, surround := range s.surrounds {
		if surround.Type != "matcher" {
			continue
		}

		surrounds = append(surrounds, surround)
	}

	s.surrounds = surrounds
}

func (s *Selection) SelectABlankWord() (bpos, epos int) {
	if s.line.Len() == 0 {
		return
	}

	bpos = s.cursor.Pos()
	spaceBefore, spaceUnder := s.spacesAroundWord(bpos)

	if !spaceUnder && !spaceBefore {
		s.cursor.Inc()
		s.cursor.Move(s.line.Backward(s.line.TokenizeSpace, s.cursor.Pos()))
		bpos = s.cursor.Pos()
	} else {
		s.cursor.ToFirstNonSpace(true)
	}

	s.cursor.Move(s.line.ForwardEnd(s.line.TokenizeSpace, s.cursor.Pos()))
	spaceAfter := s.cursor.Pos() < s.line.Len()-1 && blankRune((*s.line)[s.cursor.Pos()+1])

	bpos = s.adjustWordSelection(spaceBefore, spaceUnder, spaceAfter, bpos)

	if !s.Active() || bpos < s.cursor.Pos() {
		s.Mark(bpos)
	}

	return bpos, s.cursor.Pos()
}

func (s *Selection) SelectAShellWord() (bpos, epos int) {
	if s.line.Len() == 0 {
		return bpos, epos
	}

	s.cursor.CheckCommand()
	s.cursor.ToFirstNonSpace(true)

	sBpos, sEpos := s.line.SurroundQuotes(true, s.cursor.Pos())
	dBpos, dEpos := s.line.SurroundQuotes(false, s.cursor.Pos())

	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	if mark == -1 && cpos == -1 {
		mark, cpos = s.line.SelectBlankWord(s.cursor.Pos())
	}

	s.cursor.Set(mark)

	for {
		spaceBefore := mark > 0 && blankRune((*s.line)[mark-1])
		if spaceBefore {
			s.cursor.Dec()
			s.cursor.ToFirstNonSpace(false)
			s.cursor.Inc()

			break
		} else if mark == 0 {
			break
		}

		s.cursor.Move(s.line.Backward(s.line.TokenizeSpace, s.cursor.Pos()))
		mark = s.cursor.Pos()
	}

	bpos = s.cursor.Pos()
	s.cursor.Set(cpos)

	for {
		spaceAfter := cpos < s.line.Len()-1 && blankRune((*s.line)[cpos+1])
		if spaceAfter || cpos == s.line.Len()-1 {
			break
		}

		s.cursor.Move(s.line.ForwardEnd(s.line.TokenizeSpace, cpos))
		cpos = s.cursor.Pos()
	}

	if !s.Active() || bpos < s.cursor.Pos() {
		s.Mark(bpos)
	}

	return bpos, cpos
}

func (s *Selection) SelectAWord() (bpos, epos int) {
	if s.line.Len() == 0 {
		return
	}

	bpos = s.cursor.Pos()
	cpos := bpos

	spaceBefore, spaceUnder := s.spacesAroundWord(bpos)

	bpos, epos = s.line.SelectWord(cpos)
	s.cursor.Set(epos)
	cpos = s.cursor.Pos()

	spaceAfter := cpos < s.line.Len()-1 && blankRune((*s.line)[cpos+1])

	bpos = s.adjustWordSelection(spaceBefore, spaceUnder, spaceAfter, bpos)

	if !s.Active() || bpos < cpos {
		s.Mark(bpos)
	}

	return bpos, epos
}

func (s *Selection) SelectKeyword(bpos, epos int, next bool) (kbpos, kepos int, match bool) {
	if s.line.Len() == 0 {
		return bpos, epos, false
	}

	selection := (*s.line)[bpos:epos]

	_, match, kbpos, kepos = s.matchKeyword(selection, bpos, next)
	if !match {
		return bpos, epos, false
	}

	kbpos, kepos, match = s.checkRange(kbpos, kepos)
	if !match {
		return bpos, epos, false
	}

	s.Mark(kbpos)

	return kbpos, kepos, true
}

func (s *Selection) Surround(bchar, echar rune) {
	if s.line.Len() == 0 || s.Len() == 0 {
		return
	}

	defer s.Reset()

	text := []rune(s.Text())
	buf := make([]rune, 0, 1+len(text)+1)
	buf = append(buf, bchar)
	buf = append(buf, text...)
	buf = append(buf, echar)

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return
	}

	s.line.InsertBetween(bpos, epos, buf...)
}

func (s *Selection) Surrounds() []Selection {
	return s.surrounds
}

func (s *Selection) Text() string {
	if s.line.Len() == 0 {
		return ""
	}

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return ""
	}

	return string((*s.line)[bpos:epos])
}

func (s *Selection) Visual(line bool) {
	s.visual = true
	s.visualLine = line
}

func (s *Selection) adjustWordSelection(_, under, after bool, bpos int) int {
	if after && !under {
		s.cursor.Inc()
		s.cursor.ToFirstNonSpace(true)
		s.cursor.Dec()
	} else if !after {
		epos := s.cursor.Pos()
		s.cursor.Set(bpos - 1)
		s.cursor.ToFirstNonSpace(false)
		s.cursor.Inc()
		bpos = s.cursor.Pos()
		s.cursor.Set(epos)
	}

	return bpos
}

func (s *Selection) checkRange(bpos, epos int) (int, int, bool) {

	switch {
	case s.line.Len() == 0:
		return -1, -1, false
	case bpos < 0 && epos < 0:
		return -1, -1, false
	case bpos > s.line.Len() && epos > s.line.Len():
		return -1, -1, false
	}

	if bpos > s.line.Len() {
		bpos = s.line.Len()
	}

	if epos > s.line.Len() {
		epos = s.line.Len()
	}

	if bpos < 0 {
		bpos, epos = epos, -1
	} else if epos < 0 {
		epos = -1
	}

	if bpos > epos && epos != -1 {
		bpos, epos = epos, bpos
	}

	return bpos, epos, true
}

func (s *Selection) cycleSubgroup(groups []int, bbpos int, next bool) (bpos, epos int, cycled bool) {
	canCycleSubgroup := (next && s.kmpos < len(groups)-1) || (!next && s.kmpos > 1)

	if !canCycleSubgroup {
		return
	}

	if next {
		s.kmpos++
	} else {
		s.kmpos--
	}

	return bbpos + groups[s.kmpos-1], bbpos + groups[s.kmpos], true
}

func (s *Selection) matchKeyword(buf []rune, bbpos int, next bool) (name string, found bool, bpos, epos int) {
	matchersNames := []string{
		"URI",
	}

	matchers := map[string]*regexp.Regexp{
		"URI": regexp.MustCompile(`https?:\/\/(?:www\.)?([-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b)*(\/[\/\d\w\.-]*)*(?:[\?])*(.+)*`),
	}

	var matcherName string
	var groups []int

	if s.kpos > 0 && s.kpos <= len(matchersNames) {
		matcherName = matchersNames[s.kpos-1]
		matcher := matchers[matcherName]

		groups = s.runMatcher(buf, matcher)
	}

	if sbpos, sepos, cycled := s.cycleSubgroup(groups, bbpos, next); cycled {
		return matcherName, true, sbpos, sepos
	}

	prevPos := s.kpos
	s.kmpos = 1

	if next {
		s.kpos++
	} else {
		s.kpos--
	}

	if s.kpos > len(matchersNames) {
		s.kpos = 1
	}

	if s.kpos < 1 {
		s.kpos = len(matchersNames)
	}

	var (
		kpos = s.kpos
		done func(i int) bool
		move func(inc int) int
	)

	if next {
		done = func(i int) bool { return i <= len(matchersNames) }
		move = func(_ int) int { return kpos + 1 }
	} else {
		done = func(i int) bool { return i > 0 }
		move = func(_ int) int { return kpos - 1 }
	}

	for done(kpos) {
		matcherName = matchersNames[kpos-1]
		matcher := matchers[matcherName]

		groups = s.runMatcher(buf, matcher)

		if len(groups) <= 1 {
			kpos = move(kpos)
			continue
		}

		if !next && prevPos == 1 {
			s.kmpos = len(groups)

			return matcherName, true, bbpos + groups[len(groups)-2], bbpos + groups[len(groups)-1]
		}

		return matcherName, true, bbpos + groups[0], bbpos + groups[1]
	}

	return "", false, -1, -1
}

func (s *Selection) runMatcher(buf []rune, matcher *regexp.Regexp) (groups []int) {
	if matcher == nil {
		return
	}

	matches := matcher.FindAllStringSubmatchIndex(string(buf), -1)
	if len(matches) == 0 {
		return
	}

	return matches[0]
}

func (s *Selection) selectToCursor(bpos int) (int, int) {
	var epos int

	if s.cursor.Pos() < bpos {
		bpos, epos = s.cursor.Pos(), bpos
	} else {
		epos = s.cursor.Pos()
	}

	if s.visualLine {

		for bpos--; bpos >= 0; bpos-- {
			if (*s.line)[bpos] == '\n' {
				bpos++
				break
			}
		}

		if bpos == -1 {
			bpos = 0
		}

		for ; epos < s.line.Len(); epos++ {
			if epos == -1 {
				epos = 0
			}

			if (*s.line)[epos] == '\n' {
				break
			}
		}
	}

	if bpos > epos {
		bpos, epos = epos, bpos
	}

	return bpos, epos
}

func (s *Selection) spacesAroundWord(cpos int) (before, under bool) {
	under = blankRune(s.cursor.Char())
	before = cpos > 0 && blankRune((*s.line)[cpos-1])

	return
}
