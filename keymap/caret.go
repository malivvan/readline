package keymap

import (
	"strings"

	"github.com/malivvan/readline/term"
)

type CursorStyle string

const (
	cursorBlock             CursorStyle = "block"
	cursorUnderline         CursorStyle = "underline"
	cursorBeam              CursorStyle = "beam"
	cursorBlinkingBlock     CursorStyle = "blinking-block"
	cursorBlinkingUnderline CursorStyle = "blinking-underline"
	cursorBlinkingBeam      CursorStyle = "blinking-beam"
	cursorUserDefault       CursorStyle = "default"
)

var cursors = map[CursorStyle]string{
	cursorBlock:             "\x1b[2 q",
	cursorUnderline:         "\x1b[4 q",
	cursorBeam:              "\x1b[6 q",
	cursorBlinkingBlock:     "\x1b[1 q",
	cursorBlinkingUnderline: "\x1b[3 q",
	cursorBlinkingBeam:      "\x1b[5 q",
	cursorUserDefault:       "\x1b[0 q",
}

var standardCursors = map[Mode]CursorStyle{
	ViInsert:  cursorBlinkingBeam,
	Vi:        cursorBlinkingBeam,
	ViCommand: cursorBlinkingBlock,
	ViOpp:     cursorBlinkingUnderline,
	Visual:    cursorBlock,
	Emacs:     cursorBlinkingBlock,
}

func (c CursorStyle) String() string {
	cursor, found := cursors[c]
	if !found {
		return string(cursorUserDefault)
	}

	return cursor
}

func (m *Engine) PrintCursor(keymap Mode) {
	var cursor CursorStyle

	cursorOptname := "cursor-" + string(keymap)
	modeSet := strings.TrimSpace(m.config.GetString(cursorOptname))

	if _, valid := cursors[CursorStyle(modeSet)]; valid {
		term.WriteString(cursors[CursorStyle(modeSet)])
		return
	}

	if defaultCur, valid := standardCursors[keymap]; valid {
		term.WriteString(cursors[defaultCur])
		return
	}

	term.WriteString(cursors[cursor])
}
