package core

import (
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
)

type Cursor struct {
	pos  int
	mark int
	line *Line
}

func CoordinatesCursor(cur *Cursor, indent int) (x, y int) {
	cur.CheckAppend()

	newlines := cur.line.newlines()
	bpos := 0
	usedY := 0

	for pos, newline := range newlines {
		switch {
		case newline[0] < cur.pos:

			line := (*cur.line)[bpos:newline[0]]
			bpos = newline[0] + 1
			_, y := strutil.LineSpan(line, pos, indent)
			usedY += y

		default:

			line := (*cur.line)[bpos:cur.pos]
			usedX, y := strutil.LineSpan(line, pos, indent)
			usedY += y

			return usedX, usedY
		}
	}

	return
}

func NewCursor(line *Line) *Cursor {
	return &Cursor{
		pos:  0,
		mark: -1,
		line: line,
	}
}

func (c *Cursor) AtBeginningOfLine() bool {
	if c.pos == 0 {
		return true
	}

	newlines := c.line.newlines()

	for line := range newlines {
		epos := newlines[line][0]
		if epos == c.pos-1 {
			return true
		}
	}

	return false
}

func (c *Cursor) AtEndOfLine() bool {
	if c.pos >= c.line.Len()-1 {
		return true
	}

	newlines := c.line.newlines()

	for line := range newlines {
		epos := newlines[line][0]
		if epos == c.pos+1 {
			return true
		}
	}

	return false
}

func (c *Cursor) BeginningOfLine() {
	defer c.CheckCommand()

	newlinePos := c.line.Find(inputrc.Newline, c.pos, false)
	if newlinePos != -1 {
		c.pos = newlinePos + 1
	} else {
		c.pos = 0
	}
}

func (c *Cursor) Char() rune {
	c.CheckAppend()

	if c.line.Len() == 0 {
		return rune(0)
	}

	if c.pos >= c.line.Len() {
		return rune(0)
	}

	return (*c.line)[c.pos]
}

func (c *Cursor) CheckAppend() {

	if c.pos < 0 {
		c.pos = 0
	}

	if c.pos > c.line.Len() {
		c.pos = c.line.Len()
	}

	if c.mark < -1 {
		c.mark = -1
	}

	if c.mark > c.line.Len()-1 {
		c.mark = -1
	}
}

func (c *Cursor) CheckCommand() {
	c.CheckAppend()

	if c.pos == c.line.Len() && !c.OnEmptyLine() {
		c.pos--
	}

	if c.line.Len() > 0 && c.pos < c.line.Len() && c.Char() == '\n' && !c.OnEmptyLine() {
		c.Dec()
	}
}

func (c *Cursor) Dec() {
	if c.pos > 0 {
		c.pos--
	}
}

func (c *Cursor) EndOfLine() {
	defer c.CheckCommand()

	if c.OnEmptyLine() {
		return
	}

	newlinePos := c.line.Find(inputrc.Newline, c.pos, true)

	if newlinePos != -1 {
		c.pos = newlinePos - 1
	} else {
		c.pos = c.line.Len() - 1
	}
}

func (c *Cursor) EndOfLineAppend() {
	defer c.CheckAppend()

	if c.OnEmptyLine() {
		return
	}

	newlinePos := c.line.Find(inputrc.Newline, c.pos-1, true)

	if newlinePos != -1 {
		c.pos = newlinePos
	} else {
		c.pos = c.line.Len()
	}
}

func (c *Cursor) Inc() {
	if c.pos < c.line.Len() {
		c.pos++
	}
}

func (c *Cursor) InsertAt(r ...rune) {
	c.CheckAppend()

	c.line.Insert(c.pos, r...)
	c.pos += len(r)
}

func (c *Cursor) LineMove(lines int) {
	c.CheckAppend()
	defer c.CheckAppend()

	newlines := c.line.newlines()
	if len(newlines) == 1 || lines == 0 {
		return
	}

	if lines < 0 {
		for range -lines {
			c.moveLineUp()
		}
	} else {
		for range lines {
			c.moveLineDown()
		}
	}
}

func (c *Cursor) LinePos() int {
	c.CheckAppend()

	newlines := c.line.newlines()

	for i, newline := range newlines {
		if newline[0] < c.pos {
			continue
		}

		return i
	}

	return len(newlines)
}

func (c *Cursor) Mark() int {
	return c.mark
}

func (c *Cursor) Move(offset int) {
	defer c.CheckAppend()
	c.pos += offset
}

func (c *Cursor) OnEmptyLine() bool {
	if c.line.Len() == 0 {
		return true
	}

	if c.pos == 0 {
		return (*c.line)[c.pos] == inputrc.Newline
	} else if c.pos == c.line.Len() {
		return (*c.line)[c.pos-1] == inputrc.Newline
	}

	under := (*c.line)[c.pos] == inputrc.Newline
	before := (*c.line)[c.pos-1] == inputrc.Newline

	return under && before
}

func (c *Cursor) Pos() int {
	c.CheckAppend()
	return c.pos
}

func (c *Cursor) ReplaceWith(char rune) {
	c.CheckAppend()

	switch {
	case c.pos == c.line.Len():
		c.line.Insert(c.line.Len(), char)
	default:
		(*c.line)[c.pos] = char
	}
}

func (c *Cursor) ResetMark() {
	c.mark = -1
}

func (c *Cursor) Set(pos int) {
	defer c.CheckAppend()

	switch {
	case pos < 0:
		c.pos = 0
	case pos > c.line.Len():
		c.pos = c.line.Len()
	default:
		c.pos = pos
	}
}

func (c *Cursor) SetMark() {
	c.CheckAppend()
	c.mark = c.pos
}

func (c *Cursor) ToFirstNonSpace(forward bool) {
	if c.line.Len() == 0 {
		return
	}

	defer c.CheckAppend()

	if c.pos >= c.line.Len() {
		forward = false
		c.pos = c.line.Len() - 1
	}

	if !forward && c.pos == 0 {
		return
	}

	for {
		if !c.onSpace() {
			return
		}

		if forward {
			c.pos++
		} else {
			c.pos--
		}

		if c.pos <= 0 {
			return
		}
	}
}

func (c *Cursor) moveLineDown() {
	var cpos, begin int
	begin = -1

	newlines := c.line.newlines()

	for line := range newlines {
		end := newlines[line][0]
		if line < c.LinePos() {
			begin = end
			continue
		}

		if line == c.LinePos() {
			cpos = c.pos - begin
			begin = end

			continue
		}

		if end-begin > cpos {
			c.pos = begin + cpos
		} else {
			c.pos = end
		}

		break
	}
}

func (c *Cursor) moveLineUp() {
	var cpos, begin int

	newlines := c.line.newlines()

	for line := len(newlines) - 1; line >= 0; line-- {
		end := newlines[line][0]

		if line > c.LinePos() {
			continue
		}

		if line > 0 {
			begin = newlines[line-1][0]
		} else {
			begin = -1
			end--
		}

		if line == c.LinePos() {
			cpos = c.pos - begin
			continue
		}

		if end-begin > cpos {
			c.pos = begin + cpos
		} else {
			c.pos = end
		}

		break
	}
}

func (c *Cursor) onSpace() bool {
	switch c.Char() {
	case inputrc.Space, inputrc.Newline, inputrc.Tab:
		return true
	default:
		return false
	}
}
