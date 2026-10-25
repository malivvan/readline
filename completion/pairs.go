package completion

import (
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/strutil"
)

func AutopairDelete(line *core.Line, cur *core.Cursor) {
	if cur.Pos() == 0 {
		return
	}

	toDelete := (*line)[cur.Pos()-1]
	isPair, _ := strutil.SurroundType(toDelete)
	matcher := strutil.IsSurround(toDelete, cur.Char())

	if isPair && matcher {
		line.CutRune(cur.Pos())
	}
}

func AutopairInsertOrJump(key rune, line *core.Line, cur *core.Cursor) (skipInsert bool) {
	matcher, closer := strutil.SurroundType(key)

	if !matcher {
		return false
	}

	if closer && cur.Char() == key {
		cur.Inc()

		return true
	}

	if key == '"' || key == '\'' {
		if unclosed, _ := strutil.GetQuotedWordStart((*line)[:cur.Pos()]); unclosed {
			return false
		}
	}

	switch {
	case closer && key != '\'' && key != '"':
		return false
	default:
		_, closeChar := strutil.MatchSurround(key)
		line.Insert(cur.Pos(), closeChar)
	}

	return false
}

func (e *Engine) CompleteSyntax(completer func([]rune, int) ([]rune, int)) {
	if completer == nil {
		return
	}

	line := []rune(*e.line)
	pos := e.cursor.Pos() - 1

	newLine, newPos := completer(line, pos)
	if string(newLine) == string(line) {
		return
	}

	newPos++

	e.line.Set(newLine...)
	e.cursor.Set(newPos)
}
