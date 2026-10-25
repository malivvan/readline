// Command multiline shows the AcceptMultiline hook.
//
// By default the shell returns the buffer as soon as Enter is pressed, even if
// it spans several lines (multi-line edition is always available: Enter inside
// the buffer adds lines, and movement commands work across them).
//
// AcceptMultiline lets the application decide when a buffer is complete. This
// example implements a shell-like rule: the buffer is accepted only when its
// quotes are balanced and it does not end with a backslash. Otherwise a newline
// is appended and edition continues on the next line.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

func acceptMultiline(line []rune) bool {
	_, err := shellquote.Split(string(line))
	if err == nil {
		return true
	}

	switch err {
	case shellquote.UnterminatedSingleQuoteError, shellquote.UnterminatedDoubleQuoteError:
		return false

	case shellquote.UnterminatedEscapeError:
		return !strings.HasSuffix(string(line), `\`)
	}

	return true
}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return color.FgGreen + "sh> " + color.Reset
	})

	// The secondary prompt is printed before every continuation line.
	shell.Prompt.Secondary(func() string {
		return color.FgGreen + "... " + color.Reset
	})

	shell.AcceptMultiline = acceptMultiline

	fmt.Println(`Enter adds a line while quotes are open or the buffer ends with a backslash`)
	fmt.Println(`try: echo "a multiline` + "\\" + `"`)

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

		fmt.Printf("--- accepted buffer ---\n%s\n", line)
	}
}
