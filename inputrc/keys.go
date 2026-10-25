package inputrc

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	Control   rune = 0x1f
	Meta      rune = 0x80
	Esc       rune = 0x1b
	Delete    rune = 0x7f
	Alert     rune = '\a'
	Backspace rune = '\b'
	Formfeed  rune = '\f'
	Newline   rune = '\n'
	Return    rune = '\r'
	Tab       rune = '\t'
	Vertical  rune = '\v'
	Space     rune = ' '
)

const (
	ErrBindMissingClosingQuote Error = `bind missing closing quote`

	ErrMissingColon Error = "missing :"

	ErrMacroMissingClosingQuote Error = `macro missing closing quote`

	ErrInvalidKeymap Error = "invalid keymap"

	ErrInvalidEditingMode Error = "invalid editing mode"

	ErrElseWithoutMatchingIf Error = "$else without matching $if"

	ErrEndifWithoutMatchingIf Error = "$endif without matching $if"

	ErrUnknownModifier Error = "unknown modifier"
)

func Decontrol(c rune) rune {
	return unicode.ToUpper(c | 0x40)
}

func Demeta(c rune) rune {
	return c & ^Meta
}

func Encontrol(c rune) rune {
	return unicode.ToUpper(c) & Control
}

func Enmeta(c rune) rune {
	return c | Meta
}

func Escape(s string) string {
	return escape(s, map[rune]string{
		Delete: `\C-?`,
		Return: `\C-M`,
	})
}

func EscapeMacro(s string) string {
	return escape(s, map[rune]string{
		Delete: `\d`,
		Return: `\r`,
	})
}

func IsControl(c rune) bool {
	return c < Space && c&Meta == 0
}

func IsMeta(c rune) bool {
	return c > Delete && c <= 0xff
}

func Unescape(s string) string {
	r := []rune(s)
	return unescapeRunes(r, 0, len(r))
}

func escape(s string, m map[rune]string) string {
	var v []string

	for _, c := range s {
		switch c {
		case Alert:
			v = append(v, `\a`)
		case Backspace:
			v = append(v, `\b`)
		case Delete:
			v = append(v, m[Delete])
		case Esc:
			v = append(v, `\e`)
		case Formfeed:
			v = append(v, `\f`)
		case Newline:
			v = append(v, `\n`)
		case Return:
			v = append(v, m[Return])
		case Tab:
			v = append(v, `\t`)
		case Vertical:
			v = append(v, `\v`)
		case '\\', '"', '\'':
			v = append(v, `\`+string(c))
		default:
			var s string
			if IsControl(c) {
				s += `\C-`
				c = Decontrol(c)
			}

			if IsMeta(c) {
				s += `\M-`
				c = Demeta(c)
			}

			if unicode.IsPrint(c) {
				s += string(c)
			} else {
				s += fmt.Sprintf(`\x%2x`, c)
			}

			v = append(v, s)
		}
	}

	return strings.Join(v, "")
}
