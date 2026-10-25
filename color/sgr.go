package color

import (
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

const (
	SGRStart = "\x1b["
	Fg       = "38;05;"
	Bg       = "48;05;"
	SGREnd   = "m"
)

func DisableEffects() {

	Reset = ""
	Bold = ""
	Dim = ""
	Underscore = ""
	Blink = ""
	Reverse = ""
	BoldReset = ""
	DimReset = ""
	UnderscoreReset = ""
	BlinkReset = ""
	ReverseReset = ""

	FgBlack = ""
	FgRed = ""
	FgGreen = ""
	FgYellow = ""
	FgBlue = ""
	FgMagenta = ""
	FgCyan = ""
	FgWhite = ""
	FgDefault = ""

	FgBlackBright = ""
	FgRedBright = ""
	FgGreenBright = ""
	FgYellowBright = ""
	FgBlueBright = ""
	FgMagentaBright = ""
	FgCyanBright = ""
	FgWhiteBright = ""

	BgBlack = ""
	BgRed = ""
	BgGreen = ""
	BgYellow = ""
	BgBlue = ""
	BgMagenta = ""
	BgCyan = ""
	BgWhite = ""
	BgDefault = ""

	BgDarkGray = ""
	BgBlueLight = ""

	BgBlackBright = ""
	BgRedBright = ""
	BgGreenBright = ""
	BgYellowBright = ""
	BgBlueBright = ""
	BgMagentaBright = ""
	BgCyanBright = ""
	BgWhiteBright = ""
}

func Fmt(color string) string {
	return SGRStart + color + SGREnd
}

func HasEffects() bool {
	if term := os.Getenv("TERM"); term == "" {
		return false
	} else if term == "dumb" {
		return false
	}

	return true
}

func Strip(str string) string {

	if strings.IndexByte(str, 0x1B) == -1 && strings.IndexByte(str, 0x9B) == -1 {
		return str
	}

	return re.ReplaceAllString(str, "")
}

// Trim truncates input to at most maxPrintableLength terminal cells. ANSI
// escape sequences are preserved and never counted as printable cells; when the
// input is truncated while an SGR style is active, the style is closed with
// Reset so that it does not bleed into whatever follows the truncated string.
func Trim(input string, maxPrintableLength int) string {
	if maxPrintableLength <= 0 {
		return ""
	}

	var (
		out   strings.Builder
		pos   int
		cells int
		style bool
	)

	for pos < len(input) {
		if loc := re.FindStringIndex(input[pos:]); loc != nil && loc[0] == 0 {
			seq := input[pos : pos+loc[1]]
			out.WriteString(seq)

			switch {
			case seq == Reset, seq == "\x1b[m":
				style = false
			case strings.HasSuffix(seq, "m"):
				style = true
			}

			pos += loc[1]

			continue
		}

		char, size := utf8.DecodeRuneInString(input[pos:])
		width := uniseg.StringWidth(string(char))

		if cells+width > maxPrintableLength {
			break
		}

		out.WriteString(input[pos : pos+size])

		cells += width
		pos += size
	}

	if pos < len(input) && style {
		out.WriteString(Reset)
	}

	return out.String()
}

func UnquoteRC(color string) string {
	color = strings.ReplaceAll(color, `\e`, "\x1b")

	if unquoted, err := strconv.Unquote(color); err == nil {
		return unquoted
	}

	return color
}
