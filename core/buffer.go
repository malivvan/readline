package core

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
	"github.com/malivvan/readline/term"
)

type Line []rune

type Tokenizer func(cursorPos int) (split []string, index int, newPos int)

func CoordinatesLine(l *Line, indent int) (int, int) {
	var usedY, usedX, lineStart, lineIdx int

	for i, r := range *l {
		if r == '\n' {
			_, y := strutil.LineSpan((*l)[lineStart:i], lineIdx, indent)
			usedY += y

			lineStart = i + 1
			lineIdx++
		}
	}

	x, y := strutil.LineSpan((*l)[lineStart:], lineIdx, indent)
	usedY += y
	usedX = x

	return usedX, usedY
}

func DisplayLine(l *Line, indent int) {
	var builtLine strings.Builder
	var lineLen int

	for _, r := range *l {
		if r == '\n' {
			builtLine.WriteString(color.BgDefault)

			if lineLen < term.GetWidth() {
				builtLine.WriteString(term.ClearLineAfter)
			}

			builtLine.WriteString(term.NewlineReturn)
			fmt.Fprintf(&builtLine, "\x1b[%dC", indent)
			builtLine.WriteString(term.ClearLineBefore)

			lineLen = 0
		} else {
			builtLine.WriteRune(r)

			lineLen++
		}
	}

	if l.Len() > 0 && (*l)[l.Len()-1] == '\n' {
		builtLine.WriteString(color.BgDefault)
		builtLine.WriteString(term.ClearLineAfter)
		builtLine.WriteString(term.NewlineReturn)
		fmt.Fprintf(&builtLine, "\x1b[%dC", indent)
		builtLine.WriteString(term.ClearLineBefore)
	}

	builtLine.WriteString(color.BgDefault)

	term.WriteString(builtLine.String())
}

func finishToken(idx, count, cpos, match int, pos map[int]int, line []rune, split []string) (int, []string) {
	if match == count {
		split = append(split, string(line[pos[count]:idx]))
		return count, split
	}

	if idx == cpos {
		start := pos[count]
		if start == 0 {
			start++
		}

		split = []string{
			string(line[:start-1]),
			string(line[pos[count]:idx]),
		}

		return count, split
	}

	count--

	return count, split
}

func alphaNumUnderscore(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		r == '_'
}

func beginToken(idx, count, cpos, match int, pos map[int]int, line []rune, split []string) (int, int, []string) {
	count++

	pos[count] = idx

	if idx != cpos {
		return count, match, split
	}

	if idx == 0 {
		idx++
	}

	match = count
	split = []string{string(line[:idx-1])}

	return count, match, split
}

func (l *Line) Backward(tokenizer Tokenizer, pos int) (adjust int) {
	split, index, pos := tokenizer(pos)

	switch {
	case len(split) == 0:
		return
	case index == 0 && pos == 0:
		return
	case pos == 0:
		adjust = len(split[index-1])
	default:
		adjust = pos
	}

	return adjust * -1
}

func (l *Line) Cut(bpos, epos int) {
	bpos, epos, valid := l.checkRange(bpos, epos)
	if !valid {
		return
	}

	switch epos {
	case -1:
		*l = slices.Delete([]rune(*l), bpos, l.Len())
	default:
		*l = slices.Delete([]rune(*l), bpos, epos)
	}
}

func (l *Line) CutRune(pos int) {
	if pos < 0 || pos > l.Len() || l.Len() == 0 {
		return
	}

	switch pos {
	case l.Len():
		*l = slices.Delete([]rune(*l), pos-1, pos)
	default:
		*l = slices.Delete([]rune(*l), pos, pos+1)
	}
}

func (l *Line) Find(char rune, pos int, forward bool) int {
	if l.Len() == 0 {
		return -1
	}

	pos = l.checkPosRange(pos)

	for {
		if forward {
			pos++
			if pos > l.Len()-1 {
				break
			}
		} else {
			pos--
			if pos < 0 {
				break
			}
		}

		if (*l)[pos] == char {
			return pos
		}
	}

	return -1
}

func (l *Line) FindSurround(char rune, pos int) (bpos, epos int, bchar, echar rune) {
	bchar, echar = strutil.MatchSurround(char)

	bpos = l.Find(bchar, pos+1, false)
	epos = l.Find(echar, pos-1, true)

	return
}

func (l *Line) Forward(tokenizer Tokenizer, pos int) (adjust int) {
	split, index, pos := tokenizer(pos)

	switch {
	case len(split) == 0:
		return
	case index+1 == len(split):
		adjust = l.Len() - pos
	default:
		adjust = len(split[index]) - pos
	}

	return
}

func (l *Line) ForwardEnd(tokenizer Tokenizer, pos int) (adjust int) {
	split, index, pos := tokenizer(pos)
	if len(split) == 0 {
		return
	}

	word := strings.TrimRightFunc(split[index], unicode.IsSpace)

	switch {
	case index == len(split)-1 && pos >= len(word)-1:
		return
	case pos >= len(word)-1:
		word = strings.TrimRightFunc(split[index+1], unicode.IsSpace)
		adjust = len(split[index]) - pos
		adjust += len(word) - 1
	default:
		adjust = len(word) - pos - 1
	}

	return
}

func (l *Line) Insert(pos int, chars ...rune) {

	end := len(chars)
	for end > 0 && chars[end-1] == 0 {
		end--
	}

	chars = chars[:end]

	if pos < 0 || pos > l.Len() {
		return
	}

	*l = slices.Insert([]rune(*l), pos, chars...)
}

func (l *Line) InsertBetween(bpos, epos int, chars ...rune) {
	bpos, epos, valid := l.checkRange(bpos, epos)
	if !valid {
		return
	}

	switch epos {
	case -1:
		l.Insert(bpos, chars...)
	default:
		*l = slices.Delete([]rune(*l), bpos, epos)
		l.Insert(bpos, chars...)
	}
}

func (l *Line) Len() int {
	return len(*l)
}

func (l *Line) Lines() int {
	var count int

	for _, r := range *l {
		if r == inputrc.Newline {
			count++
		}
	}

	return count
}

func (l *Line) SelectBlankWord(pos int) (bpos, epos int) {
	if l.Len() == 0 {
		return bpos, epos
	}

	pos = l.checkPosRange(pos)
	if pos == l.Len() {
		pos--
	}

	bpos, epos = pos, pos

	isInWord := func(r rune) bool {
		return !unicode.IsSpace(r)
	}
	if unicode.IsSpace((*l)[pos]) {
		isInWord = unicode.IsSpace
	}

	for bpos > 0 && isInWord((*l)[bpos-1]) {
		bpos--
	}

	for epos < l.Len()-1 && isInWord((*l)[epos+1]) {
		epos++
	}

	return bpos, epos
}

func (l *Line) SelectWord(pos int) (bpos, epos int) {
	if l.Len() == 0 {
		return bpos, epos
	}

	pos = l.checkPosRange(pos)
	if pos == l.Len() {
		pos--
	}

	bpos, epos = pos, pos

	isInWord := alphaNumUnderscore
	if !alphaNumUnderscore((*l)[pos]) {
		isInWord = unicode.IsSpace
	}

	for bpos > 0 && isInWord((*l)[bpos-1]) {
		bpos--
	}

	for epos < l.Len()-1 && isInWord((*l)[epos+1]) {
		epos++
	}

	return bpos, epos
}

func (l *Line) Set(chars ...rune) {
	*l = chars
}

func (l *Line) SurroundQuotes(single bool, pos int) (bpos, epos int) {
	var bchar, echar rune

	if single {
		bchar, echar = '\'', '\''
	} else {
		bchar, echar = '"', '"'
	}

	var before, after int

	bpos = l.Find(bchar, pos+1, false)
	epos = l.Find(echar, pos, true)

	next, prev := epos, bpos

	for {
		if prev != -1 {
			before++
		}

		if next != -1 {
			after++
		}

		if prev == -1 || next == -1 {
			break
		}

		prev = l.Find(bchar, prev, false)
		next = l.Find(echar, next, true)
	}

	if before%2 == 0 && after%2 == 0 {
		return -1, -1
	}

	return bpos, epos
}

func (l *Line) Tokenize(cpos int) ([]string, int, int) {
	line := *l

	if line.Len() == 0 {
		return nil, 0, 0
	}

	cpos = l.checkPosRange(cpos)

	var index, pos int
	var punc bool

	split := []string{""}

	for i, char := range line {
		switch {
		case unicode.IsPunct(char):
			if i > 0 && line[i-1] != char {
				split = append(split, "")
			}

			split[len(split)-1] += string(char)
			punc = true

		case char == ' ' || char == '\t':
			split[len(split)-1] += string(char)
			punc = true

		case char == '\n':

			if i > 0 && line[i-1] == char {
				split = append(split, "")
			}

			split[len(split)-1] += string(char)
			punc = true

		default:
			if punc {
				split = append(split, "")
			}

			split[len(split)-1] += string(char)
			punc = false
		}

		if i == cpos {
			index = len(split) - 1
			pos = len(split[index]) - 1
		}
	}

	if cpos == len(line) {
		index = len(split) - 1
		pos = len(split[index])
	}

	return split, index, pos
}

func (l *Line) TokenizeBlock(cpos int) ([]string, int, int) {
	line := *l

	if line.Len() == 0 {
		return nil, 0, 0
	}

	cpos = l.checkPosRange(cpos)
	if cpos == l.Len() {
		cpos--
	}

	var (
		opener, closer rune
		split          []string
		count          int
		pos            = make(map[int]int)
		match          int
		single, double bool
	)

	switch line[cpos] {
	case '(', ')', '{', '[', '}', ']':
		opener, closer = strutil.MatchSurround(line[cpos])

	default:
		return nil, 0, 0
	}

	for idx := range line {
		switch line[idx] {
		case '\'':
			if !single {
				double = !double
			}

		case '"':
			if !double {
				single = !single
			}

		case opener:
			if !single && !double {
				count, match, split = beginToken(idx, count, cpos, match, pos, line, split)
			} else if idx == cpos {
				return nil, 0, 0
			}

		case closer:
			if !single && !double {
				count, split = finishToken(idx, count, cpos, match, pos, line, split)

				if match == count {
					return split, 1, 0
				} else if idx == cpos {
					return split, 1, len(split[1])
				}
			} else if idx == cpos {
				return nil, 0, 0
			}
		}
	}

	return nil, 0, 0
}

func (l *Line) TokenizeSpace(cpos int) ([]string, int, int) {
	line := *l

	if line.Len() == 0 {
		return nil, 0, 0
	}

	cpos = l.checkPosRange(cpos)

	var index, pos int

	split := []string{""}
	var newline bool

	for i, char := range line {
		switch char {
		case ' ', '\t':
			split[len(split)-1] += string(char)
			newline = false

		case '\n':

			if i > 0 && line[i-1] == char {
				split = append(split, "")
			}

			split[len(split)-1] += string(char)
			newline = true

		default:
			if (i > 0 && (line[i-1] == ' ' || line[i-1] == '\t')) || newline {
				split = append(split, "")
			}

			newline = false
			split[len(split)-1] += string(char)
		}

		if i == cpos {
			index = len(split) - 1
			pos = len(split[index]) - 1
		}
	}

	if cpos == len(line) {
		index = len(split) - 1
		pos = len(split[index])
	}

	return split, index, pos
}

func (l *Line) checkPosRange(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos > l.Len() {
		return l.Len()
	}

	return pos
}

func (l *Line) checkRange(bpos, epos int) (int, int, bool) {
	if bpos == -1 && epos == -1 {
		return -1, -1, false
	}

	if epos > l.Len() {
		epos = l.Len()
	}

	if bpos < 0 {
		bpos = 0
	}

	if epos > -1 && epos < bpos {
		bpos, epos = epos, bpos
	}

	return bpos, epos, true
}

func (l *Line) newlines() [][]int {
	var indices [][]int

	for i, r := range *l {
		if r == inputrc.Newline {
			indices = append(indices, []int{i, i + 1})
		}
	}

	indices = append(indices, []int{l.Len(), l.Len() + 1})

	return indices
}
