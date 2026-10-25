// Command highlighting shows the SyntaxHighlighter hook.
//
// The callback receives the current input buffer and returns the string to
// render: it may embed any ANSI escape sequence. Readline keeps its own
// overlays (visual selection, matching parens, ...) on top of it.
//
// The highlighter below colors:
//   - known commands (first word) in bold green, unknown ones in red,
//   - quoted strings in yellow,
//   - comments starting with "#" in dim gray.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

var commands = []string{"echo", "help", "list", "quit", "show"}

func isCommand(word string) bool {
	for _, command := range commands {
		if command == word {
			return true
		}
	}

	return false
}

func highlight(line []rune) string {
	var (
		out         strings.Builder
		quoted      bool
		quote       rune
		comment     bool
		word        strings.Builder
		atFirstWord = true
	)

	// flush styles the word that just ended, when it is the command word.
	flush := func() {
		if !atFirstWord || word.Len() == 0 {
			return
		}

		if isCommand(word.String()) {
			out.WriteString(color.Bold + color.FgGreen)
		} else {
			out.WriteString(color.FgRed)
		}

		out.WriteString(word.String())
		out.WriteString(color.Reset)

		word.Reset()
		atFirstWord = false
	}

	for _, char := range line {
		switch {
		case comment:
			out.WriteRune(char)

		case quoted && char == quote:
			out.WriteRune(char)
			out.WriteString(color.Reset)
			quoted = false

		case quoted:
			out.WriteRune(char)

		case char == '"' || char == '\'':
			flush()
			quoted, quote = true, char
			out.WriteString(color.FgYellow)
			out.WriteRune(char)

		case char == '#' && word.Len() == 0:
			comment = true
			out.WriteString(color.Dim)
			out.WriteRune(char)

		case char == ' ' || char == '\t':
			flush()
			out.WriteRune(char)

		default:
			word.WriteRune(char)
		}
	}

	flush()

	if comment {
		out.WriteString(color.Reset)
	}

	return out.String()
}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "highlight> "
	})

	shell.SyntaxHighlighter = highlight

	fmt.Println("try: echo \"hello world\" # a comment")

	for {
		line, err := shell.Readline()

		switch {
		case errors.Is(err, readline.ErrInterrupt):
			continue

		case errors.Is(err, io.EOF):
			return

		case err != nil:
			fmt.Printf("error: %v\n", err)
			return
		}

		if line == "quit" {
			return
		}

		fmt.Println(color.Strip(highlight([]rune(line))))
	}
}
