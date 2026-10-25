package term

const (
	NewlineReturn = "\r\n"

	ClearLineAfter   = "\x1b[0K"
	ClearLineBefore  = "\x1b[1K"
	ClearLine        = "\x1b[2K"
	ClearScreenBelow = "\x1b[0J"
	ClearScreen      = "\x1b[2J"
	ClearDisplay     = "\x1b[3J"

	CursorTopLeft    = "\x1b[H"
	SaveCursorPos    = "\x1b7"
	RestoreCursorPos = "\x1b8"
	HideCursor       = "\x1b[?25l"
	ShowCursor       = "\x1b[?25h"

	BracketedPasteStart = "\x1b[?2004h"
	BracketedPasteEnd   = "\x1b[?2004l"
)

var (
	ArrowUp    = string([]byte{27, 91, 65})
	ArrowDown  = string([]byte{27, 91, 66})
	ArrowRight = string([]byte{27, 91, 67})
	ArrowLeft  = string([]byte{27, 91, 68})
)

func MoveCursorBackwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dD", i)
}

func MoveCursorDown(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dB", i)
}

func MoveCursorForwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dC", i)
}

func MoveCursorUp(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dA", i)
}
