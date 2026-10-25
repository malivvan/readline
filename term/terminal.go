package term

import (
	"os"
)

var standardTermWidth = 80

// termFile is the file used to query the terminal size. When nil, a standard
// width is assumed.
var termFile *os.File = os.Stderr

// SetTerminalFile sets the file used to query the terminal size.
// A nil file makes size queries fall back to the standard width.
func SetTerminalFile(f *os.File) {
	termFile = f
}

// hasTerminalFile reports whether terminal size queries are available.
func hasTerminalFile() bool {
	return termFile != nil
}

func DisableBracketedPaste() {
	WriteString(BracketedPasteEnd)
}

func EnableBracketedPaste() {
	WriteString(BracketedPasteStart)
}

func GetLength() int {
	if !hasTerminalFile() {
		return standardTermWidth
	}

	termFd := int(termFile.Fd())

	_, length, err := GetSize(termFd)
	if err != nil || length == 0 {
		return standardTermWidth
	}

	return length
}

func GetWidth() (termWidth int) {
	if !hasTerminalFile() {
		return standardTermWidth
	}

	var err error
	fd := int(termFile.Fd())
	termWidth, _, err = GetSize(fd)

	if err != nil || termWidth == 0 {
		termWidth = standardTermWidth
	}

	return
}
