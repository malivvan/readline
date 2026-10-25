package strutil

import (
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/term"
	"github.com/rivo/uniseg"
)

func FormatTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "     ")
}

func LineSpan(line []rune, idx, indent int) (x, y int) {
	termWidth := term.GetWidth()
	lineLen := RealLength(string(line))
	lineLen += indent

	cursorY := lineLen / termWidth
	cursorX := lineLen % termWidth

	if idx != 0 {
		cursorY++
	}

	return cursorX, cursorY
}

func RealLength(s string) int {
	colors := color.Strip(s)
	tabs := strings.ReplaceAll(colors, "\t", "     ")

	return uniseg.StringWidth(tabs)
}
