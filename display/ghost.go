package display

import (
	"strings"

	"github.com/malivvan/readline/core"
)

func (e *Engine) inlineAcceptable(line *core.Line, cursor *core.Cursor) bool {
	return e.inline != "" &&
		cursor.Pos() == line.Len() &&
		strings.HasPrefix(e.inline, string(*line)) &&
		len([]rune(e.inline)) > line.Len()
}
